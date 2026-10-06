package tui

import (
	"bytes"
	"strings"
	"testing"

	"mitt/internal/ui"
)

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		data ui.DashboardData
		// want contains fragments that must appear in the output.
		want []string
		// notWant contains fragments that must not appear.
		notWant []string
	}{
		{
			name: "empty dashboard shows headers and idle states",
			data: ui.DashboardData{},
			want: []string{
				"OPEN TABLES",
				"TODAY",
				"CATALOG",
				"No open tables",
				"$ 0.00",
				"All products available",
			},
			notWant: []string{"OUT OF STOCK"},
		},
		{
			name: "populated dashboard shows tables totals and out of stock",
			data: ui.DashboardData{
				OpenTabs: []ui.OpenTabView{
					{TableID: "t1", ItemCount: 3, TotalCents: 5500},
					{TableID: "t2", ItemCount: 1, TotalCents: 1800},
				},
				TodaySalesCents: 12000,
				ProductCount:    3,
				OutOfStock:      []string{"Aperol"},
			},
			want: []string{
				"OPEN TABLES (2)",
				"t1",
				"t2",
				"$ 55.00",
				"$ 18.00",
				"$ 73.00",
				"TODAY",
				"$ 120.00",
				"CATALOG (3",
				"OUT OF STOCK",
				"Aperol",
			},
			notWant: []string{"No open tables", "All products available"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Render(&buf, tt.data)
			got := buf.String()
			for _, fragment := range tt.want {
				if !strings.Contains(got, fragment) {
					t.Fatalf("Render() missing %q in:\n%s", fragment, got)
				}
			}
			for _, fragment := range tt.notWant {
				if strings.Contains(got, fragment) {
					t.Fatalf("Render() unexpectedly contains %q in:\n%s", fragment, got)
				}
			}
		})
	}
}
