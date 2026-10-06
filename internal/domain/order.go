// Package domain holds the core bar POS types.
package domain

import "fmt"

// ErrInvalidQuantity is returned when an order quantity is not positive.
var ErrInvalidQuantity = fmt.Errorf("quantity must be greater than zero")

// ErrNegativeUnitPrice is returned when an order line price is negative.
var ErrNegativeUnitPrice = fmt.Errorf("unit price must not be negative")

// OrderItem is one line on a tab. Name and UnitPriceCents are snapshots
// taken from the product at order time, so later catalog edits do not
// rewrite history. Money is in integer cents.
type OrderItem struct {
	ProductID      string
	Name           string
	UnitPriceCents int64
	Qty            int
}

// LineTotal returns the line amount in cents.
func (it OrderItem) LineTotal() int64 {
	return it.UnitPriceCents * int64(it.Qty)
}

// Validate reports whether the order line holds usable data.
func (it OrderItem) Validate() error {
	if it.Qty <= 0 {
		return fmt.Errorf("validate item for product %q: %w", it.ProductID, ErrInvalidQuantity)
	}
	if it.UnitPriceCents < 0 {
		return fmt.Errorf("validate item for product %q: %w", it.ProductID, ErrNegativeUnitPrice)
	}
	return nil
}
