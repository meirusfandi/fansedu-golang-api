package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func AdminAuditLogsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		rows, err := deps.DB.Query(r.Context(), `
			SELECT id, admin_user_id::text, role, method, path, status_code, duration_ms, request_id, created_at
			FROM admin_audit_logs
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		defer rows.Close()
		out := make([]map[string]interface{}, 0, limit)
		for rows.Next() {
			var id, adminUserID, role, method, path string
			var statusCode, durationMs int
			var requestID *string
			var createdAt time.Time
			if err := rows.Scan(&id, &adminUserID, &role, &method, &path, &statusCode, &durationMs, &requestID, &createdAt); err != nil {
				writeError(w, http.StatusInternalServerError, "server_error", err.Error())
				return
			}
			out = append(out, map[string]interface{}{
				"id":            id,
				"admin_user_id": adminUserID,
				"role":          role,
				"method":        method,
				"path":          path,
				"status_code":   statusCode,
				"duration_ms":   durationMs,
				"request_id":    requestID,
				"created_at":    createdAt.Format(time.RFC3339),
			})
		}
		if err := rows.Err(); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"page":  page,
			"limit": limit,
			"data":  out,
		})
	}
}
