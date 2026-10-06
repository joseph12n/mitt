// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"mitt/internal/store"
)

// Server wires domain-validated HTTP handlers over the SQLite store.
type Server struct {
	store *store.Store
}

// New builds the LAN handler: stdlib ServeMux with method patterns wrapped
// in pairing-token auth. Every response is JSON, including auth failures.
func New(s *store.Store, token string) http.Handler {
	srv := &Server{store: s}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", srv.handleHealth)
	mux.HandleFunc("GET /api/products", srv.handleProductsList)
	mux.HandleFunc("POST /api/products", srv.handleProductCreate)
	mux.HandleFunc("PATCH /api/products/{id}", srv.handleProductPatch)
	mux.HandleFunc("POST /api/tabs", srv.handleTabOpen)
	mux.HandleFunc("GET /api/tabs/open", srv.handleTabsOpen)
	mux.HandleFunc("POST /api/tabs/{id}/items", srv.handleTabAddItem)
	mux.HandleFunc("POST /api/tabs/{id}/close", srv.handleTabClose)
	mux.HandleFunc("GET /api/expenses", srv.handleExpensesList)
	mux.HandleFunc("POST /api/expenses", srv.handleExpenseCreate)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "unknown endpoint")
	})
	return AuthMiddleware(token, mux)
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
