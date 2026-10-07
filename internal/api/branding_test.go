// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"mitt/internal/domain"
)

// brandingShape asserts the GET /api/branding JSON keys.
func brandingShape(t *testing.T, body map[string]any) {
	t.Helper()
	for _, key := range []string{"shop_name", "primary", "accent", "background", "has_logo", "updated_at"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("branding body missing %q: %v", key, body)
		}
	}
}

func TestBrandingGetPublic(t *testing.T) {
	h, _ := openTestAPI(t, testToken)
	rec := doRequest(t, h, http.MethodGet, "/api/branding", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/branding = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	brandingShape(t, body)
	if body["shop_name"] != "mitt" {
		t.Fatalf("shop_name = %v, want mitt", body["shop_name"])
	}
	if body["has_logo"] != false {
		t.Fatalf("has_logo = %v, want false", body["has_logo"])
	}
}

func TestBrandingPatchAuthAndValidation(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	if rec := doRequest(t, h, http.MethodPatch, "/api/branding",
		`{"shop_name":"X"}`, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("PATCH without token = %d, want 401", rec.Code)
	}

	if rec := doRequest(t, h, http.MethodPatch, "/api/branding",
		`{"primary":"nope"}`, testToken); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("PATCH bad color = %d, want 422 (%s)", rec.Code, rec.Body.String())
	} else {
		decodeBody(t, rec)
	}

	rec := doRequest(t, h, http.MethodPatch, "/api/branding",
		`{"shop_name":"Bar Central","primary":"#112233","accent":"#445566","background":"#778899"}`, testToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH colors = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	if body["shop_name"] != "Bar Central" || body["primary"] != "#112233" {
		t.Fatalf("patched branding = %v, want merged colors", body)
	}

	// Absent keys keep stored values.
	rec = doRequest(t, h, http.MethodPatch, "/api/branding",
		`{"accent":"#AABBCC"}`, testToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH merge = %d, want 200", rec.Code)
	}
	if body := decodeBody(t, rec); body["shop_name"] != "Bar Central" || body["accent"] != "#AABBCC" {
		t.Fatalf("merged branding = %v, want kept name and new accent", body)
	}
}

func TestBrandingLogoFlow(t *testing.T) {
	h, _ := openTestAPI(t, testToken)

	if rec := doRequest(t, h, http.MethodGet, "/api/branding/logo", "", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("GET logo unset = %d, want 404", rec.Code)
	}

	raw := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x01}
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(raw)
	rec := doRequest(t, h, http.MethodPatch, "/api/branding",
		`{"logo_data_url":`+strconvQuote(dataURL)+`}`, testToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH logo = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	if body := decodeBody(t, rec); body["has_logo"] != true {
		t.Fatalf("has_logo = %v, want true", body["has_logo"])
	}

	logo := doRequest(t, h, http.MethodGet, "/api/branding/logo", "", "")
	if logo.Code != http.StatusOK {
		t.Fatalf("GET logo = %d, want 200", logo.Code)
	}
	if ct := logo.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("logo Content-Type = %q, want image/png", ct)
	}
	if etag := logo.Header().Get("ETag"); etag == "" {
		t.Fatalf("logo ETag missing")
	}
	if cc := logo.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Fatalf("logo Cache-Control = %q, want immutable", cc)
	}
	if logo.Body.String() != string(raw) {
		t.Fatalf("logo bytes differ, want exact roundtrip")
	}

	// Null clears the logo.
	rec = doRequest(t, h, http.MethodPatch, "/api/branding",
		`{"logo_data_url":null}`, testToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH clear logo = %d, want 200", rec.Code)
	}
	if rec := doRequest(t, h, http.MethodGet, "/api/branding/logo", "", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("GET logo after clear = %d, want 404", rec.Code)
	}

	// Oversized payloads are rejected before decoding.
	big := strings.Repeat("A", domain.MaxLogoBytes*4/3+101)
	rec = doRequest(t, h, http.MethodPatch, "/api/branding",
		`{"logo_data_url":"data:image/png;base64,`+big+`"}`, testToken)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("PATCH oversize logo = %d, want 413", rec.Code)
	}
}

// strconvQuote quotes s for embedding in a JSON test body.
func strconvQuote(s string) string {
	out := make([]byte, 0, len(s)+2)
	out = append(out, '"')
	for i := 0; i < len(s); i++ {
		if s[i] == '"' || s[i] == '\\' {
			out = append(out, '\\')
		}
		out = append(out, s[i])
	}
	return string(append(out, '"'))
}
