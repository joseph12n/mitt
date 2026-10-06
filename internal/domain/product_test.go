package domain

import (
	"errors"
	"testing"
)

func TestProductValidate(t *testing.T) {
	cases := []struct {
		name    string
		product Product
		wantErr error
	}{
		{
			name:    "valid available product",
			product: Product{ID: "p1", Name: "Quilmes", PriceCents: 1500, Available: true},
			wantErr: nil,
		},
		{
			name:    "valid zero price product",
			product: Product{ID: "p2", Name: "Free tapa", PriceCents: 0, Available: true},
			wantErr: nil,
		},
		{
			name:    "empty name",
			product: Product{ID: "p3", Name: "", PriceCents: 1500, Available: true},
			wantErr: ErrEmptyProductName,
		},
		{
			name:    "negative price",
			product: Product{ID: "p4", Name: "Fernet", PriceCents: -1, Available: true},
			wantErr: ErrNegativePrice,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.Validate()
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

func TestTableValidate(t *testing.T) {
	cases := []struct {
		name    string
		table   Table
		wantErr error
	}{
		{
			name:    "valid free table",
			table:   Table{ID: "t1", Label: "Mesa 1", Status: TableFree},
			wantErr: nil,
		},
		{
			name:    "valid occupied table",
			table:   Table{ID: "t2", Label: "Barra 3", Status: TableOccupied},
			wantErr: nil,
		},
		{
			name:    "empty label",
			table:   Table{ID: "t3", Label: "", Status: TableFree},
			wantErr: ErrEmptyTableLabel,
		},
		{
			name:    "unknown status",
			table:   Table{ID: "t4", Label: "Mesa 4", Status: TableStatus("dirty")},
			wantErr: ErrUnknownTableStatus,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.table.Validate()
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
