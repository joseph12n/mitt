package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
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
	// Vite build emits <div id="root"> as the React mount point; section
	// titles render client-side, so the shell marker is the stable assert.
	body := rec.Body.String()
	if !strings.Contains(body, `<div id="root"`) {
		t.Fatalf("dashboard body misses root mount marker (%d bytes)", len(body))
	}
	if !strings.Contains(body, "/assets/") {
		t.Fatalf("dashboard body misses bundled asset reference (%d bytes)", len(body))
	}
}

// TestHandlerServesAssets checks a bundled file is reachable with caching.
func TestHandlerServesAssets(t *testing.T) {
	// Find one real bundled asset so the test tracks the build output.
	var asset string
	_ = fs.WalkDir(distFS, "static/dist/assets", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && asset == "" {
			asset = "/" + strings.TrimPrefix(path, "static/dist/")
		}
		return nil
	})
	if asset == "" {
		t.Skip("no bundled assets present; run make web-build first")
	}
	req := httptest.NewRequest(http.MethodGet, asset, nil)
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", asset, rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Fatalf("Cache-Control = %q, want immutable", cc)
	}
}

// TestHandlerMissingIndexFallsBack503 covers a binary built from a checkout
// where static/dist holds only .gitkeep (fresh clone without web-build).
func TestHandlerMissingIndexFallsBack503(t *testing.T) {
	h := handlerForFS(fstest.MapFS{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("GET / with empty FS = %d, want 503", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "dashboard not built") {
		t.Fatalf("503 body = %q, want build hint", body)
	}
}
