// Package ui implements the pluggable dashboard port over the store.
package ui

import (
	"context"
	"fmt"

	"mitt/internal/store"
)

// Service is the default DashboardService backed by the SQLite store.
type Service struct {
	store *store.Store
}

// Compile-time proof that Service satisfies the seam every shell uses.
var _ DashboardService = (*Service)(nil)

// New wraps s for dashboard snapshots. It keeps the store reference; the
// caller owns the store lifecycle.
func New(s *store.Store) *Service {
	return &Service{store: s}
}

// Snapshot reads open tabs plus the catalog and folds them into one
// DashboardData. Open tabs arrive oldest first from the store and keep that
// order here. ProductCount counts the full catalog; OutOfStock names only
// unavailable products, ordered by name (store order).
func (sv *Service) Snapshot(ctx context.Context) (DashboardData, error) {
	tabs, err := sv.store.ListOpenTabs(ctx)
	if err != nil {
		return DashboardData{}, fmt.Errorf("snapshot open tabs: %w", err)
	}
	products, err := sv.store.ListProducts(ctx, false)
	if err != nil {
		return DashboardData{}, fmt.Errorf("snapshot products: %w", err)
	}
	views := make([]OpenTabView, 0, len(tabs))
	for _, t := range tabs {
		qty := 0
		for _, it := range t.Items {
			qty += it.Qty
		}
		views = append(views, OpenTabView{
			TableID:    t.TableID,
			ItemCount:  qty,
			TotalCents: t.TotalCents(),
		})
	}
	out := make([]string, 0)
	for _, p := range products {
		if !p.Available {
			out = append(out, p.Name)
		}
	}
	// TODO: replace the hardcoded 0 with a future store day-sales query
	// (see DashboardData.TodaySalesCents). No sales reader exists yet.
	return DashboardData{
		OpenTabs:        views,
		TodaySalesCents: 0,
		ProductCount:    len(products),
		OutOfStock:      out,
	}, nil
}
