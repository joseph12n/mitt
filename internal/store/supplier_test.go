package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"mitt/internal/domain"
)

func TestUpsertGetSupplierRoundtrip(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	want := domain.Supplier{ID: "s1", Name: "Distribuidora Sur", Phone: "555-1234", Note: "Entrega martes"}
	if err := s.UpsertSupplier(ctx, want); err != nil {
		t.Fatalf("UpsertSupplier() = %v, want nil", err)
	}
	got, err := s.GetSupplier(ctx, "s1")
	if err != nil {
		t.Fatalf("GetSupplier(s1) = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("GetSupplier(s1) = %+v, want %+v", got, want)
	}

	// Upserting the same id replaces the row.
	updated := domain.Supplier{ID: "s1", Name: "Distribuidora Norte", Phone: "", Note: ""}
	if err := s.UpsertSupplier(ctx, updated); err != nil {
		t.Fatalf("second UpsertSupplier() = %v, want nil", err)
	}
	got, err = s.GetSupplier(ctx, "s1")
	if err != nil {
		t.Fatalf("GetSupplier(s1) after update = %v, want nil", err)
	}
	if got != updated {
		t.Fatalf("GetSupplier(s1) after update = %+v, want %+v", got, updated)
	}
}

func TestListSuppliersByName(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	for _, sup := range []domain.Supplier{
		{ID: "s2", Name: "Zeta"},
		{ID: "s1", Name: "Alfa"},
	} {
		if err := s.UpsertSupplier(ctx, sup); err != nil {
			t.Fatalf("UpsertSupplier(%q) = %v, want nil", sup.ID, err)
		}
	}
	got, err := s.ListSuppliers(ctx)
	if err != nil {
		t.Fatalf("ListSuppliers() = %v, want nil", err)
	}
	if len(got) != 2 || got[0].Name != "Alfa" || got[1].Name != "Zeta" {
		t.Fatalf("ListSuppliers() = %+v, want Alfa then Zeta", got)
	}
}

func TestDeleteSupplier(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	if err := s.UpsertSupplier(ctx, domain.Supplier{ID: "s1", Name: "Hielo"}); err != nil {
		t.Fatalf("UpsertSupplier() = %v, want nil", err)
	}
	if err := s.DeleteSupplier(ctx, "s1"); err != nil {
		t.Fatalf("DeleteSupplier(s1) = %v, want nil", err)
	}
	if _, err := s.GetSupplier(ctx, "s1"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetSupplier after delete = %v, want sql.ErrNoRows", err)
	}
	if err := s.DeleteSupplier(ctx, "ghost"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("DeleteSupplier(ghost) = %v, want sql.ErrNoRows", err)
	}
}

func TestGetSupplierMissing(t *testing.T) {
	s := openTestStore(t)
	if _, err := s.GetSupplier(context.Background(), "ghost"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetSupplier(ghost) = %v, want sql.ErrNoRows", err)
	}
}
