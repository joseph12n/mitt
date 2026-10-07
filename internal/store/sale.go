package store

import (
	"context"
	"fmt"
	"time"

	"mitt/internal/domain"
)

// SalesSummary is the one-row aggregate for today's closed sales. Date is
// a YYYY-MM-DD calendar day; TotalCents is in integer cents.
type SalesSummary struct {
	Date       string
	Count      int
	TotalCents int64
}

// defaultSalesLimit applies when callers pass a non-positive limit.
const defaultSalesLimit = 50

// SaveSale records one closed sale snapshot and replaces its items inside
// one transaction, so a retry never leaves half-written lines behind. The
// sale id mirrors the closed tab id, making repeated closes idempotent.
// Times are stored as UTC RFC3339 text, matching tabs and expenses.
func (s *Store) SaveSale(ctx context.Context, sale domain.Sale) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("save sale %q: begin transaction: %w", sale.ID, err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO sales(id, table_id, total_cents, closed_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			table_id = excluded.table_id,
			total_cents = excluded.total_cents,
			closed_at = excluded.closed_at`,
		sale.ID, sale.TableID, sale.TotalCents, sale.ClosedAt.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("save sale %q: upsert header: %w", sale.ID, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM sale_items WHERE sale_id = ?`, sale.ID); err != nil {
		return fmt.Errorf("save sale %q: replace items: %w", sale.ID, err)
	}
	for _, it := range sale.Items {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO sale_items(sale_id, product_id, name, unit_price_cents, qty)
			VALUES(?, ?, ?, ?, ?)`,
			sale.ID, it.ProductID, it.Name, it.UnitPriceCents, it.Qty)
		if err != nil {
			return fmt.Errorf("save sale %q: insert item for product %q: %w", sale.ID, it.ProductID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("save sale %q: commit: %w", sale.ID, err)
	}
	return nil
}

// loadSaleItems returns the order lines of one sale ordered by product id.
func (s *Store) loadSaleItems(ctx context.Context, saleID string) ([]domain.OrderItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT product_id, name, unit_price_cents, qty
		FROM sale_items WHERE sale_id = ? ORDER BY product_id`, saleID)
	if err != nil {
		return nil, fmt.Errorf("load items for sale %q: %w", saleID, err)
	}
	defer rows.Close()
	var items []domain.OrderItem
	for rows.Next() {
		var it domain.OrderItem
		if err := rows.Scan(&it.ProductID, &it.Name, &it.UnitPriceCents, &it.Qty); err != nil {
			return nil, fmt.Errorf("scan item for sale %q: %w", saleID, err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load items for sale %q: %w", saleID, err)
	}
	return items, nil
}

// ListSales returns closed sales newest first, capped at limit rows. A
// non-positive limit falls back to the default page size. Headers are
// fully read before items are loaded because the pool holds a single
// connection: querying items while rows are open would deadlock.
func (s *Store) ListSales(ctx context.Context, limit int) ([]domain.Sale, error) {
	if limit <= 0 {
		limit = defaultSalesLimit
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, table_id, total_cents, closed_at
		FROM sales ORDER BY closed_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list sales: %w", err)
	}
	type header struct {
		id, tableID, closedAt string
		total                 int64
	}
	var headers []header
	for rows.Next() {
		var h header
		if err := rows.Scan(&h.id, &h.tableID, &h.total, &h.closedAt); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan sale: %w", err)
		}
		headers = append(headers, h)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("list sales: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("list sales: %w", err)
	}
	sales := make([]domain.Sale, 0, len(headers))
	for _, h := range headers {
		closedAt, err := time.Parse(time.RFC3339, h.closedAt)
		if err != nil {
			return nil, fmt.Errorf("parse sale %q closed_at: %w", h.id, err)
		}
		items, err := s.loadSaleItems(ctx, h.id)
		if err != nil {
			return nil, err
		}
		sales = append(sales, domain.Sale{
			ID:         h.id,
			TableID:    h.tableID,
			Items:      items,
			TotalCents: h.total,
			ClosedAt:   closedAt,
		})
	}
	return sales, nil
}

// TodaySummary aggregates today's closed sales in SQL: the calendar day,
// the sale count, and the summed total. Closed_at is RFC3339 text, so
// strftime parses it directly; both sides use UTC, matching the stored
// values. An empty day returns the date with zero count and total.
func (s *Store) TodaySummary(ctx context.Context) (SalesSummary, error) {
	var out SalesSummary
	err := s.db.QueryRowContext(ctx, `
		SELECT strftime('%Y-%m-%d', 'now'),
			COUNT(*),
			COALESCE(SUM(total_cents), 0)
		FROM sales
		WHERE strftime('%Y-%m-%d', closed_at) = strftime('%Y-%m-%d', 'now')`).
		Scan(&out.Date, &out.Count, &out.TotalCents)
	if err != nil {
		return SalesSummary{}, fmt.Errorf("today sales summary: %w", err)
	}
	return out, nil
}
