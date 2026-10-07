package store

import (
	"bytes"
	"context"
	"testing"

	"mitt/internal/domain"
)

// TestBrandingDefaultsAndRoundtrip proves the single-row branding store:
// defaults when absent, exact bytes on roundtrip, overwrite on update.
func TestBrandingDefaultsAndRoundtrip(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	got, err := s.GetBranding(ctx)
	if err != nil {
		t.Fatalf("GetBranding() = %v, want nil", err)
	}
	if want := domain.DefaultBranding(); got.ShopName != want.ShopName ||
		got.Primary != want.Primary || got.Accent != want.Accent ||
		got.Background != want.Background || len(got.Logo) != 0 || got.LogoMIME != "" {
		t.Fatalf("GetBranding() absent = %+v, want defaults %+v", got, want)
	}

	logo := append(append([]byte{}, 0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A), 0x01, 0x02)
	want := domain.Branding{
		ShopName:   "Bar Central",
		Primary:    "#112233",
		Accent:     "#445566",
		Background: "#778899",
		Logo:       logo,
		LogoMIME:   domain.LogoMIMEPNG,
	}
	if err := s.UpsertBranding(ctx, want); err != nil {
		t.Fatalf("UpsertBranding() = %v, want nil", err)
	}
	got, err = s.GetBranding(ctx)
	if err != nil {
		t.Fatalf("GetBranding() = %v, want nil", err)
	}
	if got.ShopName != want.ShopName || got.Primary != want.Primary ||
		got.Accent != want.Accent || got.Background != want.Background ||
		got.LogoMIME != want.LogoMIME || !bytes.Equal(got.Logo, want.Logo) {
		t.Fatalf("GetBranding() = %+v, want %+v", got, want)
	}

	row, err := s.GetBrandingRow(ctx)
	if err != nil {
		t.Fatalf("GetBrandingRow() = %v, want nil", err)
	}
	if row.UpdatedAt.IsZero() {
		t.Fatalf("GetBrandingRow().UpdatedAt is zero, want stored timestamp")
	}

	updated := want
	updated.Primary = "#AABBCC"
	updated.Logo, updated.LogoMIME = nil, ""
	if err := s.UpsertBranding(ctx, updated); err != nil {
		t.Fatalf("UpsertBranding() update = %v, want nil", err)
	}
	got, err = s.GetBranding(ctx)
	if err != nil {
		t.Fatalf("GetBranding() = %v, want nil", err)
	}
	if got.Primary != "#AABBCC" || len(got.Logo) != 0 || got.LogoMIME != "" {
		t.Fatalf("GetBranding() after update = %+v, want cleared logo and new primary", got)
	}

	if err := s.UpsertBranding(ctx, domain.Branding{ShopName: ""}); err == nil {
		t.Fatalf("UpsertBranding() invalid = nil, want domain error")
	}
}
