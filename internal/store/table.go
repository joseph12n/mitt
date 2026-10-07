package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"mitt/internal/domain"
)

// ErrTableHasOpenTab is returned when deleting a table that still has an
// open tab. Close the tab first so the running bill is never orphaned.
var ErrTableHasOpenTab = fmt.Errorf("table has an open tab")

// ErrUnknownTable is returned when a tab references a table row that does
// not exist. The hub owns the tables, so clients must create the table
// first via POST /api/tables.
var ErrUnknownTable = fmt.Errorf("unknown table")

// UpsertTable inserts a table row or replaces it when the id exists.
func (s *Store) UpsertTable(ctx context.Context, t domain.Table) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tables_tbl(id, label, status)
		VALUES(?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			label = excluded.label,
			status = excluded.status`,
		t.ID, t.Label, string(t.Status))
	if err != nil {
		return fmt.Errorf("upsert table %q: %w", t.ID, err)
	}
	return nil
}

// GetTable returns one table by id. It returns an error wrapping
// sql.ErrNoRows when the id is unknown.
func (s *Store) GetTable(ctx context.Context, id string) (domain.Table, error) {
	var t domain.Table
	var status string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, label, status FROM tables_tbl WHERE id = ?`, id).
		Scan(&t.ID, &t.Label, &status)
	if err != nil {
		return domain.Table{}, fmt.Errorf("get table %q: %w", id, err)
	}
	t.Status = domain.TableStatus(status)
	return t, nil
}

// ListTables returns every table ordered by label.
func (s *Store) ListTables(ctx context.Context) ([]domain.Table, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, label, status FROM tables_tbl ORDER BY label`)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()
	var tables []domain.Table
	for rows.Next() {
		var t domain.Table
		var status string
		if err := rows.Scan(&t.ID, &t.Label, &status); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		t.Status = domain.TableStatus(status)
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	return tables, nil
}

// DeleteTable removes one table row. It refuses with ErrTableHasOpenTab
// while an open tab references the table, and wraps sql.ErrNoRows when the
// id is unknown. The check and the delete run in one transaction.
func (s *Store) DeleteTable(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("delete table %q: begin transaction: %w", id, err)
	}
	defer tx.Rollback()
	var open string
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM tabs WHERE table_id = ? AND status = ? LIMIT 1`,
		id, string(domain.TabOpen)).Scan(&open)
	if err == nil {
		return fmt.Errorf("delete table %q: %w", id, ErrTableHasOpenTab)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("delete table %q: check open tabs: %w", id, err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM tables_tbl WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete table %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete table %q: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("delete table %q: %w", id, sql.ErrNoRows)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("delete table %q: commit: %w", id, err)
	}
	return nil
}
