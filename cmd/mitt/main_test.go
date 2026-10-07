package main

import (
	"testing"
)

// TestResolveTokenPrecedence checks flag beats env beats random, and reports
// the source so logs can tell stable tokens apart from one-run ones.
func TestResolveTokenPrecedence(t *testing.T) {
	t.Setenv("MITT_TOKEN", "env-token-1234567890")

	token, source, err := resolveToken("flag-token-1234567890")
	if err != nil || token != "flag-token-1234567890" || source != "flag" {
		t.Fatalf("flag: got (%q, %q, %v)", token, source, err)
	}

	token, source, err = resolveToken("")
	if err != nil || token != "env-token-1234567890" || source != "env" {
		t.Fatalf("env: got (%q, %q, %v)", token, source, err)
	}
}
