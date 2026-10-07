// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"fmt"
	"net/http"
	"testing"
)

// closeTabFlow seeds a product, table, and tab, adds one line, and closes
// it, returning the sale body.
func closeTabFlow(t *testing.T, h http.Handler, tableLabel string, priceCents, qty int) map[string]any {
	t.Helper()
	beer := seedProduct(t, h, fmt.Sprintf(`{"name":%q,"price_cents":%d}`, "Beer-"+tableLabel, priceCents))
	tabID := openTab(t, h, seedTable(t, h, tableLabel))
	if rec := doRequest(t, h, http.MethodPost, "/api/tabs/"+tabID+"/items",
		fmt.Sprintf(`{"product_id":%q,"qty":%d}`, beer, qty), testToken); rec.Code != http.StatusOK {
		t.Fatalf("add item = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	closed := doRequest(t, h, http.MethodPost, "/api/tabs/"+tabID+"/close", "", testToken)
	if closed.Code != http.StatusOK {
		t.Fatalf("close = %d, want 200 (%s)", closed.Code, closed.Body.String())
	}
	return decodeBody(t, closed)
}

func TestSalesListMatchesCloseShape(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	empty := doRequest(t, h, http.MethodGet, "/api/sales", "", testToken)
	if empty.Code != http.StatusOK {
		t.Fatalf("list empty = %d, want 200", empty.Code)
	}
	if sales := decodeBody(t, empty)["sales"].([]any); len(sales) != 0 {
		t.Fatalf("sales = %d rows, want 0", len(sales))
	}

	closed := closeTabFlow(t, h, "T1", 1500, 2)

	list := doRequest(t, h, http.MethodGet, "/api/sales", "", testToken)
	if list.Code != http.StatusOK {
		t.Fatalf("list sales = %d, want 200", list.Code)
	}
	sales := decodeBody(t, list)["sales"].([]any)
	if len(sales) != 1 {
		t.Fatalf("sales = %d rows, want 1", len(sales))
	}
	row := sales[0].(map[string]any)
	// Same field names as the close-sale snapshot.
	for _, key := range []string{"id", "table_id", "items", "total_cents", "closed_at"} {
		if _, ok := row[key]; !ok {
			t.Fatalf("sale row missing %q: %v", key, row)
		}
	}
	if row["id"] != closed["id"] || row["total_cents"] != closed["total_cents"] {
		t.Fatalf("sale row = %v, want close snapshot %v", row, closed)
	}
	items := row["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("sale items = %d rows, want 1", len(items))
	}
}

func TestSalesListLimit(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	closeTabFlow(t, h, "T1", 1500, 1)
	closeTabFlow(t, h, "T2", 2500, 1)

	one := doRequest(t, h, http.MethodGet, "/api/sales?limit=1", "", testToken)
	if one.Code != http.StatusOK {
		t.Fatalf("list limit=1 = %d, want 200", one.Code)
	}
	if sales := decodeBody(t, one)["sales"].([]any); len(sales) != 1 {
		t.Fatalf("sales limit=1 = %d rows, want 1", len(sales))
	}

	capped := doRequest(t, h, http.MethodGet, "/api/sales?limit=999", "", testToken)
	if capped.Code != http.StatusOK {
		t.Fatalf("list limit=999 = %d, want 200 (capped)", capped.Code)
	}
	if sales := decodeBody(t, capped)["sales"].([]any); len(sales) != 2 {
		t.Fatalf("sales limit=999 = %d rows, want 2 (capped, not rejected)", len(sales))
	}

	for _, query := range []string{"?limit=abc", "?limit=0", "?limit=-3"} {
		t.Run("bad limit "+query, func(t *testing.T) {
			rec := doRequest(t, h, http.MethodGet, "/api/sales"+query, "", testToken)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("list %s = %d, want 400 (%s)", query, rec.Code, rec.Body.String())
			}
			decodeBody(t, rec)
		})
	}
}

func TestSalesToday(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	closed := closeTabFlow(t, h, "T1", 1500, 2) // total 3000

	rec := doRequest(t, h, http.MethodGet, "/api/sales/today", "", testToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("today = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	if body["count"] != float64(1) {
		t.Fatalf("count = %v, want 1", body["count"])
	}
	if body["total_cents"] != closed["total_cents"] {
		t.Fatalf("total_cents = %v, want %v", body["total_cents"], closed["total_cents"])
	}
	date, _ := body["date"].(string)
	if len(date) != 10 || date[4] != '-' || date[7] != '-' {
		t.Fatalf("date = %q, want YYYY-MM-DD", body["date"])
	}
}

func TestSalesAuth(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	for _, path := range []string{"/api/sales", "/api/sales/today"} {
		if rec := doRequest(t, h, http.MethodGet, path, "", ""); rec.Code != http.StatusUnauthorized {
			t.Fatalf("GET %s without token = %d, want 401", path, rec.Code)
		}
	}
}
