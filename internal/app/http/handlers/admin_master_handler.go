package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/cache"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var slugClean = regexp.MustCompile(`[^a-z0-9-]+`)

func slugFromName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	s = slugClean.ReplaceAllString(s, "")
	return s
}

func ensureSlug(slug, name string) string {
	if strings.TrimSpace(slug) != "" {
		return strings.ToLower(strings.TrimSpace(slug))
	}
	return slugFromName(name)
}

// --- Roles ---
func AdminListRoles(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.RoleRepo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if list == nil {
			list = []domain.Role{}
		}
		out := make([]dto.RoleResponse, len(list))
		for i := range list {
			out[i] = roleToResp(list[i])
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(out)
	}
}
func AdminGetRole(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.RoleRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "role not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(roleToResp(e))
	}
}
func AdminCreateRole(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.RoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		slug := ensureSlug(req.Slug, req.Name)
		e := domain.Role{Name: req.Name, Slug: slug, Description: req.Description, IconURL: req.IconURL}
		if req.UserRoleCode != nil {
			e.UserRoleCode = strings.TrimSpace(*req.UserRoleCode)
		}
		created, err := deps.RoleRepo.Create(r.Context(), e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(roleToResp(created))
	}
}
func AdminUpdateRole(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.RoleRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "role not found", http.StatusNotFound)
			return
		}
		var req dto.RoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name != "" {
			e.Name = req.Name
		}
		if req.Slug != "" {
			e.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		}
		if req.Description != nil {
			e.Description = req.Description
		}
		if req.IconURL != nil {
			e.IconURL = req.IconURL
		}
		if req.UserRoleCode != nil {
			e.UserRoleCode = strings.TrimSpace(*req.UserRoleCode)
		}
		if err := deps.RoleRepo.Update(r.Context(), e); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
func AdminDeleteRole(deps *Deps) http.HandlerFunc {
	return deleteMaster(deps.RoleRepo.Delete)
}
func roleToResp(e domain.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID: e.ID, Name: e.Name, Slug: e.Slug, UserRoleCode: e.UserRoleCode,
		Description: e.Description, IconURL: e.IconURL,
		CreatedAt: e.CreatedAt.Format(time.RFC3339), UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}

