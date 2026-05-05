package repo

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type landingPackageRepoPg struct {
	pool *pgxpool.Pool
}

// NewLandingPackageRepoPg returns a repo that reads from the packages table (database/landing_schema.sql).
func NewLandingPackageRepoPg(pool *pgxpool.Pool) LandingPackageRepo {
	return &landingPackageRepoPg{pool: pool}
}

func parseJSONArray(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	var out []string
	_ = json.Unmarshal(b, &out)
	return out
}

func (r *landingPackageRepoPg) List(ctx context.Context) ([]domain.LandingPackage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, short_description, price_early_bird, price_normal,
		       cta_label, wa_message_template, cta_url, is_open, is_bundle, bundle_subtitle, durasi,
		       COALESCE(materi, '[]'::jsonb), COALESCE(fasilitas, '[]'::jsonb), COALESCE(bonus, '[]'::jsonb)
		FROM packages
		ORDER BY created_at
	`)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, nil
		}
		if errors.As(err, &pgErr) && pgErr.Code == "42703" {
			return r.listWithoutNumericPrice(ctx)
		}
		return nil, err
	}
	defer rows.Close()
	var list []domain.LandingPackage
	for rows.Next() {
		var p domain.LandingPackage
		var shortDesc, waTpl, ctaURL, bundleSub, durasi *string
		var earlyBirdVal, normalVal *int64
		var materiJSON, fasilitasJSON, bonusJSON []byte
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &shortDesc, &earlyBirdVal, &normalVal,
			&p.CTALabel, &waTpl, &ctaURL, &p.IsOpen, &p.IsBundle, &bundleSub, &durasi,
			&materiJSON, &fasilitasJSON, &bonusJSON,
		); err != nil {
			return nil, err
		}
		p.ShortDescription = shortDesc
		p.PriceEarlyBird = earlyBirdVal
		p.PriceNormal = normalVal
		p.WAMessageTemplate = waTpl
		p.CTAURL = ctaURL
		p.BundleSubtitle = bundleSub
		p.Durasi = durasi
		p.Materi = parseJSONArray(materiJSON)
		p.Fasilitas = parseJSONArray(fasilitasJSON)
		p.Bonus = parseJSONArray(bonusJSON)
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.attachLinkedCourses(ctx, list)
}

func (r *landingPackageRepoPg) listWithoutNumericPrice(ctx context.Context) ([]domain.LandingPackage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, short_description, cta_label, wa_message_template, cta_url, is_open, is_bundle, bundle_subtitle, durasi,
		       COALESCE(materi, '[]'::jsonb), COALESCE(fasilitas, '[]'::jsonb), COALESCE(bonus, '[]'::jsonb)
		FROM packages
		ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.LandingPackage
	for rows.Next() {
		var p domain.LandingPackage
		var shortDesc, waTpl, ctaURL, bundleSub, durasi *string
		var materiJSON, fasilitasJSON, bonusJSON []byte
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &shortDesc,
			&p.CTALabel, &waTpl, &ctaURL, &p.IsOpen, &p.IsBundle, &bundleSub, &durasi,
			&materiJSON, &fasilitasJSON, &bonusJSON,
		); err != nil {
			return nil, err
		}
		p.ShortDescription = shortDesc
		p.WAMessageTemplate = waTpl
		p.CTAURL = ctaURL
		p.BundleSubtitle = bundleSub
		p.Durasi = durasi
		p.Materi = parseJSONArray(materiJSON)
		p.Fasilitas = parseJSONArray(fasilitasJSON)
		p.Bonus = parseJSONArray(bonusJSON)
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.attachLinkedCourses(ctx, list)
}

func (r *landingPackageRepoPg) GetBySlug(ctx context.Context, slug string) (domain.LandingPackage, error) {
	var p domain.LandingPackage
	var shortDesc, waTpl, ctaURL, bundleSub, durasi *string
	var earlyBirdVal, normalVal *int64
	var materiJSON, fasilitasJSON, bonusJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, short_description, price_early_bird, price_normal,
		       cta_label, wa_message_template, cta_url, is_open, is_bundle, bundle_subtitle, durasi,
		       COALESCE(materi, '[]'::jsonb), COALESCE(fasilitas, '[]'::jsonb), COALESCE(bonus, '[]'::jsonb)
		FROM packages
		WHERE slug = $1
	`, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &shortDesc, &earlyBirdVal, &normalVal,
		&p.CTALabel, &waTpl, &ctaURL, &p.IsOpen, &p.IsBundle, &bundleSub, &durasi,
		&materiJSON, &fasilitasJSON, &bonusJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LandingPackage{}, ErrPackageNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42703" {
			return r.getBySlugWithoutNumericPrice(ctx, slug)
		}
		return domain.LandingPackage{}, err
	}
	p.ShortDescription = shortDesc
	p.PriceEarlyBird = earlyBirdVal
	p.PriceNormal = normalVal
	p.WAMessageTemplate = waTpl
	p.CTAURL = ctaURL
	p.BundleSubtitle = bundleSub
	p.Durasi = durasi
	p.Materi = parseJSONArray(materiJSON)
	p.Fasilitas = parseJSONArray(fasilitasJSON)
	p.Bonus = parseJSONArray(bonusJSON)
	p.LinkedCourses, _ = r.ListLinkedCourses(ctx, p.ID)
	return p, nil
}

