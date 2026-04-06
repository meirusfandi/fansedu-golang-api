package config

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EnvSettingKeys are settings.key values (same names as env vars) that override Config when value is non-empty.
// Bootstrap-only (never read from DB): ENV, PORT, DATABASE_URL.
var EnvSettingKeys = []string{
	"JWT_SECRET",
	"OPENAI_API_KEY",
	"APP_URL",
	"ADMIN_PASSWORD_BYPASS_KEY",
	"MIGRATE_BYPASS_KEY",
	"REDIS_URL",
	"GEO_UPSTREAM_BASE_URL",
	"GEO_CACHE_TTL_SECONDS",
	"LEADERBOARD_CACHE_TTL_SECONDS",
	"SCHOOL_LIST_CACHE_SECONDS",
	"PACKAGES_LIST_CACHE_SECONDS",
	"SMTP_HOST",
	"SMTP_PORT",
	"SMTP_USER",
	"SMTP_PASSWORD",
	"BREVO_SMTP_KEY",
	"SMTP_FROM",
	"MIDTRANS_SERVER_KEY",
	"MIDTRANS_IS_PRODUCTION",
	"MIDTRANS_SNAP_BASE_URL",
}

// IsSensitiveSettingKey returns true if list responses should mask the value (admin still sees full value on GET by id).
func IsSensitiveSettingKey(key string) bool {
	k := strings.ToUpper(strings.TrimSpace(key))
	switch {
	case strings.Contains(k, "SECRET"),
		strings.Contains(k, "PASSWORD"),
		strings.HasSuffix(k, "_KEY"),
		k == "BREVO_SMTP_KEY",
		k == "OPENAI_API_KEY":
		return true
	default:
		return false
	}
}

// ApplySettingsOverrides merges non-empty settings rows into cfg (DB wins over env for those keys).
func ApplySettingsOverrides(ctx context.Context, pool *pgxpool.Pool, cfg Config) Config {
	if pool == nil {
		return cfg
	}
	rows, err := pool.Query(ctx, `
		SELECT key, value FROM settings WHERE key = ANY($1::text[])
	`, EnvSettingKeys)
	if err != nil {
		log.Printf("warning: settings override query failed (using env only): %v", err)
		return cfg
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var k string
		var v *string
		if err := rows.Scan(&k, &v); err != nil {
			continue
		}
		if v == nil {
			continue
		}
		s := strings.TrimSpace(*v)
		if s == "" {
			continue
		}
		m[strings.TrimSpace(k)] = s
	}
	if err := rows.Err(); err != nil {
		log.Printf("warning: settings override rows: %v", err)
		return cfg
	}

	if len(m) > 0 {
		log.Printf("config: merged %d non-empty setting(s) from database over env defaults", len(m))
	}

	set := func(envKey string, dest *string) {
		if s, ok := m[envKey]; ok {
			*dest = s
		}
	}

	set("JWT_SECRET", &cfg.JWTSecret)
	set("OPENAI_API_KEY", &cfg.OpenAIAPIKey)
	set("APP_URL", &cfg.AppURL)
	set("ADMIN_PASSWORD_BYPASS_KEY", &cfg.AdminPasswordBypassKey)
	set("MIGRATE_BYPASS_KEY", &cfg.MigrateBypassKey)
	set("REDIS_URL", &cfg.RedisURL)
	set("GEO_UPSTREAM_BASE_URL", &cfg.GeoUpstreamBaseURL)

	if s, ok := m["GEO_CACHE_TTL_SECONDS"]; ok {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			cfg.GeoCacheTTLSeconds = n
		}
	}
	if s, ok := m["LEADERBOARD_CACHE_TTL_SECONDS"]; ok {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			cfg.LeaderboardCacheTTLSeconds = n
		}
	}
	if s, ok := m["SCHOOL_LIST_CACHE_SECONDS"]; ok {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			cfg.SchoolListCacheSeconds = n
		}
	}
	if s, ok := m["PACKAGES_LIST_CACHE_SECONDS"]; ok {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			cfg.PackagesListCacheSeconds = n
		}
	}

	set("SMTP_HOST", &cfg.SMTPHost)
	if s, ok := m["SMTP_PORT"]; ok {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			cfg.SMTPPort = n
		}
	}
	set("SMTP_USER", &cfg.SMTPUser)
	if s, ok := m["SMTP_PASSWORD"]; ok {
		cfg.SMTPPassword = s
	} else if s, ok := m["BREVO_SMTP_KEY"]; ok {
		cfg.SMTPPassword = s
	}
	set("SMTP_FROM", &cfg.SMTPFrom)

	set("MIDTRANS_SERVER_KEY", &cfg.MidtransServerKey)
	if s, ok := m["MIDTRANS_IS_PRODUCTION"]; ok {
		cfg.MidtransIsProduction = strings.EqualFold(s, "true") || s == "1"
	}
	set("MIDTRANS_SNAP_BASE_URL", &cfg.MidtransSnapBaseURL)

	return cfg
}
