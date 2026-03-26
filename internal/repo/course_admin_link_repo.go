package repo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CourseLinkedPackageRow — paket landing yang memuat course ini (package_courses).
type CourseLinkedPackageRow struct {
	ID   string
	Name string
	Slug string
}

// CourseLinkedTryoutRow — tryout terhubung ke course (course_tryouts).
type CourseLinkedTryoutRow struct {
	ID        string
	Title     string
	Status    string
	OpensAt   string
	ClosesAt  string
	SortOrder int
}

// CourseAdminLinkRepo tautan course ↔ packages (landing) dan course ↔ tryouts.
type CourseAdminLinkRepo interface {
	ListPackagesForCourse(ctx context.Context, courseID string) ([]CourseLinkedPackageRow, error)
	ReplacePackagesForCourse(ctx context.Context, courseID string, packageIDs []string) error
	ListTryoutsForCourse(ctx context.Context, courseID string) ([]CourseLinkedTryoutRow, error)
	ReplaceTryoutsForCourse(ctx context.Context, courseID string, tryoutIDs []string) error
}

type courseAdminLinkRepo struct{ pool *pgxpool.Pool }

func NewCourseAdminLinkRepo(pool *pgxpool.Pool) CourseAdminLinkRepo {
	return &courseAdminLinkRepo{pool: pool}
}

func (r *courseAdminLinkRepo) ListPackagesForCourse(ctx context.Context, courseID string) ([]CourseLinkedPackageRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id::text, p.name, p.slug
		FROM package_courses pc
		JOIN packages p ON p.id = pc.package_id
		WHERE pc.course_id = $1::uuid
		ORDER BY pc.sort_order ASC, p.name ASC
	`, courseID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "42P01" || pgErr.Code == "42703") {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	var out []CourseLinkedPackageRow
	for rows.Next() {
		var row CourseLinkedPackageRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Slug); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *courseAdminLinkRepo) ReplacePackagesForCourse(ctx context.Context, courseID string, packageIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM package_courses WHERE course_id = $1::uuid`, courseID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil
		}
		return err
	}

	seen := map[string]struct{}{}
	for _, raw := range packageIDs {
		pid := strings.TrimSpace(raw)
		if pid == "" {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		_, err := tx.Exec(ctx, `
			INSERT INTO package_courses (package_id, course_id, sort_order)
			SELECT $1::uuid, $2::uuid, COALESCE((SELECT MAX(pc.sort_order) FROM package_courses pc WHERE pc.package_id = $1::uuid), -1) + 1
		`, pid, courseID)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *courseAdminLinkRepo) ListTryoutsForCourse(ctx context.Context, courseID string) ([]CourseLinkedTryoutRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id::text, t.title, t.status::text, t.opens_at, t.closes_at, ct.sort_order
		FROM course_tryouts ct
		JOIN tryout_sessions t ON t.id = ct.tryout_session_id
		WHERE ct.course_id = $1::uuid
		ORDER BY ct.sort_order ASC, t.title ASC
	`, courseID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	var out []CourseLinkedTryoutRow
	for rows.Next() {
		var row CourseLinkedTryoutRow
		var opensAt, closesAt time.Time
		if err := rows.Scan(&row.ID, &row.Title, &row.Status, &opensAt, &closesAt, &row.SortOrder); err != nil {
			return nil, err
		}
		row.OpensAt = opensAt.Format(time.RFC3339)
		row.ClosesAt = closesAt.Format(time.RFC3339)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *courseAdminLinkRepo) ReplaceTryoutsForCourse(ctx context.Context, courseID string, tryoutIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM course_tryouts WHERE course_id = $1::uuid`, courseID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil
		}
		return err
	}

	seen := map[string]struct{}{}
	order := 0
	for _, raw := range tryoutIDs {
		tid := strings.TrimSpace(raw)
		if tid == "" {
			continue
		}
		if _, ok := seen[tid]; ok {
			continue
		}
		seen[tid] = struct{}{}
		_, err := tx.Exec(ctx, `
			INSERT INTO course_tryouts (course_id, tryout_session_id, sort_order) VALUES ($1::uuid, $2::uuid, $3)
		`, courseID, tid, order)
		if err != nil {
			return err
		}
		order++
	}
	return tx.Commit(ctx)
}
