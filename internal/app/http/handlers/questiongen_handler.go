package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/usecase/questiongen"
)

func GenerateQuestions(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.QuestionGenUsecase == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "question generator unavailable")
			return
		}
		var req dto.GenerateQuestionsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		list, err := deps.QuestionGenUsecase.GenerateQuestions(r.Context(), questiongen.GenerateQuestionsRequest{
			Subject:    req.Subject,
			Grade:      req.Grade,
			Topic:      req.Topic,
			Difficulty: req.Difficulty,
			Count:      req.Count,
		})
		if err != nil {
			if errors.Is(err, questiongen.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, "validation_error", "subject, grade, topic, count required")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": list})
	}
}

func SubmitAnswer(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.QuestionGenUsecase == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "question generator unavailable")
			return
		}
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		var req dto.SubmitAnswerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		out, err := deps.QuestionGenUsecase.SubmitAnswer(r.Context(), userID, questiongen.SubmitAnswerRequest{
			QuestionID:  req.QuestionID,
			Answer:      req.Answer,
			TimeSpentMs: req.TimeSpentMs,
		})
		if err != nil {
			if errors.Is(err, questiongen.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, "validation_error", "questionId and answer required")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func GetAnalysis(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.QuestionGenUsecase == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "question generator unavailable")
			return
		}
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		topic := strings.TrimSpace(r.URL.Query().Get("topic"))
		grade := strings.TrimSpace(r.URL.Query().Get("grade"))
		out, err := deps.QuestionGenUsecase.Analysis(r.Context(), userID, topic, grade)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func GetRanking(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.QuestionGenUsecase == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "question generator unavailable")
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		out, err := deps.QuestionGenUsecase.Ranking(r.Context(), limit)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}
}

func ListQuestions(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.QuestionGenUsecase == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "question generator unavailable")
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		out, err := deps.QuestionGenUsecase.Questions(r.Context(), questiongen.ListQuestionsRequest{
			Subject:    strings.TrimSpace(strings.ToLower(r.URL.Query().Get("subject"))),
			Grade:      strings.TrimSpace(strings.ToLower(r.URL.Query().Get("grade"))),
			Topic:      strings.TrimSpace(strings.ToLower(r.URL.Query().Get("topic"))),
			Difficulty: strings.TrimSpace(strings.ToLower(r.URL.Query().Get("difficulty"))),
			Limit:      limit,
		})
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}
}

func CreateSubscription(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.QuestionGenUsecase == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "question generator unavailable")
			return
		}
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		var req dto.CreateSubscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		var startAt *time.Time
		var endAt *time.Time
		if req.StartAt != nil && strings.TrimSpace(*req.StartAt) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.StartAt))
			if err != nil {
				writeError(w, http.StatusBadRequest, "validation_error", "startAt must be RFC3339")
				return
			}
			startAt = &t
		}
		if req.EndAt != nil && strings.TrimSpace(*req.EndAt) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.EndAt))
			if err != nil {
				writeError(w, http.StatusBadRequest, "validation_error", "endAt must be RFC3339")
				return
			}
			endAt = &t
		}
		out, err := deps.QuestionGenUsecase.Subscribe(r.Context(), userID, questiongen.CreateSubscriptionRequest{
			PlanCode: req.PlanCode,
			StartAt:  startAt,
			EndAt:    endAt,
		})
		if err != nil {
			if errors.Is(err, questiongen.ErrInvalidInput) {
				writeError(w, http.StatusBadRequest, "validation_error", "planCode required")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(out)
	}
}

