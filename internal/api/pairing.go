package api

import (
	"net"
	"net/http"
	"strings"

	"mitt/internal/pair"
)

// bearerToken extracts the request-scoped pairing token from the
// Authorization header using the same Bearer scheme the auth middleware
// validates. It returns "" when the header is missing or malformed.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, authScheme) {
		return ""
	}
	return strings.TrimPrefix(h, authScheme)
}

// handlePairing serves GET /api/pairing with the LAN base URL and QR payload.
// The advertise override wins verbatim so tests and fixed-IP bars skip LAN
// detection; otherwise the first LAN IPv4 replaces the wildcard listen host.
// AuthMiddleware already validated the token, so the request header value is
// echoed into the pairing code with the stored token as fallback.
func (s *Server) handlePairing(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if token == "" {
		token = s.token
	}
	if adv := strings.TrimSpace(s.advertise); adv != "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"url":          adv,
			"pairing_code": pair.Code(adv, token),
		})
		return
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "no_network", "no LAN IPv4 address available")
		return
	}
	ip, err := pair.LanIPv4(addrs)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "no_network", "no LAN IPv4 address available")
		return
	}
	base := pair.BaseURL("", s.addr, ip)
	writeJSON(w, http.StatusOK, map[string]any{
		"url":          base,
		"pairing_code": pair.Code(base, token),
	})
}
