// Package tui renders the dashboard through the standard library only.
// It proves the ui.DashboardService seam: any graphical shell can replace
// this renderer without touching the port.
package tui

import (
	"fmt"
	"io"
	"text/tabwriter"

	"mitt/internal/ui"
)

// formatMoney renders integer cents as a big readable "$ D.CC" figure.
func formatMoney(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s$ %d.%02d", sign, cents/100, cents%100)
}

// Render prints the dashboard snapshot as plain text. Section labels are
// English (OPEN TABLES / TODAY / CATALOG). Status always pairs a symbol
// with words — never color-only (plain text has no color).
func Render(w io.Writer, d ui.DashboardData) {
	fmt.Fprintln(w, "MITT — BAR DASHBOARD")
	fmt.Fprintln(w, "====================")
	fmt.Fprintln(w, "")

	fmt.Fprintf(w, "OPEN TABLES (%d) [* occupied]\n", len(d.OpenTabs))
	if len(d.OpenTabs) == 0 {
		fmt.Fprintln(w, "No open tables — all tables FREE [o].")
	} else {
		tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "TABLE\tITEMS\tTOTAL")
		var openTotal int64
		for _, t := range d.OpenTabs {
			fmt.Fprintf(tw, "%s\t%d\t%s\n", t.TableID, t.ItemCount, formatMoney(t.TotalCents))
			openTotal += t.TotalCents
		}
		_ = tw.Flush()
		fmt.Fprintf(w, "Open total: %s across %d table(s).\n", formatMoney(openTotal), len(d.OpenTabs))
	}
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, "TODAY [$ day sales]")
	fmt.Fprintf(w, "Sales today: %s\n", formatMoney(d.TodaySalesCents))
	if d.TodaySalesCents == 0 {
		fmt.Fprintln(w, "No closed sales counted yet — first close appears here. [o idle]")
	}
	fmt.Fprintln(w, "")

	fmt.Fprintf(w, "CATALOG (%d product(s))\n", d.ProductCount)
	if len(d.OutOfStock) == 0 {
		fmt.Fprintln(w, "[OK] All products available.")
	} else {
		fmt.Fprintf(w, "[!] OUT OF STOCK (%d):\n", len(d.OutOfStock))
		for _, name := range d.OutOfStock {
			fmt.Fprintf(w, "  - %s [! OUT]\n", name)
		}
	}
}
