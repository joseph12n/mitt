package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"mitt/internal/domain"
)

// tabFixture builds an open tab with two lines at a fixed second.
func tabFixture() domain.Tab {
	return domain.Tab{
		ID:       "tab1",
		TableID:  "t1",
		Status:   domain.TabOpen,
		OpenedAt: time.Date(2026, time.October, 6, 20, 0, 0, 0, time.UTC),
		Items: []domain.OrderItem{
			{ProductID: "p1", Name: "Quilmes", UnitPriceCents: 1500, Qty: 2},
			{ProductID: "p2", Name: "Fernet", UnitPriceCents: 2500, Qty: 1},
		},
	}
}

func TestSaveGetTabRoundtrip(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	want := tabFixture()

	if err := s.SaveTab(ctx, want); err != nil {
		t.Fatalf("SaveTab() = %v, want nil", err)
	}
	got, err := s.GetOpenTabByTable(ctx, want.TableID)
	if err != nil {
		t.Fatalf("GetOpenTabByTable(%q) = %v, want nil", want.TableID, err)
	}
	if got.ID != want.ID || got.TableID != want.TableID || got.Status != want.Status {
		t.Fatalf("GetOpenTabByTable() header = %+v, want %+v", got, want)
	}
	if !got.OpenedAt.Equal(want.OpenedAt) {
		t.Fatalf("OpenedAt = %v, want %v", got.OpenedAt, want.OpenedAt)
	}
	if len(got.Items) != len(want.Items) {
		t.Fatalf("items = %d rows, want %d", len(got.Items), len(want.Items))
	}
	for i, item := range want.Items {
		if got.Items[i] != item {
			t.Fatalf("items[%d] = %+v, want %+v", i, got.Items[i], item)
		}
	}
	if got.TotalCents() != want.TotalCents() {
		t.Fatalf("TotalCents() = %d, want %d", got.TotalCents(), want.TotalCents())
	}
}

func TestSaveTabReplacesItems(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	tab := tabFixture()

	if err := s.SaveTab(ctx, tab); err != nil {
		t.Fatalf("first SaveTab() = %v, want nil", err)
	}
	tab.Items = []domain.OrderItem{
		{ProductID: "p3", Name: "Aperol", UnitPriceCents: 1800, Qty: 3},
	}
	if err := s.SaveTab(ctx, tab); err != nil {
		t.Fatalf("second SaveTab() = %v, want nil", err)
	}
	got, err := s.GetOpenTabByTable(ctx, tab.TableID)
	if err != nil {
		t.Fatalf("GetOpenTabByTable() = %v, want nil", err)
	}
	if len(got.Items) != 1 || got.Items[0] != tab.Items[0] {
		t.Fatalf("items after replace = %+v, want %+v", got.Items, tab.Items)
	}
	if got.TotalCents() != tab.TotalCents() {
		t.Fatalf("TotalCents() = %d, want %d", got.TotalCents(), tab.TotalCents())
	}
}

func TestGetOpenTabMissing(t *testing.T) {
	s := openTestStore(t)
	_, err := s.GetOpenTabByTable(context.Background(), "no-table")
	if err == nil {
		t.Fatal("GetOpenTabByTable(no-table) = nil, want error")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("GetOpenTabByTable(no-table) = %v, want error wrapping sql.ErrNoRows", err)
	}
}

func TestListOpenTabs(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	older := tabFixture()
	older.ID = "tab-old"
	older.TableID = "t1"
	older.OpenedAt = time.Date(2026, time.October, 6, 19, 0, 0, 0, time.UTC)
	newer := tabFixture()
	newer.ID = "tab-new"
	newer.TableID = "t2"
	newer.OpenedAt = time.Date(2026, time.October, 6, 21, 0, 0, 0, time.UTC)
	closed := tabFixture()
	closed.ID = "tab-closed"
	closed.TableID = "t3"
	closed.Status = domain.TabClosed

	for _, tab := range []domain.Tab{newer, closed, older} {
		if err := s.SaveTab(ctx, tab); err != nil {
			t.Fatalf("SaveTab(%q) = %v, want nil", tab.ID, err)
		}
	}

	got, err := s.ListOpenTabs(ctx)
	if err != nil {
		t.Fatalf("ListOpenTabs() = %v, want nil", err)
	}
	wantIDs := []string{"tab-old", "tab-new"}
	if len(got) != len(wantIDs) {
		t.Fatalf("ListOpenTabs() has %d rows, want %d", len(got), len(wantIDs))
	}
	for i, id := range wantIDs {
		if got[i].ID != id {
			t.Fatalf("ListOpenTabs()[%d].ID = %q, want %q", i, got[i].ID, id)
		}
		if len(got[i].Items) != 2 {
			t.Fatalf("ListOpenTabs()[%d] has %d items, want 2", i, len(got[i].Items))
		}
	}
}
