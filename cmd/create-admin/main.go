package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/meirusfandi/fansedu-golang-api/internal/config"
	"github.com/meirusfandi/fansedu-golang-api/internal/db"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

func getRequiredEnv(key string) (string, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return "", errors.New(key + " is required")
	}
	return v, nil
}

func main() {
	config.LoadEnvFile()
	cfg := config.Load()
	adminEmail, err := getRequiredEnv("BOOTSTRAP_ADMIN_EMAIL")
	if err != nil {
		log.Fatal(err)
	}
	adminPassword, err := getRequiredEnv("BOOTSTRAP_ADMIN_PASSWORD")
	if err != nil {
		log.Fatal(err)
	}
	adminName := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_NAME"))
	if adminName == "" {
		adminName = "Administrator"
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required. Set it in .env or .env.dev")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	userRepo := repo.NewUserRepo(pool)
	existing, err := userRepo.FindByEmail(ctx, adminEmail)
	if err == nil {
		log.Printf("admin already exists: id=%s email=%s", existing.ID, existing.Email)
		os.Exit(0)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	u := domain.User{
		Email:        adminEmail,
		PasswordHash: string(hash),
		Name:         adminName,
		Role:         domain.UserRoleAdmin,
	}
	created, err := userRepo.Create(ctx, u)
	if err != nil {
		log.Fatalf("create admin: %v", err)
	}
	log.Printf("admin created: id=%s email=%s name=%s", created.ID, created.Email, created.Name)
}
