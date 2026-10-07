// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"mitt/internal/store"
)

const testToken = "test-pairing-token"

// openTestAPI builds a handler over a TempDir database for API tests.
func openTestAPI(t *testing.T, token string) (http.Handler, *store.Store) {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open() = %v, want nil", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Fatalf("Close() = %v, want nil", err)
		}
	})
	return New(s, token, "", ":8080"), s
}

// doRequest performs one JSON request against the handler.
func doRequest(t *testing.T, h http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// decodeBody parses a JSON response object into a generic map.
func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	var v map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return v
}

func TestHealthNoAuth(t *testing.T) {
	h, _ := openTestAPI(t, testToken)
	rec := doRequest(t, h, http.MethodGet, "/api/health", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/health = %d, want 200", rec.Code)
	}
	if got := decodeBody(t, rec)["status"]; got != "ok" {
		t.Fatalf("status = %v, want ok", got)
	}
}

func TestUnauthorized(t *testing.T) {
	h, _ := openTestAPI(t, testToken)
	cases := []struct {
		name   string
		header string
	}{
		{"missing header", ""},
		{"wrong token", "Bearer wrong-token"},
		{"wrong scheme", "Token " + testToken},
		{"empty bearer", "Bearer "},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
			body := decodeBody(t, rec)
			errObj, ok := body["error"].(map[string]any)
			if !ok || errObj["code"] != "unauthorized" {
				t.Fatalf("error envelope = %v, want unauthorized code", body)
			}
		})
	}
}

func TestProductCRUD(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	created := doRequest(t, h, http.MethodPost, "/api/products",
		`{"name":"Quilmes","price_cents":1500}`, testToken)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201 (%s)", created.Code, created.Body.String())
	}
	product := decodeBody(t, created)
	id, _ := product["id"].(string)
	if id == "" || product["name"] != "Quilmes" || product["price_cents"] != float64(1500) {
		t.Fatalf("created product = %v, want Quilmes/1500 with id", product)
	}

	list := doRequest(t, h, http.MethodGet, "/api/products", "", testToken)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", list.Code)
	}
	if products := decodeBody(t, list)["products"].([]any); len(products) != 1 {
		t.Fatalf("products = %d rows, want 1", len(products))
	}

	patched := doRequest(t, h, http.MethodPatch, "/api/products/"+id,
		`{"price_cents":1800}`, testToken)
	if patched.Code != http.StatusOK {
		t.Fatalf("patch status = %d, want 200", patched.Code)
	}
	if got := decodeBody(t, patched)["price_cents"]; got != float64(1800) {
		t.Fatalf("patched price = %v, want 1800", got)
	}

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"empty name rejected", http.MethodPost, "/api/products", `{"name":"","price_cents":100}`, http.StatusUnprocessableEntity},
		{"negative price rejected", http.MethodPost, "/api/products", `{"name":"X","price_cents":-1}`, http.StatusUnprocessableEntity},
		{"malformed JSON rejected", http.MethodPost, "/api/products", `{oops`, http.StatusBadRequest},
		{"patch unknown id", http.MethodPatch, "/api/products/ghost", `{"name":"X"}`, http.StatusNotFound},
		{"patch negative price", http.MethodPatch, "/api/products/" + id, `{"price_cents":-5}`, http.StatusUnprocessableEntity},
		{"patch empty name", http.MethodPatch, "/api/products/" + id, `{"name":""}`, http.StatusUnprocessableEntity},
		{"bad filter value", http.MethodGet, "/api/products?available=maybe", "", http.StatusBadRequest},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, h, tt.method, tt.path, tt.body, testToken)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s %s = %d, want %d (%s)", tt.method, tt.path, rec.Code, tt.wantStatus, rec.Body.String())
			}
			decodeBody(t, rec) // every response stays a JSON envelope
		})
	}
}

func TestProductAvailableFilter(t *testing.T) {
	h, _ := openTestAPI(t, testToken)
	for _, body := range []string{
		`{"name":"Quilmes","price_cents":1500}`,
		`{"name":"Seasonal","price_cents":999,"available":false}`,
	} {
		if rec := doRequest(t, h, http.MethodPost, "/api/products", body, testToken); rec.Code != http.StatusCreated {
			t.Fatalf("create = %d, want 201", rec.Code)
		}
	}
	cases := []struct {
		name  string
		query string
		want  int
	}{
		{"all products", "", 2},
		{"only available", "?available=true", 1},
		{"only unavailable", "?available=false", 1},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, h, http.MethodGet, "/api/products"+tt.query, "", testToken)
			if rec.Code != http.StatusOK {
				t.Fatalf("list = %d, want 200", rec.Code)
			}
			if got := decodeBody(t, rec)["products"].([]any); len(got) != tt.want {
				t.Fatalf("products = %d rows, want %d", len(got), tt.want)
			}
		})
	}
}

