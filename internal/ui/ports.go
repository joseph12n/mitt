// Package ui defines the pluggable UI seam for mitt.
//
// Any native shell (stdlib text dashboard, Fyne, Wails, or a future
// graphical frontend) renders from ui.DashboardData obtained through the
// DashboardService interface. Shells must depend only on this port, never
// on the store or API adapters directly.
package ui

import "context"

// OpenTabView is one occupied table summarized for dashboard rendering.
type OpenTabView struct {
	// TableID labels the physical table (for example "t1").
	TableID string
	// ItemCount is the total ordered units across all lines of the tab.
	ItemCount int
	// TotalCents is the running tab total in integer cents.
	TotalCents int64
}

// DashboardData is everything a dashboard shell needs in one snapshot.
type DashboardData struct {
	// OpenTabs holds every open tab, oldest first.
	OpenTabs []OpenTabView
	// TodaySalesCents is today's closed revenue in integer cents.
	//
	// Gap: the store currently exposes no sales reader (sales tables
	// exist in the schema but closed tabs persist back into tabs, and no
	// day aggregate query exists), so this field is always 0 until a
	// dedicated sales query lands.
	// TODO: wire a future store day-sales query here (for example
	// Store.SalesToday) and drop the hardcoded 0. Do NOT expand store
	// surfaces ad hoc for this; the query belongs to a storage task.
	TodaySalesCents int64
	// ProductCount is the total catalog size.
	ProductCount int
	// OutOfStock lists names of unavailable products, ordered by name.
	OutOfStock []string
}

// DashboardService is the port every UI shell programs against.
type DashboardService interface {
	// Snapshot reads one consistent dashboard view from the store.
	Snapshot(ctx context.Context) (DashboardData, error)
}
