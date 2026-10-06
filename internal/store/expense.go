package store

import (
	"context"
	"fmt"
	"time"

	"mitt/internal/domain"
)

// AddExpense records one bar purchase or running cost.
func (s *Store) AddExpense(ctx context.Context, e domain.Expense) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO expenses(id, description, qty, cost_cents, date)
		VALUES(?, ?, ?, ?, ?)`,
		e.ID, e.Description, e.Qty, e.CostCents, e.Date.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("add expense %q: %w", e.ID, err)
	}
	return nil
}

// ListExpenses returns every expense ordered by date, oldest first.
func (s *Store) ListExpenses(ctx context.Context) ([]domain.Expense, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, description, qty, cost_cents, date
		FROM expenses ORDER BY date ASC`)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()
	var expenses []domain.Expense
	for rows.Next() {
		var e domain.Expense
		var date string
		if err := rows.Scan(&e.ID, &e.Description, &e.Qty, &e.CostCents, &date); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		parsed, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return nil, fmt.Errorf("parse expense %q date: %w", e.ID, err)
		}
		e.Date = parsed
		expenses = append(expenses, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	return expenses, nil
}
