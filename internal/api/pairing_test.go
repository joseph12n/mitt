package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"mitt/internal/store"
)

// openPairingAPI builds a handler with an advertise override so pairing
// tests stay deterministic without LAN detection.
func openPairingAPI(t *testing.T, token, advertise, addr string) http.Handler {
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
	return New(s, token, advertise, addr)
}

func TestPairingAdvertiseOverride(t *testing.T) {
	h := openPairingAPI(t, testToken, "http://192.168.1.20:8080", ":8080")
	rec := doRequest(t, h, http.MethodGet, "/api/pairing", "", testToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/pairing = %d, want 200 (%s)", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	if body["url"] != "http://192.168.1.20:8080" {
		t.Fatalf("url = %v, want advertise override", body["url"])
	}
	code, _ := body["pairing_code"].(string)
	if !strings.HasPrefix(code, "mitt://pair?") || !strings.Contains(code, "url=") || !strings.Contains(code, testToken) {
		t.Fatalf("pairing_code = %q, want mitt://pair with url and token", code)
	}
}

func TestPairingUnauthorized(t *testing.T) {
	h := openPairingAPI(t, testToken, "http://192.168.1.20:8080", ":8080")
	req := httptest.NewRequest(http.MethodGet, "/api/pairing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/pairing without token = %d, want 401", rec.Code)
	}
}