// seedProduct creates a product and returns its id.
func seedProduct(t *testing.T, h http.Handler, body string) string {
	t.Helper()
	rec := doRequest(t, h, http.MethodPost, "/api/products", body, testToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed product %s = %d, want 201", body, rec.Code)
	}
	id, _ := decodeBody(t, rec)["id"].(string)
	if id == "" {
		t.Fatalf("seeded product has no id: %s", rec.Body.String())
	}
	return id
}

// seedTable creates a hub table and returns its id.
func seedTable(t *testing.T, h http.Handler, label string) string {
	t.Helper()
	rec := doRequest(t, h, http.MethodPost, "/api/tables",
		fmt.Sprintf(`{"label":%q}`, label), testToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed table %s = %d, want 201 (%s)", label, rec.Code, rec.Body.String())
	}
	id, _ := decodeBody(t, rec)["id"].(string)
	if id == "" {
		t.Fatalf("seeded table has no id: %s", rec.Body.String())
	}
	return id
}

// openTab opens a tab for the table and returns its id.
func openTab(t *testing.T, h http.Handler, table string) string {
	t.Helper()
	rec := doRequest(t, h, http.MethodPost, "/api/tabs",
		fmt.Sprintf(`{"table_id":%q}`, table), testToken)
	if rec.Code != http.StatusCreated && rec.Code != http.StatusOK {
		t.Fatalf("open tab = %d, want 201 (%s)", rec.Code, rec.Body.String())
	}
	id, _ := decodeBody(t, rec)["id"].(string)
	if id == "" {
		t.Fatalf("opened tab has no id: %s", rec.Body.String())
	}
	return id
}

func TestTabOpenAddCloseTotals(t *testing.T) {
	h, _ := openTestAPI(t, testToken)
	beer := seedProduct(t, h, `{"name":"Quilmes","price_cents":1500}`)
	fernet := seedProduct(t, h, `{"name":"Fernet","price_cents":2500}`)

	tabID := openTab(t, h, seedTable(t, h, "T1"))

	for _, item := range []struct {
		product string
		qty     int
	}{
		{beer, 2},
		{fernet, 1},
	} {
		rec := doRequest(t, h, http.MethodPost, "/api/tabs/"+tabID+"/items",
			fmt.Sprintf(`{"product_id":%q,"qty":%d}`, item.product, item.qty), testToken)
		if rec.Code != http.StatusOK {
			t.Fatalf("add item = %d, want 200 (%s)", rec.Code, rec.Body.String())
		}
	}

	open := doRequest(t, h, http.MethodGet, "/api/tabs/open", "", testToken)
	if open.Code != http.StatusOK {
		t.Fatalf("list open = %d, want 200", open.Code)
	}
	if tabs := decodeBody(t, open)["tabs"].([]any); len(tabs) != 1 {
		t.Fatalf("open tabs = %d, want 1", len(tabs))
	}

	// Catalog edits after ordering must not rewrite the running bill.
	if rec := doRequest(t, h, http.MethodPatch, "/api/products/"+beer, `{"price_cents":9999}`, testToken); rec.Code != http.StatusOK {
		t.Fatalf("reprice = %d, want 200", rec.Code)
	}

	closed := doRequest(t, h, http.MethodPost, "/api/tabs/"+tabID+"/close", "", testToken)
	if closed.Code != http.StatusOK {
		t.Fatalf("close = %d, want 200 (%s)", closed.Code, closed.Body.String())
	}
	sale := decodeBody(t, closed)
	// 2x1500 + 1x2500 at order-time prices, not the repriced 9999.
	if sale["total_cents"] != float64(5500) {
		t.Fatalf("total_cents = %v, want 5500", sale["total_cents"])
	}
	if sale["id"] != tabID {
		t.Fatalf("sale id = %v, want tab %s", sale["id"], tabID)
	}

	// Closed tabs leave the open list and reject further items with 404.
	if rec := doRequest(t, h, http.MethodGet, "/api/tabs/open", "", testToken); rec.Code != http.StatusOK {
		t.Fatalf("list open after close = %d, want 200", rec.Code)
	} else if tabs := decodeBody(t, rec)["tabs"].([]any); len(tabs) != 0 {
		t.Fatalf("open tabs after close = %d, want 0", len(tabs))
	}
	if rec := doRequest(t, h, http.MethodPost, "/api/tabs/"+tabID+"/items",
		fmt.Sprintf(`{"product_id":%q,"qty":1}`, fernet), testToken); rec.Code != http.StatusNotFound {
		t.Fatalf("add to closed tab = %d, want 404", rec.Code)
	}
}

