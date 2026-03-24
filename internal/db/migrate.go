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

// Migrate runs every embedded SQL file under internal/db/migrations/ (bundled at build time via //go:embed).
// Files are executed in lexical order (001_…, 002_…, …, 043_…). The returned slice lists filenames in that order.
// Safe to call multiple times only if migrations are idempotent; 001_init.sql is not (run once on fresh DB).
func Migrate(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	applied := make([]string, 0, len(names))
	for _, name := range names {
		body, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return applied, err
		}
		log.Printf("db: running migration %s", name)
		_, err = pool.Exec(ctx, string(body))
		if err != nil {
			return applied, err
		}
		applied = append(applied, name)
	}
	return applied, nil
}
