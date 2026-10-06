// Package domain holds the core bar POS types.
package domain

import "fmt"

// TableStatus is the occupancy state of a physical table.
type TableStatus string

// Supported table states.
const (
	TableFree     TableStatus = "free"
	TableOccupied TableStatus = "occupied"
)

// ErrEmptyTableLabel is returned when a table has no label.
var ErrEmptyTableLabel = fmt.Errorf("table label must not be empty")

// ErrUnknownTableStatus is returned for unrecognized table states.
var ErrUnknownTableStatus = fmt.Errorf("unknown table status")

// Table is a physical table customers sit at.
type Table struct {
	ID     string
	Label  string
	Status TableStatus
}

// Validate reports whether the table holds usable data.
func (t Table) Validate() error {
	if t.Label == "" {
		return fmt.Errorf("validate table %q: %w", t.ID, ErrEmptyTableLabel)
	}
	switch t.Status {
	case TableFree, TableOccupied:
		return nil
	default:
		return fmt.Errorf("validate table %q: %w: %q", t.ID, ErrUnknownTableStatus, t.Status)
	}
}