// --- Schools ---
func AdminListSchools(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.SchoolRepo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.SchoolResponse, len(list))
		for i := range list {
			out[i] = schoolToResp(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
func AdminGetSchool(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.SchoolRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "school not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(schoolToResp(e))
	}
}
func AdminCreateSchool(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.SchoolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		slug := ensureSlug(req.Slug, req.Name)
		e := domain.School{Name: req.Name, Slug: slug, Description: req.Description, Address: req.Address, LogoURL: req.LogoURL}
		created, err := deps.SchoolRepo.Create(r.Context(), e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cache.InvalidateSchoolList(r.Context(), deps.Redis)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(schoolToResp(created))
	}
}
func AdminUpdateSchool(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.SchoolRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "school not found", http.StatusNotFound)
			return
		}
		var req dto.SchoolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name != "" {
			e.Name = req.Name
		}
		if req.Slug != "" {
			e.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		}
		if req.Description != nil {
			e.Description = req.Description
		}
		if req.Address != nil {
			e.Address = req.Address
		}
		if req.LogoURL != nil {
			e.LogoURL = req.LogoURL
		}
		if err := deps.SchoolRepo.Update(r.Context(), e); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cache.InvalidateSchoolList(r.Context(), deps.Redis)
		w.WriteHeader(http.StatusOK)
	}
}
func AdminDeleteSchool(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := deps.SchoolRepo.Delete(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cache.InvalidateSchoolList(r.Context(), deps.Redis)
		w.WriteHeader(http.StatusNoContent)
	}
}
func schoolToResp(e domain.School) dto.SchoolResponse {
	return dto.SchoolResponse{
		ID: e.ID, Name: e.Name, Slug: e.Slug,
		Description: e.Description, Address: e.Address, LogoURL: e.LogoURL,
		CreatedAt: e.CreatedAt.Format(time.RFC3339), UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}

// --- Settings ---
func AdminListSettings(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.SettingRepo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.SettingResponse, len(list))
		for i := range list {
			out[i] = settingToResp(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
func AdminGetSetting(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.SettingRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "setting not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(settingToResp(e))
	}
}
func AdminCreateSetting(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.SettingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Key == "" {
			http.Error(w, "key required", http.StatusBadRequest)
			return
		}
		slug := ensureSlug(req.Slug, req.Key)
		var valueJSON json.RawMessage
		if req.ValueJSON != nil {
			valueJSON, _ = json.Marshal(req.ValueJSON)
		}
		e := domain.Setting{Key: req.Key, Slug: slug, Value: req.Value, ValueJSON: valueJSON, Description: req.Description}
		created, err := deps.SettingRepo.Create(r.Context(), e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(settingToResp(created))
	}
}
func AdminUpdateSetting(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.SettingRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "setting not found", http.StatusNotFound)
			return
		}
		var req dto.SettingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Key != "" {
			e.Key = req.Key
		}
		if req.Slug != "" {
			e.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		}
		if req.Value != nil {
			e.Value = req.Value
		}
		if req.ValueJSON != nil {
			e.ValueJSON, _ = json.Marshal(req.ValueJSON)
		}
		if req.Description != nil {
			e.Description = req.Description
		}
		if err := deps.SettingRepo.Update(r.Context(), e); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
func AdminDeleteSetting(deps *Deps) http.HandlerFunc {
	return deleteMaster(deps.SettingRepo.Delete)
}
func settingToResp(e domain.Setting) dto.SettingResponse {
	var valueJSON interface{}
	if len(e.ValueJSON) > 0 {
		_ = json.Unmarshal(e.ValueJSON, &valueJSON)
	}
	return dto.SettingResponse{
		ID: e.ID, Key: e.Key, Slug: e.Slug,
		Value: e.Value, ValueJSON: valueJSON, Description: e.Description,
		CreatedAt: e.CreatedAt.Format(time.RFC3339), UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}

// --- Events ---
func AdminListEvents(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.EventRepo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.EventResponse, len(list))
		for i := range list {
			out[i] = eventToResp(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
func AdminGetEvent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.EventRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(eventToResp(e))
	}
}
func AdminCreateEvent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.EventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Title == "" {
			http.Error(w, "title required", http.StatusBadRequest)
			return
		}
		startAt, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			http.Error(w, "start_at required (RFC3339)", http.StatusBadRequest)
			return
		}
		endAt, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			http.Error(w, "end_at required (RFC3339)", http.StatusBadRequest)
			return
		}
		slug := ensureSlug(req.Slug, req.Title)
		status := req.Status
		if status == "" {
			status = "draft"
		}
		e := domain.Event{Title: req.Title, Slug: slug, Description: req.Description, StartAt: startAt, EndAt: endAt, ThumbnailURL: req.ThumbnailURL, Status: status}
		created, err := deps.EventRepo.Create(r.Context(), e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(eventToResp(created))
	}
}
func AdminUpdateEvent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.EventRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		var req dto.EventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Title != "" {
			e.Title = req.Title
		}
		if req.Slug != "" {
			e.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		}
		if req.Description != nil {
			e.Description = req.Description
		}
		if req.StartAt != "" {
			e.StartAt, _ = time.Parse(time.RFC3339, req.StartAt)
		}
		if req.EndAt != "" {
			e.EndAt, _ = time.Parse(time.RFC3339, req.EndAt)
		}
		if req.ThumbnailURL != nil {
			e.ThumbnailURL = req.ThumbnailURL
		}
		if req.Status != "" {
			e.Status = req.Status
		}
		if err := deps.EventRepo.Update(r.Context(), e); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