func (r *landingPackageRepoPg) getBySlugWithoutNumericPrice(ctx context.Context, slug string) (domain.LandingPackage, error) {
	var p domain.LandingPackage
	var shortDesc, waTpl, ctaURL, bundleSub, durasi *string
	var materiJSON, fasilitasJSON, bonusJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, short_description, cta_label, wa_message_template, cta_url, is_open, is_bundle, bundle_subtitle, durasi,
		       COALESCE(materi, '[]'::jsonb), COALESCE(fasilitas, '[]'::jsonb), COALESCE(bonus, '[]'::jsonb)
		FROM packages
		WHERE slug = $1
	`, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &shortDesc,
		&p.CTALabel, &waTpl, &ctaURL, &p.IsOpen, &p.IsBundle, &bundleSub, &durasi,
		&materiJSON, &fasilitasJSON, &bonusJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LandingPackage{}, ErrPackageNotFound
		}
		return domain.LandingPackage{}, err
	}
	p.ShortDescription = shortDesc
	p.WAMessageTemplate = waTpl
	p.CTAURL = ctaURL
	p.BundleSubtitle = bundleSub
	p.Durasi = durasi
	p.Materi = parseJSONArray(materiJSON)
	p.Fasilitas = parseJSONArray(fasilitasJSON)
	p.Bonus = parseJSONArray(bonusJSON)
	p.LinkedCourses, _ = r.ListLinkedCourses(ctx, p.ID)
	return p, nil
}

func (r *landingPackageRepoPg) attachLinkedCourses(ctx context.Context, list []domain.LandingPackage) ([]domain.LandingPackage, error) {
	for i := range list {
		lc, err := r.ListLinkedCourses(ctx, list[i].ID)
		if err != nil {
			return nil, err
		}
		list[i].LinkedCourses = lc
	}
	return list, nil
}

func (r *landingPackageRepoPg) GetByID(ctx context.Context, id string) (domain.LandingPackage, error) {
	var p domain.LandingPackage
	var shortDesc, waTpl, ctaURL, bundleSub, durasi *string
	var earlyBirdVal, normalVal *int64
	var materiJSON, fasilitasJSON, bonusJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, short_description, price_early_bird, price_normal,
		       cta_label, wa_message_template, cta_url, is_open, is_bundle, bundle_subtitle, durasi,
		       COALESCE(materi, '[]'::jsonb), COALESCE(fasilitas, '[]'::jsonb), COALESCE(bonus, '[]'::jsonb)
		FROM packages
		WHERE id = $1::uuid
	`, id).Scan(
		&p.ID, &p.Name, &p.Slug, &shortDesc, &earlyBirdVal, &normalVal,
		&p.CTALabel, &waTpl, &ctaURL, &p.IsOpen, &p.IsBundle, &bundleSub, &durasi,
		&materiJSON, &fasilitasJSON, &bonusJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LandingPackage{}, ErrPackageNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42703" {
			return r.getByIDWithoutNumericPrice(ctx, id)
		}
		return domain.LandingPackage{}, err
	}
	p.ShortDescription = shortDesc
	p.PriceEarlyBird = earlyBirdVal
	p.PriceNormal = normalVal
	p.WAMessageTemplate = waTpl
	p.CTAURL = ctaURL
	p.BundleSubtitle = bundleSub
	p.Durasi = durasi
	p.Materi = parseJSONArray(materiJSON)
	p.Fasilitas = parseJSONArray(fasilitasJSON)
	p.Bonus = parseJSONArray(bonusJSON)
	p.LinkedCourses, _ = r.ListLinkedCourses(ctx, p.ID)
	return p, nil
}

