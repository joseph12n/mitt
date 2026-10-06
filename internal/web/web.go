// Package web serves the embedded single-page dashboard for bar staff.
// It is pure HTML+JS with no external dependencies (the bar LAN may be
// offline). All API calls go to the same origin the page was served from.
package web

import (
	_ "embed"
	"net/http"
)

//go:embed static/index.html
var indexPage []byte

// Handler returns the dashboard page with an explicit HTML content type.
// Exact-path guarding ("/" only) is the caller's job in internal/api.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexPage)
	})
}
