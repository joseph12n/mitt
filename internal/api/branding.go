// Package api exposes the LAN HTTP API for mobile clients.
package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mitt/internal/domain"
	"mitt/internal/store"
)

// errLogoTooLarge maps oversized logo payloads to 413.
var errLogoTooLarge = errors.New("logo exceeds 512KiB")

// brandingDTO is the paintable identity exchanged over the LAN API.
// Raw logo bytes never travel here; fetch them from /api/branding/logo.
type brandingDTO struct {
	ShopName   string `json:"shop_name"`
	Primary    string `json:"primary"`
	Accent     string `json:"accent"`
	Background string `json:"background"`
	HasLogo    bool   `json:"has_logo"`
	UpdatedAt  string `json:"updated_at"`
}

// toBrandingDTO maps a stored row to its snake_case wire form.
func toBrandingDTO(row store.BrandingRow) brandingDTO {
	updated := ""
	if !row.UpdatedAt.IsZero() {
		updated = row.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	return brandingDTO{
		ShopName:   row.Branding.ShopName,
		Primary:    row.Branding.Primary,
		Accent:     row.Branding.Accent,
		Background: row.Branding.Background,
		HasLogo:    len(row.Branding.Logo) > 0,
		UpdatedAt:  updated,
	}
}

// handleBrandingGet serves public GET /api/branding so the login-less page
// can paint before pairing.
func (s *Server) handleBrandingGet(w http.ResponseWriter, r *http.Request) {
	row, err := s.store.GetBrandingRow(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "load branding")
		return
	}
	writeJSON(w, http.StatusOK, toBrandingDTO(row))
}

// handleBrandingLogo serves public GET /api/branding/logo as raw bytes.
// The ETag tracks the row timestamp, so clients cache it immutably.
func (s *Server) handleBrandingLogo(w http.ResponseWriter, r *http.Request) {
	row, err := s.store.GetBrandingRow(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "load logo")
		return
	}
	if len(row.Branding.Logo) == 0 {
		writeError(w, http.StatusNotFound, "not_found", "no logo")
		return
	}
	w.Header().Set("Content-Type", row.Branding.LogoMIME)
	w.Header().Set("ETag", strconv.Quote(row.UpdatedAt.UTC().Format(time.RFC3339Nano)))
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(row.Branding.Logo)
}

// patchBrandingRequest is the PATCH /api/branding body. Every field is
// optional; nil fields keep their stored value. LogoDataURL is a plain
// RawMessage (not a pointer) so explicit null (clear) differs from an
// absent key (keep): absent stays nil while null decodes to "null".
type patchBrandingRequest struct {
	ShopName    *string         `json:"shop_name"`
	Primary     *string         `json:"primary"`
	Accent      *string         `json:"accent"`
	Background  *string         `json:"background"`
	LogoDataURL json.RawMessage `json:"logo_data_url"`
}

// handleBrandingPatch serves authed PATCH /api/branding by merging the
// present keys onto the stored row. Domain failures are a 422 and an
// oversized logo is a 413.
func (s *Server) handleBrandingPatch(w http.ResponseWriter, r *http.Request) {
	var req patchBrandingRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	merged, err := s.store.GetBranding(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "load branding")
		return
	}
	if req.ShopName != nil {
		merged.ShopName = *req.ShopName
	}
	if req.Primary != nil {
		merged.Primary = *req.Primary
	}
	if req.Accent != nil {
		merged.Accent = *req.Accent
	}
	if req.Background != nil {
		merged.Background = *req.Background
	}
	if req.LogoDataURL != nil {
		logo, mime, clear, err := parseLogoField(req.LogoDataURL)
		if err != nil {
			if errors.Is(err, errLogoTooLarge) {
				writeError(w, http.StatusRequestEntityTooLarge, "too_large", err.Error())
				return
			}
			writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
			return
		}
		if clear {
			merged.Logo, merged.LogoMIME = nil, ""
		} else {
			merged.Logo, merged.LogoMIME = logo, mime
		}
	}
	if err := merged.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	if err := s.store.UpsertBranding(r.Context(), merged); err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "save branding")
		return
	}
	row, err := s.store.GetBrandingRow(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "load branding")
		return
	}
	writeJSON(w, http.StatusOK, toBrandingDTO(row))
}

// parseLogoField decodes one logo_data_url value: null clears, a data URL
// sets. It reports clear=true for null.
func parseLogoField(raw json.RawMessage) (logo []byte, mime string, clear bool, err error) {
	if string(raw) == "null" {
		return nil, "", true, nil
	}
	var dataURL string
	if err := json.Unmarshal(raw, &dataURL); err != nil {
		return nil, "", false, fmt.Errorf("logo_data_url must be a data URL string or null")
	}
	logo, mime, err = parseLogoDataURL(dataURL)
	if err != nil {
		return nil, "", false, err
	}
	return logo, mime, false, nil
}

// parseLogoDataURL decodes a data:<mime>;base64,... logo. Payloads that are
// certainly over the cap are rejected before base64 decoding.
func parseLogoDataURL(s string) ([]byte, string, error) {
	const prefix = "data:"
	if !strings.HasPrefix(s, prefix) {
		return nil, "", fmt.Errorf("logo_data_url must start with data:")
	}
	rest := s[len(prefix):]
	idx := strings.Index(rest, ";base64,")
	if idx < 0 {
		return nil, "", fmt.Errorf("logo_data_url must be base64")
	}
	mime := rest[:idx]
	if mime != domain.LogoMIMEPNG && mime != domain.LogoMIMEJPEG && mime != domain.LogoMIMESVG {
		return nil, "", fmt.Errorf("unsupported logo mime %q", mime)
	}
	payload := rest[idx+len(";base64,"):]
	if len(payload) > domain.MaxLogoBytes*4/3+100 {
		return nil, "", errLogoTooLarge
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", fmt.Errorf("logo is not valid base64")
	}
	if len(raw) > domain.MaxLogoBytes {
		return nil, "", errLogoTooLarge
	}
	if len(raw) == 0 {
		return nil, "", fmt.Errorf("logo is empty")
	}
	return raw, mime, nil
}
