// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"database/sql"
	"errors"
	"net/http"

	"mitt/internal/domain"
)

// productDTO is the catalog representation exchanged over the LAN API.
type productDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Available  bool   `json:"available"`
}

// toProductDTO maps the domain type to its snake_case wire form.
func toProductDTO(p domain.Product) productDTO {
	return productDTO{ID: p.ID, Name: p.Name, PriceCents: p.PriceCents, Available: p.Available}
}

// createProductRequest is the POST /api/products body. Available defaults to
// true when omitted so new catalog entries are sellable right away.
type createProductRequest struct {
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Available  *bool  `json:"available"`
}

// patchProductRequest is the PATCH /api/products/{id} body. Every field is
// optional; nil fields keep their stored value.
type patchProductRequest struct {
	Name       *string `json:"name"`
	PriceCents *int64  `json:"price_cents"`
	Available  *bool   `json:"available"`
}

// handleProductsList serves GET /api/products with an optional
// available=true|false filter. Unknown filter values are a 400.
func (s *Server) handleProductsList(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("available")
	var onlyAvailable, unavailableOnly bool
	switch raw {
	case "":
	case "true":
		onlyAvailable = true
	case "false":
		unavailableOnly = true
	default:
		writeError(w, http.StatusBadRequest, "bad_request", "available must be true or false")
		return
	}
	products, err := s.store.ListProducts(r.Context(), onlyAvailable)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "list products")
		return
	}
	if unavailableOnly {
		filtered := products[:0]
		for _, p := range products {
			if !p.Available {
				filtered = append(filtered, p)
			}
		}
		products = filtered
	}
	out := make([]productDTO, 0, len(products))
	for _, p := range products {
		out = append(out, toProductDTO(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"products": out})
}

// handleProductCreate serves POST /api/products. Domain validation failures
// are a 422.
func (s *Server) handleProductCreate(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	available := true
	if req.Available != nil {
		available = *req.Available
	}
	id, err := newID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "generate product id")
		return
	}
	p := domain.Product{ID: id, Name: req.Name, PriceCents: req.PriceCents, Available: available}
	if err := p.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	if err := s.store.UpsertProduct(r.Context(), p); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "save product")
		return
	}
	writeJSON(w, http.StatusCreated, toProductDTO(p))
}

// handleProductPatch serves PATCH /api/products/{id}. Unknown ids are a 404
// and domain validation failures are a 422.
func (s *Server) handleProductPatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := s.store.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "unknown product")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "load product")
		return
	}
	var req patchProductRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.PriceCents != nil {
		p.PriceCents = *req.PriceCents
	}
	if req.Available != nil {
		p.Available = *req.Available
	}
	if err := p.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	if err := s.store.UpsertProduct(r.Context(), p); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "save product")
		return
	}
	writeJSON(w, http.StatusOK, toProductDTO(p))
}