func AdminDeleteEvent(deps *Deps) http.HandlerFunc {
	return deleteMaster(deps.EventRepo.Delete)
}
func eventToResp(e domain.Event) dto.EventResponse {
	return dto.EventResponse{
		ID: e.ID, Title: e.Title, Slug: e.Slug,
		Description: e.Description, StartAt: e.StartAt.Format(time.RFC3339), EndAt: e.EndAt.Format(time.RFC3339),
		ThumbnailURL: e.ThumbnailURL, Status: e.Status,
		CreatedAt: e.CreatedAt.Format(time.RFC3339), UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}

// --- Subjects ---
func AdminListSubjects(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.SubjectRepo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.SubjectResponse, len(list))
		for i := range list {
			out[i] = subjectToResp(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
func AdminGetSubject(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.SubjectRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "subject not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(subjectToResp(e))
	}
}
func AdminCreateSubject(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.SubjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		slug := ensureSlug(req.Slug, req.Name)
		e := domain.Subject{Name: req.Name, Slug: slug, Description: req.Description, IconURL: req.IconURL, SortOrder: req.SortOrder}
		created, err := deps.SubjectRepo.Create(r.Context(), e)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				http.Error(w, "subject with this slug already exists", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(subjectToResp(created))
	}
}
func AdminUpdateSubject(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.SubjectRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "subject not found", http.StatusNotFound)
			return
		}
		var req dto.SubjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name != "" {
			e.Name = req.Name
		}
		if req.Slug != "" {
			e.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		}
		if req.Description != nil {
			e.Description = req.Description
		}
		if req.IconURL != nil {
			e.IconURL = req.IconURL
		}
		e.SortOrder = req.SortOrder
		if err := deps.SubjectRepo.Update(r.Context(), e); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
func AdminDeleteSubject(deps *Deps) http.HandlerFunc {
	return deleteMaster(deps.SubjectRepo.Delete)
}
func subjectToResp(e domain.Subject) dto.SubjectResponse {
	return dto.SubjectResponse{
		ID: e.ID, Name: e.Name, Slug: e.Slug,
		Description: e.Description, IconURL: e.IconURL, SortOrder: e.SortOrder,
		CreatedAt: e.CreatedAt.Format(time.RFC3339), UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}

// --- Levels (jenjang pendidikan: SD, SMP, SMA) ---
func AdminListLevels(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.LevelRepo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if list == nil {
			list = []domain.Level{}
		}
		out := make([]dto.LevelResponse, len(list))
		for i := range list {
			out[i] = levelToResp(list[i])
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(out)
	}
}
func AdminGetLevel(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.LevelRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "level not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(levelToResp(e))
	}
}
func AdminCreateLevel(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.LevelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		slug := ensureSlug(req.Slug, req.Name)
		e := domain.Level{Name: req.Name, Slug: slug, Description: req.Description, SortOrder: req.SortOrder, IconURL: req.IconURL}
		created, err := deps.LevelRepo.Create(r.Context(), e)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				http.Error(w, "level with this slug already exists", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(levelToResp(created))
	}
}
func AdminUpdateLevel(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		e, err := deps.LevelRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "level not found", http.StatusNotFound)
			return
		}
		var req dto.LevelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name != "" {
			e.Name = req.Name
		}
		if req.Slug != "" {
			e.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
		}
		if req.Description != nil {
			e.Description = req.Description
		}
		e.SortOrder = req.SortOrder
		if req.IconURL != nil {
			e.IconURL = req.IconURL
		}
		if err := deps.LevelRepo.Update(r.Context(), e); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
func AdminDeleteLevel(deps *Deps) http.HandlerFunc {
	return deleteMaster(deps.LevelRepo.Delete)
}
func levelToResp(e domain.Level) dto.LevelResponse {
	return dto.LevelResponse{
		ID: e.ID, Name: e.Name, Slug: e.Slug,
		Description: e.Description, SortOrder: e.SortOrder, IconURL: e.IconURL,
		CreatedAt: e.CreatedAt.Format(time.RFC3339), UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}

// LevelWithSubjects returns level dengan daftar bidang/mata pelajaran (public atau admin).
func LevelWithSubjects(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "levelId")
		if id == "" {
			id = chi.URLParam(r, "id")
		}
		level, err := deps.LevelRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "level not found", http.StatusNotFound)
			return
		}
		subjectIDs, err := deps.LevelRepo.ListSubjectIDsByLevel(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var subjects []dto.SubjectResponse
		for _, sid := range subjectIDs {
			s, err := deps.SubjectRepo.GetByID(r.Context(), sid)
			if err != nil {
				continue
			}
			subjects = append(subjects, subjectToResp(s))
		}
		resp := dto.LevelWithSubjectsResponse{
			LevelResponse: levelToResp(level),
			Subjects:      subjects,
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func deleteMaster(del func(context.Context, string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := del(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
