package repo

import (
	"context"
	"encoding/json"
	"errors"

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
		SELECT id, name, slug, short_description, price_display, price_early_bird, price_normal,
		       cta_label, wa_message_template, cta_url, is_open, is_bundle, bundle_subtitle, durasi,
		       COALESCE(materi, '[]'::jsonb), COALESCE(fasilitas, '[]'::jsonb), COALESCE(bonus, '[]'::jsonb)
		FROM packages
		ORDER BY created_at
	`)
	if err != nil {
		// If table does not exist yet (landing_schema.sql not run), return empty list
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	var list []domain.LandingPackage
	for rows.Next() {
		var p domain.LandingPackage
		var shortDesc, priceDisp, priceEarly, priceNorm, waTpl, ctaURL, bundleSub, durasi *string
		var materiJSON, fasilitasJSON, bonusJSON []byte
		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &shortDesc, &priceDisp, &priceEarly, &priceNorm,
			&p.CTALabel, &waTpl, &ctaURL, &p.IsOpen, &p.IsBundle, &bundleSub, &durasi,
			&materiJSON, &fasilitasJSON, &bonusJSON,
		)
		if err != nil {
			return nil, err
		}
		p.ShortDescription = shortDesc
		p.PriceDisplay = priceDisp
		p.PriceEarlyBird = priceEarly
		p.PriceNormal = priceNorm
		p.WAMessageTemplate = waTpl
		p.CTAURL = ctaURL
		p.BundleSubtitle = bundleSub
		p.Durasi = durasi
		p.Materi = parseJSONArray(materiJSON)
		p.Fasilitas = parseJSONArray(fasilitasJSON)
		p.Bonus = parseJSONArray(bonusJSON)
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *landingPackageRepoPg) GetBySlug(ctx context.Context, slug string) (domain.LandingPackage, error) {
	var p domain.LandingPackage
	var shortDesc, priceDisp, priceEarly, priceNorm, waTpl, ctaURL, bundleSub, durasi *string
	var materiJSON, fasilitasJSON, bonusJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, short_description, price_display, price_early_bird, price_normal,
		       cta_label, wa_message_template, cta_url, is_open, is_bundle, bundle_subtitle, durasi,
		       COALESCE(materi, '[]'::jsonb), COALESCE(fasilitas, '[]'::jsonb), COALESCE(bonus, '[]'::jsonb)
		FROM packages
		WHERE slug = $1
	`, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &shortDesc, &priceDisp, &priceEarly, &priceNorm,
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
	p.PriceDisplay = priceDisp
	p.PriceEarlyBird = priceEarly
	p.PriceNormal = priceNorm
	p.WAMessageTemplate = waTpl
	p.CTAURL = ctaURL
	p.BundleSubtitle = bundleSub
	p.Durasi = durasi
	p.Materi = parseJSONArray(materiJSON)
	p.Fasilitas = parseJSONArray(fasilitasJSON)
	p.Bonus = parseJSONArray(bonusJSON)
	return p, nil
}
