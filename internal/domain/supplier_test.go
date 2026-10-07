package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestSupplierValidate(t *testing.T) {
	longName := strings.Repeat("n", 80)
	tooLongName := strings.Repeat("n", 81)
	longPhone := strings.Repeat("1", 40)
	tooLongPhone := strings.Repeat("1", 41)
	longNote := strings.Repeat("x", 200)
	tooLongNote := strings.Repeat("x", 201)

	cases := []struct {
		name     string
		supplier Supplier
		wantErr  error
	}{
		{
			name:     "valid name only",
			supplier: Supplier{ID: "s1", Name: "Distribuidora Sur"},
			wantErr:  nil,
		},
		{
			name:     "valid full record",
			supplier: Supplier{ID: "s2", Name: "Hielo Barato", Phone: "555-1234", Note: "Entrega martes y jueves"},
			wantErr:  nil,
		},
		{
			name:     "valid name at max length",
			supplier: Supplier{ID: "s3", Name: longName},
			wantErr:  nil,
		},
		{
			name:     "valid phone and note at max length",
			supplier: Supplier{ID: "s4", Name: "X", Phone: longPhone, Note: longNote},
			wantErr:  nil,
		},
		{
			name:     "empty name",
			supplier: Supplier{ID: "s5", Name: ""},
			wantErr:  ErrEmptySupplierName,
		},
		{
			name:     "blank name",
			supplier: Supplier{ID: "s6", Name: "   "},
			wantErr:  ErrEmptySupplierName,
		},
		{
			name:     "name too long",
			supplier: Supplier{ID: "s7", Name: tooLongName},
			wantErr:  ErrSupplierNameLong,
		},
		{
			name:     "phone too long",
			supplier: Supplier{ID: "s8", Name: "X", Phone: tooLongPhone},
			wantErr:  ErrSupplierPhoneLong,
		},
		{
			name:     "note too long",
			supplier: Supplier{ID: "s9", Name: "X", Note: tooLongNote},
			wantErr:  ErrSupplierNoteLong,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.supplier.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want %v", tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, want error wrapping %v", err, tt.wantErr)
			}
		})
	}
}
