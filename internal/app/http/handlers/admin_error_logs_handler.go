package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

func errorLogToMap(e domain.ApplicationErrorLog) map[string]interface{} {
	m := map[string]interface{}{
		"id":         e.ID,
		"createdAt":  e.CreatedAt.UTC().Format(time.RFC3339),
		"errorType":  e.ErrorType,
		"message":    e.Message,
		"httpStatus": e.HTTPStatus,
		"method":     e.Method,
		"path":       e.Path,
	}
	if e.ErrorCode != nil {
		m["errorCode"] = *e.ErrorCode
	}
	if e.QueryString != nil {
		m["queryString"] = *e.QueryString
	}
	if e.UserID != nil {
		m["userId"] = *e.UserID
	}
	if e.UserRole != nil {
		m["userRole"] = *e.UserRole
	}
	if e.RequestID != nil {
		m["requestId"] = *e.RequestID
	}
	if e.IPAddress != nil {
		m["ipAddress"] = *e.IPAddress
	}
	if e.UserAgent != nil {
		m["userAgent"] = *e.UserAgent
	}
	if e.StackTrace != nil {
		m["stackTrace"] = *e.StackTrace
	}
	if e.Meta != nil {
		m["meta"] = e.Meta
	}
	if e.ResolvedAt != nil {
		m["resolvedAt"] = e.ResolvedAt.UTC().Format(time.RFC3339)
	}
	if e.ResolvedBy != nil {
		m["resolvedBy"] = *e.ResolvedBy
	}
	if e.AdminNote != nil {
		m["adminNote"] = *e.AdminNote
	}
	return m
}

// AdminErrorLogsList GET /api/v1/admin/error-logs
func AdminErrorLogsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.ApplicationErrorLogRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "error log store unavailable")
			return
		}
		page := 1
		if v := r.URL.Query().Get("page"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				page = n
			}
		}
		limit := 20
		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		offset := (page - 1) * limit
		filter := repo.ApplicationErrorLogListFilter{
			ErrorType:  r.URL.Query().Get("error_type"),
			Offset:     offset,
			Limit:      limit,
			Search:     r.URL.Query().Get("q"),
		}
		if v := r.URL.Query().Get("http_status"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				filter.HTTPStatus = n
			}
		}
		if v := r.URL.Query().Get("resolved"); v != "" {
			if v == "true" {
				t := true
				filter.Resolved = &t
			} else if v == "false" {
				f := false
				filter.Resolved = &f
			}
		}
		if v := r.URL.Query().Get("from"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				filter.From = &t
			}
		}
		if v := r.URL.Query().Get("to"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				filter.To = &t
			}
		}
		list, total, err := deps.ApplicationErrorLogRepo.List(r.Context(), filter)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		out := make([]map[string]interface{}, 0, len(list))
		for i := range list {
			out = append(out, errorLogToMap(list[i]))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
			"data":       out,
		})
	}
}

// AdminErrorLogsAnalytics GET /api/v1/admin/error-logs/analytics
func AdminErrorLogsAnalytics(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.ApplicationErrorLogRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "error log store unavailable")
			return
		}
		to := time.Now().UTC()
		if v := r.URL.Query().Get("to"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				to = t
			}
		}
		from := to.AddDate(0, 0, -7)
		if v := r.URL.Query().Get("from"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				from = t
			}
		}
		if !from.Before(to) {
			from = to.AddDate(0, 0, -1)
		}
		summary, err := deps.ApplicationErrorLogRepo.AnalyticsSummary(r.Context(), from, to)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"from":    from.Format(time.RFC3339),
			"to":      to.Format(time.RFC3339),
			"summary": summary,
		})
	}
}

// AdminErrorLogGet GET /api/v1/admin/error-logs/{id}
func AdminErrorLogGet(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.ApplicationErrorLogRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "error log store unavailable")
			return
		}
		id := chi.URLParam(r, "id")
		e, err := deps.ApplicationErrorLogRepo.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusNotFound, "not_found", "error log not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(errorLogToMap(e))
	}
}

// AdminErrorLogPatch PATCH /api/v1/admin/error-logs/{id}
func AdminErrorLogPatch(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.ApplicationErrorLogRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "error log store unavailable")
			return
		}
		adminID, ok := middleware.GetUserID(r.Context())
		if !ok || adminID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not logged in")
			return
		}
		id := chi.URLParam(r, "id")
		var req dto.ErrorLogPatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "invalid body")
			return
		}
		if req.Resolved == nil && req.AdminNote == nil {
			writeError(w, http.StatusBadRequest, "validation_error", "provide resolved and/or adminNote")
			return
		}
		ctx := r.Context()
		if req.Resolved != nil {
			if *req.Resolved {
				if err := deps.ApplicationErrorLogRepo.MarkResolved(ctx, id, adminID, req.AdminNote); err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						writeError(w, http.StatusNotFound, "not_found", "not found or already resolved")
						return
					}
					writeError(w, http.StatusInternalServerError, "server_error", err.Error())
					return
				}
			} else {
				if err := deps.ApplicationErrorLogRepo.MarkUnresolved(ctx, id); err != nil {
					if errors.Is(err, pgx.ErrNoRows) {
						writeError(w, http.StatusNotFound, "not_found", "error log not found")
						return
					}
					writeError(w, http.StatusInternalServerError, "server_error", err.Error())
					return
				}
			}
		}
		if req.Resolved == nil && req.AdminNote != nil {
			if err := deps.ApplicationErrorLogRepo.UpdateAdminNote(ctx, id, *req.AdminNote); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					writeError(w, http.StatusNotFound, "not_found", "error log not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "server_error", err.Error())
				return
			}
		}
		e, err := deps.ApplicationErrorLogRepo.GetByID(ctx, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(errorLogToMap(e))
	}
}
