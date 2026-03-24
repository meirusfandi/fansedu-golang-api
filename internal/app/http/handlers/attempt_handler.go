package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/cache"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

func AttemptListByUser(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		list, err := deps.AttemptService.ListByUser(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.AttemptResponse, len(list))
		for i := range list {
			out[i] = attemptToDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AttemptGetByID(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attemptID := chi.URLParam(r, "attemptId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		a, err := deps.AttemptService.GetByID(r.Context(), attemptID, userID)
		if err != nil {
			if err == service.ErrAttemptNotFound || err == service.ErrNotYourAttempt {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(attemptToDTO(a))
	}
}

func AttemptGetQuestions(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attemptID := chi.URLParam(r, "attemptId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		attempt, err := deps.AttemptService.GetByID(r.Context(), attemptID, userID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		questions, err := deps.QuestionRepo.ListByTryoutSessionID(r.Context(), attempt.TryoutSessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.QuestionResponse, len(questions))
		for i := range questions {
			out[i] = questionToDTO(questions[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AttemptPutAnswer(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attemptID := chi.URLParam(r, "attemptId")
		questionID := chi.URLParam(r, "questionId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, err := deps.AttemptService.GetByID(r.Context(), attemptID, userID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var req dto.AnswerPutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		aa := domain.AttemptAnswer{
			AttemptID:      attemptID,
			QuestionID:     questionID,
			AnswerText:     req.AnswerText,
			SelectedOption: req.SelectedOption,
			IsMarked:       false,
		}
		if req.IsMarked != nil {
			aa.IsMarked = *req.IsMarked
		}
		if err := deps.AttemptAnswerRepo.Upsert(r.Context(), aa); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func AttemptSubmit(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attemptID := chi.URLParam(r, "attemptId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		a, fb, err := deps.AttemptService.Submit(r.Context(), attemptID, userID)
		if err != nil {
			if err == service.ErrAttemptNotFound || err == service.ErrNotYourAttempt {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			if err == service.ErrAlreadySubmitted {
				http.Error(w, "already submitted", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if a.Score != nil {
			_ = cache.LeaderboardZAdd(r.Context(), deps.Redis, a.TryoutSessionID, userID, *a.Score)
		}

		// Progress notification -> notify all trainers (guru/instructor) linked to this student.
		// We trigger it when an attempt is successfully submitted.
		if a.Status == domain.AttemptStatusSubmitted && a.SubmittedAt != nil {
			student, _ := deps.UserRepo.FindByID(r.Context(), userID)
			tryout, _ := deps.TryoutService.GetByID(r.Context(), a.TryoutSessionID)
			trainers, _ := deps.TrainerRepo.ListTrainersByStudent(r.Context(), userID)
			for _, t := range trainers {
				_, _ = deps.NotificationRepo.Create(r.Context(), domain.Notification{
					UserID: t.ID,
					Title:  "Progress Siswa",
					Body: fmt.Sprintf(
						"%s menyelesaikan tryout %s. Skor: %.0f",
						student.Name,
						tryout.Title,
						func() float64 {
							if a.Score != nil {
								return *a.Score
							}
							return 0
						}(),
					),
					Type: "progress_update",
				})
			}
		}

		score, percentile := 0.0, 0.0
		if a.Score != nil {
			score = *a.Score
		}
		if a.Percentile != nil {
			percentile = *a.Percentile
		}
		resp := dto.SubmitResponse{
			AttemptID:  a.ID,
			Score:      score,
			Percentile: percentile,
		}
		if fb != nil {
			resp.Feedback = &dto.FeedbackResponse{
				Summary:           fb.Summary,
				Recap:             fb.Recap,
				RecommendationText: fb.RecommendationText,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func attemptToDTO(a domain.Attempt) dto.AttemptResponse {
	return dto.AttemptResponse{
		ID:               a.ID,
		UserID:           a.UserID,
		TryoutSessionID:  a.TryoutSessionID,
		StartedAt:        a.StartedAt,
		SubmittedAt:      a.SubmittedAt,
		Status:           a.Status,
		Score:            a.Score,
		MaxScore:         a.MaxScore,
		Percentile:       a.Percentile,
		TimeSecondsSpent: a.TimeSecondsSpent,
	}
}

func questionToDTO(q domain.Question) dto.QuestionResponse {
	var opts interface{}
	if len(q.Options) > 0 {
		_ = json.Unmarshal(q.Options, &opts)
	}
	var imageURLs []string
	if len(q.ImageURLs) > 0 {
		_ = json.Unmarshal(q.ImageURLs, &imageURLs)
	}
	return dto.QuestionResponse{
		ID:               q.ID,
		TryoutSessionID:  q.TryoutSessionID,
		SortOrder:        q.SortOrder,
		Type:             q.Type,
		Body:             q.Body,
		ImageURL:         q.ImageURL,
		ImageURLs:        imageURLs,
		Options:          opts,
		MaxScore:         q.MaxScore,
	}
}
