package main

import (
	"context"
	"log"
	"os"

	"github.com/meirusfandi/fansedu-golang-api/internal/config"
	"github.com/meirusfandi/fansedu-golang-api/internal/db"
)

func main() {
	config.LoadEnvFile()
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required to run migrations")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	applied, err := db.Migrate(ctx, pool)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Printf("migrations completed (%d files): %v", len(applied), applied)
	os.Exit(0)
}
