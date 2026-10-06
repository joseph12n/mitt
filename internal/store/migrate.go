package store

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed migrations/schema.sql
var schemaSQL string

// schemaVersion is the latest migration level applied by migrate.
const schemaVersion = 1

// migrate applies the embedded schema when the database is unversioned.
// It is idempotent: a database already at schemaVersion is left alone.
func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version >= schemaVersion {
		return nil
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("apply schema v%d: %w", schemaVersion, err)
	}
	if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, schemaVersion)); err != nil {
		return fmt.Errorf("record schema version: %w", err)
	}
	return nil
}
