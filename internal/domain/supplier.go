// Package domain holds the core bar POS types.
package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Supplier errors.
var (
	ErrEmptySupplierName = fmt.Errorf("supplier name must not be empty")
	ErrSupplierNameLong  = fmt.Errorf("supplier name must be at most 80 characters")
	ErrSupplierPhoneLong = fmt.Errorf("supplier phone must be at most 40 characters")
	ErrSupplierNoteLong  = fmt.Errorf("supplier note must be at most 200 characters")
)

// Supplier length bounds in characters, so hub, web, and mobile show
// one short record everywhere.
const (
	maxSupplierNameLen  = 80
	maxSupplierPhoneLen = 40
	maxSupplierNoteLen  = 200
)

// Supplier is a goods or services provider. Phone and note are optional
// contact details; nothing references suppliers yet, so deletes are
// unrestricted.
type Supplier struct {
	ID    string
	Name  string
	Phone string
	Note  string
}

// Validate reports whether the supplier holds usable data. The name is
// trimmed and must hold 1..80 characters; phone and note are optional
// and capped at 40 and 200 characters.
func (s Supplier) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("validate supplier %q: %w", s.ID, ErrEmptySupplierName)
	}
	if utf8.RuneCountInString(strings.TrimSpace(s.Name)) > maxSupplierNameLen {
		return fmt.Errorf("validate supplier %q: %w", s.ID, ErrSupplierNameLong)
	}
	if utf8.RuneCountInString(s.Phone) > maxSupplierPhoneLen {
		return fmt.Errorf("validate supplier %q: %w", s.ID, ErrSupplierPhoneLong)
	}
	if utf8.RuneCountInString(s.Note) > maxSupplierNoteLen {
		return fmt.Errorf("validate supplier %q: %w", s.ID, ErrSupplierNoteLong)
	}
	return nil
}
