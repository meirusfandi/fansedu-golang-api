package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type QuestionRepo interface {
	Create(ctx context.Context, q domain.Question) (domain.Question, error)
	GetByID(ctx context.Context, id string) (domain.Question, error)
	ListByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Question, error)
	Update(ctx context.Context, q domain.Question) error
	Delete(ctx context.Context, id string) error
}

type questionRepo struct{ pool *pgxpool.Pool }

func NewQuestionRepo(pool *pgxpool.Pool) QuestionRepo { return &questionRepo{pool: pool} }

func (r *questionRepo) Create(ctx context.Context, q domain.Question) (domain.Question, error) {
	imageURLs := q.ImageURLs
	if imageURLs == nil {
		imageURLs = []byte("[]")
	}
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO questions (tryout_session_id, sort_order, type, body, image_url, image_urls, options, max_score)
		VALUES ($1::uuid, $2, $3::question_type, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`, q.TryoutSessionID, q.SortOrder, q.Type, q.Body, q.ImageURL, imageURLs, q.Options, q.MaxScore).Scan(&id, &q.CreatedAt)
	if err != nil {
		return domain.Question{}, err
	}
	q.ID = id
	return q, nil
}

func (r *questionRepo) GetByID(ctx context.Context, id string) (domain.Question, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tryout_session_id, sort_order, type, body, image_url, COALESCE(image_urls, '[]'::jsonb), options, max_score, created_at
		FROM questions WHERE id = $1::uuid
	`, id)
	var q domain.Question
	err := row.Scan(&q.ID, &q.TryoutSessionID, &q.SortOrder, &q.Type, &q.Body, &q.ImageURL, &q.ImageURLs, &q.Options, &q.MaxScore, &q.CreatedAt)
	return q, err
}

func (r *questionRepo) ListByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Question, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tryout_session_id, sort_order, type, body, image_url, COALESCE(image_urls, '[]'::jsonb), options, max_score, created_at
		FROM questions WHERE tryout_session_id = $1::uuid ORDER BY sort_order
	`, tryoutSessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Question
	for rows.Next() {
		var q domain.Question
		if err := rows.Scan(&q.ID, &q.TryoutSessionID, &q.SortOrder, &q.Type, &q.Body, &q.ImageURL, &q.ImageURLs, &q.Options, &q.MaxScore, &q.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, q)
	}
	return list, rows.Err()
}

func (r *questionRepo) Update(ctx context.Context, q domain.Question) error {
	imageURLs := q.ImageURLs
	if imageURLs == nil {
		imageURLs = []byte("[]")
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE questions SET sort_order=$2, type=$3::question_type, body=$4, image_url=$5, image_urls=$6, options=$7, max_score=$8 WHERE id = $1::uuid
	`, q.ID, q.SortOrder, q.Type, q.Body, q.ImageURL, imageURLs, q.Options, q.MaxScore)
	return err
}

func (r *questionRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM questions WHERE id = $1::uuid`, id)
	return err
}
