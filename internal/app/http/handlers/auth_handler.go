package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

func AuthRegister(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Password == "" || req.Name == "" {
			http.Error(w, "name, email, password required", http.StatusBadRequest)
			return
		}
		role := strings.TrimSpace(strings.ToLower(req.Role))
		if role != "" && !isValidRegisterRole(role) {
			http.Error(w, "role must be student, instructor, or guru", http.StatusBadRequest)
			return
		}
		u, token, err := deps.AuthService.Register(r.Context(), req.Name, req.Email, req.Password, role)
		if err != nil {
			if err == service.ErrEmailExists {
				http.Error(w, "email already registered", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Auto-daftarkan siswa ke semua tryout yang akan datang (bukan draft)
		if u.Role == domain.UserRoleStudent {
			_ = deps.TryoutRegistrationRepo.EnsureStudentForAllOpenTryouts(r.Context(), u.ID)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.AuthResponse{
			User:  userToMap(u),
			Token: token,
		})
	}
}

func AuthLogin(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Password == "" {
			http.Error(w, "email and password required", http.StatusBadRequest)
			return
		}
		u, token, err := deps.AuthService.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if err == service.ErrInvalidCreds {
				http.Error(w, "invalid email or password", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.AuthResponse{
			User:  userToMap(u),
			Token: token,
		})
	}
}

func AuthLogout(_ *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

// AuthMe returns current user from JWT. GET /api/v1/auth/me
func AuthMe(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not logged in")
			return
		}
		u, err := deps.UserRepo.FindByID(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		role := u.Role
		if role == "guru" {
			role = "instructor"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.AuthUserResponse{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Role:  role,
		})
	}
}

func AuthForgotPassword(_ *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Stub: always return ok (jangan bocorkan apakah email ada)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func AuthResetPassword(_ *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Stub: terima token + new_password, return ok
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func userToMap(u domain.User) map[string]interface{} {
	m := map[string]interface{}{
		"id":    u.ID,
		"name":  u.Name,
		"email": u.Email,
		"role":  u.Role,
	}
	if u.AvatarURL != nil {
		m["avatar_url"] = *u.AvatarURL
	}
	return m
}

// isValidRegisterRole: student, instructor, atau guru.
func isValidRegisterRole(role string) bool {
	return role == domain.UserRoleStudent || role == "siswa" || role == domain.UserRoleGuru || role == "instructor"
}
