package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
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
			ActiveTryouts:     overview.ActiveTryouts,
			AvgScore:          overview.AvgScore,
			TotalCertificates: overview.TotalCertificates,
		})
	}
}

func AdminCreateTryout(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.TryoutCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		t := domain.TryoutSession{
			Title:           req.Title,
			ShortTitle:      req.ShortTitle,
			Description:     req.Description,
			DurationMinutes: req.DurationMinutes,
			QuestionsCount:  req.QuestionsCount,
			Level:           req.Level,
			OpensAt:         req.OpensAt,
			ClosesAt:        req.ClosesAt,
			MaxParticipants: req.MaxParticipants,
			Status:          req.Status,
		}
		if t.Status == "" {
			t.Status = domain.TryoutStatusOpen
		}
		created, err := deps.AdminService.CreateTryout(r.Context(), t)
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
		var req dto.TryoutCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		t := domain.TryoutSession{
			ID:               id,
			Title:            req.Title,
			ShortTitle:       req.ShortTitle,
			Description:      req.Description,
			DurationMinutes:  req.DurationMinutes,
			QuestionsCount:   req.QuestionsCount,
			Level:            req.Level,
			OpensAt:          req.OpensAt,
			ClosesAt:         req.ClosesAt,
			MaxParticipants:  req.MaxParticipants,
			Status:           req.Status,
		}
		if err := deps.AdminService.UpdateTryout(r.Context(), t); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
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

func AdminCreateQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		var req dto.QuestionCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		opts, _ := json.Marshal(req.Options)
		q := domain.Question{
			TryoutSessionID: tryoutID,
			SortOrder:        req.SortOrder,
			Type:             req.Type,
			Body:             req.Body,
			Options:          opts,
			MaxScore:         req.MaxScore,
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
		_ = json.NewEncoder(w).Encode(questionToDTO(created))
	}
}

func AdminUpdateQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionID := chi.URLParam(r, "questionId")
		var req dto.QuestionUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		// Load existing then apply patch
		q, err := deps.QuestionRepo.GetByID(r.Context(), questionID)
		if err != nil {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if req.SortOrder != nil {
			q.SortOrder = *req.SortOrder
		}
		if req.Type != nil {
			q.Type = *req.Type
		}
		if req.Body != nil {
			q.Body = *req.Body
		}
		if req.Options != nil {
			opts, _ := json.Marshal(req.Options)
			q.Options = opts
		}
		if req.MaxScore != nil {
			q.MaxScore = *req.MaxScore
		}
		if err := deps.AdminService.UpdateQuestion(r.Context(), q); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func AdminDeleteQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		questionID := chi.URLParam(r, "questionId")
		if err := deps.AdminService.DeleteQuestion(r.Context(), questionID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminCreateCourse(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.CourseCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		c := domain.Course{
			Title:       req.Title,
			Description: req.Description,
		}
		created, err := deps.AdminService.CreateCourse(r.Context(), c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.CourseResponse{
			ID:          created.ID,
			Title:       created.Title,
			Description: created.Description,
			CreatedBy:   created.CreatedBy,
		})
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
			"id":         created.ID,
			"user_id":    created.UserID,
			"issued_at":  created.IssuedAt.Format(time.RFC3339),
		})
	}
}
