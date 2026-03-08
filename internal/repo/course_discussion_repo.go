package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type CourseDiscussionRepo interface {
	Create(ctx context.Context, d domain.CourseDiscussion) (domain.CourseDiscussion, error)
	GetByID(ctx context.Context, id string) (domain.CourseDiscussion, error)
	ListByCourseID(ctx context.Context, courseID string, limit int) ([]domain.CourseDiscussion, error)
}

type courseDiscussionRepo struct{ pool *pgxpool.Pool }

func NewCourseDiscussionRepo(pool *pgxpool.Pool) CourseDiscussionRepo {
	return &courseDiscussionRepo{pool: pool}
}

func (r *courseDiscussionRepo) Create(ctx context.Context, d domain.CourseDiscussion) (domain.CourseDiscussion, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO course_discussions (id, course_id, user_id, title, body)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5)
		RETURNING created_at, updated_at
	`, id, d.CourseID, d.UserID, d.Title, d.Body).Scan(&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return domain.CourseDiscussion{}, err
	}
	d.ID = id
	return d, nil
}

func (r *courseDiscussionRepo) GetByID(ctx context.Context, id string) (domain.CourseDiscussion, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, course_id, user_id, title, body, created_at, updated_at
		FROM course_discussions WHERE id = $1::uuid
	`, id)
	var d domain.CourseDiscussion
	err := row.Scan(&d.ID, &d.CourseID, &d.UserID, &d.Title, &d.Body, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

func (r *courseDiscussionRepo) ListByCourseID(ctx context.Context, courseID string, limit int) ([]domain.CourseDiscussion, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, course_id, user_id, title, body, created_at, updated_at
		FROM course_discussions WHERE course_id = $1::uuid ORDER BY created_at DESC LIMIT $2
	`, courseID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.CourseDiscussion
	for rows.Next() {
		var d domain.CourseDiscussion
		if err := rows.Scan(&d.ID, &d.CourseID, &d.UserID, &d.Title, &d.Body, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

type CourseDiscussionReplyRepo interface {
	Create(ctx context.Context, r domain.CourseDiscussionReply) (domain.CourseDiscussionReply, error)
	ListByDiscussionID(ctx context.Context, discussionID string, limit int) ([]domain.CourseDiscussionReply, error)
}

type courseDiscussionReplyRepo struct{ pool *pgxpool.Pool }

func NewCourseDiscussionReplyRepo(pool *pgxpool.Pool) CourseDiscussionReplyRepo {
	return &courseDiscussionReplyRepo{pool: pool}
}

func (r *courseDiscussionReplyRepo) Create(ctx context.Context, reply domain.CourseDiscussionReply) (domain.CourseDiscussionReply, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO course_discussion_replies (id, discussion_id, user_id, body)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4)
		RETURNING created_at, updated_at
	`, id, reply.DiscussionID, reply.UserID, reply.Body).Scan(&reply.CreatedAt, &reply.UpdatedAt)
	if err != nil {
		return domain.CourseDiscussionReply{}, err
	}
	reply.ID = id
	return reply, nil
}

func (r *courseDiscussionReplyRepo) ListByDiscussionID(ctx context.Context, discussionID string, limit int) ([]domain.CourseDiscussionReply, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, discussion_id, user_id, body, created_at, updated_at
		FROM course_discussion_replies WHERE discussion_id = $1::uuid ORDER BY created_at ASC LIMIT $2
	`, discussionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.CourseDiscussionReply
	for rows.Next() {
		var reply domain.CourseDiscussionReply
		if err := rows.Scan(&reply.ID, &reply.DiscussionID, &reply.UserID, &reply.Body, &reply.CreatedAt, &reply.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, reply)
	}
	return list, rows.Err()
}
