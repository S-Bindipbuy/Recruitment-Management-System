package database

import (
	"embed"
	"log"

	"github.com/pressly/goose/v3"
)

// migrations contains the libSQL/Turso schema. libSQL is SQLite-compatible,
// so goose's SQLite querier is used to apply it to local files, replicas and
// remote Turso databases alike.
//
//go:embed migrations/turso/*.sql
var migrations embed.FS

// RunMigrations applies the goose migrations. Versions are tracked in the
// goose_db_version table, so this is a cheap no-op on subsequent boots.
func RunMigrations() {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("turso"); err != nil {
		log.Fatalf("goose dialect %q: %v", "turso", err)
	}

	const dir = "migrations/turso"
	if err := goose.Up(DB.DB, dir); err != nil {
		log.Fatalf("migrations from %s failed: %v", dir, err)
	}
}
