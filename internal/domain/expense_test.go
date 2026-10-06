package domain

import (
	"errors"
	"testing"
	"time"
)

func TestExpenseValidate(t *testing.T) {
	cases := []struct {
		name    string
		expense Expense
		wantErr error
	}{
		{
			name:    "valid expense",
			expense: Expense{ID: "e1", Description: "Hielo 10kg", Qty: 10, CostCents: 5000, Date: time.Now()},
			wantErr: nil,
		},
		{
			name:    "valid fractional quantity",
			expense: Expense{ID: "e2", Description: "Fernet 0.75L", Qty: 0.75, CostCents: 8000, Date: time.Now()},
			wantErr: nil,
		},
		{
			name:    "valid zero cost",
			expense: Expense{ID: "e3", Description: "Cortesia proveedor", Qty: 1, CostCents: 0, Date: time.Now()},
			wantErr: nil,
		},
		{
			name:    "empty description",
			expense: Expense{ID: "e4", Description: "", Qty: 1, CostCents: 1000, Date: time.Now()},
			wantErr: ErrEmptyExpenseDescription,
		},
		{
			name:    "zero quantity",
			expense: Expense{ID: "e5", Description: "Hielo", Qty: 0, CostCents: 1000, Date: time.Now()},
			wantErr: ErrInvalidExpenseQty,
		},
		{
			name:    "negative quantity",
			expense: Expense{ID: "e6", Description: "Hielo", Qty: -2, CostCents: 1000, Date: time.Now()},
			wantErr: ErrInvalidExpenseQty,
		},
		{
			name:    "negative cost",
			expense: Expense{ID: "e7", Description: "Hielo", Qty: 1, CostCents: -1, Date: time.Now()},
			wantErr: ErrNegativeExpenseCost,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.expense.Validate()
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
