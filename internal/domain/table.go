// Package domain holds the core bar POS types.
package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// TableStatus is the occupancy state of a physical table.
type TableStatus string

// Supported table states.
const (
	TableFree     TableStatus = "free"
	TableOccupied TableStatus = "occupied"
)

// ErrEmptyTableLabel is returned when a table has no label.
var ErrEmptyTableLabel = fmt.Errorf("table label must not be empty")

// ErrTableLabelTooLong is returned when a table label exceeds the limit.
var ErrTableLabelTooLong = fmt.Errorf("table label must be at most 40 characters")

// maxTableLabelLen bounds table labels so hub, web, and mobile show one
// short name everywhere.
const maxTableLabelLen = 40

// ErrUnknownTableStatus is returned for unrecognized table states.
var ErrUnknownTableStatus = fmt.Errorf("unknown table status")

// Table is a physical table customers sit at.
type Table struct {
	ID     string
	Label  string
	Status TableStatus
}

// Validate reports whether the table holds usable data. The label is
// trimmed and must hold 1..40 characters.
func (t Table) Validate() error {
	label := strings.TrimSpace(t.Label)
	if label == "" {
		return fmt.Errorf("validate table %q: %w", t.ID, ErrEmptyTableLabel)
	}
	if utf8.RuneCountInString(label) > maxTableLabelLen {
		return fmt.Errorf("validate table %q: %w", t.ID, ErrTableLabelTooLong)
	}
	switch t.Status {
	case TableFree, TableOccupied:
		return nil
	default:
		return fmt.Errorf("validate table %q: %w: %q", t.ID, ErrUnknownTableStatus, t.Status)
	}
}
