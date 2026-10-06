package ui

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"mitt/internal/domain"
	"mitt/internal/store"
)

// openTestStore opens a fresh migrated database inside a temp dir.
func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open() = %v, want nil", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Fatalf("Close() = %v, want nil", err)
		}
	})
	return s
}

// seedProducts upserts every product in ps.
func seedProducts(t *testing.T, s *store.Store, ps []domain.Product) {
	t.Helper()
	ctx := context.Background()
	for _, p := range ps {
		if err := s.UpsertProduct(ctx, p); err != nil {
			t.Fatalf("UpsertProduct(%q) = %v, want nil", p.ID, err)
		}
	}
}

// seedTabs saves every tab in tabs.
func seedTabs(t *testing.T, s *store.Store, tabs []domain.Tab) {
	t.Helper()
	ctx := context.Background()
	for _, tab := range tabs {
		if err := s.SaveTab(ctx, tab); err != nil {
			t.Fatalf("SaveTab(%q) = %v, want nil", tab.ID, err)
		}
	}
}

// openTab builds an open tab fixture at a fixed time.
func openTab(id, tableID string, items []domain.OrderItem) domain.Tab {
	return domain.Tab{
		ID:       id,
		TableID:  tableID,
		Status:   domain.TabOpen,
		OpenedAt: time.Date(2026, time.October, 6, 20, 0, 0, 0, time.UTC),
		Items:    items,
	}
}

func TestSnapshot(t *testing.T) {
	beer := domain.Product{ID: "p1", Name: "Quilmes", PriceCents: 1500, Available: true}
	fernet := domain.Product{ID: "p2", Name: "Fernet", PriceCents: 2500, Available: true}
	aperol := domain.Product{ID: "p3", Name: "Aperol", PriceCents: 1800, Available: false}

	tests := []struct {
		name         string
		products     []domain.Product
		tabs         []domain.Tab
		wantOpen     int
		wantProducts int
		wantOut      []string
		wantViews    map[string]OpenTabView
	}{
		{
			name:         "empty store",
			products:     nil,
			tabs:         nil,
			wantOpen:     0,
			wantProducts: 0,
			wantOut:      nil,
			wantViews:    map[string]OpenTabView{},
		},
		{
			name:     "open tab with catalog and out of stock",
			products: []domain.Product{beer, fernet, aperol},
			tabs: []domain.Tab{openTab("tab1", "t1", []domain.OrderItem{
				{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 2},
				{ProductID: "p2", Name: "Fernet", UnitPriceCents: 2500, Qty: 1},
			})},
			wantOpen:     1,
			wantProducts: 3,
			wantOut:      []string{"Aperol"},
			wantViews: map[string]OpenTabView{
				"t1": {TableID: "t1", ItemCount: 3, TotalCents: 5500},
			},
		},
		{
			name:     "closed tabs excluded",
			products: []domain.Product{beer},
			tabs: []domain.Tab{
				openTab("tab-open", "t1", []domain.OrderItem{
					{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 1},
				}),
				{
					ID:       "tab-closed",
					TableID:  "t2",
					Status:   domain.TabClosed,
					OpenedAt: time.Date(2026, time.October, 6, 19, 0, 0, 0, time.UTC),
					Items: []domain.OrderItem{
						{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 9},
					},
				},
			},
			wantOpen:     1,
			wantProducts: 1,
			wantOut:      []string{},
			wantViews: map[string]OpenTabView{
				"t1": {TableID: "t1", ItemCount: 1, TotalCents: 1500},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := openTestStore(t)
			seedProducts(t, s, tt.products)
			seedTabs(t, s, tt.tabs)

			got, err := New(s).Snapshot(context.Background())
			if err != nil {
				t.Fatalf("Snapshot() = %v, want nil", err)
			}
			if len(got.OpenTabs) != tt.wantOpen {
				t.Fatalf("OpenTabs has %d rows, want %d", len(got.OpenTabs), tt.wantOpen)
			}
			if got.ProductCount != tt.wantProducts {
				t.Fatalf("ProductCount = %d, want %d", got.ProductCount, tt.wantProducts)
			}
			if len(got.OutOfStock) != len(tt.wantOut) {
				t.Fatalf("OutOfStock = %v, want %v", got.OutOfStock, tt.wantOut)
			}
			for i, name := range tt.wantOut {
				if got.OutOfStock[i] != name {
					t.Fatalf("OutOfStock[%d] = %q, want %q", i, got.OutOfStock[i], name)
				}
			}
			for _, view := range got.OpenTabs {
				want, ok := tt.wantViews[view.TableID]
				if !ok {
					t.Fatalf("unexpected open table %q in %+v", view.TableID, got.OpenTabs)
				}
				if view != want {
					t.Fatalf("OpenTabView(%q) = %+v, want %+v", view.TableID, view, want)
				}
			}
			// Documented gap: no store sales reader exists yet, so the
			// seam reports 0 until a day-sales query lands.
			if got.TodaySalesCents != 0 {
				t.Fatalf("TodaySalesCents = %d, want 0 (no sales reader yet)", got.TodaySalesCents)
			}
		})
	}
}
