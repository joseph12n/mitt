package domain

import (
	"errors"
	"testing"
)

func TestOrderItemValidate(t *testing.T) {
	cases := []struct {
		name    string
		item    OrderItem
		wantErr error
	}{
		{
			name:    "valid item",
			item:    OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 2},
			wantErr: nil,
		},
		{
			name:    "zero quantity",
			item:    OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 0},
			wantErr: ErrInvalidQuantity,
		},
		{
			name:    "negative quantity",
			item:    OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: -1},
			wantErr: ErrInvalidQuantity,
		},
		{
			name:    "negative price",
			item:    OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: -50, Qty: 1},
			wantErr: ErrNegativeUnitPrice,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
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

func TestOrderItemLineTotal(t *testing.T) {
	cases := []struct {
		name string
		item OrderItem
		want int64
	}{
		{
			name: "two units",
			item: OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 2},
			want: 3000,
		},
		{
			name: "zero price",
			item: OrderItem{ProductID: "p2", Name: "Tapa", UnitPriceCents: 0, Qty: 5},
			want: 0,
		},
		{
			name: "single unit",
			item: OrderItem{ProductID: "p3", Name: "Fernet", UnitPriceCents: 2500, Qty: 1},
			want: 2500,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.LineTotal(); got != tt.want {
				t.Fatalf("LineTotal() = %d, want %d", got, tt.want)
			}
		})
	}
}
