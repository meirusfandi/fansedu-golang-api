package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/meirusfandi/fansedu-golang-api/internal/cache"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type landingResourceDef struct {
	Table      string
	Columns    []string
	InsertCols []string
	Required   map[string]struct{}
	OrderBy    string
}

var landingResourceDefs = map[string]landingResourceDef{
	"hero-badges": {
		Table:      "hero_badges",
		Columns:    []string{"id", "label", "sort_order"},
		InsertCols: []string{"label", "sort_order"},
		Required:   map[string]struct{}{"label": {}},
		OrderBy:    "sort_order ASC",
	},
	"social-proof-stats": {
		Table:      "social_proof_stats",
		Columns:    []string{"id", "value", "label", "sort_order"},
		InsertCols: []string{"value", "label", "sort_order"},
		Required:   map[string]struct{}{"value": {}, "label": {}},
		OrderBy:    "sort_order ASC",
	},
	"masalah-items": {
		Table:      "masalah_items",
		Columns:    []string{"id", "title", "description", "sort_order"},
		InsertCols: []string{"title", "description", "sort_order"},
		Required:   map[string]struct{}{"title": {}, "description": {}},
		OrderBy:    "sort_order ASC",
	},
	"solusi-highlights": {
		Table:      "solusi_highlights",
		Columns:    []string{"id", "label", "sort_order"},
		InsertCols: []string{"label", "sort_order"},
		Required:   map[string]struct{}{"label": {}},
		OrderBy:    "sort_order ASC",
	},
	"features": {
		Table:      "features",
		Columns:    []string{"id", "title", "description", "sort_order"},
		InsertCols: []string{"title", "description", "sort_order"},
		Required:   map[string]struct{}{"title": {}, "description": {}},
		OrderBy:    "sort_order ASC",
	},
	"testimonials": {
		Table:      "testimonials",
		Columns:    []string{"id", "quote", "author", "role", "is_active", "sort_order", "created_at"},
		InsertCols: []string{"quote", "author", "role", "is_active", "sort_order"},
		Required:   map[string]struct{}{"quote": {}, "author": {}, "role": {}},
		OrderBy:    "sort_order ASC",
	},
	"tryout-steps": {
		Table:      "tryout_steps",
		Columns:    []string{"id", "step_number", "title", "description", "sort_order"},
		InsertCols: []string{"step_number", "title", "description", "sort_order"},
		Required:   map[string]struct{}{"step_number": {}, "title": {}, "description": {}},
		OrderBy:    "sort_order ASC",
	},
	"contact-links": {
		Table:      "contact_links",
		Columns:    []string{"id", "title", "subtitle", "href", "sort_order"},
		InsertCols: []string{"title", "subtitle", "href", "sort_order"},
		Required:   map[string]struct{}{"title": {}, "href": {}},
		OrderBy:    "sort_order ASC",
	},
	"nav-items": {
		Table:      "nav_items",
		Columns:    []string{"id", "label", "anchor", "sort_order", "is_visible"},
		InsertCols: []string{"label", "anchor", "sort_order", "is_visible"},
		Required:   map[string]struct{}{"label": {}, "anchor": {}},
		OrderBy:    "sort_order ASC",
	},
	"about-highlights": {
		Table:      "about_highlights",
		Columns:    []string{"id", "label", "sort_order"},
		InsertCols: []string{"label", "sort_order"},
		Required:   map[string]struct{}{"label": {}},
		OrderBy:    "sort_order ASC",
	},
	"services": {
		Table:      "services",
		Columns:    []string{"id", "title", "description", "sort_order"},
		InsertCols: []string{"title", "description", "sort_order"},
		Required:   map[string]struct{}{"title": {}, "description": {}},
		OrderBy:    "sort_order ASC",
	},
}

func AdminLandingSiteSettingsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var raw json.RawMessage
		err := deps.DB.QueryRow(r.Context(), `SELECT COALESCE(json_object_agg(key, value), '{}'::json) FROM site_settings`).Scan(&raw)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	}
}

