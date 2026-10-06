package domain

import (
	"errors"
	"testing"
	"time"
)

func openTab() *Tab {
	return &Tab{ID: "tab1", TableID: "t1", Status: TabOpen, OpenedAt: time.Now()}
}

func TestTabAddItem(t *testing.T) {
	cases := []struct {
		name      string
		setup     func() *Tab
		item      OrderItem
		available bool
		wantErr   error
		wantItems int
	}{
		{
			name:      "happy path appends item",
			setup:     openTab,
			item:      OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 2},
			available: true,
			wantErr:   nil,
			wantItems: 1,
		},
		{
			name: "reject when tab closed",
			setup: func() *Tab {
				tab := openTab()
				tab.Status = TabClosed
				return tab
			},
			item:      OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 1},
			available: true,
			wantErr:   ErrTabClosed,
			wantItems: 0,
		},
		{
			name:      "reject unavailable product",
			setup:     openTab,
			item:      OrderItem{ProductID: "p9", Name: "IPA", UnitPriceCents: 2000, Qty: 1},
			available: false,
			wantErr:   ErrProductUnavailable,
			wantItems: 0,
		},
		{
			name:      "reject zero quantity",
			setup:     openTab,
			item:      OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 0},
			available: true,
			wantErr:   ErrInvalidQuantity,
			wantItems: 0,
		},
		{
			name:      "reject negative quantity",
			setup:     openTab,
			item:      OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: -2},
			available: true,
			wantErr:   ErrInvalidQuantity,
			wantItems: 0,
		},
		{
			name:      "reject negative price",
			setup:     openTab,
			item:      OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: -10, Qty: 1},
			available: true,
			wantErr:   ErrNegativeUnitPrice,
			wantItems: 0,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tab := tt.setup()
			err := tab.AddItem(tt.item, tt.available)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("AddItem() = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Fatalf("AddItem() = nil, want %v", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("AddItem() = %v, want error wrapping %v", err, tt.wantErr)
				}
			}
			if len(tab.Items) != tt.wantItems {
				t.Fatalf("len(Items) = %d, want %d", len(tab.Items), tt.wantItems)
			}
		})
	}
}

func TestTabTotalCents(t *testing.T) {
	cases := []struct {
		name  string
		items []OrderItem
		want  int64
	}{
		{
			name:  "empty tab totals zero",
			items: nil,
			want:  0,
		},
		{
			name: "single line",
			items: []OrderItem{
				{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 2},
			},
			want: 3000,
		},
		{
			name: "multiple lines",
			items: []OrderItem{
				{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 2},
				{ProductID: "p3", Name: "Fernet", UnitPriceCents: 2500, Qty: 1},
				{ProductID: "p4", Name: "Mani", UnitPriceCents: 400, Qty: 3},
			},
			want: 3000 + 2500 + 1200,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tab := openTab()
			tab.Items = tt.items
			if got := tab.TotalCents(); got != tt.want {
				t.Fatalf("TotalCents() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTabClose(t *testing.T) {
	t.Run("close snapshots items total and time", func(t *testing.T) {
		tab := openTab()
		items := []OrderItem{
			{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 2},
			{ProductID: "p3", Name: "Fernet", UnitPriceCents: 2500, Qty: 1},
		}
		for _, it := range items {
			if err := tab.AddItem(it, true); err != nil {
				t.Fatalf("AddItem() = %v, want nil", err)
			}
		}

		before := time.Now()
		sale, err := tab.Close()
		after := time.Now()
		if err != nil {
			t.Fatalf("Close() = %v, want nil", err)
		}
		if tab.Status != TabClosed {
			t.Fatalf("tab status = %q, want closed", tab.Status)
		}
		if sale.ID != tab.ID {
			t.Fatalf("sale ID = %q, want %q", sale.ID, tab.ID)
		}
		if sale.TableID != tab.TableID {
			t.Fatalf("sale TableID = %q, want %q", sale.TableID, tab.TableID)
		}
		if sale.TotalCents != 5500 {
			t.Fatalf("sale total = %d, want 5500", sale.TotalCents)
		}
		if len(sale.Items) != len(items) {
			t.Fatalf("len(sale.Items) = %d, want %d", len(sale.Items), len(items))
		}
		for i := range items {
			if sale.Items[i] != items[i] {
				t.Fatalf("sale.Items[%d] = %+v, want %+v", i, sale.Items[i], items[i])
			}
		}
		if sale.ClosedAt.Before(before) || sale.ClosedAt.After(after) {
			t.Fatalf("sale ClosedAt = %v, want within [%v, %v]", sale.ClosedAt, before, after)
		}

		// Mutating the tab afterwards must not alter the snapshot.
		tab.Items[0].Qty = 99
		if sale.Items[0].Qty == 99 {
			t.Fatal("sale items share backing array with tab items")
		}
	})

	t.Run("close empty tab fails", func(t *testing.T) {
		tab := openTab()
		if _, err := tab.Close(); !errors.Is(err, ErrTabEmpty) {
			t.Fatalf("Close() = %v, want error wrapping %v", err, ErrTabEmpty)
		}
	})

	t.Run("close twice fails", func(t *testing.T) {
		tab := openTab()
		if err := tab.AddItem(OrderItem{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 1}, true); err != nil {
			t.Fatalf("AddItem() = %v, want nil", err)
		}
		if _, err := tab.Close(); err != nil {
			t.Fatalf("first Close() = %v, want nil", err)
		}
		if _, err := tab.Close(); !errors.Is(err, ErrTabClosed) {
			t.Fatalf("second Close() = %v, want error wrapping %v", err, ErrTabClosed)
		}
	})
}
