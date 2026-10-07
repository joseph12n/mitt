// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"mitt/internal/domain"
	"mitt/internal/store"
)

// tableDTO is the hub-owned table representation. Occupied is derived from
// open tabs at read time, never stored, so the hub stays the single source
// of truth without cross-table triggers.
type tableDTO struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Occupied bool   `json:"occupied"`
}

// createTableRequest is the POST /api/tables body.
type createTableRequest struct {
	Label string `json:"label"`
}

// toTableDTO maps a domain table to its wire form. Tables with an open tab
// read as occupied via the reused TableOccupied state.
func toTableDTO(t domain.Table, occupied map[string]bool) tableDTO {
	dto := tableDTO{ID: t.ID, Label: t.Label}
	status := domain.TableFree
	if occupied[t.ID] {
		status = domain.TableOccupied
	}
	dto.Occupied = status == domain.TableOccupied
	return dto
}

// openTableSet returns the ids of tables holding an open tab.
func (s *Server) openTableSet(r *http.Request) (map[string]bool, error) {
	tabs, err := s.store.ListOpenTabs(r.Context())
	if err != nil {
		return nil, err
	}
	occupied := make(map[string]bool, len(tabs))
	for _, t := range tabs {
		occupied[t.TableID] = true
	}
	return occupied, nil
}

// handleTablesList serves GET /api/tables ordered by label.
func (s *Server) handleTablesList(w http.ResponseWriter, r *http.Request) {
	tables, err := s.store.ListTables(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "list tables")
		return
	}
	occupied, err := s.openTableSet(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "list open tabs")
		return
	}
	out := make([]tableDTO, 0, len(tables))
	for _, t := range tables {
		out = append(out, toTableDTO(t, occupied))
	}
	writeJSON(w, http.StatusOK, map[string]any{"tables": out})
}

// handleTableCreate serves POST /api/tables. Labels are trimmed and must
// hold 1..40 characters per domain.Table.
func (s *Server) handleTableCreate(w http.ResponseWriter, r *http.Request) {
	var req createTableRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	id, err := newID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "generate table id")
		return
	}
	t := domain.Table{ID: id, Label: strings.TrimSpace(req.Label), Status: domain.TableFree}
	if err := t.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	if err := s.store.UpsertTable(r.Context(), t); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "save table")
		return
	}
	writeJSON(w, http.StatusCreated, toTableDTO(t, map[string]bool{}))
}

// handleTableDelete serves DELETE /api/tables/{id}. Tables with an open tab
// are a 409, unknown ids are a 404.
func (s *Server) handleTableDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteTable(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrTableHasOpenTab) {
			writeError(w, http.StatusConflict, "table_occupied", "table has an open tab")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "unknown table")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "delete table")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