func TestTabValidation(t *testing.T) {
	h, _ := openTestAPI(t, testToken)
	beer := seedProduct(t, h, `{"name":"Quilmes","price_cents":1500}`)
	stale := seedProduct(t, h, `{"name":"Seasonal","price_cents":999,"available":false}`)

	if rec := doRequest(t, h, http.MethodPost, "/api/tabs", `{"table_id":""}`, testToken); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("open empty table = %d, want 422", rec.Code)
	}

	emptyTable := seedTable(t, h, "T-empty")
	emptyTab := openTab(t, h, emptyTable)
	table := seedTable(t, h, "T1")
	tab := openTab(t, h, table)
	// Idempotent reopen returns the same tab instead of duplicating it.
	if rec := doRequest(t, h, http.MethodPost, "/api/tabs", fmt.Sprintf(`{"table_id":%q}`, table), testToken); rec.Code != http.StatusOK {
		t.Fatalf("reopen = %d, want 200", rec.Code)
	} else if id, _ := decodeBody(t, rec)["id"].(string); id != tab {
		t.Fatalf("reopen id = %q, want %q", id, tab)
	}

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"zero qty rejected", http.MethodPost, "/api/tabs/" + tab + "/items", fmt.Sprintf(`{"product_id":%q,"qty":0}`, beer), http.StatusUnprocessableEntity},
		{"unavailable product rejected", http.MethodPost, "/api/tabs/" + tab + "/items", fmt.Sprintf(`{"product_id":%q,"qty":1}`, stale), http.StatusUnprocessableEntity},
		{"unknown product", http.MethodPost, "/api/tabs/" + tab + "/items", `{"product_id":"ghost","qty":1}`, http.StatusNotFound},
		{"unknown tab", http.MethodPost, "/api/tabs/ghost/items", fmt.Sprintf(`{"product_id":%q,"qty":1}`, beer), http.StatusNotFound},
		{"close empty tab", http.MethodPost, "/api/tabs/" + emptyTab + "/close", "", http.StatusUnprocessableEntity},
		{"close unknown tab", http.MethodPost, "/api/tabs/ghost/close", "", http.StatusNotFound},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, h, tt.method, tt.path, tt.body, testToken)
			if rec.Code != tt.wantStatus {
				t.Fatalf("%s %s = %d, want %d (%s)", tt.method, tt.path, rec.Code, tt.wantStatus, rec.Body.String())
			}
			decodeBody(t, rec)
		})
	}
}

func TestExpenseAddList(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	for _, body := range []string{
		`{"description":"Ice bags","qty":10,"cost_cents":5000}`,
		`{"description":"Limes","qty":2.5,"cost_cents":1200}`,
	} {
		if rec := doRequest(t, h, http.MethodPost, "/api/expenses", body, testToken); rec.Code != http.StatusCreated {
			t.Fatalf("create expense = %d, want 201 (%s)", rec.Code, rec.Body.String())
		}
	}

	list := doRequest(t, h, http.MethodGet, "/api/expenses", "", testToken)
	if list.Code != http.StatusOK {
		t.Fatalf("list expenses = %d, want 200", list.Code)
	}
	expenses := decodeBody(t, list)["expenses"].([]any)
	if len(expenses) != 2 {
		t.Fatalf("expenses = %d rows, want 2", len(expenses))
	}

	cases := []struct {
		name string
		body string
	}{
		{"empty description", `{"description":"","qty":1,"cost_cents":100}`},
		{"zero qty", `{"description":"X","qty":0,"cost_cents":100}`},
		{"negative qty", `{"description":"X","qty":-2,"cost_cents":100}`},
		{"negative cost", `{"description":"X","qty":1,"cost_cents":-1}`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, h, http.MethodPost, "/api/expenses", tt.body, testToken)
			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("create %s = %d, want 422 (%s)", tt.name, rec.Code, rec.Body.String())
			}
			decodeBody(t, rec)
		})
	}
}

// TestDashboardRootPublicUnknownNotFound covers server-level routing: the
// dashboard page is public HTML on GET / while unknown paths stay 404.
func TestDashboardRootPublicUnknownNotFound(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	root := httptest.NewRequest(http.MethodGet, "/", nil)
	rootRec := httptest.NewRecorder()
	h.ServeHTTP(rootRec, root)
	if rootRec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", rootRec.Code)
	}
	if ct := rootRec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("GET / Content-Type = %q, want text/html", ct)
	}

	miss := doRequest(t, h, http.MethodGet, "/nope", "", testToken)
	if miss.Code != http.StatusNotFound {
		t.Fatalf("GET /nope = %d, want 404", miss.Code)
	}

	// Bundled assets must load without a token or the page stays blank:
	// the test bundle is unbuilt, so any non-401 (here 404) proves the
	// request reached the public static handler instead of the auth wall.
	asset := doRequest(t, h, http.MethodGet, "/assets/index-test.js", "", "")
	if asset.Code == http.StatusUnauthorized {
		t.Fatalf("GET /assets/* without token = 401, want public (non-401)")
	}

	favicon := doRequest(t, h, http.MethodGet, "/favicon.ico", "", "")
	if favicon.Code != http.StatusNoContent {
		t.Fatalf("GET /favicon.ico = %d, want 204", favicon.Code)
	}
}
