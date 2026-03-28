package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/db"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// registerRequiresPhoneOrWhatsapp: siswa & guru (termasuk enum legacy instructor) wajib punya kontak.
func registerRequiresPhoneOrWhatsapp(roleCode string) bool {
	c := strings.TrimSpace(strings.ToLower(roleCode))
	return c == domain.UserRoleStudent || c == domain.UserRoleGuru || c == "instructor"
}

func AuthRegister(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "invalid body")
			return
		}
		if req.Email == "" || req.Password == "" || req.Name == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "name, email, password required")
			return
		}
		if deps.RoleRepo == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "roles store unavailable")
			return
		}
		ctx := r.Context()
		var roleCode string
		slug := strings.TrimSpace(req.Role)
		var err error
		if slug == "" {
			roleCode, err = defaultUserRoleCode(ctx, deps.RoleRepo)
			if errors.Is(err, errRolesTableEmpty) {
				roleCode, err = domain.UserRoleStudent, nil
			}
		} else {
			roleCode, err = resolveUserRoleCodeForUserTable(ctx, deps.RoleRepo, slug)
			if errors.Is(err, errUnknownRoleSlug) {
				if fb, ok := registerFallbackUserRoleCode(slug); ok {
					roleCode, err = fb, nil
				}
			}
		}
		if err != nil {
			if errors.Is(err, errUnknownRoleSlug) {
				writeError(w, http.StatusBadRequest, "validation_error", "unknown role: use a value from table roles or student/guru/trainer (instructor accepted as alias for guru)")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		phoneTrim := strings.TrimSpace(req.Phone)
		waTrim := strings.TrimSpace(req.Whatsapp)
		if registerRequiresPhoneOrWhatsapp(roleCode) && phoneTrim == "" && waTrim == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "phone or whatsapp is required for student and guru registration")
			return
		}
		var phonePtr, waPtr *string
		if phoneTrim != "" {
			phonePtr = &phoneTrim
		}
		if waTrim != "" {
			waPtr = &waTrim
		}
		u, token, err := deps.AuthService.Register(ctx, req.Name, req.Email, req.Password, roleCode, phonePtr, waPtr)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		// Auto-daftarkan siswa ke semua tryout yang akan datang (bukan draft)
		if u.Role == domain.UserRoleStudent && deps.TryoutRegistrationRepo != nil {
			_ = deps.TryoutRegistrationRepo.EnsureStudentForAllOpenTryouts(r.Context(), u.ID)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := dto.AuthResponse{
			User:            userAuthMap(r.Context(), deps.RoleRepo, u),
			Token:           token,
			MustSetPassword: u.MustSetPassword,
		}
		if u.MustSetPassword {
			resp.NextAction = "SET_PASSWORD"
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// AuthRegisterWithInvite: POST /api/v1/auth/register-with-invite
// Body: { "token", "email", "name", "password" } — token dari link email checkout.
func AuthRegisterWithInvite(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Token    string `json:"token"`
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Token == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "token, email, and password required", http.StatusBadRequest)
			return
		}
		u, token, err := deps.AuthService.RegisterWithInvite(r.Context(), req.Token, req.Email, req.Name, req.Password)
		if err != nil {
			if err == service.ErrInviteInvalid {
				writeError(w, http.StatusBadRequest, "invalid_invite", "Token tidak valid atau sudah kadaluarsa")
				return
			}
			if err == service.ErrInviteAlreadyUsed {
				writeError(w, http.StatusBadRequest, "invite_used", "Link registrasi sudah digunakan")
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := dto.AuthResponse{
			User:            userAuthMap(r.Context(), deps.RoleRepo, u),
			Token:           token,
			MustSetPassword: u.MustSetPassword,
		}
		if u.MustSetPassword {
			resp.NextAction = "SET_PASSWORD"
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func AuthLogin(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "invalid body")
			return
		}
		if req.Email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "email and password required")
			return
		}
		u, token, err := deps.AuthService.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if err == service.ErrInvalidCreds {
				writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := dto.AuthResponse{
			User:            userAuthMap(r.Context(), deps.RoleRepo, u),
			Token:           token,
			MustSetPassword: u.MustSetPassword,
		}
		if u.MustSetPassword {
			resp.NextAction = "SET_PASSWORD"
		}
		_ = json.NewEncoder(w).Encode(resp)
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
	return UserProfileGet(deps)
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

// AuthAdminPasswordBypass resets admin password using emergency bypass key.
// Endpoint: POST /api/v1/auth/admin/password-bypass
// Header: X-Admin-Bypass-Key: <ADMIN_PASSWORD_BYPASS_KEY>
// Body: { "email": "...", "newPassword": "..." }
func AuthAdminPasswordBypass(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bypassKey := strings.TrimSpace(deps.AdminPasswordBypassKey)
		if bypassKey == "" {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "admin password bypass is disabled")
			return
		}
		reqKey := strings.TrimSpace(r.Header.Get("X-Admin-Bypass-Key"))
		if reqKey == "" || reqKey != bypassKey {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid bypass key")
			return
		}
		var req struct {
			Email       string `json:"email"`
			NewPassword string `json:"newPassword"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "invalid body")
			return
		}
		email := strings.TrimSpace(strings.ToLower(req.Email))
		if email == "" || len(req.NewPassword) < 6 {
			writeError(w, http.StatusBadRequest, "validation_error", "email and newPassword (min 6 chars) required")
			return
		}
		u, err := deps.UserRepo.FindByEmail(r.Context(), email)
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "admin user not found")
			return
		}
		if !isAdminRoleForBypass(u.Role) {
			writeError(w, http.StatusForbidden, "forbidden", "target user is not admin")
			return
		}
		if err := deps.AuthService.SetPassword(r.Context(), u.ID, req.NewPassword); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "admin password updated",
		})
	}
}

// AuthRunMigrateBypass runs embedded DB migrations with bypass key.
// Endpoint: POST /api/v1/auth/admin/run-migrate
// Header: X-Migrate-Bypass-Key: <MIGRATE_BYPASS_KEY>
func AuthRunMigrateBypass(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bypassKey := strings.TrimSpace(deps.MigrateBypassKey)
		if bypassKey == "" {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "migrate bypass is disabled")
			return
		}
		reqKey := strings.TrimSpace(r.Header.Get("X-Migrate-Bypass-Key"))
		if reqKey == "" || reqKey != bypassKey {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid bypass key")
			return
		}
		if deps.DB == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "database is not configured")
			return
		}
		applied, err := db.Migrate(r.Context(), deps.DB)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "migrations completed",
			"migrations": applied,
			"count":      len(applied),
		})
	}
}

// AdminGeneratePasswordHash returns bcrypt hash for a raw password.
// Endpoint: POST /api/v1/admin/tools/hash-password (admin + permission protected)
func AdminGeneratePasswordHash(_ *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.PasswordHashRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "invalid body")
			return
		}
		if strings.TrimSpace(req.Password) == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "password is required")
			return
		}
		hash, err := service.GeneratePasswordHash(req.Password)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.PasswordHashResponse{
			Algorithm: "bcrypt",
			Hash:      hash,
		})
	}
}

// AuthVerifyEmail verifies email using token from JSON body: POST /api/v1/auth/verify-email
func AuthVerifyEmail(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Token) == "" {
			http.Error(w, "token required", http.StatusBadRequest)
			return
		}
		if err := deps.AuthService.VerifyEmail(r.Context(), req.Token); err != nil {
			http.Error(w, "verification_link_expired_or_invalid", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "email verified, you can now login",
		})
	}
}

// AuthResendVerification handles resend verification link: POST /api/v1/auth/resend-verification
func AuthResendVerification(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		email := strings.TrimSpace(req.Email)
		if email == "" {
			http.Error(w, "email required", http.StatusBadRequest)
			return
		}
		err := deps.AuthService.ResendVerification(r.Context(), email)
		if err == service.ErrAlreadyVerified {
			http.Error(w, "already_verified", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "verification_link_sent",
		})
	}
}

func isAdminRoleForBypass(role string) bool {
	switch role {
	case domain.UserRoleAdmin,
		domain.UserRoleSuperAdmin,
		domain.UserRoleFinanceAdmin,
		domain.UserRoleAcademicAdmin,
		domain.UserRoleContentAdmin:
		return true
	default:
		return false
	}
}

// AuthSetPassword handles setting password for users with mustSetPassword=true
// POST /api/v1/auth/set-password (Bearer required)
func AuthSetPassword(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not logged in")
			return
		}

		var req dto.SetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		if len(req.NewPassword) < 6 {
			writeError(w, http.StatusBadRequest, "bad_request", "password must be at least 6 characters")
			return
		}

		if err := deps.AuthService.SetPassword(r.Context(), userID, req.NewPassword); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.SetPasswordResponse{
			Message:         "Password berhasil diatur",
			MustSetPassword: false,
		})
	}
}
