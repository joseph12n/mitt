// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"mitt/internal/domain"
	"mitt/internal/store"
)

// tabItemDTO is one order line on the wire. Unit price is a snapshot taken
// at order time, so later catalog edits never rewrite history.
type tabItemDTO struct {
	ProductID      string `json:"product_id"`
	Name           string `json:"name"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	Qty            int    `json:"qty"`
	LineTotalCents int64  `json:"line_total_cents"`
}

// tabDTO is the running bill (cuenta) representation.
type tabDTO struct {
	ID         string       `json:"id"`
	TableID    string       `json:"table_id"`
	Status     string       `json:"status"`
	OpenedAt   time.Time    `json:"opened_at"`
	Items      []tabItemDTO `json:"items"`
	TotalCents int64        `json:"total_cents"`
}

// saleDTO is the immutable snapshot returned when a tab closes.
type saleDTO struct {
	ID         string       `json:"id"`
	TableID    string       `json:"table_id"`
	Items      []tabItemDTO `json:"items"`
	TotalCents int64        `json:"total_cents"`
	ClosedAt   time.Time    `json:"closed_at"`
}

// toTabDTO maps a domain tab to its wire form, always with a non-nil list.
func toTabDTO(t domain.Tab) tabDTO {
	items := make([]tabItemDTO, 0, len(t.Items))
	for _, it := range t.Items {
		items = append(items, tabItemDTO{
			ProductID:      it.ProductID,
			Name:           it.Name,
			UnitPriceCents: it.UnitPriceCents,
			Qty:            it.Qty,
			LineTotalCents: it.LineTotal(),
		})
	}
	return tabDTO{
		ID:         t.ID,
		TableID:    t.TableID,
		Status:     string(t.Status),
		OpenedAt:   t.OpenedAt,
		Items:      items,
		TotalCents: t.TotalCents(),
	}
}

// openTabRequest is the POST /api/tabs body.
type openTabRequest struct {
	TableID string `json:"table_id"`
}

// addItemRequest is the POST /api/tabs/{id}/items body.
type addItemRequest struct {
	ProductID string `json:"product_id"`
	Qty       int    `json:"qty"`
}

// handleTabOpen serves POST /api/tabs. The table_id must reference a table
// row created first via POST /api/tables: the hub owns the tables, so an
// unknown id is a 422 unknown_table. Opening an already-open table is
// idempotent and returns the existing tab with 200 instead of duplicating
// it.
func (s *Server) handleTabOpen(w http.ResponseWriter, r *http.Request) {
	var req openTabRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := (domain.Table{ID: req.TableID, Label: req.TableID, Status: domain.TableOccupied}).Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	if _, err := s.store.GetTable(r.Context(), req.TableID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnprocessableEntity, "unknown_table", "unknown table: create it first via POST /api/tables")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "lookup table")
		return
	}
	if existing, err := s.store.GetOpenTabByTable(r.Context(), req.TableID); err == nil {
		writeJSON(w, http.StatusOK, toTabDTO(existing))
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "internal", "lookup open tab")
		return
	}
	id, err := newID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "generate tab id")
		return
	}
	tab := domain.Tab{ID: id, TableID: req.TableID, Status: domain.TabOpen, OpenedAt: time.Now().UTC()}
	if err := s.store.SaveTab(r.Context(), tab); err != nil {
		if errors.Is(err, store.ErrUnknownTable) {
			writeError(w, http.StatusUnprocessableEntity, "unknown_table", "unknown table: create it first via POST /api/tables")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "save tab")
		return
	}
	writeJSON(w, http.StatusCreated, toTabDTO(tab))
}

// handleTabsOpen serves GET /api/tabs/open, oldest first.
func (s *Server) handleTabsOpen(w http.ResponseWriter, r *http.Request) {
	tabs, err := s.store.ListOpenTabs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "list open tabs")
		return
	}
	out := make([]tabDTO, 0, len(tabs))
	for _, t := range tabs {
		out = append(out, toTabDTO(t))
	}
	writeJSON(w, http.StatusOK, map[string]any{"tabs": out})
}

// findOpenTabByID locates one open tab by id. The store only indexes open
// tabs by table, so this scans the open list; missing or already-closed tabs
// surface as sql.ErrNoRows to keep the 404 contract in one place.
func (s *Server) findOpenTabByID(r *http.Request, id string) (domain.Tab, error) {
	tabs, err := s.store.ListOpenTabs(r.Context())
	if err != nil {
		return domain.Tab{}, err
	}
	for _, t := range tabs {
		if t.ID == id {
			return t, nil
		}
	}
	return domain.Tab{}, sql.ErrNoRows
}

// handleTabAddItem serves POST /api/tabs/{id}/items. It snapshots the
// product name and price onto the order line, rejects unavailable products
// and closed tabs via domain.Tab.AddItem (422), and persists the tab.
func (s *Server) handleTabAddItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req addItemRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	product, err := s.store.GetProduct(r.Context(), req.ProductID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "unknown product")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "load product")
		return
	}
	tab, err := s.findOpenTabByID(r, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "unknown tab")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "load tab")
		return
	}
	item := domain.OrderItem{
		ProductID:      product.ID,
		Name:           product.Name,
		UnitPriceCents: product.PriceCents,
		Qty:            req.Qty,
	}
	if err := tab.AddItem(item, product.Available); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	// The tab_items primary key is (tab_id, product_id), so repeated orders
	// of the same product merge into one line instead of duplicating rows.
	tab.Items = coalesceItems(tab.Items)
	if err := s.store.SaveTab(r.Context(), tab); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "save tab")
		return
	}
	writeJSON(w, http.StatusOK, toTabDTO(tab))
}

// handleTabClose serves POST /api/tabs/{id}/close. It settles the tab via
// domain.Tab.Close (422 when already closed or empty), persists the closed
// header, and returns the sale snapshot with its total.
func (s *Server) handleTabClose(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tab, err := s.findOpenTabByID(r, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "unknown tab")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "load tab")
		return
	}
	sale, err := tab.Close()
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	// The sale snapshot is archived before the tab flips to closed: the
	// sale id mirrors the tab id, so a retry after a partial failure
	// upserts the same row instead of losing the bill to a 404.
	if err := s.store.SaveSale(r.Context(), sale); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "archive sale")
		return
	}
	if err := s.store.SaveTab(r.Context(), tab); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "close tab")
		return
	}
	items := make([]tabItemDTO, 0, len(sale.Items))
	for _, it := range sale.Items {
		items = append(items, tabItemDTO{
			ProductID:      it.ProductID,
			Name:           it.Name,
			UnitPriceCents: it.UnitPriceCents,
			Qty:            it.Qty,
			LineTotalCents: it.LineTotal(),
		})
	}
	writeJSON(w, http.StatusOK, saleDTO{
		ID:         sale.ID,
		TableID:    sale.TableID,
		Items:      items,
		TotalCents: sale.TotalCents,
		ClosedAt:   sale.ClosedAt,
	})
}

// coalesceItems merges lines sharing a product id by summing quantities,
// keeping first-seen order.
func coalesceItems(items []domain.OrderItem) []domain.OrderItem {
	index := make(map[string]int, len(items))
	out := make([]domain.OrderItem, 0, len(items))
	for _, it := range items {
		if i, ok := index[it.ProductID]; ok {
			out[i].Qty += it.Qty
			continue
		}
		index[it.ProductID] = len(out)
		out = append(out, it)
	}
	return out
}
