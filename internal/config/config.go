package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// Default JWT secret for development only.
const defaultDevJWTSecret = "dev-secret-change-me"

type Config struct {
	Env                    string // "development" | "production"
	Port                   string
	DatabaseURL            string
	JWTSecret              string
	OpenAIAPIKey           string // opsional: untuk generate feedback dengan AI saat submit tryout
	AppURL                 string // URL frontend/app untuk link di email (contoh: https://app.fansedu.com)
	AdminPasswordBypassKey string // opsional: kunci khusus untuk emergency reset password admin
	MigrateBypassKey       string // opsional: kunci khusus untuk emergency run migrate via API

	// Redis (opsional): cache geo wilayah, dll.
	RedisURL string
	// Geo: upstream format emsifa (default https://www.emsifa.com/api-wilayah-indonesia/api)
	GeoUpstreamBaseURL string
	// GeoCacheTTLSeconds TTL cache Redis untuk data provinsi/kabkota (default 30 hari).
	GeoCacheTTLSeconds int
	// LeaderboardCacheTTLSeconds TTL cache Redis untuk GET leaderboard per tryout (default 1 jam; di-invalidate saat submit/register).
	LeaderboardCacheTTLSeconds int
}

// LoadEnvFile loads .env for production (when ENV=production) or .env.dev for development.
// Call once at startup before Load(). In production, the host typically sets ENV=production.
func LoadEnvFile() {
	if strings.TrimSpace(os.Getenv("ENV")) == EnvProduction {
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("warning: loading .env: %v (using system env)", err)
		}
	} else {
		if err := godotenv.Load(".env.dev"); err != nil {
			log.Printf("warning: loading .env.dev: %v (using system env)", err)
		}
	}
}

func Load() Config {
	env := strings.ToLower(strings.TrimSpace(getenv("ENV", EnvDevelopment)))
	if env != EnvDevelopment && env != EnvProduction {
		env = EnvDevelopment
	}

	geoTTL := 30 * 24 * 3600 // 30 days
	if v := getenv("GEO_CACHE_TTL_SECONDS", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			geoTTL = n
		}
	}

	lbTTL := 3600 // 1 hour (invalidated on leaderboard-changing events)
	if v := getenv("LEADERBOARD_CACHE_TTL_SECONDS", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			lbTTL = n
		}
	}

	cfg := Config{
		Env:                    env,
		Port:                   getenv("PORT", "8080"),
		DatabaseURL:            getenv("DATABASE_URL", ""),
		JWTSecret:              getenv("JWT_SECRET", defaultDevJWTSecret),
		OpenAIAPIKey:           getenv("OPENAI_API_KEY", ""),
		AppURL:                 getenv("APP_URL", "http://localhost:5173"),
		AdminPasswordBypassKey: getenv("ADMIN_PASSWORD_BYPASS_KEY", ""),
		MigrateBypassKey:       getenv("MIGRATE_BYPASS_KEY", ""),
		RedisURL:               getenv("REDIS_URL", ""),
		GeoUpstreamBaseURL:     getenv("GEO_UPSTREAM_BASE_URL", "https://www.emsifa.com/api-wilayah-indonesia/api"),
		GeoCacheTTLSeconds:     geoTTL,
		LeaderboardCacheTTLSeconds: lbTTL,
	}

	if cfg.Env == EnvProduction {
		validateProduction(cfg)
	} else {
		logDevWarnings(cfg)
	}

	return cfg
}

func validateProduction(cfg Config) {
	if cfg.DatabaseURL == "" {
		log.Fatal("production: DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" || cfg.JWTSecret == defaultDevJWTSecret || cfg.JWTSecret == "change-me" {
		log.Fatal("production: JWT_SECRET must be set to a strong random value (not dev default)")
	}
	log.Printf("config: env=production port=%s", cfg.Port)
}

func logDevWarnings(cfg Config) {
	if cfg.DatabaseURL == "" {
		log.Printf("warning: DATABASE_URL is empty (db features will fail)")
	}
	if cfg.JWTSecret == defaultDevJWTSecret || cfg.JWTSecret == "change-me" {
		log.Printf("warning: using default JWT_SECRET (dev only)")
	}
	log.Printf("config: env=development port=%s", cfg.Port)
}

func (c Config) HTTPAddr() string {
	return ":" + c.Port
}

func (c Config) IsDevelopment() bool { return c.Env == EnvDevelopment }
func (c Config) IsProduction() bool  { return c.Env == EnvProduction }

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
