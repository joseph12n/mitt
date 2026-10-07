// Package web serves the embedded Vite + React dashboard for bar staff.
// The UI is built offline-safe (no CDN, system fonts only) by
// `make web-build` into static/dist and embedded in the binary, so the bar
// LAN needs no internet. All API calls go to the same origin.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static/dist
var distFS embed.FS

// Handler returns the dashboard page plus its bundled assets. Exact-path
// guarding ("/" only) is the caller's job in internal/api.
func Handler() http.Handler {
	return handlerForFS(distFS)
}

// handlerForFS serves the dashboard from an arbitrary filesystem rooted like
// the embed (static/dist/index.html plus static/dist/assets/*), so tests can
// cover the unbuilt-bundle fallback without touching the real embed.
func handlerForFS(fsys fs.FS) http.Handler {
	assets, err := fs.Sub(fsys, "static/dist/assets")
	index, indexErr := fs.ReadFile(fsys, "static/dist/index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bundled JS/CSS under /assets/* ship with long-lived caching.
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			http.StripPrefix("/assets/", http.FileServer(http.FS(assets))).ServeHTTP(w, r)
			return
		}
		// The SPA entry point; without a built bundle there is no dashboard.
		if indexErr != nil {
			http.Error(w, "dashboard not built — run make web-build", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(index)
	})
}
