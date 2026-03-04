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
		out := dto.DashboardResponse{
			Summary: dto.DashboardSummary{
				TotalAttempts:  resp.Summary.TotalAttempts,
				AvgScore:       resp.Summary.AvgScore,
				AvgPercentile:  resp.Summary.AvgPercentile,
			},
			OpenTryouts:      openTryouts,
			RecentAttempts:   recentAttempts,
			StrengthAreas:    resp.StrengthAreas,
			ImprovementAreas: resp.ImprovementAreas,
			Recommendation:   resp.Recommendation,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
