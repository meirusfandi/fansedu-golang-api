package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient membuat koneksi Redis dari konfigurasi.
//
// Prioritas:
//  1. REDIS_URL (disarankan), contoh:
//     redis://localhost:6379/0
//     redis://:password@redis:6379/0
//  2. Jika REDIS_URL kosong: REDIS_ADDR + opsional REDIS_PASSWORD + REDIS_DB (default 0).
//
// Jika tidak ada REDIS_URL dan tidak ada REDIS_ADDR → (nil, nil) — cache dinonaktifkan.
func NewRedisClient(cfg Config) (*redis.Client, error) {
	if u := strings.TrimSpace(cfg.RedisURL); u != "" {
		opt, err := redis.ParseURL(u)
		if err != nil {
			return nil, fmt.Errorf("redis: parse REDIS_URL: %w", err)
		}
		return redis.NewClient(opt), nil
	}

	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr == "" {
		return nil, nil
	}

	db := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			db = n
		}
	}

	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}), nil
}

// PingRedis opsional: cek koneksi saat startup (log saja, tidak fatal).
func PingRedis(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}
	return rdb.Ping(ctx).Err()
}
