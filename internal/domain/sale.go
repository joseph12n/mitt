// Package domain holds the core bar POS types.
package domain

import "time"

// Sale is the immutable snapshot produced when a tab closes. ID mirrors
// the closed tab ID. Items is a copy, so later tab mutation cannot alter
// history. Money is in integer cents.
type Sale struct {
	ID         string
	TableID    string
	Items      []OrderItem
	TotalCents int64
	ClosedAt   time.Time
}
