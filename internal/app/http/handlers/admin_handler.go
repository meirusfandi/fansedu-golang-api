package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

func AdminOverview(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overview, err := deps.AdminService.Overview(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.AdminOverviewResponse{
			TotalStudents:     overview.TotalStudents,
			TotalUsers:        overview.TotalUsers,
			ActiveTryouts:     overview.ActiveTryouts,
			TotalCourses:      overview.TotalCourses,
			TotalEnrollments:  overview.TotalEnrollments,
			AvgScore:          overview.AvgScore,
			TotalCertificates: overview.TotalCertificates,
		})
	}
}

func AdminListUsers(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := r.URL.Query().Get("role")
		list, err := deps.AdminService.ListUsers(r.Context(), role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.UserListResponse, len(list))
		for i := range list {
			out[i] = dto.UserListResponse{
				ID:        list[i].ID,
				Email:     list[i].Email,
				Name:      list[i].Name,
				Role:      list[i].Role,
				AvatarURL: list[i].AvatarURL,
				SchoolID:  list[i].SchoolID,
				SubjectID: list[i].SubjectID,
				CreatedAt: list[i].CreatedAt.Format(time.RFC3339),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminGetUser(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "userId")
		u, err := deps.AdminService.GetUser(r.Context(), id)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		resp := dto.UserDetailResponse{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Role:      u.Role,
			AvatarURL: u.AvatarURL,
			SchoolID:  u.SchoolID,
			SubjectID: u.SubjectID,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		}
		if u.SchoolID != nil && *u.SchoolID != "" {
			if school, err := deps.SchoolRepo.GetByID(r.Context(), *u.SchoolID); err == nil {
				resp.School = &dto.SchoolResponse{
					ID: school.ID, Name: school.Name, Slug: school.Slug,
					Description: school.Description, Address: school.Address, LogoURL: school.LogoURL,
					CreatedAt: school.CreatedAt.Format(time.RFC3339), UpdatedAt: school.UpdatedAt.Format(time.RFC3339),
				}
			}
		}
		if u.SubjectID != nil && *u.SubjectID != "" {
			if subj, err := deps.SubjectRepo.GetByID(r.Context(), *u.SubjectID); err == nil {
				resp.Subject = &dto.SubjectResponse{
					ID: subj.ID, Name: subj.Name, Slug: subj.Slug,
					Description: subj.Description, IconURL: subj.IconURL, SortOrder: subj.SortOrder,
					CreatedAt: subj.CreatedAt.Format(time.RFC3339), UpdatedAt: subj.UpdatedAt.Format(time.RFC3339),
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func AdminCreateUser(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.UserCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Password == "" || req.Name == "" {
			http.Error(w, "email, password, name required", http.StatusBadRequest)
			return
		}
		if deps.RoleRepo == nil {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		var roleCode string
		slug := strings.TrimSpace(req.Role)
		if slug == "" {
			var err error
			roleCode, err = defaultUserRoleCode(r.Context(), deps.RoleRepo)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			var err error
			roleCode, err = resolveUserRoleCodeForUserTable(r.Context(), deps.RoleRepo, slug)
			if err != nil {
				if errors.Is(err, errUnknownRoleSlug) {
					http.Error(w, "invalid role: must match table roles (slug or user_role_code)", http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		u := domain.User{Email: req.Email, Name: req.Name, Role: roleCode, AvatarURL: req.AvatarURL, SchoolID: req.SchoolID, SubjectID: req.SubjectID}
		created, err := deps.AdminService.CreateUser(r.Context(), u, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := dto.UserDetailResponse{
			ID:        created.ID,
			Email:     created.Email,
			Name:      created.Name,
			Role:      created.Role,
			AvatarURL: created.AvatarURL,
			SchoolID:  created.SchoolID,
			SubjectID: created.SubjectID,
			CreatedAt: created.CreatedAt.Format(time.RFC3339),
			UpdatedAt: created.UpdatedAt.Format(time.RFC3339),
		}
		if created.SchoolID != nil && *created.SchoolID != "" {
			if school, err := deps.SchoolRepo.GetByID(r.Context(), *created.SchoolID); err == nil {
				resp.School = &dto.SchoolResponse{
					ID: school.ID, Name: school.Name, Slug: school.Slug,
					Description: school.Description, Address: school.Address, LogoURL: school.LogoURL,
					CreatedAt: school.CreatedAt.Format(time.RFC3339), UpdatedAt: school.UpdatedAt.Format(time.RFC3339),
				}
			}
		}
		if created.SubjectID != nil && *created.SubjectID != "" {
			if subj, err := deps.SubjectRepo.GetByID(r.Context(), *created.SubjectID); err == nil {
				resp.Subject = &dto.SubjectResponse{
					ID: subj.ID, Name: subj.Name, Slug: subj.Slug,
					Description: subj.Description, IconURL: subj.IconURL, SortOrder: subj.SortOrder,
					CreatedAt: subj.CreatedAt.Format(time.RFC3339), UpdatedAt: subj.UpdatedAt.Format(time.RFC3339),
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func AdminUpdateUser(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "userId")
		u, err := deps.AdminService.GetUser(r.Context(), id)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		var req dto.UserUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name != nil {
			u.Name = *req.Name
		}
		if req.Email != nil {
			u.Email = *req.Email
		}
		if req.Role != nil {
			if deps.RoleRepo == nil {
				http.Error(w, "service unavailable", http.StatusServiceUnavailable)
				return
			}
			roleCode, err := resolveUserRoleCodeForUserTable(r.Context(), deps.RoleRepo, *req.Role)
			if err != nil {
				if errors.Is(err, errUnknownRoleSlug) {
					http.Error(w, "invalid role: must match table roles (slug or user_role_code)", http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			u.Role = roleCode
		}
		if req.AvatarURL != nil {
			u.AvatarURL = req.AvatarURL
		}
		if req.SchoolID != nil {
			if *req.SchoolID == "" {
				u.SchoolID = nil
			} else {
				u.SchoolID = req.SchoolID
			}
		}
		if req.SubjectID != nil {
			if *req.SubjectID == "" {
				u.SubjectID = nil
			} else {
				u.SubjectID = req.SubjectID
			}
		}
		if err := deps.AdminService.UpdateUser(r.Context(), u, req.Password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func AdminListTryouts(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.AdminService.ListTryouts(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.TryoutResponse, len(list))
		for i := range list {
			out[i] = tryoutToDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// AdminGetTryout returns one tryout (admin). GET /api/v1/admin/tryouts/{tryoutId}
func AdminGetTryout(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "tryoutId")
		t, err := deps.TryoutService.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "tryout not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tryoutToDTO(t))
	}
}

func AdminCreateTryout(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.TryoutCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.SubjectID != nil && strings.TrimSpace(*req.SubjectID) == "" {
			req.SubjectID = nil
		}
		t := domain.TryoutSession{
			Title:           req.Title,
			ShortTitle:      req.ShortTitle,
			Description:     req.Description,
			DurationMinutes: req.DurationMinutes,
			QuestionsCount:  req.QuestionsCount,
			Level:           req.Level,
			SubjectID:       req.SubjectID,
			OpensAt:         req.OpensAt,
			ClosesAt:        req.ClosesAt,
			MaxParticipants: req.MaxParticipants,
			Status:          req.Status,
			GradingMode:     normalizeTryoutGradingMode(req.GradingMode),
		}
		if err := validateTryoutForCreate(&t); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		created, err := deps.AdminService.CreateTryout(r.Context(), t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		created, err = deps.TryoutService.GetByID(r.Context(), created.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(tryoutToDTO(created))
	}
}

func AdminUpdateTryout(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "tryoutId")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		existing, err := deps.TryoutService.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "tryout not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orig := existing
		if err := mergeTryoutSessionFromJSON(body, &existing); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		restoreTryoutFieldsIfEmptyPatch(&existing, orig)
		if err := validateTryoutAfterAdminUpdate(&existing); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		if existing.GradingMode == domain.TryoutGradingModeAuto {
			if err := deps.AdminService.ValidateTryoutAutoGradingPrerequisites(r.Context(), id); err != nil {
				writeError(w, http.StatusBadRequest, "validation_error", err.Error())
				return
			}
		}
		if err := deps.AdminService.UpdateTryout(r.Context(), existing); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		updated, err := deps.TryoutService.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "tryout not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		payload, err := json.Marshal(tryoutToDTO(updated))
		if err != nil {
			http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}
}

func AdminDeleteTryout(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "tryoutId")
		if err := deps.AdminService.DeleteTryout(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminListQuestions(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		list, err := deps.AdminService.ListQuestions(r.Context(), tryoutID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.QuestionResponse, len(list))
		for i := range list {
			out[i] = questionToAdminDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminGetQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		questionID := chi.URLParam(r, "questionId")
		q, err := deps.AdminService.GetQuestion(r.Context(), questionID)
		if err != nil {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if q.TryoutSessionID != tryoutID {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(questionToAdminDTO(q))
	}
}

func AdminCreateQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		var req dto.QuestionCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Body soal tidak valid: "+err.Error())
			return
		}
		req.Type = strings.TrimSpace(req.Type)
		req.Body = strings.TrimSpace(req.Body)
		if req.Type == "" || req.Body == "" {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "type dan body wajib diisi.")
			return
		}
		if req.MaxScore < 0 {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "maxScore harus >= 0.")
			return
		}
		opts, _ := json.Marshal(req.Options)
		imageURLs := []byte("[]")
		if len(req.ImageURLs) > 0 {
			imageURLs, _ = json.Marshal(req.ImageURLs)
		}
		tagsJSON := []byte("[]")
		if len(req.Tags) > 0 {
			tagsJSON, _ = json.Marshal(req.Tags)
		}
		q := domain.Question{
			TryoutSessionID: tryoutID,
			SortOrder:       req.SortOrder,
			Type:            req.Type,
			Body:            req.Body,
			ImageURL:        req.ImageURL,
			ImageURLs:       imageURLs,
			Options:         opts,
			MaxScore:        req.MaxScore,
			ModuleID:        req.ModuleID,
			ModuleTitle:     req.ModuleTitle,
			Bidang:          req.Bidang,
			Tags:            tagsJSON,
			CorrectOption:   req.CorrectOption,
			CorrectText:     req.CorrectText,
		}
		if q.MaxScore == 0 {
			q.MaxScore = 1
		}
		created, err := deps.AdminService.CreateQuestion(r.Context(), q)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(questionToAdminDTO(created))
	}
}

func AdminUpdateQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		questionID := chi.URLParam(r, "questionId")
		var req dto.QuestionUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Body soal tidak valid: "+err.Error())
			return
		}
		if req.MaxScore != nil && *req.MaxScore < 0 {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "maxScore harus >= 0.")
			return
		}
		q, err := deps.QuestionRepo.GetByID(r.Context(), questionID)
		if err != nil {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if q.TryoutSessionID != tryoutID {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if req.SortOrder != nil {
			q.SortOrder = *req.SortOrder
		}
		if req.Type != nil {
			q.Type = strings.TrimSpace(*req.Type)
		}
		if req.Body != nil {
			q.Body = strings.TrimSpace(*req.Body)
		}
		if req.ImageURL != nil {
			q.ImageURL = req.ImageURL
		}
		if req.ImageURLs != nil {
			imageURLs, _ := json.Marshal(*req.ImageURLs)
			q.ImageURLs = imageURLs
		}
		if req.Options != nil {
			opts, _ := json.Marshal(req.Options)
			q.Options = opts
		}
		if req.MaxScore != nil {
			q.MaxScore = *req.MaxScore
		}
		if req.ModuleID != nil {
			q.ModuleID = req.ModuleID
		}
		if req.ModuleTitle != nil {
			q.ModuleTitle = req.ModuleTitle
		}
		if req.Bidang != nil {
			q.Bidang = req.Bidang
		}
		if req.Tags != nil {
			tagsJSON, _ := json.Marshal(*req.Tags)
			q.Tags = tagsJSON
		}
		if req.CorrectOption != nil {
			q.CorrectOption = req.CorrectOption
		}
		if req.CorrectText != nil {
			q.CorrectText = req.CorrectText
		}
		if err := deps.AdminService.UpdateQuestion(r.Context(), q); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fresh, gerr := deps.QuestionRepo.GetByID(r.Context(), questionID)
		if gerr != nil {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(questionToAdminDTO(fresh))
	}
}

func AdminDeleteQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		questionID := chi.URLParam(r, "questionId")
		q, err := deps.AdminService.GetQuestion(r.Context(), questionID)
		if err != nil {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if q.TryoutSessionID != tryoutID {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if err := deps.AdminService.DeleteQuestion(r.Context(), questionID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminGetQuestionStats(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		questionID := chi.URLParam(r, "questionId")
		stats, err := deps.AdminService.GetQuestionStats(r.Context(), tryoutID, questionID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	}
}

func AdminGetTryoutQuestionStatsBulk(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		stats, err := deps.AdminService.GetTryoutQuestionStatsBulk(r.Context(), tryoutID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	}
}

func AdminGetTryoutAnalysis(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		analysis, err := deps.AdminService.GetTryoutAnalysis(r.Context(), tryoutID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(analysis)
	}
}

func AdminListTryoutStudents(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		list, err := deps.AdminService.ListTryoutStudents(r.Context(), tryoutID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}
}

func AdminGetAttemptAIAnalysis(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		analysis, err := deps.AdminService.GetAttemptAIAnalysis(r.Context(), tryoutID, attemptID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(analysis)
	}
}

func AdminListCourses(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.AdminService.ListCourses(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.CourseResponse, len(list))
		for i := range list {
			tt := list[i].TrackType
			if tt == "" {
				tt = domain.CourseTrackMeetings
			}
			out[i] = dto.CourseResponse{
				ID:          list[i].ID,
				Title:       list[i].Title,
				Slug:        list[i].Slug,
				Description: list[i].Description,
				Price:       list[i].Price,
				Thumbnail:   list[i].Thumbnail,
				SubjectID:   list[i].SubjectID,
				CreatedBy:   list[i].CreatedBy,
				TrackType:   tt,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminGetCourse(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "courseId")
		c, err := deps.AdminService.GetCourseByID(r.Context(), id)
		if err != nil {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		tt := c.TrackType
		if tt == "" {
			tt = domain.CourseTrackMeetings
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.CourseResponse{
			ID:          c.ID,
			Title:       c.Title,
			Slug:        c.Slug,
			Description: c.Description,
			Price:       c.Price,
			Thumbnail:   c.Thumbnail,
			SubjectID:   c.SubjectID,
			CreatedBy:   c.CreatedBy,
			TrackType:   tt,
		})
	}
}

func AdminCreateCourse(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.CourseCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		createdBy, ok := middleware.GetUserID(r.Context())
		var createdByPtr *string
		if ok && createdBy != "" {
			createdByPtr = &createdBy
		}
		track := domain.CourseTrackMeetings
		if req.TrackType != nil && strings.TrimSpace(*req.TrackType) != "" {
			t := strings.TrimSpace(strings.ToLower(*req.TrackType))
			if t != domain.CourseTrackMeetings && t != domain.CourseTrackTryout {
				writeError(w, http.StatusBadRequest, "validation_error", "trackType must be \"meetings\" or \"tryout\"")
				return
			}
			track = t
		}
		c := domain.Course{
			Title:       req.Title,
			Slug:        req.Slug,
			Description: req.Description,
			Thumbnail:   req.Thumbnail,
			SubjectID:   req.SubjectID,
			CreatedBy:   createdByPtr,
			TrackType:   track,
		}
		if req.Price != nil {
			c.Price = *req.Price
		}
		created, err := deps.AdminService.CreateCourse(r.Context(), c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(req.LinkedTryoutIds) > 0 {
			if deps.CourseAdminLinkRepo == nil {
				writeError(w, http.StatusServiceUnavailable, "service_unavailable", "course admin link repo not configured")
				return
			}
			if err := deps.CourseAdminLinkRepo.ReplaceTryoutsForCourse(r.Context(), created.ID, req.LinkedTryoutIds); err != nil {
				writeInternalError(w, r, err)
				return
			}
		}

		if courseCreateNeedsProgramSave(req, track) {
			if deps.CourseProgramService == nil {
				writeError(w, http.StatusServiceUnavailable, "service_unavailable", "course program service not configured")
				return
			}
			meetings := make([]domain.CourseProgramMeeting, 0, len(req.Meetings))
			for _, it := range req.Meetings {
				meetings = append(meetings, domain.CourseProgramMeeting{
					MeetingNumber:  it.MeetingNumber,
					Title:          it.Title,
					DetailText:     it.DetailText,
					PdfURL:         it.PdfURL,
					PptURL:         it.PptURL,
					PrTitle:        it.PrTitle,
					PrDescription:  it.PrDescription,
					LiveClassURL:   it.LiveClassURL,
				})
			}
			if err := deps.CourseProgramService.SaveProgram(r.Context(), created.ID, track, meetings, req.PretestTryoutSessionID); err != nil {
				if errors.Is(err, repo.ErrCourseProgramValidation) {
					writeError(w, http.StatusBadRequest, "validation_error", err.Error())
					return
				}
				writeInternalError(w, r, err)
				return
			}
		}

		// Reload agar track_type + konsisten pasca SaveProgram
		created, err = deps.AdminService.GetCourseByID(r.Context(), created.ID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}

		tt := created.TrackType
		if tt == "" {
			tt = domain.CourseTrackMeetings
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.CourseResponse{
			ID:          created.ID,
			Title:       created.Title,
			Slug:        created.Slug,
			Description: created.Description,
			Price:       created.Price,
			Thumbnail:   created.Thumbnail,
			SubjectID:   created.SubjectID,
			CreatedBy:   created.CreatedBy,
			TrackType:   tt,
		})
	}
}

// courseCreateNeedsProgramSave true jika body minta journey/program disinkronkan (bukan hanya metadata kelas).
func courseCreateNeedsProgramSave(req dto.CourseCreateRequest, track string) bool {
	if len(req.Meetings) > 0 {
		return true
	}
	if req.PretestTryoutSessionID != nil && strings.TrimSpace(*req.PretestTryoutSessionID) != "" {
		return true
	}
	if track == domain.CourseTrackTryout && len(req.LinkedTryoutIds) > 0 {
		return true
	}
	return false
}

func AdminUpdateCourse(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "courseId")
		c, err := deps.AdminService.GetCourseByID(r.Context(), id)
		if err != nil {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		var req dto.CourseUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		c.Title = req.Title
		c.Description = req.Description
		if req.SubjectID != nil {
			c.SubjectID = req.SubjectID
		}
		if req.Slug != nil {
			c.Slug = req.Slug
		}
		if req.Price != nil {
			c.Price = *req.Price
		}
		if req.Thumbnail != nil {
			c.Thumbnail = req.Thumbnail
		}
		if err := deps.AdminService.UpdateCourse(r.Context(), c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func AdminListEnrollments(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := chi.URLParam(r, "courseId")
		list, err := deps.AdminService.ListEnrollmentsByCourse(r.Context(), courseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.EnrollmentResponse, len(list))
		for i := range list {
			out[i] = dto.EnrollmentResponse{
				ID:         list[i].ID,
				UserID:     list[i].UserID,
				CourseID:   list[i].CourseID,
				Status:     list[i].Status,
				EnrolledAt: list[i].EnrolledAt.Format(time.RFC3339),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminIssueCertificate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.CertificateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		c := domain.Certificate{
			UserID:          req.UserID,
			TryoutSessionID: req.TryoutSessionID,
			CourseID:        req.CourseID,
		}
		created, err := deps.AdminService.IssueCertificate(r.Context(), c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        created.ID,
			"userId":    created.UserID,
			"issuedAt":  created.IssuedAt.Format(time.RFC3339),
		})
	}
}

func courseContentToDTO(c domain.CourseContent) dto.CourseContentResponse {
	var content interface{}
	if len(c.Content) > 0 {
		_ = json.Unmarshal(c.Content, &content)
	}
	return dto.CourseContentResponse{
		ID:          c.ID,
		CourseID:    c.CourseID,
		Title:       c.Title,
		Description: c.Description,
		SortOrder:   c.SortOrder,
		Type:        c.Type,
		Content:     content,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
	}
}

func AdminListCourseContents(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := chi.URLParam(r, "courseId")
		list, err := deps.AdminService.ListCourseContents(r.Context(), courseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.CourseContentResponse, len(list))
		for i := range list {
			out[i] = courseContentToDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminCreateCourseContent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := chi.URLParam(r, "courseId")
		var req dto.CourseContentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Title == "" || req.Type == "" {
			http.Error(w, "title and type required", http.StatusBadRequest)
			return
		}
		content, _ := json.Marshal(req.Content)
		c := domain.CourseContent{
			CourseID:    courseID,
			Title:       req.Title,
			Description: req.Description,
			SortOrder:   req.SortOrder,
			Type:        req.Type,
			Content:     content,
		}
		created, err := deps.AdminService.CreateCourseContent(r.Context(), c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(courseContentToDTO(created))
	}
}

func AdminUpdateCourseContent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentID := chi.URLParam(r, "contentId")
		c, err := deps.AdminService.GetCourseContent(r.Context(), contentID)
		if err != nil {
			http.Error(w, "content not found", http.StatusNotFound)
			return
		}
		var req dto.CourseContentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Title != "" {
			c.Title = req.Title
		}
		if req.Description != nil {
			c.Description = req.Description
		}
		// allow 0
		c.SortOrder = req.SortOrder
		if req.Type != "" {
			c.Type = req.Type
		}
		if req.Content != nil {
			content, _ := json.Marshal(req.Content)
			c.Content = content
		}
		if err := deps.AdminService.UpdateCourseContent(r.Context(), c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func AdminDeleteCourseContent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentID := chi.URLParam(r, "contentId")
		if err := deps.AdminService.DeleteCourseContent(r.Context(), contentID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminListPayments(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := parseInt(l); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		list, err := deps.AdminService.ListPayments(r.Context(), limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.PaymentResponse, len(list))
		for i := range list {
			var paidAt *string
			if list[i].PaidAt != nil {
				s := list[i].PaidAt.Format(time.RFC3339)
				paidAt = &s
			}
			out[i] = dto.PaymentResponse{
				ID:     list[i].ID,
				UserID: list[i].UserID,
				Amount: list[i].Amount,
				Currency:    list[i].Currency,
				Status:      list[i].Status,
				Type:        list[i].Type,
				PaidAt:      paidAt,
				CreatedAt:   list[i].CreatedAt.Format(time.RFC3339),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func AdminCreatePayment(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.PaymentCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.UserID == "" || req.Amount <= 0 {
			http.Error(w, "user_id and amount required", http.StatusBadRequest)
			return
		}
		status := req.Status
		if status == "" {
			status = domain.PaymentStatusPending
		}
		ptype := req.Type
		if ptype == "" {
			ptype = domain.PaymentTypeCoursePurchase
		}
		var paidAt *time.Time
		if req.PaidAt != nil && strings.TrimSpace(*req.PaidAt) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.PaidAt))
			if err != nil {
				http.Error(w, "paidAt must be RFC3339", http.StatusBadRequest)
				return
			}
			paidAt = &t
		} else if status == domain.PaymentStatusPaid {
			now := time.Now()
			paidAt = &now
		}
		currency := req.Currency
		if currency == "" {
			currency = "IDR"
		}
		p := domain.Payment{
			UserID:      req.UserID,
			Amount:      req.Amount,
			Currency:    currency,
			Status:      status,
			Type:        ptype,
			ReferenceID: req.ReferenceID,
			Description: req.Description,
			OrderID:     req.OrderID,
			ProofURL:    req.ProofURL,
			PaidAt:      paidAt,
		}
		created, err := deps.AdminService.CreatePayment(r.Context(), p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		var paidAtStr *string
		if created.PaidAt != nil {
			s := created.PaidAt.Format(time.RFC3339)
			paidAtStr = &s
		}
		_ = json.NewEncoder(w).Encode(dto.PaymentResponse{
			ID:        created.ID,
			UserID:    created.UserID,
			Amount:    created.Amount,
			Currency:  created.Currency,
			Status:    created.Status,
			Type:      created.Type,
			PaidAt:    paidAtStr,
			CreatedAt: created.CreatedAt.Format(time.RFC3339),
		})
	}
}

func AdminConfirmPayment(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		paymentID := chi.URLParam(r, "paymentId")
		var req struct {
			Confirmed     bool    `json:"confirmed"`
			RejectionNote *string `json:"rejectionNote"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if err := deps.AdminService.ConfirmPayment(r.Context(), paymentID, req.Confirmed, adminID, req.RejectionNote); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "payment updated"})
	}
}

// AdminConfirmPaymentByAction: POST /api/v1/admin/payments/:paymentId/confirm
func AdminConfirmPaymentByAction(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		paymentID := chi.URLParam(r, "paymentId")
		if err := deps.AdminService.ConfirmPayment(r.Context(), paymentID, true, adminID, nil); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "payment confirmed"})
	}
}

// AdminRejectPaymentByAction: POST /api/v1/admin/payments/:paymentId/reject
// Body optional: { "reason": "..." }
func AdminRejectPaymentByAction(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		paymentID := chi.URLParam(r, "paymentId")
		var req struct {
			Reason *string `json:"reason"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if err := deps.AdminService.ConfirmPayment(r.Context(), paymentID, false, adminID, req.Reason); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "payment rejected"})
	}
}

// AdminTransactionDetail: GET /api/v1/admin/transactions/:orderId
// Returns transaction detail for admin page (order, buyer, items, and latest payment by order).
func AdminTransactionDetail(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "orderId")
		if orderID == "" {
			http.Error(w, "orderId required", http.StatusBadRequest)
			return
		}
		order, err := deps.OrderRepo.GetByID(r.Context(), orderID)
		if err != nil {
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		}
		user, err := deps.UserRepo.FindByID(r.Context(), order.UserID)
		if err != nil {
			http.Error(w, "buyer not found", http.StatusNotFound)
			return
		}
		items, err := deps.OrderItemRepo.ListByOrderID(r.Context(), orderID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		type itemResp struct {
			OrderItemID string `json:"orderItemId"`
			CourseID    string `json:"courseId"`
			CourseTitle string `json:"courseTitle"`
			CourseSlug  string `json:"courseSlug,omitempty"`
			Price       int    `json:"price"`
		}
		itemOut := make([]itemResp, 0, len(items))
		for _, it := range items {
			courseTitle := ""
			var courseSlug string
			if c, err := deps.CourseRepo.GetByID(r.Context(), it.CourseID); err == nil {
				courseTitle = c.Title
				if c.Slug != nil {
					courseSlug = *c.Slug
				}
			}
			itemOut = append(itemOut, itemResp{
				OrderItemID: it.ID,
				CourseID:    it.CourseID,
				CourseTitle: courseTitle,
				CourseSlug:  courseSlug,
				Price:       it.Price,
			})
		}

		var paymentOut any
		if p, err := deps.PaymentRepo.GetByOrderID(r.Context(), orderID); err == nil {
			var paidAt *string
			if p.PaidAt != nil {
				s := p.PaidAt.Format(time.RFC3339)
				paidAt = &s
			}
			var confirmedAt *string
			if p.ConfirmedAt != nil {
				s := p.ConfirmedAt.Format(time.RFC3339)
				confirmedAt = &s
			}
			paymentOut = map[string]any{
				"id":              p.ID,
				"amount":          p.Amount,
				"currency":        p.Currency,
				"status":          p.Status,
				"type":            p.Type,
				"gateway":         p.Gateway,
				"transactionId":   p.TransactionID,
				"proofUrl":        p.ProofURL,
				"confirmedBy":     p.ConfirmedBy,
				"confirmedAt":     confirmedAt,
				"rejectionNote":   p.RejectionNote,
				"paidAt":          paidAt,
				"createdAt":       p.CreatedAt.Format(time.RFC3339),
				"updatedAt":       p.UpdatedAt.Format(time.RFC3339),
			}
		}

		var discountPercent *float64
		if order.DiscountPercent != nil {
			discountPercent = order.DiscountPercent
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"order": map[string]any{
				"id":                 order.ID,
				"status":             order.Status,
				"totalPrice":         order.TotalPrice,
				"normalPrice":        order.NormalPrice,
				"discount":           order.Discount,
				"discountPercent":    discountPercent,
				"promoCode":          order.PromoCode,
				"confirmationCode":   order.ConfirmationCode,
				"paymentMethod":      order.PaymentMethod,
				"paymentReference":   order.PaymentReference,
				"paymentProofUrl":    order.PaymentProofURL,
				"paymentProofAt":     order.PaymentProofAt,
				"senderAccountNo":    order.SenderAccountNo,
				"senderName":         order.SenderName,
				"roleHint":           order.RoleHint,
				"buyerEmail":         order.BuyerEmail,
				"createdAt":          order.CreatedAt.Format(time.RFC3339),
				"updatedAt":          order.UpdatedAt.Format(time.RFC3339),
			},
			"buyer": map[string]any{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
			"items":   itemOut,
			"payment": paymentOut,
		})
	}
}

// AdminVerifyOrder: PUT /api/v1/admin/orders/:orderId/verify — verifikasi pembayaran, enroll user, kirim email.
// Body opsional: { "purchasedAt": "2026-01-15T10:00:00Z" } untuk menyelaraskan tanggal pembelian.
func AdminVerifyOrder(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "orderId")
		if orderID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "orderId required")
			return
		}
		var body dto.AdminVerifyOrderRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		var purchasedAt *time.Time
		if body.PurchasedAt != nil && strings.TrimSpace(*body.PurchasedAt) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*body.PurchasedAt))
			if err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "purchasedAt harus RFC3339")
				return
			}
			purchasedAt = &t
		}
		if err := deps.CheckoutService.VerifyOrder(r.Context(), orderID, purchasedAt); err != nil {
			if err == service.ErrOrderNotFound {
				writeError(w, http.StatusNotFound, "order_not_found", "order tidak ditemukan")
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Pembayaran terverifikasi, user ter-enroll"})
	}
}

func AdminReportMonthly(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		year, month := time.Now().Year(), int(time.Now().Month())
		if y := r.URL.Query().Get("year"); y != "" {
			if n, err := parseInt(y); err == nil && n >= 2020 && n <= 2100 {
				year = n
			}
		}
		if m := r.URL.Query().Get("month"); m != "" {
			if n, err := parseInt(m); err == nil && n >= 1 && n <= 12 {
				month = n
			}
		}
		report, err := deps.AdminService.ReportMonthly(r.Context(), year, month)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.MonthlyReportResponse{
			Year:              report.Year,
			Month:             report.Month,
			NewEnrollments:    report.NewEnrollments,
			PaymentsCount:     report.PaymentsCount,
			TotalRevenue: report.TotalRevenue,
		})
	}
}

func AdminCourseReport(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := chi.URLParam(r, "courseId")
		if courseID == "" {
			http.Error(w, "course_id required", http.StatusBadRequest)
			return
		}
		report, err := deps.AdminService.GetCourseReport(r.Context(), courseID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				http.Error(w, "course not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		students := make([]dto.CourseReportStudentRow, 0, len(report.Students))
		for _, st := range report.Students {
			scores := make([]dto.CourseReportTryoutScore, 0, len(st.TryoutScores))
			for _, sc := range st.TryoutScores {
				var submittedAt *string
				if sc.SubmittedAt != nil {
					s := sc.SubmittedAt.Format(time.RFC3339)
					submittedAt = &s
				}
				scores = append(scores, dto.CourseReportTryoutScore{
					TryoutID:    sc.TryoutID,
					TryoutTitle: sc.TryoutTitle,
					AttemptID:   sc.AttemptID,
					Score:       sc.Score,
					MaxScore:    sc.MaxScore,
					Percentile:  sc.Percentile,
					SubmittedAt: submittedAt,
				})
			}
			var completedAt *string
			if st.Progress.CompletedAt != nil {
				s := st.Progress.CompletedAt.Format(time.RFC3339)
				completedAt = &s
			}
			var lastActivityAt *string
			if st.Attendance.LastActivityAt != nil {
				s := st.Attendance.LastActivityAt.Format(time.RFC3339)
				lastActivityAt = &s
			}
			students = append(students, dto.CourseReportStudentRow{
				StudentID:        st.StudentID,
				StudentName:      st.StudentName,
				StudentEmail:     st.StudentEmail,
				EnrolledAt:       st.EnrolledAt.Format(time.RFC3339),
				EnrollmentStatus: st.EnrollmentStatus,
				Progress: dto.CourseReportProgress{
					Status:      st.Progress.Status,
					CompletedAt: completedAt,
				},
				TryoutScores: scores,
				Attendance: dto.CourseReportAttendance{
					TryoutsParticipated: st.Attendance.TryoutsParticipated,
					LastActivityAt:      lastActivityAt,
				},
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.CourseReportResponse{
			Course: dto.CourseReportCourseInfo{
				ID:          report.Course.ID,
				Title:       report.Course.Title,
				Description: report.Course.Description,
			},
			GeneratedAt: report.GeneratedAt.Format(time.RFC3339),
			Students:    students,
		})
	}
}
