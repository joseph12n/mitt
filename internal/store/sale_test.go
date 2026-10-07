package store

import (
	"context"
	"testing"
	"time"

	"mitt/internal/domain"
)

// saleFixture builds a sale snapshot with one line at a fixed instant.
func saleFixture(id, tableID string, total int64, closedAt time.Time) domain.Sale {
	return domain.Sale{
		ID:      id,
		TableID: tableID,
		Items: []domain.OrderItem{
			{ProductID: "p1", Name: "Quilmes", UnitPriceCents: total, Qty: 1},
		},
		TotalCents: total,
		ClosedAt:   closedAt,
	}
}

func TestSaveListSalesOrderAndLimit(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	older := saleFixture("sale-old", "t1", 1500, time.Date(2026, time.October, 5, 20, 0, 0, 0, time.UTC))
	newer := saleFixture("sale-new", "t1", 5500, time.Date(2026, time.October, 6, 20, 0, 0, 0, time.UTC))
	for _, sale := range []domain.Sale{older, newer} {
		if err := s.SaveSale(ctx, sale); err != nil {
			t.Fatalf("SaveSale(%q) = %v, want nil", sale.ID, err)
		}
	}

	got, err := s.ListSales(ctx, 50)
	if err != nil {
		t.Fatalf("ListSales() = %v, want nil", err)
	}
	if len(got) != 2 || got[0].ID != "sale-new" || got[1].ID != "sale-old" {
		t.Fatalf("ListSales() order = %+v, want newest first", got)
	}
	if len(got[0].Items) != 1 || got[0].Items[0].Name != "Quilmes" {
		t.Fatalf("ListSales()[0].Items = %+v, want one Quilmes line", got[0].Items)
	}
	if !got[0].ClosedAt.Equal(newer.ClosedAt) {
		t.Fatalf("ListSales()[0].ClosedAt = %v, want %v", got[0].ClosedAt, newer.ClosedAt)
	}

	limited, err := s.ListSales(ctx, 1)
	if err != nil {
		t.Fatalf("ListSales(1) = %v, want nil", err)
	}
	if len(limited) != 1 || limited[0].ID != "sale-new" {
		t.Fatalf("ListSales(1) = %+v, want only sale-new", limited)
	}
}

func TestSaveSaleIdempotent(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	sale := saleFixture("sale-1", "t1", 1500, time.Now().UTC())
	if err := s.SaveSale(ctx, sale); err != nil {
		t.Fatalf("first SaveSale() = %v, want nil", err)
	}
	sale.TotalCents = 2500
	sale.Items[0].UnitPriceCents = 2500
	if err := s.SaveSale(ctx, sale); err != nil {
		t.Fatalf("second SaveSale() = %v, want nil", err)
	}
	got, err := s.ListSales(ctx, 50)
	if err != nil {
		t.Fatalf("ListSales() = %v, want nil", err)
	}
	if len(got) != 1 || got[0].TotalCents != 2500 || len(got[0].Items) != 1 {
		t.Fatalf("ListSales() after re-save = %+v, want one sale at 2500", got)
	}
}

func TestTodaySummaryMath(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	now := time.Now().UTC()
	yesterday := now.Add(-24 * time.Hour)
	for _, sale := range []domain.Sale{
		saleFixture("sale-today-1", "t1", 1000, now),
		saleFixture("sale-today-2", "t2", 2500, now),
		saleFixture("sale-old", "t1", 9999, yesterday),
	} {
		if err := s.SaveSale(ctx, sale); err != nil {
			t.Fatalf("SaveSale(%q) = %v, want nil", sale.ID, err)
		}
	}

	summary, err := s.TodaySummary(ctx)
	if err != nil {
		t.Fatalf("TodaySummary() = %v, want nil", err)
	}
	if summary.Date != now.Format("2006-01-02") {
		t.Fatalf("TodaySummary().Date = %q, want %q", summary.Date, now.Format("2006-01-02"))
	}
	if summary.Count != 2 {
		t.Fatalf("TodaySummary().Count = %d, want 2", summary.Count)
	}
	if summary.TotalCents != 3500 {
		t.Fatalf("TodaySummary().TotalCents = %d, want 3500", summary.TotalCents)
	}
}

func TestTodaySummaryEmpty(t *testing.T) {
	s := openTestStore(t)
	summary, err := s.TodaySummary(context.Background())
	if err != nil {
		t.Fatalf("TodaySummary() = %v, want nil", err)
	}
	if summary.Date != time.Now().UTC().Format("2006-01-02") {
		t.Fatalf("TodaySummary().Date = %q, want today", summary.Date)
	}
	if summary.Count != 0 || summary.TotalCents != 0 {
		t.Fatalf("TodaySummary() = %+v, want zero count and total", summary)
	}
}
