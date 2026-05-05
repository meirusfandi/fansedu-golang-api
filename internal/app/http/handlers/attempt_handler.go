package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

func AttemptListByUser(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		list, err := deps.AttemptService.ListByUser(r.Context(), userID)
		if err != nil {
			writeInternalError(w, r, err)
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
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		a, err := deps.AttemptService.GetByID(r.Context(), attemptID, userID)
		if err != nil {
			if err == service.ErrAttemptNotFound || err == service.ErrNotYourAttempt {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Attempt tidak ditemukan.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		resp := attemptToDTO(a)
		if a.Status == domain.AttemptStatusSubmitted {
			if analysis, aerr := deps.AttemptService.TryoutAnalysisForAttempt(r.Context(), a.ID, a.TryoutSessionID); aerr == nil && analysis != nil {
				rev, mod, overall := tryoutAnalysisToDTO(analysis)
				resp.Review = rev
				resp.ModuleAnalysis = mod
				resp.OverallAnalysis = overall
				if len(mod) > 0 {
					resp.ModuleSummary = append([]dto.ModuleAnalysisRow(nil), mod...)
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func AttemptGetQuestions(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attemptID := chi.URLParam(r, "attemptId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		attempt, err := deps.AttemptService.GetByID(r.Context(), attemptID, userID)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Attempt tidak ditemukan.")
			return
		}
		questions, err := deps.QuestionRepo.ListByTryoutSessionID(r.Context(), attempt.TryoutSessionID)
		if err != nil {
			writeInternalError(w, r, err)
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

// StudentTryoutAttemptPaper returns the question set for an attempt under a tryout (student namespace).
// GET /api/v1/student/tryouts/{tryoutId}/attempts/{attemptId}/paper — same JSON array as GET /attempts/{attemptId}/questions.
func StudentTryoutAttemptPaper(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" || tryoutID == "" || attemptID == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "tryoutId, attemptId, dan autentikasi wajib.")
			return
		}
		attempt, err := deps.AttemptService.GetByID(r.Context(), attemptID, userID)
		if err != nil {
			if err == service.ErrAttemptNotFound || err == service.ErrNotYourAttempt {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Attempt tidak ditemukan.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		if attempt.TryoutSessionID != tryoutID {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Attempt tidak sesuai tryout ini.")
			return
		}
		t, err := deps.TryoutService.GetByID(r.Context(), tryoutID)
		if err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		if role, _ := middleware.GetRole(r.Context()); domain.IsStudentRoleCode(role) {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, uerr := deps.UserRepo.FindByID(r.Context(), userID)
				if uerr != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
					return
				}
			}
		}
		questions, err := deps.QuestionRepo.ListByTryoutSessionID(r.Context(), tryoutID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		out := make([]dto.QuestionResponse, len(questions))
		for i := range questions {
			out[i] = questionToDTO(questions[i])
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AttemptPutAnswer(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attemptID := chi.URLParam(r, "attemptId")
		questionID := chi.URLParam(r, "questionId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		_, err := deps.AttemptService.GetByID(r.Context(), attemptID, userID)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Attempt tidak ditemukan.")
			return
		}
		var req dto.AnswerPutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Body permintaan tidak valid.")
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
			writeInternalError(w, r, err)
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
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		a, fb, analysis, err := deps.AttemptService.Submit(r.Context(), attemptID, userID)
		if err != nil {
			if err == service.ErrAttemptNotFound || err == service.ErrNotYourAttempt {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Attempt tidak ditemukan.")
				return
			}
			if err == service.ErrAlreadySubmitted {
				writeError(w, http.StatusConflict, "ALREADY_SUBMITTED", "Attempt sudah dikirim.")
				return
			}
			writeInternalError(w, r, err)
			return
		}

		if a.Status == domain.AttemptStatusSubmitted {
			ReconcileTryoutLeaderboardRedis(r.Context(), deps, a.TryoutSessionID)
		}

		// Progress notification -> notify all trainers/guru linked to this student.
		// We trigger it when an attempt is successfully submitted.
		if a.Status == domain.AttemptStatusSubmitted && a.SubmittedAt != nil {
			student, _ := deps.UserRepo.FindByID(r.Context(), userID)
			tryout, _ := deps.TryoutService.GetByID(r.Context(), a.TryoutSessionID)
			trainers, _ := deps.TrainerRepo.ListTrainersByStudent(r.Context(), userID)
			for _, t := range trainers {
				body := fmt.Sprintf("%s menyelesaikan tryout %s.", student.Name, tryout.Title)
				if tryout.GradingMode == domain.TryoutGradingModeManual {
					body += " Menunggu penilaian manual."
				} else if a.Score != nil {
					body += fmt.Sprintf(" Skor: %.0f", *a.Score)
				}
				_, _ = deps.NotificationRepo.Create(r.Context(), domain.Notification{
					UserID: t.ID,
					Title:  "Progress Siswa",
					Body:   body,
					Type:   "progressUpdate",
				})
			}
		}

		maxScore := 0.0
		if a.MaxScore != nil {
			maxScore = *a.MaxScore
		}
		tryoutMeta, _ := deps.TryoutService.GetByID(r.Context(), a.TryoutSessionID)
		pending := tryoutMeta.GradingMode == domain.TryoutGradingModeManual
		resp := dto.SubmitResponse{
			AttemptID:      a.ID,
			Score:          a.Score,
			MaxScore:       maxScore,
			GradingPending: pending,
			Percentile:     a.Percentile,
		}
		if fb != nil {
			resp.Feedback = &dto.FeedbackResponse{
				Summary:            fb.Summary,
				Recap:              fb.Recap,
				RecommendationText: fb.RecommendationText,
			}
		}
		if analysis != nil {
			rev, mod, overall := tryoutAnalysisToDTO(analysis)
			resp.Review = rev
			resp.ModuleAnalysis = mod
			resp.OverallAnalysis = overall
			if len(mod) > 0 {
				resp.ModuleSummary = append([]dto.ModuleAnalysisRow(nil), mod...)
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

func tryoutAnalysisToDTO(analysis *service.TryoutSubmitAnalysis) (review []dto.AttemptReviewRow, modules []dto.ModuleAnalysisRow, overall *dto.TryoutOverallAnalysisRow) {
	if analysis == nil {
		return nil, nil, nil
	}
	for _, o := range analysis.Review {
		review = append(review, dto.AttemptReviewRow{
			QuestionID:        o.QuestionID,
			SortOrder:         o.SortOrder,
			QuestionType:      o.QuestionType,
			QuestionTypeLabel: o.QuestionTypeLabel,
			QuestionBody:      o.QuestionBody,
			AnswerText:        o.AnswerText,
			SelectedOption:    o.SelectedOption,
			CorrectOption:     o.CorrectOption,
			CorrectText:       o.CorrectText,
			IsCorrect:         o.IsCorrect,
			ScoreGot:          o.ScoreGot,
			MaxScore:          o.MaxScore,
			AnalysisSummary:   o.AnalysisSummary,
			AnalysisDetail:    o.AnalysisDetail,
			ModuleKey:         o.ModuleKey,
			ModuleLabel:       o.ModuleLabel,
			ModuleID:          o.ModuleID,
			ModuleTitle:       o.ModuleTitle,
			Bidang:            o.Bidang,
			Tags:              o.Tags,
		})
	}
	for _, m := range analysis.Modules {
		modules = append(modules, dto.ModuleAnalysisRow{
			ModuleKey:     m.ModuleKey,
			ModuleLabel:   m.ModuleLabel,
			QuestionCount: m.QuestionCount,
			CorrectCount:  m.CorrectCount,
			WrongCount:    m.WrongCount,
			UnscoredCount: m.UnscoredCount,
		})
	}
	overall = overallTryoutAnalysisToDTO(analysis.Overall)
	return review, modules, overall
}

func overallTryoutAnalysisToDTO(o service.TryoutOverallAnalysis) *dto.TryoutOverallAnalysisRow {
	by := make([]dto.QuestionTypeStatRow, len(o.ByQuestionType))
	for i, t := range o.ByQuestionType {
		by[i] = dto.QuestionTypeStatRow{
			Type: t.Type, Label: t.Label, Total: t.Total,
			Correct: t.Correct, Wrong: t.Wrong, Unscored: t.Unscored,
			ScoreGot: t.ScoreGot, MaxScore: t.MaxScore,
		}
	}
	return &dto.TryoutOverallAnalysisRow{
		TotalQuestions:  o.TotalQuestions,
		AnsweredCount:   o.AnsweredCount,
		UnansweredCount: o.UnansweredCount,
		CorrectCount:    o.CorrectCount,
		WrongCount:      o.WrongCount,
		UnscoredCount:   o.UnscoredCount,
		ScorePercent:    o.ScorePercent,
		ScoreGot:        o.ScoreGot,
		MaxScore:        o.MaxScore,
		ByQuestionType:  by,
		Summary:         o.Summary,
	}
}

// questionToDTO soal untuk siswa / ujian (tanpa kunci jawaban).
func questionToDTO(q domain.Question) dto.QuestionResponse {
	return questionToDTOWithSecrets(q, false)
}

// questionToAdminDTO soal untuk admin (termasuk kunci PG / isian jika ada).
func questionToAdminDTO(q domain.Question) dto.QuestionResponse {
	return questionToDTOWithSecrets(q, true)
}

func questionToDTOWithSecrets(q domain.Question, includeKeys bool) dto.QuestionResponse {
	var opts interface{}
	if len(q.Options) > 0 {
		_ = json.Unmarshal(q.Options, &opts)
	}
	var imageURLs []string
	if len(q.ImageURLs) > 0 {
		_ = json.Unmarshal(q.ImageURLs, &imageURLs)
	}
	var tags []string
	if len(q.Tags) > 0 {
		_ = json.Unmarshal(q.Tags, &tags)
	}
	r := dto.QuestionResponse{
		ID:              q.ID,
		TryoutSessionID: q.TryoutSessionID,
		SortOrder:       q.SortOrder,
		Type:            q.Type,
		Body:            q.Body,
		ImageURL:        q.ImageURL,
		ImageURLs:       imageURLs,
		Options:         opts,
		MaxScore:        q.MaxScore,
		ModuleID:        q.ModuleID,
		ModuleTitle:     q.ModuleTitle,
		Bidang:          q.Bidang,
		Tags:            tags,
	}
	if includeKeys {
		r.CorrectOption = q.CorrectOption
		r.CorrectText = q.CorrectText
	}
	return r
}
