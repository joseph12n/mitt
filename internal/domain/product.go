// Package domain holds the core bar POS types.
//
// The core depends only on the standard library. Adapters (storage,
// transport, UI) plug in around it and must not leak inward.
package domain

import "fmt"

// ErrEmptyProductName is returned when a product has no name.
var ErrEmptyProductName = fmt.Errorf("product name must not be empty")

// ErrNegativePrice is returned when product money is negative.
var ErrNegativePrice = fmt.Errorf("price must not be negative")

// Product is a sellable catalog entry. Money is in integer cents.
type Product struct {
	ID         string
	Name       string
	PriceCents int64
	Available  bool
}

// Validate reports whether the product holds usable catalog data.
func (p Product) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("validate product %q: %w", p.ID, ErrEmptyProductName)
	}
	if p.PriceCents < 0 {
		return fmt.Errorf("validate product %q: %w", p.ID, ErrNegativePrice)
	}
	return nil
}