func AdminLandingSiteSettingsUpsert(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		for k, v := range payload {
			key := strings.TrimSpace(k)
			if key == "" {
				continue
			}
			val := fmt.Sprint(v)
			_, err := deps.DB.Exec(r.Context(), `
				INSERT INTO site_settings (key, value)
				VALUES ($1, $2)
				ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
			`, key, val)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error", err.Error())
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminLandingResourceList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		def, ok := getLandingDef(chi.URLParam(r, "resource"))
		if !ok {
			writeError(w, http.StatusNotFound, "not_found", "resource not found")
			return
		}
		query := fmt.Sprintf(
			`SELECT COALESCE(json_agg(t), '[]'::json) FROM (SELECT %s FROM %s ORDER BY %s) t`,
			strings.Join(def.Columns, ", "),
			def.Table,
			def.OrderBy,
		)
		var raw json.RawMessage
		if err := deps.DB.QueryRow(r.Context(), query).Scan(&raw); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	}
}

func AdminLandingResourceCreate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		def, ok := getLandingDef(chi.URLParam(r, "resource"))
		if !ok {
			writeError(w, http.StatusNotFound, "not_found", "resource not found")
			return
		}
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		for req := range def.Required {
			if _, ok := payload[req]; !ok {
				writeError(w, http.StatusBadRequest, "validation_error", req+" is required")
				return
			}
		}
		cols := []string{"id"}
		vals := []interface{}{uuid.NewString()}
		args := []string{"$1"}
		for i, col := range def.InsertCols {
			cols = append(cols, col)
			vals = append(vals, normalizeLandingValue(col, payload[col]))
			args = append(args, fmt.Sprintf("$%d", i+2))
		}
		query := fmt.Sprintf(
			`INSERT INTO %s (%s) VALUES (%s)`,
			def.Table,
			strings.Join(cols, ", "),
			strings.Join(args, ", "),
		)
		if _, err := deps.DB.Exec(r.Context(), query, vals...); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func AdminLandingResourceUpdate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		def, ok := getLandingDef(chi.URLParam(r, "resource"))
		if !ok {
			writeError(w, http.StatusNotFound, "not_found", "resource not found")
			return
		}
		id := strings.TrimSpace(chi.URLParam(r, "id"))
		if id == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "id is required")
			return
		}
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		sets := make([]string, 0, len(def.InsertCols))
		values := make([]interface{}, 0, len(def.InsertCols)+1)
		arg := 1
		for _, col := range def.InsertCols {
			v, ok := payload[col]
			if !ok {
				continue
			}
			sets = append(sets, fmt.Sprintf("%s = $%d", col, arg))
			values = append(values, normalizeLandingValue(col, v))
			arg++
		}
		if len(sets) == 0 {
			writeError(w, http.StatusBadRequest, "validation_error", "no updatable fields provided")
			return
		}
		values = append(values, id)
		query := fmt.Sprintf(`UPDATE %s SET %s WHERE id = $%d`, def.Table, strings.Join(sets, ", "), arg)
		ct, err := deps.DB.Exec(r.Context(), query, values...)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		if ct.RowsAffected() == 0 {
			writeError(w, http.StatusNotFound, "not_found", "resource item not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminLandingResourceDelete(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		def, ok := getLandingDef(chi.URLParam(r, "resource"))
		if !ok {
			writeError(w, http.StatusNotFound, "not_found", "resource not found")
			return
		}
		id := strings.TrimSpace(chi.URLParam(r, "id"))
		if id == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "id is required")
			return
		}
		query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, def.Table)
		ct, err := deps.DB.Exec(r.Context(), query, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		if ct.RowsAffected() == 0 {
			writeError(w, http.StatusNotFound, "not_found", "resource item not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

type adminLandingPackageRequest struct {
	Name              string   `json:"name"`
	Slug              string   `json:"slug"`
	ShortDescription  *string  `json:"short_description"`
	PriceEarlyBird    *int64   `json:"price_early_bird"`
	PriceNormal       *int64   `json:"price_normal"`
	CTALabel          *string  `json:"cta_label"`
	WAMessageTemplate *string  `json:"wa_message_template"`
	CTAURL            *string  `json:"cta_url"`
	IsOpen            *bool    `json:"is_open"`
	IsBundle          *bool    `json:"is_bundle"`
	BundleSubtitle    *string  `json:"bundle_subtitle"`
	Durasi            *string  `json:"durasi"`
	Materi            []string `json:"materi"`
	Fasilitas         []string `json:"fasilitas"`
	Bonus             []string `json:"bonus"`
	// LinkedCourseIDs urutan = sort_order; omit pada update jika tidak mengubah relasi kelas.
	LinkedCourseIDs *[]string `json:"linked_course_ids,omitempty"`
}

func AdminLandingPackagesList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.LandingPackageRepo.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		if list == nil {
			list = []domain.LandingPackage{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}
}

func AdminLandingPackageCreate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req adminLandingPackageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		req.Slug = strings.TrimSpace(req.Slug)
		if req.Name == "" || req.Slug == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "name and slug are required")
			return
		}
		ctaLabel := "Daftar"
		if req.CTALabel != nil && strings.TrimSpace(*req.CTALabel) != "" {
			ctaLabel = strings.TrimSpace(*req.CTALabel)
		}
		isOpen := true
		if req.IsOpen != nil {
			isOpen = *req.IsOpen
		}
		isBundle := false
		if req.IsBundle != nil {
			isBundle = *req.IsBundle
		}
		materiJSON, _ := json.Marshal(req.Materi)
		fasilitasJSON, _ := json.Marshal(req.Fasilitas)
		bonusJSON, _ := json.Marshal(req.Bonus)
		var newID string
		err := deps.DB.QueryRow(r.Context(), `
			INSERT INTO packages (
				name, slug, short_description, price_early_bird, price_normal,
				cta_label, wa_message_template, cta_url, is_open, is_bundle, bundle_subtitle, durasi,
				materi, fasilitas, bonus, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9, $10, $11, $12,
				$13::jsonb, $14::jsonb, $15::jsonb, NOW()
			) RETURNING id::text
		`, req.Name, req.Slug, req.ShortDescription, req.PriceEarlyBird, req.PriceNormal, ctaLabel, req.WAMessageTemplate, req.CTAURL, isOpen, isBundle, req.BundleSubtitle, req.Durasi, materiJSON, fasilitasJSON, bonusJSON).Scan(&newID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		if deps.LandingPackageRepo != nil {
			ids := []string{}
			if req.LinkedCourseIDs != nil {
				ids = *req.LinkedCourseIDs
			}
			_ = deps.LandingPackageRepo.ReplaceLinkedCourses(r.Context(), newID, ids)
		}
		cache.InvalidatePackagesList(r.Context(), deps.Redis)
		w.WriteHeader(http.StatusCreated)
	}
}

