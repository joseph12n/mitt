// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"net/http"
	"strconv"

	"mitt/internal/domain"
)

// salesDefaultLimit is the page size when ?limit= is omitted. salesMaxLimit
// caps it so a client cannot page the whole history in one response.
const (
	salesDefaultLimit = 50
	salesMaxLimit     = 500
)

// toSaleDTO maps a domain sale to the same snapshot shape returned when a
// tab closes, always with a non-nil item list.
func toSaleDTO(sale domain.Sale) saleDTO {
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
	return saleDTO{
		ID:         sale.ID,
		TableID:    sale.TableID,
		Items:      items,
		TotalCents: sale.TotalCents,
		ClosedAt:   sale.ClosedAt,
	}
}

// handleSalesList serves GET /api/sales newest first. The limit query
// parameter defaults to 50 and is capped at 500; non-numeric or
// non-positive values are a 400.
func (s *Server) handleSalesList(w http.ResponseWriter, r *http.Request) {
	limit := salesDefaultLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "bad_request", "limit must be a positive integer")
			return
		}
		limit = n
		if limit > salesMaxLimit {
			limit = salesMaxLimit
		}
	}
	sales, err := s.store.ListSales(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "list sales")
		return
	}
	out := make([]saleDTO, 0, len(sales))
	for _, sale := range sales {
		out = append(out, toSaleDTO(sale))
	}
	writeJSON(w, http.StatusOK, map[string]any{"sales": out})
}

// handleSalesToday serves GET /api/sales/today with the UTC calendar day,
// the sale count, and the summed total in cents.
func (s *Server) handleSalesToday(w http.ResponseWriter, r *http.Request) {
	summary, err := s.store.TodaySummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "sales today summary")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"date":        summary.Date,
		"count":       summary.Count,
		"total_cents": summary.TotalCents,
	})
}
