// Package domain holds the core bar POS types.
package domain

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// validPNG returns minimal PNG bytes (signature plus one filler byte).
func validPNG() []byte {
	return append(append([]byte{}, pngMagic...), 0x00)
}

// TestBrandingValidateTable covers the white-label validation boundary.
func TestBrandingValidateTable(t *testing.T) {
	oversize := append(append([]byte{}, pngMagic...), bytes.Repeat([]byte{0x00}, MaxLogoBytes+1-len(pngMagic))...)

	cases := []struct {
		name    string
		mutate  func(*Branding)
		wantErr error
	}{
		{"defaults valid", func(*Branding) {}, nil},
		{"bad hex digits", func(b *Branding) { b.Primary = "#zzzzzz" }, ErrInvalidBrandColor},
		{"short hex", func(b *Branding) { b.Accent = "#fff" }, ErrInvalidBrandColor},
		{"missing hash", func(b *Branding) { b.Background = "121417" }, ErrInvalidBrandColor},
		{"blank shop name", func(b *Branding) { b.ShopName = "   " }, ErrInvalidShopName},
		{"empty shop name", func(b *Branding) { b.ShopName = "" }, ErrInvalidShopName},
		{"long shop name", func(b *Branding) { b.ShopName = strings.Repeat("x", 61) }, ErrInvalidShopName},
		{"oversize logo", func(b *Branding) {
			b.Logo, b.LogoMIME = oversize, LogoMIMEPNG
		}, ErrLogoTooLarge},
		{"svg with script", func(b *Branding) {
			b.Logo = []byte(`<svg><script>alert(1)</script></svg>`)
			b.LogoMIME = LogoMIMESVG
		}, ErrUnsafeSVG},
		{"svg with uppercase script", func(b *Branding) {
			b.Logo = []byte(`<svg><SCRIPT src="x"/></svg>`)
			b.LogoMIME = LogoMIMESVG
		}, ErrUnsafeSVG},
		{"mime mismatch", func(b *Branding) {
			b.Logo, b.LogoMIME = validPNG(), LogoMIMEJPEG
		}, ErrLogoMIMEMismatch},
		{"unsupported kind", func(b *Branding) {
			b.Logo = []byte("GIF89a not a logo")
			b.LogoMIME = "image/gif"
		}, ErrUnsupportedLogo},
		{"logo without mime", func(b *Branding) {
			b.Logo, b.LogoMIME = validPNG(), ""
		}, ErrLogoMIMEMismatch},
		{"mime without logo", func(b *Branding) {
			b.Logo, b.LogoMIME = nil, LogoMIMEPNG
		}, ErrLogoMIMEMismatch},
		{"valid jpeg", func(b *Branding) {
			b.Logo = []byte{0xFF, 0xD8, 0xFF, 0x00}
			b.LogoMIME = LogoMIMEJPEG
		}, nil},
		{"valid svg with prolog", func(b *Branding) {
			b.Logo = []byte("  \n" + `<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"></svg>`)
			b.LogoMIME = LogoMIMESVG
		}, nil},
		{"60 char name ok", func(b *Branding) {
			b.ShopName = strings.Repeat("y", 60)
		}, nil},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			b := DefaultBranding()
			tt.mutate(&b)
			err := b.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
