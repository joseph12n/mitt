// Package domain holds the core bar POS types.
package domain

import (
	"fmt"
	"time"
)

// TabStatus is the lifecycle state of a table tab.
type TabStatus string

// Supported tab states.
const (
	TabOpen   TabStatus = "open"
	TabClosed TabStatus = "closed"
)

// Tab errors.
var (
	ErrTabClosed          = fmt.Errorf("tab is closed")
	ErrProductUnavailable = fmt.Errorf("product is not available")
	ErrTabEmpty           = fmt.Errorf("tab has no items")
)

// Tab is the running bill (cuenta) for one table. Money is in
// integer cents. Use AddItem to append lines and Close to settle it.
type Tab struct {
	ID       string
	TableID  string
	Items    []OrderItem
	Status   TabStatus
	OpenedAt time.Time
}

// AddItem appends one order line. It rejects lines with a non-positive
// quantity, lines added to a closed tab, and lines for products that
// are currently unavailable.
func (t *Tab) AddItem(item OrderItem, productAvailable bool) error {
	if t.Status == TabClosed {
		return fmt.Errorf("add item to tab %q: %w", t.ID, ErrTabClosed)
	}
	if !productAvailable {
		return fmt.Errorf("add item to tab %q: %w", t.ID, ErrProductUnavailable)
	}
	if err := item.Validate(); err != nil {
		return fmt.Errorf("add item to tab %q: %w", t.ID, err)
	}
	t.Items = append(t.Items, item)
	return nil
}

// TotalCents returns the sum of all line totals in cents.
func (t Tab) TotalCents() int64 {
	var total int64
	for _, it := range t.Items {
		total += it.LineTotal()
	}
	return total
}

// Close settles the tab and returns an immutable sale snapshot.
// It fails when the tab is already closed or holds no items.
func (t *Tab) Close() (Sale, error) {
	if t.Status == TabClosed {
		return Sale{}, fmt.Errorf("close tab %q: %w", t.ID, ErrTabClosed)
	}
	if len(t.Items) == 0 {
		return Sale{}, fmt.Errorf("close tab %q: %w", t.ID, ErrTabEmpty)
	}
	t.Status = TabClosed
	items := make([]OrderItem, len(t.Items))
	copy(items, t.Items)
	return Sale{
		ID:         t.ID,
		TableID:    t.TableID,
		Items:      items,
		TotalCents: t.TotalCents(),
		ClosedAt:   time.Now(),
	}, nil
}
