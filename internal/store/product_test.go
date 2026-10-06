package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"mitt/internal/domain"
)

func TestUpsertGetProduct(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	cases := []struct {
		name    string
		upserts []domain.Product
		getID   string
		want    domain.Product
	}{
		{
			name:    "insert available product",
			upserts: []domain.Product{{ID: "p1", Name: "Quilmes", PriceCents: 1500, Available: true}},
			getID:   "p1",
			want:    domain.Product{ID: "p1", Name: "Quilmes", PriceCents: 1500, Available: true},
		},
		{
			name: "upsert overwrites price and availability",
			upserts: []domain.Product{
				{ID: "p2", Name: "Fernet", PriceCents: 2000, Available: true},
				{ID: "p2", Name: "Fernet", PriceCents: 2500, Available: false},
			},
			getID: "p2",
			want:  domain.Product{ID: "p2", Name: "Fernet", PriceCents: 2500, Available: false},
		},
		{
			name:    "unavailable flag survives roundtrip",
			upserts: []domain.Product{{ID: "p3", Name: "Seasonal", PriceCents: 999, Available: false}},
			getID:   "p3",
			want:    domain.Product{ID: "p3", Name: "Seasonal", PriceCents: 999, Available: false},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			for _, p := range tt.upserts {
				if err := s.UpsertProduct(ctx, p); err != nil {
					t.Fatalf("UpsertProduct(%v) = %v, want nil", p, err)
				}
			}
			got, err := s.GetProduct(ctx, tt.getID)
			if err != nil {
				t.Fatalf("GetProduct(%q) = %v, want nil", tt.getID, err)
			}
			if got != tt.want {
				t.Fatalf("GetProduct(%q) = %+v, want %+v", tt.getID, got, tt.want)
			}
		})
	}
}

func TestGetProductMissing(t *testing.T) {
	s := openTestStore(t)
	_, err := s.GetProduct(context.Background(), "ghost")
	if err == nil {
		t.Fatal("GetProduct(ghost) = nil, want error")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetProduct(ghost) = %v, want error wrapping sql.ErrNoRows", err)
	}
}

func TestListProductsFilter(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	seed := []domain.Product{
		{ID: "p1", Name: "Quilmes", PriceCents: 1500, Available: true},
		{ID: "p2", Name: "Fernet", PriceCents: 2500, Available: false},
		{ID: "p3", Name: "Aperol", PriceCents: 1800, Available: true},
	}
	for _, p := range seed {
		if err := s.UpsertProduct(ctx, p); err != nil {
			t.Fatalf("UpsertProduct(%v) = %v, want nil", p, err)
		}
	}

	cases := []struct {
		name          string
		onlyAvailable bool
		wantIDs       []string
	}{
		{name: "all products", onlyAvailable: false, wantIDs: []string{"p3", "p2", "p1"}},
		{name: "only available", onlyAvailable: true, wantIDs: []string{"p3", "p1"}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.ListProducts(ctx, tt.onlyAvailable)
			if err != nil {
				t.Fatalf("ListProducts(%v) = %v, want nil", tt.onlyAvailable, err)
			}
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("ListProducts(%v) has %d rows, want %d", tt.onlyAvailable, len(got), len(tt.wantIDs))
			}
			for i, id := range tt.wantIDs {
				if got[i].ID != id {
					t.Fatalf("ListProducts(%v)[%d].ID = %q, want %q", tt.onlyAvailable, i, got[i].ID, id)
				}
			}
		})
	}
}
