// Package domain holds the core bar POS types.
package domain

import (
	"fmt"
	"strings"
)

// White-label branding limits and logo kinds.
const (
	// MaxLogoBytes caps stored logos so modest bar PCs stay lean.
	MaxLogoBytes = 512 * 1024

	// LogoMIMEPNG is the only accepted raster loss-less logo kind.
	LogoMIMEPNG = "image/png"
	// LogoMIMEJPEG is the accepted photographic logo kind.
	LogoMIMEJPEG = "image/jpeg"
	// LogoMIMESVG is the accepted vector logo kind.
	LogoMIMESVG = "image/svg+xml"
)

// Branding errors.
var (
	ErrInvalidShopName   = fmt.Errorf("shop name must be 1..60 characters")
	ErrInvalidBrandColor = fmt.Errorf("branding color must match #rrggbb")
	ErrLogoTooLarge      = fmt.Errorf("logo must not exceed 512KiB")
	ErrUnsupportedLogo   = fmt.Errorf("logo must be PNG, JPEG, or SVG")
	ErrLogoMIMEMismatch  = fmt.Errorf("logo mime does not match image kind")
	ErrUnsafeSVG         = fmt.Errorf("svg logo must not contain scripts")
)

// Branding is the white-label identity painted by every client.
// Colors are #rrggbb hex; Logo is optional raw file bytes.
type Branding struct {
	ShopName   string
	Primary    string
	Accent     string
	Background string
	Logo       []byte
	LogoMIME   string
}

// DefaultBranding returns the night-bar-first identity from
// docs/design-tokens.md: surface #1C1F24 as primary, accent #E8A33D,
// background bg #121417.
func DefaultBranding() Branding {
	return Branding{
		ShopName:   "mitt",
		Primary:    "#1C1F24",
		Accent:     "#E8A33D",
		Background: "#121417",
	}
}

// Validate reports whether the branding holds paintable data.
func (b Branding) Validate() error {
	name := strings.TrimSpace(b.ShopName)
	if n := len([]rune(name)); n < 1 || n > 60 {
		return fmt.Errorf("validate branding: %w", ErrInvalidShopName)
	}
	for _, c := range []struct {
		role string
		hex  string
	}{
		{"primary", b.Primary},
		{"accent", b.Accent},
		{"background", b.Background},
	} {
		if !validHexColor(c.hex) {
			return fmt.Errorf("validate branding %s: %w", c.role, ErrInvalidBrandColor)
		}
	}
	if len(b.Logo) == 0 {
		if b.LogoMIME != "" {
			return fmt.Errorf("validate branding logo: %w", ErrLogoMIMEMismatch)
		}
		return nil
	}
	if len(b.Logo) > MaxLogoBytes {
		return fmt.Errorf("validate branding logo: %w", ErrLogoTooLarge)
	}
	kind := logoKind(b.Logo)
	if kind == "" {
		return fmt.Errorf("validate branding logo: %w", ErrUnsupportedLogo)
	}
	if kind == LogoMIMESVG && hasScriptTag(b.Logo) {
		return fmt.Errorf("validate branding logo: %w", ErrUnsafeSVG)
	}
	if b.LogoMIME != kind {
		return fmt.Errorf("validate branding logo: %w", ErrLogoMIMEMismatch)
	}
	return nil
}

// validHexColor reports whether s matches ^#[0-9a-fA-F]{6}$.
func validHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		if c < '0' || (c > '9' && c < 'A') || (c > 'F' && c < 'a') || c > 'f' {
			return false
		}
	}
	return true
}

// pngMagic is the 8-byte PNG signature.
var pngMagic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

// logoKind sniffs PNG, JPEG, and SVG bytes, returning the matching MIME
// type or "" when the bytes match nothing supported.
func logoKind(logo []byte) string {
	if len(logo) >= 8 && string(logo[:8]) == string(pngMagic) {
		return LogoMIMEPNG
	}
	if len(logo) >= 3 && logo[0] == 0xFF && logo[1] == 0xD8 && logo[2] == 0xFF {
		return LogoMIMEJPEG
	}
	if isSVG(logo) {
		return LogoMIMESVG
	}
	return ""
}

// isSVG reports whether logo starts with <svg after optional whitespace
// and one optional XML prolog.
func isSVG(logo []byte) bool {
	s := string(logo)
	i := skipSpace(s, 0)
	if strings.HasPrefix(s[i:], "<?xml") {
		end := strings.Index(s[i:], "?>")
		if end < 0 {
			return false
		}
		i = skipSpace(s, i+end+2)
	}
	if len(s)-i < 5 || !strings.EqualFold(s[i:i+4], "<svg") {
		return false
	}
	switch c := s[i+4]; c {
	case ' ', '\t', '\n', '\r', '>', '/':
		return true
	default:
		return false
	}
}

// hasScriptTag reports whether raw SVG contains a <script tag,
// case-insensitive.
func hasScriptTag(logo []byte) bool {
	return strings.Contains(strings.ToLower(string(logo)), "<script")
}

// skipSpace advances past ASCII whitespace.
func skipSpace(s string, i int) int {
	for i < len(s) {
		switch s[i] {
		case ' ', '\t', '\n', '\r', '\f', '\v':
			i++
		default:
			return i
		}
	}
	return i
}
