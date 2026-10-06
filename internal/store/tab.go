package store

import (
	"context"
	"fmt"
	"time"

	"mitt/internal/domain"
)

// SaveTab upserts the tab header and replaces its items inside one
// transaction, so a retry never leaves half-written lines behind.
func (s *Store) SaveTab(ctx context.Context, t domain.Tab) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("save tab %q: begin transaction: %w", t.ID, err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO tabs(id, table_id, status, opened_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			table_id = excluded.table_id,
			status = excluded.status,
			opened_at = excluded.opened_at`,
		t.ID, t.TableID, string(t.Status), t.OpenedAt.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("save tab %q: upsert header: %w", t.ID, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM tab_items WHERE tab_id = ?`, t.ID); err != nil {
		return fmt.Errorf("save tab %q: replace items: %w", t.ID, err)
	}
	for _, it := range t.Items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO tab_items(tab_id, product_id, name, unit_price_cents, qty)
			VALUES(?, ?, ?, ?, ?)`,
			t.ID, it.ProductID, it.Name, it.UnitPriceCents, it.Qty)
		if err != nil {
			return fmt.Errorf("save tab %q: insert item for product %q: %w", t.ID, it.ProductID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("save tab %q: commit: %w", t.ID, err)
	}
	return nil
}

// scanTab parses one tab header row. Times are stored as RFC3339 text.
func scanTab(id, tableID, status, openedAt string) (domain.Tab, error) {
	at, err := time.Parse(time.RFC3339, openedAt)
	if err != nil {
		return domain.Tab{}, fmt.Errorf("parse tab %q opened_at: %w", id, err)
	}
	return domain.Tab{
		ID:       id,
		TableID:  tableID,
		Status:   domain.TabStatus(status),
		OpenedAt: at,
	}, nil
}

// loadItems returns the order lines of one tab ordered by product id.
func (s *Store) loadItems(ctx context.Context, tabID string) ([]domain.OrderItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT product_id, name, unit_price_cents, qty
		FROM tab_items WHERE tab_id = ? ORDER BY product_id`, tabID)
	if err != nil {
		return nil, fmt.Errorf("load items for tab %q: %w", tabID, err)
	}
	defer rows.Close()
	var items []domain.OrderItem
	for rows.Next() {
		var it domain.OrderItem
		if err := rows.Scan(&it.ProductID, &it.Name, &it.UnitPriceCents, &it.Qty); err != nil {
			return nil, fmt.Errorf("scan item for tab %q: %w", tabID, err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load items for tab %q: %w", tabID, err)
	}
	return items, nil
}

// GetOpenTabByTable returns the open tab for one table with its items. It
// returns an error wrapping sql.ErrNoRows when the table has no open tab.
func (s *Store) GetOpenTabByTable(ctx context.Context, tableID string) (domain.Tab, error) {
	var id, tbl, status, openedAt string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, table_id, status, opened_at
		FROM tabs WHERE table_id = ? AND status = ?
		ORDER BY opened_at DESC LIMIT 1`, tableID, string(domain.TabOpen)).
		Scan(&id, &tbl, &status, &openedAt)
	if err != nil {
		return domain.Tab{}, fmt.Errorf("get open tab for table %q: %w", tableID, err)
	}
	tab, err := scanTab(id, tbl, status, openedAt)
	if err != nil {
		return domain.Tab{}, err
	}
	items, err := s.loadItems(ctx, tab.ID)
	if err != nil {
		return domain.Tab{}, err
	}
	tab.Items = items
	return tab, nil
}

// ListOpenTabs returns every open tab with its items, oldest first.
// Headers are fully read before items are loaded because the pool holds a
// single connection: querying items while rows are open would deadlock.
func (s *Store) ListOpenTabs(ctx context.Context) ([]domain.Tab, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, table_id, status, opened_at
		FROM tabs WHERE status = ? ORDER BY opened_at ASC`, string(domain.TabOpen))
	if err != nil {
		return nil, fmt.Errorf("list open tabs: %w", err)
	}
	type header struct {
		id, tableID, status, openedAt string
	}
	var headers []header
	for rows.Next() {
		var h header
		if err := rows.Scan(&h.id, &h.tableID, &h.status, &h.openedAt); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan tab: %w", err)
		}
		headers = append(headers, h)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("list open tabs: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("list open tabs: %w", err)
	}
	var tabs []domain.Tab
	for _, h := range headers {
		tab, err := scanTab(h.id, h.tableID, h.status, h.openedAt)
		if err != nil {
			return nil, err
		}
		items, err := s.loadItems(ctx, tab.ID)
		if err != nil {
			return nil, err
		}
		tab.Items = items
		tabs = append(tabs, tab)
	}
	return tabs, nil
}
