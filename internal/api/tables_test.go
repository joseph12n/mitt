// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// errorCode extracts the {"error":{"code"}} value from a JSON response.
func errorCode(t *testing.T, recBody map[string]any) string {
	t.Helper()
	errObj, ok := recBody["error"].(map[string]any)
	if !ok {
		t.Fatalf("error envelope = %v, want error object", recBody)
	}
	code, _ := errObj["code"].(string)
	return code
}

func TestTablesCRUD(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	created := doRequest(t, h, http.MethodPost, "/api/tables", `{"label":"T1"}`, testToken)
	if created.Code != http.StatusCreated {
		t.Fatalf("create table = %d, want 201 (%s)", created.Code, created.Body.String())
	}
	table := decodeBody(t, created)
	id, _ := table["id"].(string)
	if id == "" || table["label"] != "T1" || table["occupied"] != false {
		t.Fatalf("created table = %v, want T1/free with id", table)
	}

	list := doRequest(t, h, http.MethodGet, "/api/tables", "", testToken)
	if list.Code != http.StatusOK {
		t.Fatalf("list tables = %d, want 200", list.Code)
	}
	if tables := decodeBody(t, list)["tables"].([]any); len(tables) != 1 {
		t.Fatalf("tables = %d rows, want 1", len(tables))
	}

	// Opening a tab for an unknown table is a 422: the hub owns the tables.
	unknown := doRequest(t, h, http.MethodPost, "/api/tabs",
		`{"table_id":"ghost"}`, testToken)
	if unknown.Code != http.StatusUnprocessableEntity {
		t.Fatalf("open tab unknown table = %d, want 422", unknown.Code)
	}
	if code := errorCode(t, decodeBody(t, unknown)); code != "unknown_table" {
		t.Fatalf("open tab unknown table code = %q, want unknown_table", code)
	}

	tabID := openTab(t, h, id)

	occupied := doRequest(t, h, http.MethodGet, "/api/tables", "", testToken)
	rows := decodeBody(t, occupied)["tables"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["occupied"] != true {
		t.Fatalf("tables after open = %v, want T1 occupied", rows)
	}

	// Deleting an occupied table is a 409, never an orphaned bill.
	busy := doRequest(t, h, http.MethodDelete, "/api/tables/"+id, "", testToken)
	if busy.Code != http.StatusConflict {
		t.Fatalf("delete occupied table = %d, want 409", busy.Code)
	}
	if code := errorCode(t, decodeBody(t, busy)); code != "table_occupied" {
		t.Fatalf("delete occupied table code = %q, want table_occupied", code)
	}

	// Settle the bill, then the table deletes cleanly.
	beer := seedProduct(t, h, `{"name":"Quilmes","price_cents":1500}`)
	if rec := doRequest(t, h, http.MethodPost, "/api/tabs/"+tabID+"/items",
		fmt.Sprintf(`{"product_id":%q,"qty":1}`, beer), testToken); rec.Code != http.StatusOK {
		t.Fatalf("add item = %d, want 200", rec.Code)
	}
	if rec := doRequest(t, h, http.MethodPost, "/api/tabs/"+tabID+"/close", "", testToken); rec.Code != http.StatusOK {
		t.Fatalf("close tab = %d, want 200", rec.Code)
	}
	if rec := doRequest(t, h, http.MethodDelete, "/api/tables/"+id, "", testToken); rec.Code != http.StatusNoContent {
		t.Fatalf("delete free table = %d, want 204 (%s)", rec.Code, rec.Body.String())
	}
	if rec := doRequest(t, h, http.MethodDelete, "/api/tables/"+id, "", testToken); rec.Code != http.StatusNotFound {
		t.Fatalf("delete missing table = %d, want 404", rec.Code)
	}
	if rec := doRequest(t, h, http.MethodGet, "/api/tables", "", testToken); rec.Code != http.StatusOK {
		t.Fatalf("list after delete = %d, want 200", rec.Code)
	} else if tables := decodeBody(t, rec)["tables"].([]any); len(tables) != 0 {
		t.Fatalf("tables after delete = %d rows, want 0", len(tables))
	}
}

func TestTableCreateValidation(t *testing.T) {
	h, _ := openTestAPI(t, testToken)
	cases := []struct {
		name string
		body string
	}{
		{"empty label", `{"label":""}`},
		{"blank label", `{"label":"   "}`},
		{"label over forty chars", fmt.Sprintf(`{"label":%q}`, strings.Repeat("m", 41))},
		{"malformed JSON", `{oops`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, h, http.MethodPost, "/api/tables", tt.body, testToken)
			want := http.StatusUnprocessableEntity
			if tt.name == "malformed JSON" {
				want = http.StatusBadRequest
			}
			if rec.Code != want {
				t.Fatalf("POST /api/tables %s = %d, want %d (%s)", tt.name, rec.Code, want, rec.Body.String())
			}
			decodeBody(t, rec)
		})
	}
}

func TestTablesUnauthorized(t *testing.T) {
	h, _ := openTestAPI(t, testToken)
	id := seedTable(t, h, "T1")
	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list tables", http.MethodGet, "/api/tables", ""},
		{"create table", http.MethodPost, "/api/tables", `{"label":"T2"}`},
		{"delete table", http.MethodDelete, "/api/tables/" + id, ""},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, h, tt.method, tt.path, tt.body, "")
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s %s = %d, want 401", tt.method, tt.path, rec.Code)
			}
			if code := errorCode(t, decodeBody(t, rec)); code != "unauthorized" {
				t.Fatalf("error code = %q, want unauthorized", code)
			}
		})
	}
}
