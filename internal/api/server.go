// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"mitt/internal/store"
	"mitt/internal/web"
)

// Server wires domain-validated HTTP handlers over the SQLite store.
type Server struct {
	store     *store.Store
	token     string
	advertise string
	addr      string
}

// New builds the LAN handler: stdlib ServeMux with method patterns wrapped
// in pairing-token auth. Every response is JSON, including auth failures.
func New(s *store.Store, token, advertise, addr string) http.Handler {
	srv := &Server{store: s, token: token, advertise: advertise, addr: addr}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", srv.handleHealth)
	mux.HandleFunc("GET /api/pairing", srv.handlePairing)
	mux.HandleFunc("GET /api/products", srv.handleProductsList)
	mux.HandleFunc("POST /api/products", srv.handleProductCreate)
	mux.HandleFunc("PATCH /api/products/{id}", srv.handleProductPatch)
	mux.HandleFunc("POST /api/tabs", srv.handleTabOpen)
	mux.HandleFunc("GET /api/tabs/open", srv.handleTabsOpen)
	mux.HandleFunc("POST /api/tabs/{id}/items", srv.handleTabAddItem)
	mux.HandleFunc("POST /api/tabs/{id}/close", srv.handleTabClose)
	mux.HandleFunc("GET /api/expenses", srv.handleExpensesList)
	mux.HandleFunc("POST /api/expenses", srv.handleExpenseCreate)
	// GET / serves the human dashboard: mux longest-match keeps every
	// /api route first, and only the exact root path gets the page while any
	// other unmatched path stays a 404. The outer handler also serves this
	// same page publicly (no token); see below.
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		web.Handler().ServeHTTP(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "unknown endpoint")
	})
	// The dashboard page and its fingerprinted bundles are public so staff
	// browsers load them without a token; the JSON API stays behind pairing
	// auth. Browsers also probe /favicon.ico on their own.
	authed := AuthMiddleware(token, mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && (r.URL.Path == "/" || strings.HasPrefix(r.URL.Path, "/assets/")) {
			web.Handler().ServeHTTP(w, r)
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/favicon.ico" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		authed.ServeHTTP(w, r)
	})
}

// handleHealth reports liveness without auth so clients can find the hub.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// writeJSON encodes v as JSON with the matching content type.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError encodes the shared {"error":{"code","message"}} envelope.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

// decodeJSON parses one JSON object body. It reports 400 on malformed input.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return false
	}
	return true
}

// newID returns a random 128-bit hex identifier for rows created via the API.
func newID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}
