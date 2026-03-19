package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// POST /api/v1/analytics/pageview
func AnalyticsTrackPageview(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Path      string `json:"path"`
			URL       string `json:"url"`
			SessionID string `json:"sessionId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		path := strings.TrimSpace(req.Path)
		if path == "" {
			path = strings.TrimSpace(req.URL)
		}
		if path == "" {
			path = r.Header.Get("Referer")
		}
		if path == "" {
			path = "/"
		}

		sessionID := strings.TrimSpace(req.SessionID)
		if sessionID == "" {
			sessionID = strings.TrimSpace(r.Header.Get("X-Session-Id"))
		}
		if sessionID == "" {
			sessionID = strings.TrimSpace(r.Header.Get("X-Request-Id"))
		}
		if sessionID == "" {
			// Fallback ringan supaya tetap bisa catat pageview
			sessionID = strings.ReplaceAll(strings.TrimSpace(r.RemoteAddr), ":", "_")
		}

		ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
		if ip == "" {
			ip = strings.TrimSpace(r.RemoteAddr)
		}
		ua := strings.TrimSpace(r.UserAgent())

		if deps.AnalyticsRepo != nil {
			_ = deps.AnalyticsRepo.TrackPageView(r.Context(), sessionID, path, ip, ua)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}
}

// GET /api/v1/admin/analytics/summary
func AdminAnalyticsSummary(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.AnalyticsRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "analytics not configured")
			return
		}
		s, err := deps.AnalyticsRepo.GetSummary(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s)
	}
}

// GET /api/v1/admin/analytics/visitors?page=1&limit=20
func AdminAnalyticsVisitors(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.AnalyticsRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "analytics not configured")
			return
		}
		page := parseQueryInt(r, "page", 1)
		limit := parseQueryInt(r, "limit", 20)
		items, total, err := deps.AnalyticsRepo.ListVisitors(r.Context(), page, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":  items,
			"total": total,
			"page":  page,
			"limit": limit,
		})
	}
}

func parseQueryInt(r *http.Request, key string, def int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}
