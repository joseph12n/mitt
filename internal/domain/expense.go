// Package domain holds the core bar POS types.
package domain

import (
	"fmt"
	"time"
)

// Expense errors.
var (
	ErrEmptyExpenseDescription = fmt.Errorf("expense description must not be empty")
	ErrInvalidExpenseQty       = fmt.Errorf("expense quantity must be greater than zero")
	ErrNegativeExpenseCost     = fmt.Errorf("expense cost must not be negative")
)

// Expense is a simple bar purchase or running cost. Qty is a measured
// amount (units, kilos, liters) and may be fractional; CostCents is the
// total money paid, in integer cents.
type Expense struct {
	ID          string
	Description string
	Qty         float64
	CostCents   int64
	Date        time.Time
}

// Validate reports whether the expense holds usable data.
func (e Expense) Validate() error {
	if e.Description == "" {
		return fmt.Errorf("validate expense %q: %w", e.ID, ErrEmptyExpenseDescription)
	}
	if e.Qty <= 0 {
		return fmt.Errorf("validate expense %q: %w", e.ID, ErrInvalidExpenseQty)
	}
	if e.CostCents < 0 {
		return fmt.Errorf("validate expense %q: %w", e.ID, ErrNegativeExpenseCost)
	}
	return nil
}
