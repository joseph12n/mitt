package store

import (
	"context"
	"testing"
	"time"

	"mitt/internal/domain"
)

func TestAddListExpensesOrder(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	cases := []struct {
		name   string
		add    domain.Expense
		wantAt int // position in the final chronological listing
	}{
		{
			name:   "middle date added first",
			add:    domain.Expense{ID: "e2", Description: "Ice", Qty: 2, CostCents: 500, Date: time.Date(2026, time.October, 5, 12, 0, 0, 0, time.UTC)},
			wantAt: 1,
		},
		{
			name:   "oldest date added second",
			add:    domain.Expense{ID: "e1", Description: "Beer keg", Qty: 1.5, CostCents: 12000, Date: time.Date(2026, time.October, 4, 9, 30, 0, 0, time.UTC)},
			wantAt: 0,
		},
		{
			name:   "newest date added last",
			add:    domain.Expense{ID: "e3", Description: "Limes", Qty: 0.75, CostCents: 800, Date: time.Date(2026, time.October, 6, 18, 15, 0, 0, time.UTC)},
			wantAt: 2,
		},
	}

	byID := make(map[string]domain.Expense, len(cases))
	for _, tt := range cases {
		t.Run("add "+tt.name, func(t *testing.T) {
			if err := s.AddExpense(ctx, tt.add); err != nil {
				t.Fatalf("AddExpense(%v) = %v, want nil", tt.add, err)
			}
		})
		byID[tt.add.ID] = tt.add
	}

	got, err := s.ListExpenses(ctx)
	if err != nil {
		t.Fatalf("ListExpenses() = %v, want nil", err)
	}
	if len(got) != len(cases) {
		t.Fatalf("ListExpenses() has %d rows, want %d", len(got), len(cases))
	}
	for _, tt := range cases {
		row := got[tt.wantAt]
		want := byID[row.ID]
		if row.ID == "" || byID[row.ID].ID == "" {
			t.Fatalf("ListExpenses()[%d] = %+v, unknown id", tt.wantAt, row)
		}
		if row.Description != want.Description || row.Qty != want.Qty || row.CostCents != want.CostCents {
			t.Fatalf("ListExpenses()[%d] = %+v, want %+v", tt.wantAt, row, want)
		}
		if !row.Date.Equal(want.Date) {
			t.Fatalf("ListExpenses()[%d].Date = %v, want %v", tt.wantAt, row.Date, want.Date)
		}
	}
	wantOrder := []string{"e1", "e2", "e3"}
	for i, id := range wantOrder {
		if got[i].ID != id {
			t.Fatalf("ListExpenses()[%d].ID = %q, want %q (chronological)", i, got[i].ID, id)
		}
	}
}
