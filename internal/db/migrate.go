package db

import (
	"context"
	"embed"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate runs all embedded SQL migrations in order (by filename).
// Safe to call multiple times only if migrations are idempotent; 001_init.sql is not (run once).
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		body, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		log.Printf("db: running migration %s", name)
		_, err = pool.Exec(ctx, string(body))
		if err != nil {
			return err
		}
	}
	return nil
}