func (r *landingPackageRepoPg) getByIDWithoutNumericPrice(ctx context.Context, id string) (domain.LandingPackage, error) {
	var p domain.LandingPackage
	var shortDesc, waTpl, ctaURL, bundleSub, durasi *string
	var materiJSON, fasilitasJSON, bonusJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, short_description, cta_label, wa_message_template, cta_url, is_open, is_bundle, bundle_subtitle, durasi,
		       COALESCE(materi, '[]'::jsonb), COALESCE(fasilitas, '[]'::jsonb), COALESCE(bonus, '[]'::jsonb)
		FROM packages
		WHERE id = $1::uuid
	`, id).Scan(
		&p.ID, &p.Name, &p.Slug, &shortDesc,
		&p.CTALabel, &waTpl, &ctaURL, &p.IsOpen, &p.IsBundle, &bundleSub, &durasi,
		&materiJSON, &fasilitasJSON, &bonusJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LandingPackage{}, ErrPackageNotFound
		}
		return domain.LandingPackage{}, err
	}
	p.ShortDescription = shortDesc
	p.WAMessageTemplate = waTpl
	p.CTAURL = ctaURL
	p.BundleSubtitle = bundleSub
	p.Durasi = durasi
	p.Materi = parseJSONArray(materiJSON)
	p.Fasilitas = parseJSONArray(fasilitasJSON)
	p.Bonus = parseJSONArray(bonusJSON)
	p.LinkedCourses, _ = r.ListLinkedCourses(ctx, p.ID)
	return p, nil
}

func (r *landingPackageRepoPg) ListLinkedCourses(ctx context.Context, packageID string) ([]domain.PackageLinkedCourse, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id::text, c.title, COALESCE(c.slug, '')
		FROM package_courses pc
		JOIN courses c ON c.id = pc.course_id
		WHERE pc.package_id = $1::uuid
		ORDER BY pc.sort_order ASC, c.title ASC
	`, packageID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	var out []domain.PackageLinkedCourse
	for rows.Next() {
		var row domain.PackageLinkedCourse
		if err := rows.Scan(&row.ID, &row.Title, &row.Slug); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *landingPackageRepoPg) ReplaceLinkedCourses(ctx context.Context, packageID string, courseIDs []string) error {
	if packageID == "" {
		return nil
	}
	// Kosong = jangan hapus baris package_courses (hindari hilangnya tautan kelas dari PUT parsial / bug client).
	if len(courseIDs) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM package_courses WHERE package_id = $1::uuid`, packageID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil
		}
		return err
	}
	for i, cid := range courseIDs {
		cid = strings.TrimSpace(cid)
		if cid == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO package_courses (package_id, course_id, sort_order) VALUES ($1::uuid, $2::uuid, $3)
		`, packageID, cid, i); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
