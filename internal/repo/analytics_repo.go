package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsSummary struct {
	TotalPageviews int `json:"totalPageviews"`
	Today          int `json:"today"`
	ThisWeek       int `json:"thisWeek"`
	ThisMonth      int `json:"thisMonth"`
	UniqueVisitors int `json:"uniqueVisitors"`
}

type AnalyticsVisitor struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId"`
	Path      string `json:"path"`
	IPAddress string `json:"ipAddress,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type AnalyticsRepo interface {
	TrackPageView(ctx context.Context, sessionID, path, ipAddress, userAgent string) error
	GetSummary(ctx context.Context) (AnalyticsSummary, error)
	ListVisitors(ctx context.Context, page, limit int) ([]AnalyticsVisitor, int, error)
}

type analyticsRepo struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepo(pool *pgxpool.Pool) AnalyticsRepo {
	return &analyticsRepo{pool: pool}
}

func (r *analyticsRepo) TrackPageView(ctx context.Context, sessionID, path, ipAddress, userAgent string) error {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO analytics_pageviews (id, session_id, path, ip_address, user_agent)
		SELECT $1::uuid, $2, $3, $4, $5
		WHERE NOT EXISTS (
			SELECT 1
			FROM analytics_pageviews ap
			WHERE ap.session_id = $2
			  AND ap.path = $3
			  AND ap.created_at > NOW() - INTERVAL '5 minutes'
		)
	`, id, sessionID, path, nullIfEmpty(ipAddress), nullIfEmpty(userAgent))
	return err
}

func (r *analyticsRepo) GetSummary(ctx context.Context) (AnalyticsSummary, error) {
	var s AnalyticsSummary
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int AS total_pageviews,
			COUNT(*) FILTER (WHERE created_at >= date_trunc('day', NOW()))::int AS today,
			COUNT(*) FILTER (WHERE created_at >= date_trunc('week', NOW()))::int AS this_week,
			COUNT(*) FILTER (WHERE created_at >= date_trunc('month', NOW()))::int AS this_month,
			COUNT(DISTINCT session_id)::int AS unique_visitors
		FROM analytics_pageviews
	`).Scan(&s.TotalPageviews, &s.Today, &s.ThisWeek, &s.ThisMonth, &s.UniqueVisitors)
	return s, err
}

func (r *analyticsRepo) ListVisitors(ctx context.Context, page, limit int) ([]AnalyticsVisitor, int, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM analytics_pageviews`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, session_id, path, COALESCE(ip_address, ''), COALESCE(user_agent, ''), created_at::text
		FROM analytics_pageviews
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]AnalyticsVisitor, 0, limit)
	for rows.Next() {
		var v AnalyticsVisitor
		if err := rows.Scan(&v.ID, &v.SessionID, &v.Path, &v.IPAddress, &v.UserAgent, &v.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}
