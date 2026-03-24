package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetJSON cache-aside GET; true jika hit.
func GetJSON(ctx context.Context, rdb *redis.Client, key string, dest interface{}) (hit bool, err error) {
	if rdb == nil {
		return false, nil
	}
	val, err := rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(val, dest); err != nil {
		return false, err
	}
	return true, nil
}

// SetJSON SET dengan TTL.
func SetJSON(ctx context.Context, rdb *redis.Client, key string, v interface{}, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, b, ttl).Err()
}

// Del hapus satu key.
func Del(ctx context.Context, rdb *redis.Client, keys ...string) {
	if rdb == nil || len(keys) == 0 {
		return
	}
	_ = rdb.Del(ctx, keys...).Err()
}

// InvalidateSchoolList hapus cache daftar sekolah (panggil setelah create/update/delete).
func InvalidateSchoolList(ctx context.Context, rdb *redis.Client) {
	Del(ctx, rdb, KeySchoolList)
}

// InvalidatePackagesList hapus cache GET /packages.
func InvalidatePackagesList(ctx context.Context, rdb *redis.Client) {
	Del(ctx, rdb, KeyPackagesList)
}
