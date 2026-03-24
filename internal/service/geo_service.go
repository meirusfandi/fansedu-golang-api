package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meirusfandi/fansedu-golang-api/internal/cache"
)

// ErrGeoNotFound is returned when province/regency data is not found upstream.
var ErrGeoNotFound = errors.New("geo: not found")

// GeoService serves Indonesian provinces/regencies (emsifa-compatible JSON) with optional Redis cache-aside.
type GeoService interface {
	ProvincesJSON(ctx context.Context) ([]byte, error)
	RegenciesJSON(ctx context.Context, provinceID string) ([]byte, error)
}

type geoService struct {
	rdb        *redis.Client
	upstream   string
	ttl        time.Duration
	httpClient *http.Client
}

// NewGeoService builds geo service. If rdb is nil, upstream is fetched on every request (no Redis).
func NewGeoService(rdb *redis.Client, upstreamBaseURL string, cacheTTL time.Duration) GeoService {
	base := strings.TrimRight(strings.TrimSpace(upstreamBaseURL), "/")
	if base == "" {
		base = "https://www.emsifa.com/api-wilayah-indonesia/api"
	}
	if cacheTTL <= 0 {
		cacheTTL = 7 * 24 * time.Hour // default 7 hari (sesuai rekomendasi cache wilayah)
	}
	return &geoService{
		rdb:      rdb,
		upstream: base,
		ttl:      cacheTTL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *geoService) ProvincesJSON(ctx context.Context) ([]byte, error) {
	return s.cacheAside(ctx, cache.KeyProvinceList, s.upstream+"/provinces.json")
}

func (s *geoService) RegenciesJSON(ctx context.Context, provinceID string) ([]byte, error) {
	pid := strings.TrimSpace(provinceID)
	if pid == "" {
		return nil, ErrGeoNotFound
	}
	key := cache.CityListKey(pid)
	url := fmt.Sprintf("%s/regencies/%s.json", s.upstream, pid)
	return s.cacheAside(ctx, key, url)
}

func (s *geoService) cacheAside(ctx context.Context, redisKey, fetchURL string) ([]byte, error) {
	if s.rdb != nil {
		val, err := s.rdb.Get(ctx, redisKey).Bytes()
		if err == nil && len(val) > 0 {
			return val, nil
		}
		if err != nil && err != redis.Nil {
			// Log and continue to upstream — degrade gracefully without cache
			_ = err
		}
	}

	body, status, err := s.httpGet(ctx, fetchURL)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, ErrGeoNotFound
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("upstream status %d", status)
	}

	if s.rdb != nil {
		_ = s.rdb.Set(ctx, redisKey, body, s.ttl).Err()
	}
	return body, nil
}

func (s *geoService) httpGet(ctx context.Context, u string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "fansedu-api-geo/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20)) // 8 MiB max
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return b, resp.StatusCode, nil
}
