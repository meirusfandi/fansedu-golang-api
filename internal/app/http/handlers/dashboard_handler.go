package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
)

// DashboardGeneral returns public/general dashboard stats (no auth). GET /api/v1/dashboard
func DashboardGeneral(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overview, _ := deps.AdminService.Overview(r.Context())
		levels, _ := deps.LevelRepo.List(r.Context())
		subjects, _ := deps.SubjectRepo.List(r.Context())
		schools, _ := deps.SchoolRepo.List(r.Context())
		siteName := "FansEdu LMS"
		if s, err := deps.SettingRepo.GetByKey(r.Context(), "site_name"); err == nil && s.Value != nil {
			siteName = *s.Value
		}
		out := dto.GeneralDashboardResponse{
			SiteName:      siteName,
			TotalLevels:   len(levels),
			TotalSubjects: len(subjects),
			TotalSchools:  len(schools),
		}
		if overview != nil {
			out.OpenTryouts = overview.ActiveTryouts
			out.TotalCourses = overview.TotalCourses
			out.TotalStudents = overview.TotalStudents
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func DashboardStudent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		u, err := deps.UserRepo.FindByID(r.Context(), userID)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		userDTO := dto.DashboardUser{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      u.Role,
			AvatarURL: u.AvatarURL,
			SchoolID:  u.SchoolID,
			SubjectID: u.SubjectID,
		}
		if u.SchoolID != nil && *u.SchoolID != "" {
			if school, err := deps.SchoolRepo.GetByID(r.Context(), *u.SchoolID); err == nil {
				userDTO.SchoolName = &school.Name
			}
		}
		if u.SubjectID != nil && *u.SubjectID != "" {
			if subj, err := deps.SubjectRepo.GetByID(r.Context(), *u.SubjectID); err == nil {
				userDTO.SubjectName = &subj.Name
			}
		}
		resp, err := deps.DashboardService.GetStudentDashboard(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		openTryouts := make([]interface{}, 0, len(resp.OpenTryouts))
		for _, t := range resp.OpenTryouts {
			openTryouts = append(openTryouts, tryoutToDTO(t))
		}
		recentAttempts := make([]interface{}, 0, len(resp.RecentAttempts))
		for _, a := range resp.RecentAttempts {
			recentAttempts = append(recentAttempts, attemptToDTO(a))
		}
		var learningEval *dto.LearningEvaluationDTO
		if resp.LearningEvaluation != nil {
			breakdown := make([]dto.QuestionScoreDetail, 0, len(resp.LearningEvaluation.AnswerBreakdown))
			for _, b := range resp.LearningEvaluation.AnswerBreakdown {
				breakdown = append(breakdown, dto.QuestionScoreDetail{
					QuestionID:   b.QuestionID,
					QuestionType: b.QuestionType,
					MaxScore:     b.MaxScore,
					ScoreGot:     b.ScoreGot,
					Status:       b.Status,
				})
			}
			learningEval = &dto.LearningEvaluationDTO{
				AttemptID:       resp.LearningEvaluation.AttemptID,
				AnswerBreakdown: breakdown,
				StrengthAreas:   resp.LearningEvaluation.StrengthAreas,
				ImprovementAreas: resp.LearningEvaluation.ImprovementAreas,
				Recommendation:  resp.LearningEvaluation.Recommendation,
			}
		}
		out := dto.DashboardResponse{
			User:               userDTO,
			Summary: dto.DashboardSummary{
				TotalAttempts:  resp.Summary.TotalAttempts,
				AvgScore:       resp.Summary.AvgScore,
				AvgPercentile:  resp.Summary.AvgPercentile,
			},
			OpenTryouts:        openTryouts,
			RecentAttempts:     recentAttempts,
			CourseProgress:    nil,
			StrengthAreas:      resp.StrengthAreas,
			ImprovementAreas:   resp.ImprovementAreas,
			Recommendation:     resp.Recommendation,
			LearningEvaluation: learningEval,
		}

		// Course progress for "kelas" section in student dashboard.
		// We derive progress from enrollment status because course-level progress
		// isn't stored as a separate numeric field yet.
		if enrollments, err := deps.EnrollmentRepo.ListByUserID(r.Context(), userID); err == nil {
			progress := make([]dto.StudentCourseItem, 0, len(enrollments))
			for _, e := range enrollments {
				c, err := deps.CourseRepo.GetByID(r.Context(), e.CourseID)
				if err != nil {
					continue
				}
				slug := ""
				if c.Slug != nil {
					slug = *c.Slug
				}
				thumb := ""
				if c.Thumbnail != nil {
					thumb = *c.Thumbnail
				}
				enrolledAt := e.EnrolledAt.Format("2006-01-02T15:04:05Z07:00")
				progressPercent := enrollmentProgressPercent(e.Status)
				progress = append(progress, dto.StudentCourseItem{
					ID:              e.ID,
					Program:         dto.StudentCourseProgram{ID: c.ID, Slug: slug, Title: c.Title, Thumbnail: thumb},
					ProgressPercent: progressPercent,
					EnrolledAt:      enrolledAt,
					LastAccessedAt:  enrolledAt,
				})
			}
			out.CourseProgress = progress
		} else {
			out.CourseProgress = []dto.StudentCourseItem{}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
