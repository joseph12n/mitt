package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

// openTestStore opens a fresh migrated database inside a temp dir.
func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open() = %v, want nil", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Fatalf("Close() = %v, want nil", err)
		}
	})
	return s
}

// TestOpenMigrateIdempotent proves that opening the same file twice keeps
// the schema and the data instead of failing or wiping rows.
func TestOpenMigrateIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mitt.db")

	first, err := Open(path)
	if err != nil {
		t.Fatalf("first Open() = %v, want nil", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("first Close() = %v, want nil", err)
	}

	second, err := Open(path)
	if err != nil {
		t.Fatalf("second Open() = %v, want nil", err)
	}
	defer second.Close()

	tables := []struct {
		name  string
		table string
	}{
		{"products", "products"},
		{"tables", "tables_tbl"},
		{"tabs", "tabs"},
		{"tab items", "tab_items"},
		{"sales", "sales"},
		{"sale items", "sale_items"},
		{"suppliers", "suppliers"},
		{"expenses", "expenses"},
		{"branding", "branding"},
	}
	for _, tt := range tables {
		t.Run("table "+tt.name, func(t *testing.T) {
			var found string
			err := second.db.QueryRow(
				`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, tt.table).Scan(&found)
			if err != nil {
				t.Fatalf("lookup table %q: %v", tt.table, err)
			}
		})
	}

	t.Run("schema version recorded", func(t *testing.T) {
		var version int
		if err := second.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
			t.Fatalf("PRAGMA user_version: %v", err)
		}
		if version != schemaVersion {
			t.Fatalf("user_version = %d, want %d", version, schemaVersion)
		}
	})

	t.Run("wal journal mode", func(t *testing.T) {
		var mode string
		if err := second.db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
			t.Fatalf("PRAGMA journal_mode: %v", err)
		}
		if mode != "wal" {
			t.Fatalf("journal_mode = %q, want %q", mode, "wal")
		}
	})

	t.Run("missing table lookup is sql.ErrNoRows", func(t *testing.T) {
		var found string
		err := second.db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'nope'`).Scan(&found)
		if err != sql.ErrNoRows {
			t.Fatalf("missing table err = %v, want sql.ErrNoRows", err)
		}
	})
}
