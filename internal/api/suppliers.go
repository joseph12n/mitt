// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"mitt/internal/domain"
)

// supplierDTO is the provider representation exchanged over the LAN API.
type supplierDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Note  string `json:"note"`
}

// toSupplierDTO maps the domain type to its snake_case wire form.
func toSupplierDTO(s domain.Supplier) supplierDTO {
	return supplierDTO{ID: s.ID, Name: s.Name, Phone: s.Phone, Note: s.Note}
}

// createSupplierRequest is the POST /api/suppliers body. Phone and note
// are optional and default to empty.
type createSupplierRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Note  string `json:"note"`
}

// patchSupplierRequest is the PATCH /api/suppliers/{id} body. Every field
// is optional; nil fields keep their stored value.
type patchSupplierRequest struct {
	Name  *string `json:"name"`
	Phone *string `json:"phone"`
	Note  *string `json:"note"`
}

// handleSuppliersList serves GET /api/suppliers ordered by name.
func (s *Server) handleSuppliersList(w http.ResponseWriter, r *http.Request) {
	suppliers, err := s.store.ListSuppliers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "list suppliers")
		return
	}
	out := make([]supplierDTO, 0, len(suppliers))
	for _, sup := range suppliers {
		out = append(out, toSupplierDTO(sup))
	}
	writeJSON(w, http.StatusOK, map[string]any{"suppliers": out})
}

// handleSupplierCreate serves POST /api/suppliers. Domain validation
// failures are a 422.
func (s *Server) handleSupplierCreate(w http.ResponseWriter, r *http.Request) {
	var req createSupplierRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	id, err := newID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "generate supplier id")
		return
	}
	sup := domain.Supplier{ID: id, Name: strings.TrimSpace(req.Name), Phone: req.Phone, Note: req.Note}
	if err := sup.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	if err := s.store.UpsertSupplier(r.Context(), sup); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "save supplier")
		return
	}
	writeJSON(w, http.StatusCreated, toSupplierDTO(sup))
}

// handleSupplierPatch serves PATCH /api/suppliers/{id}. Unknown ids are a
// 404 and domain validation failures are a 422.
func (s *Server) handleSupplierPatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sup, err := s.store.GetSupplier(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "unknown supplier")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "load supplier")
		return
	}
	var req patchSupplierRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name != nil {
		sup.Name = strings.TrimSpace(*req.Name)
	}
	if req.Phone != nil {
		sup.Phone = *req.Phone
	}
	if req.Note != nil {
		sup.Note = *req.Note
	}
	if err := sup.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	if err := s.store.UpsertSupplier(r.Context(), sup); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "save supplier")
		return
	}
	writeJSON(w, http.StatusOK, toSupplierDTO(sup))
}

// handleSupplierDelete serves DELETE /api/suppliers/{id}. Unknown ids
// are a 404. Nothing references suppliers yet, so no guards apply.
func (s *Server) handleSupplierDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.DeleteSupplier(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "unknown supplier")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "delete supplier")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
