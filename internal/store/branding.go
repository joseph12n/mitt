package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"mitt/internal/domain"
)

// BrandingRow pairs stored branding with its last-write timestamp.
// UpdatedAt is zero when the row was never stored.
type BrandingRow struct {
	Branding  domain.Branding
	UpdatedAt time.Time
}

// GetBrandingRow returns the stored branding with its timestamp, or the
// defaults with a zero timestamp when no row was stored yet.
func (s *Store) GetBrandingRow(ctx context.Context) (BrandingRow, error) {
	var row BrandingRow
	var shop, primary, accent, background, mime, updated string
	var logo []byte
	err := s.db.QueryRowContext(ctx, `
		SELECT shop_name, "primary", accent, background, logo, logo_mime, updated_at
		FROM branding WHERE id = 1`).Scan(&shop, &primary, &accent, &background, &logo, &mime, &updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return BrandingRow{Branding: domain.DefaultBranding()}, nil
		}
		return BrandingRow{}, fmt.Errorf("get branding: %w", err)
	}
	at, err := time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		at, err = time.Parse(time.RFC3339, updated)
		if err != nil {
			return BrandingRow{}, fmt.Errorf("parse branding updated_at: %w", err)
		}
	}
	row.Branding = domain.Branding{
		ShopName:   shop,
		Primary:    primary,
		Accent:     accent,
		Background: background,
		Logo:       logo,
		LogoMIME:   mime,
	}
	row.UpdatedAt = at
	return row, nil
}

// GetBranding returns the stored branding, or the defaults when no row
// was stored yet.
func (s *Store) GetBranding(ctx context.Context) (domain.Branding, error) {
	row, err := s.GetBrandingRow(ctx)
	if err != nil {
		return domain.Branding{}, err
	}
	return row.Branding, nil
}

// UpsertBranding validates then replaces the single branding row inside
// one transaction.
func (s *Store) UpsertBranding(ctx context.Context, b domain.Branding) error {
	if err := b.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("upsert branding: begin transaction: %w", err)
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO branding(id, shop_name, "primary", accent, background, logo, logo_mime, updated_at)
		VALUES(1, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			shop_name = excluded.shop_name,
			"primary" = excluded."primary",
			accent = excluded.accent,
			background = excluded.background,
			logo = excluded.logo,
			logo_mime = excluded.logo_mime,
			updated_at = excluded.updated_at`,
		b.ShopName, b.Primary, b.Accent, b.Background,
		b.Logo, b.LogoMIME, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("upsert branding: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("upsert branding: commit: %w", err)
	}
	return nil
}
