package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// ProgramsList returns programs (courses) with pagination. GET /api/v1/programs
func ProgramsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 {
			limit = 12
		}
		search := strings.TrimSpace(r.URL.Query().Get("search"))
		category := strings.TrimSpace(r.URL.Query().Get("category"))

		var list []domain.Course
		var err error
		if category != "" {
			subj, subjErr := deps.SubjectRepo.GetBySlug(r.Context(), category)
			if subjErr == nil {
				list, err = deps.CourseRepo.ListBySubjectID(r.Context(), &subj.ID)
			} else {
				list, err = deps.CourseRepo.List(r.Context())
			}
		} else {
			list, err = deps.CourseRepo.List(r.Context())
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		if search != "" {
			filtered := list[:0]
			for _, c := range list {
				titleMatch := strings.Contains(strings.ToLower(c.Title), strings.ToLower(search))
				descMatch := c.Description != nil && strings.Contains(strings.ToLower(*c.Description), strings.ToLower(search))
				if titleMatch || descMatch {
					filtered = append(filtered, c)
				}
			}
			list = filtered
		}
		total := len(list)
		totalPages := (total + limit - 1) / limit
		if totalPages < 1 {
			totalPages = 1
		}
		start := (page - 1) * limit
		if start > total {
			start = total
		}
		end := start + limit
		if end > total {
			end = total
		}
		pageList := list[start:end]

		data := make([]dto.ProgramListItem, 0, len(pageList))
		for _, c := range pageList {
			slug := ""
			if c.Slug != nil {
				slug = *c.Slug
			}
			shortDesc := ""
			if c.Description != nil {
				shortDesc = *c.Description
				if len(shortDesc) > 160 {
					shortDesc = shortDesc[:157] + "..."
				}
			}
			thumb := ""
			if c.Thumbnail != nil {
				thumb = *c.Thumbnail
			}
			instructor := dto.ProgramInstructor{ID: "", Name: "Instructor", Avatar: ""}
			if c.CreatedBy != nil && *c.CreatedBy != "" {
				if u, uErr := deps.UserRepo.FindByID(r.Context(), *c.CreatedBy); uErr == nil {
					instructor.ID = u.ID
					instructor.Name = u.Name
					if u.AvatarURL != nil {
						instructor.Avatar = *u.AvatarURL
					}
				}
			}
			cat := ""
			if c.SubjectID != nil && *c.SubjectID != "" {
				if subj, subjErr := deps.SubjectRepo.GetByID(r.Context(), *c.SubjectID); subjErr == nil {
					cat = subj.Name
				}
			}
			data = append(data, dto.ProgramListItem{
				ID:               c.ID,
				Slug:             slug,
				Title:            c.Title,
				ShortDescription: shortDesc,
				Thumbnail:        thumb,
				Price:            c.PriceCents / 100,
				PriceDisplay:     formatRupiah(c.PriceCents),
				Instructor:       instructor,
				Category:         cat,
				Level:            "beginner",
				Duration:         "-",
				Rating:           4.9,
				ReviewCount:      0,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.ProgramsListResponse{
			Data:       data,
			Total:      total,
			Page:       page,
			TotalPages: totalPages,
		})
	}
}

// ProgramBySlug returns program detail by slug with modules/lessons. GET /api/v1/programs/:slug
func ProgramBySlug(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "slug required")
			return
		}
		c, err := deps.CourseRepo.GetBySlug(r.Context(), slug)
		if err != nil {
			// Fallback: slug mungkin dari tabel packages (landing)
			if deps.LandingPackageRepo != nil {
				pkg, pkgErr := deps.LandingPackageRepo.GetBySlug(r.Context(), slug)
				if pkgErr == nil {
					shortDesc := ""
					if pkg.ShortDescription != nil {
						shortDesc = *pkg.ShortDescription
					}
					priceDisplay := ""
					if pkg.PriceDisplay != nil {
						priceDisplay = *pkg.PriceDisplay
					} else if pkg.PriceEarlyBird != nil {
						priceDisplay = *pkg.PriceEarlyBird
					}
					dur := "-"
					if pkg.Durasi != nil {
						dur = *pkg.Durasi
					}
					modules := make([]dto.ProgramModule, 0)
					if len(pkg.Materi) > 0 {
						lessons := make([]dto.ProgramLesson, 0, len(pkg.Materi))
						for _, m := range pkg.Materi {
							lessons = append(lessons, dto.ProgramLesson{ID: "", Title: m, Duration: ""})
						}
						modules = append(modules, dto.ProgramModule{ID: "materi", Title: "Yang akan kamu kuasai", Lessons: lessons})
					}
					if len(pkg.Fasilitas) > 0 {
						lessons := make([]dto.ProgramLesson, 0, len(pkg.Fasilitas))
						for _, f := range pkg.Fasilitas {
							lessons = append(lessons, dto.ProgramLesson{ID: "", Title: f, Duration: ""})
						}
						modules = append(modules, dto.ProgramModule{ID: "fasilitas", Title: "Fasilitas", Lessons: lessons})
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(dto.ProgramDetailResponse{
						ID:               pkg.ID,
						Slug:             pkg.Slug,
						Title:            pkg.Name,
						ShortDescription: shortDesc,
						Description:      shortDesc,
						Thumbnail:        "",
						Price:            0,
						PriceDisplay:     priceDisplay,
						Instructor:       dto.ProgramInstructor{},
						Category:         "",
						Level:            "beginner",
						Duration:         dur,
						Rating:           4.9,
						ReviewCount:      0,
						Modules:          modules,
						Reviews:          []dto.ProgramReview{},
					})
					return
				}
			}
			writeError(w, http.StatusNotFound, "not_found", "program not found")
			return
		}
		shortDesc := ""
		desc := ""
		if c.Description != nil {
			desc = *c.Description
			shortDesc = desc
			if len(shortDesc) > 160 {
				shortDesc = shortDesc[:157] + "..."
			}
		}
		slugStr := ""
		if c.Slug != nil {
			slugStr = *c.Slug
		}
		thumb := ""
		if c.Thumbnail != nil {
			thumb = *c.Thumbnail
		}
		instructor := dto.ProgramInstructor{ID: "", Name: "Instructor", Avatar: ""}
		if c.CreatedBy != nil && *c.CreatedBy != "" {
			if u, uErr := deps.UserRepo.FindByID(r.Context(), *c.CreatedBy); uErr == nil {
				instructor.ID = u.ID
				instructor.Name = u.Name
				if u.AvatarURL != nil {
					instructor.Avatar = *u.AvatarURL
				}
			}
		}
		cat := ""
		if c.SubjectID != nil && *c.SubjectID != "" {
			if subj, subjErr := deps.SubjectRepo.GetByID(r.Context(), *c.SubjectID); subjErr == nil {
				cat = subj.Name
			}
		}
		contents, _ := deps.CourseContentRepo.ListByCourseID(r.Context(), c.ID)
		modules := make([]dto.ProgramModule, 0, len(contents))
		for _, cc := range contents {
			lessons := []dto.ProgramLesson{{ID: cc.ID, Title: cc.Title, Duration: "-"}}
			modules = append(modules, dto.ProgramModule{
				ID:      cc.ID,
				Title:   cc.Title,
				Lessons: lessons,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.ProgramDetailResponse{
			ID:               c.ID,
			Slug:             slugStr,
			Title:            c.Title,
			ShortDescription: shortDesc,
			Description:      desc,
			Thumbnail:        thumb,
			Price:            c.PriceCents / 100,
			PriceDisplay:     formatRupiah(c.PriceCents),
			Instructor:       instructor,
			Category:         cat,
			Level:            "beginner",
			Duration:         "-",
			Rating:           4.9,
			ReviewCount:      0,
			Modules:          modules,
			Reviews:          []dto.ProgramReview{},
		})
	}
}
