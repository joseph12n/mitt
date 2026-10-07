package store

import (
	"context"
	"database/sql"
	"fmt"

	"mitt/internal/domain"
)

// UpsertSupplier inserts a supplier row or replaces it when the id
// already exists.
func (s *Store) UpsertSupplier(ctx context.Context, sup domain.Supplier) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO suppliers(id, name, phone, note)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			phone = excluded.phone,
			note = excluded.note`,
		sup.ID, sup.Name, sup.Phone, sup.Note)
	if err != nil {
		return fmt.Errorf("upsert supplier %q: %w", sup.ID, err)
	}
	return nil
}

// GetSupplier returns one supplier by id. It returns an error wrapping
// sql.ErrNoRows when the id is unknown.
func (s *Store) GetSupplier(ctx context.Context, id string) (domain.Supplier, error) {
	var sup domain.Supplier
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, phone, note FROM suppliers WHERE id = ?`, id).
		Scan(&sup.ID, &sup.Name, &sup.Phone, &sup.Note)
	if err != nil {
		return domain.Supplier{}, fmt.Errorf("get supplier %q: %w", id, err)
	}
	return sup, nil
}

// ListSuppliers returns every supplier ordered by name.
func (s *Store) ListSuppliers(ctx context.Context) ([]domain.Supplier, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, phone, note FROM suppliers ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list suppliers: %w", err)
	}
	defer rows.Close()
	var suppliers []domain.Supplier
	for rows.Next() {
		var sup domain.Supplier
		if err := rows.Scan(&sup.ID, &sup.Name, &sup.Phone, &sup.Note); err != nil {
			return nil, fmt.Errorf("scan supplier: %w", err)
		}
		suppliers = append(suppliers, sup)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list suppliers: %w", err)
	}
	return suppliers, nil
}

// DeleteSupplier removes one supplier row. It returns an error wrapping
// sql.ErrNoRows when the id is unknown. Nothing references suppliers
// yet, so no guards apply.
func (s *Store) DeleteSupplier(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM suppliers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete supplier %q: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete supplier %q: %w", id, err)
	}
	if n == 0 {
		return fmt.Errorf("delete supplier %q: %w", id, sql.ErrNoRows)
	}
	return nil
}
