package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
)

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
