// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

// authScheme is the only accepted Authorization scheme.
const authScheme = "Bearer "

// GenerateToken returns a cryptographically random 32-byte pairing token
// hex-encoded for display. Callers print it once and never commit it.
func GenerateToken() (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate pairing token: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}

// AuthMiddleware enforces Bearer token auth with constant-time comparison.
// GET /api/health stays public so clients can discover the hub.
func AuthMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}
		if !validToken(r.Header.Get("Authorization"), token) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid pairing token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// validToken compares the Authorization header against the pairing token.
// An empty configured token never matches, so the API fails closed.
func validToken(header, token string) bool {
	if !strings.HasPrefix(header, authScheme) {
		return false
	}
	got := strings.TrimPrefix(header, authScheme)
	if got == "" || token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(token)) == 1
}
