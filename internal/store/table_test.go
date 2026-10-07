package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"mitt/internal/domain"
)

// seedTable upserts one free table row for tab tests.
func seedTable(t *testing.T, s *Store, id, label string) {
	t.Helper()
	if err := s.UpsertTable(context.Background(),
		domain.Table{ID: id, Label: label, Status: domain.TableFree}); err != nil {
		t.Fatalf("UpsertTable(%q) = %v, want nil", id, err)
	}
}

func TestUpsertListTableRoundtrip(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	for _, tbl := range []domain.Table{
		{ID: "t2", Label: "Barra", Status: domain.TableFree},
		{ID: "t1", Label: "T1", Status: domain.TableFree},
	} {
		if err := s.UpsertTable(ctx, tbl); err != nil {
			t.Fatalf("UpsertTable(%q) = %v, want nil", tbl.ID, err)
		}
	}
	got, err := s.ListTables(ctx)
	if err != nil {
		t.Fatalf("ListTables() = %v, want nil", err)
	}
	if len(got) != 2 || got[0].Label != "Barra" || got[1].Label != "T1" {
		t.Fatalf("ListTables() = %+v, want Barra then T1", got)
	}

	relabel := domain.Table{ID: "t1", Label: "Terraza", Status: domain.TableFree}
	if err := s.UpsertTable(ctx, relabel); err != nil {
		t.Fatalf("second UpsertTable() = %v, want nil", err)
	}
	one, err := s.GetTable(ctx, "t1")
	if err != nil {
		t.Fatalf("GetTable(t1) = %v, want nil", err)
	}
	if one.Label != "Terraza" {
		t.Fatalf("GetTable(t1).Label = %q, want Terraza", one.Label)
	}
}

func TestDeleteTable(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	seedTable(t, s, "t1", "T1")

	if err := s.DeleteTable(ctx, "t1"); err != nil {
		t.Fatalf("DeleteTable(t1) = %v, want nil", err)
	}
	if _, err := s.GetTable(ctx, "t1"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetTable after delete = %v, want sql.ErrNoRows", err)
	}
	if err := s.DeleteTable(ctx, "ghost"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("DeleteTable(ghost) = %v, want sql.ErrNoRows", err)
	}
}

func TestDeleteTableOccupiedGuard(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	seedTable(t, s, "t1", "T1")

	tab := tabFixture()
	if err := s.SaveTab(ctx, tab); err != nil {
		t.Fatalf("SaveTab() = %v, want nil", err)
	}
	if err := s.DeleteTable(ctx, "t1"); !errors.Is(err, ErrTableHasOpenTab) {
		t.Fatalf("DeleteTable(occupied) = %v, want ErrTableHasOpenTab", err)
	}

	tab.Status = domain.TabClosed
	if err := s.SaveTab(ctx, tab); err != nil {
		t.Fatalf("close SaveTab() = %v, want nil", err)
	}
	if err := s.DeleteTable(ctx, "t1"); err != nil {
		t.Fatalf("DeleteTable after close = %v, want nil", err)
	}
}

func TestSaveTabUnknownTable(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	open := domain.Tab{
		ID:       "tab-ghost",
		TableID:  "no-table",
		Status:   domain.TabOpen,
		OpenedAt: time.Date(2026, time.October, 6, 20, 0, 0, 0, time.UTC),
	}
	if err := s.SaveTab(ctx, open); !errors.Is(err, ErrUnknownTable) {
		t.Fatalf("SaveTab(open, unknown table) = %v, want ErrUnknownTable", err)
	}

	// Closing never strands the bill, even for a legacy tab row whose
	// table predates hub-owned tables.
	closed := open
	closed.Status = domain.TabClosed
	if err := s.SaveTab(ctx, closed); err != nil {
		t.Fatalf("SaveTab(closed, unknown table) = %v, want nil", err)
	}
}