func AdminLandingPackageUpdate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(chi.URLParam(r, "id"))
		if id == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "id is required")
			return
		}
		var req adminLandingPackageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		sets := []string{}
		values := []interface{}{}
		arg := 1
		if strings.TrimSpace(req.Name) != "" {
			sets = append(sets, fmt.Sprintf("name = $%d", arg))
			values = append(values, strings.TrimSpace(req.Name))
			arg++
		}
		if strings.TrimSpace(req.Slug) != "" {
			sets = append(sets, fmt.Sprintf("slug = $%d", arg))
			values = append(values, strings.TrimSpace(req.Slug))
			arg++
		}
		if req.ShortDescription != nil {
			sets = append(sets, fmt.Sprintf("short_description = $%d", arg))
			values = append(values, req.ShortDescription)
			arg++
		}
		if req.PriceEarlyBird != nil {
			sets = append(sets, fmt.Sprintf("price_early_bird = $%d", arg))
			values = append(values, req.PriceEarlyBird)
			arg++
		}
		if req.PriceNormal != nil {
			sets = append(sets, fmt.Sprintf("price_normal = $%d", arg))
			values = append(values, req.PriceNormal)
			arg++
		}
		if req.CTALabel != nil {
			sets = append(sets, fmt.Sprintf("cta_label = $%d", arg))
			values = append(values, req.CTALabel)
			arg++
		}
		if req.WAMessageTemplate != nil {
			sets = append(sets, fmt.Sprintf("wa_message_template = $%d", arg))
			values = append(values, req.WAMessageTemplate)
			arg++
		}
		if req.CTAURL != nil {
			sets = append(sets, fmt.Sprintf("cta_url = $%d", arg))
			values = append(values, req.CTAURL)
			arg++
		}
		if req.IsOpen != nil {
			sets = append(sets, fmt.Sprintf("is_open = $%d", arg))
			values = append(values, req.IsOpen)
			arg++
		}
		if req.IsBundle != nil {
			sets = append(sets, fmt.Sprintf("is_bundle = $%d", arg))
			values = append(values, req.IsBundle)
			arg++
		}
		if req.BundleSubtitle != nil {
			sets = append(sets, fmt.Sprintf("bundle_subtitle = $%d", arg))
			values = append(values, req.BundleSubtitle)
			arg++
		}
		if req.Durasi != nil {
			sets = append(sets, fmt.Sprintf("durasi = $%d", arg))
			values = append(values, req.Durasi)
			arg++
		}
		if req.Materi != nil {
			raw, _ := json.Marshal(req.Materi)
			sets = append(sets, fmt.Sprintf("materi = $%d::jsonb", arg))
			values = append(values, raw)
			arg++
		}
		if req.Fasilitas != nil {
			raw, _ := json.Marshal(req.Fasilitas)
			sets = append(sets, fmt.Sprintf("fasilitas = $%d::jsonb", arg))
			values = append(values, raw)
			arg++
		}
		if req.Bonus != nil {
			raw, _ := json.Marshal(req.Bonus)
			sets = append(sets, fmt.Sprintf("bonus = $%d::jsonb", arg))
			values = append(values, raw)
			arg++
		}
		if len(sets) == 0 {
			writeError(w, http.StatusBadRequest, "validation_error", "no updatable fields provided")
			return
		}
		sets = append(sets, "updated_at = NOW()")
		values = append(values, id)
		query := fmt.Sprintf("UPDATE packages SET %s WHERE id = $%d::uuid", strings.Join(sets, ", "), arg)
		ct, err := deps.DB.Exec(r.Context(), query, values...)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		if ct.RowsAffected() == 0 {
			writeError(w, http.StatusNotFound, "not_found", "package not found")
			return
		}
		if deps.LandingPackageRepo != nil && req.LinkedCourseIDs != nil {
			_ = deps.LandingPackageRepo.ReplaceLinkedCourses(r.Context(), id, *req.LinkedCourseIDs)
		}
		cache.InvalidatePackagesList(r.Context(), deps.Redis)
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminLandingPackageDelete(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(chi.URLParam(r, "id"))
		if id == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "id is required")
			return
		}
		ct, err := deps.DB.Exec(r.Context(), `DELETE FROM packages WHERE id = $1::uuid`, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		if ct.RowsAffected() == 0 {
			writeError(w, http.StatusNotFound, "not_found", "package not found")
			return
		}
		cache.InvalidatePackagesList(r.Context(), deps.Redis)
		w.WriteHeader(http.StatusNoContent)
	}
}

func getLandingDef(resource string) (landingResourceDef, bool) {
	def, ok := landingResourceDefs[strings.TrimSpace(resource)]
	return def, ok
}

func normalizeLandingValue(col string, v interface{}) interface{} {
	switch col {
	case "sort_order", "step_number":
		switch n := v.(type) {
		case float64:
			return int(n)
		case float32:
			return int(n)
		case int:
			return n
		case int64:
			return int(n)
		case string:
			i, err := strconv.Atoi(strings.TrimSpace(n))
			if err == nil {
				return i
			}
		}
	case "is_active", "is_visible":
		switch b := v.(type) {
		case bool:
			return b
		case string:
			s := strings.TrimSpace(strings.ToLower(b))
			return s == "1" || s == "true" || s == "yes"
		}
	}
	return v
}
