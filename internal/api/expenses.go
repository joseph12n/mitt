// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"net/http"
	"time"

	"mitt/internal/domain"
)

// expenseDTO is the cost entry representation exchanged over the LAN API.
type expenseDTO struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Qty         float64   `json:"qty"`
	CostCents   int64     `json:"cost_cents"`
	Date        time.Time `json:"date"`
}

// createExpenseRequest is the POST /api/expenses body.
type createExpenseRequest struct {
	Description string  `json:"description"`
	Qty         float64 `json:"qty"`
	CostCents   int64   `json:"cost_cents"`
}

// handleExpensesList serves GET /api/expenses, oldest first.
func (s *Server) handleExpensesList(w http.ResponseWriter, r *http.Request) {
	expenses, err := s.store.ListExpenses(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "list expenses")
		return
	}
	out := make([]expenseDTO, 0, len(expenses))
	for _, e := range expenses {
		out = append(out, expenseDTO{
			ID:          e.ID,
			Description: e.Description,
			Qty:         e.Qty,
			CostCents:   e.CostCents,
			Date:        e.Date,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"expenses": out})
}

// handleExpenseCreate serves POST /api/expenses. Domain validation failures
// are a 422.
func (s *Server) handleExpenseCreate(w http.ResponseWriter, r *http.Request) {
	var req createExpenseRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	id, err := newID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "generate expense id")
		return
	}
	expense := domain.Expense{
		ID:          id,
		Description: req.Description,
		Qty:         req.Qty,
		CostCents:   req.CostCents,
		Date:        time.Now().UTC(),
	}
	if err := expense.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	if err := s.store.AddExpense(r.Context(), expense); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "save expense")
		return
	}
	writeJSON(w, http.StatusCreated, expenseDTO{
		ID:          expense.ID,
		Description: expense.Description,
		Qty:         expense.Qty,
		CostCents:   expense.CostCents,
		Date:        expense.Date,
	})
}
