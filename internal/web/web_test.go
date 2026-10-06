package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandlerServesDashboard checks the page is public HTML for bar staff.
func TestHandlerServesDashboard(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
	if body := rec.Body.String(); !strings.Contains(body, "MESAS") {
		t.Fatalf("dashboard body misses MESAS section (%d bytes)", len(body))
	}
}
