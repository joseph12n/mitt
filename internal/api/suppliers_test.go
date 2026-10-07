// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestSupplierCRUD(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	created := doRequest(t, h, http.MethodPost, "/api/suppliers",
		`{"name":"Distribuidora Sur","phone":"555-1234","note":"Entrega martes"}`, testToken)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201 (%s)", created.Code, created.Body.String())
	}
	supplier := decodeBody(t, created)
	id, _ := supplier["id"].(string)
	if id == "" || supplier["name"] != "Distribuidora Sur" || supplier["phone"] != "555-1234" {
		t.Fatalf("created supplier = %v, want Sur/555-1234 with id", supplier)
	}

	list := doRequest(t, h, http.MethodGet, "/api/suppliers", "", testToken)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", list.Code)
	}
	if suppliers := decodeBody(t, list)["suppliers"].([]any); len(suppliers) != 1 {
		t.Fatalf("suppliers = %d rows, want 1", len(suppliers))
	}

	patched := doRequest(t, h, http.MethodPatch, "/api/suppliers/"+id,
		`{"phone":"555-9999"}`, testToken)
	if patched.Code != http.StatusOK {
		t.Fatalf("patch status = %d, want 200 (%s)", patched.Code, patched.Body.String())
	}
	got := decodeBody(t, patched)
	if got["phone"] != "555-9999" || got["name"] != "Distribuidora Sur" {
		t.Fatalf("patched supplier = %v, want merged phone", got)
	}

	deleted := doRequest(t, h, http.MethodDelete, "/api/suppliers/"+id, "", testToken)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204 (%s)", deleted.Code, deleted.Body.String())
	}
	if rec := doRequest(t, h, http.MethodPatch, "/api/suppliers/"+id,
		`{"phone":"1"}`, testToken); rec.Code != http.StatusNotFound {
		t.Fatalf("patch after delete = %d, want 404", rec.Code)
	}
}

func TestSupplierValidation(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	longName := strings.Repeat("n", 81)
	longPhone := strings.Repeat("1", 41)
	longNote := strings.Repeat("x", 201)

	seeded := doRequest(t, h, http.MethodPost, "/api/suppliers", `{"name":"Hielo"}`, testToken)
	if seeded.Code != http.StatusCreated {
		t.Fatalf("seed supplier = %d, want 201", seeded.Code)
	}
	id, _ := decodeBody(t, seeded)["id"].(string)

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"empty name rejected", http.MethodPost, "/api/suppliers", `{"name":""}`, http.StatusUnprocessableEntity},
		{"blank name rejected", http.MethodPost, "/api/suppliers", `{"name":"   "}`, http.StatusUnprocessableEntity},
		{"long name rejected", http.MethodPost, "/api/suppliers", fmt.Sprintf(`{"name":%q}`, longName), http.StatusUnprocessableEntity},
		{"long phone rejected", http.MethodPost, "/api/suppliers", fmt.Sprintf(`{"name":"X","phone":%q}`, longPhone), http.StatusUnprocessableEntity},
		{"long note rejected", http.MethodPost, "/api/suppliers", fmt.Sprintf(`{"name":"X","note":%q}`, longNote), http.StatusUnprocessableEntity},
		{"malformed JSON rejected", http.MethodPost, "/api/suppliers", `{oops`, http.StatusBadRequest},
		{"patch unknown id", http.MethodPatch, "/api/suppliers/ghost", `{"phone":"1"}`, http.StatusNotFound},
		{"delete unknown id", http.MethodDelete, "/api/suppliers/ghost", "", http.StatusNotFound},
		{"patch empty name", http.MethodPatch, "/api/suppliers/" + id, `{"name":""}`, http.StatusUnprocessableEntity},
		{"patch long phone", http.MethodPatch, "/api/suppliers/" + id, fmt.Sprintf(`{"phone":%q}`, longPhone), http.StatusUnprocessableEntity},
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

func TestSupplierAuth(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	for _, tt := range []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"list", http.MethodGet, "/api/suppliers", ""},
		{"create", http.MethodPost, "/api/suppliers", `{"name":"X"}`},
		{"patch", http.MethodPatch, "/api/suppliers/ghost", `{"phone":"1"}`},
		{"delete", http.MethodDelete, "/api/suppliers/ghost", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if rec := doRequest(t, h, tt.method, tt.path, tt.body, ""); rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s %s without token = %d, want 401", tt.method, tt.path, rec.Code)
			}
			if rec := doRequest(t, h, tt.method, tt.path, tt.body, "wrong-token"); rec.Code != http.StatusUnauthorized {
				t.Fatalf("%s %s with wrong token = %d, want 401", tt.method, tt.path, rec.Code)
			}
		})
	}
}
