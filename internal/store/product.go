package store

import (
	"context"
	"fmt"

	"mitt/internal/domain"
)

// boolToInt maps a boolean to the 0/1 integer stored in SQLite.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// UpsertProduct inserts a catalog product or replaces it when the id
// already exists.
func (s *Store) UpsertProduct(ctx context.Context, p domain.Product) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO products(id, name, price_cents, available)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			price_cents = excluded.price_cents,
			available = excluded.available`,
		p.ID, p.Name, p.PriceCents, boolToInt(p.Available))
	if err != nil {
		return fmt.Errorf("upsert product %q: %w", p.ID, err)
	}
	return nil
}

// GetProduct returns one catalog product by id. It returns an error
// wrapping sql.ErrNoRows when the id is unknown.
func (s *Store) GetProduct(ctx context.Context, id string) (domain.Product, error) {
	var p domain.Product
	var available int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, price_cents, available FROM products WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.PriceCents, &available)
	if err != nil {
		return domain.Product{}, fmt.Errorf("get product %q: %w", id, err)
	}
	p.Available = available != 0
	return p, nil
}

// ListProducts returns the catalog ordered by name. When onlyAvailable is
// true, unavailable products are excluded.
func (s *Store) ListProducts(ctx context.Context, onlyAvailable bool) ([]domain.Product, error) {
	query := `SELECT id, name, price_cents, available FROM products ORDER BY name`
	if onlyAvailable {
		query = `SELECT id, name, price_cents, available FROM products WHERE available = 1 ORDER BY name`
	}
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()
	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		var available int
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceCents, &available); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		p.Available = available != 0
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	return products, nil
}
