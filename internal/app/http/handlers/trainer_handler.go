package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// TrainerProfileGet returns profile (sama bentuknya dengan /student/profile dan /auth/me).
// GET /api/v1/trainer/profile
func TrainerProfileGet(deps *Deps) http.HandlerFunc {
	return UserProfileGet(deps)
}

// TrainerProfileUpdate updates guru profile (name, school_id). PUT /api/v1/trainer/profile
func TrainerProfileUpdate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		var req dto.TrainerProfileUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		u, err := deps.UserRepo.FindByID(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		if err := ApplyUserProfileUpdate(r.Context(), deps, &u, &req); err != nil {
			writeErrorFromProfileApply(w, r, err)
			return
		}
		if err := deps.UserRepo.Update(r.Context(), u); err != nil {
			writeErrorFromUserRepoUpdate(w, r, err)
			return
		}
		u2, school, err := deps.UserRepo.FindByIDProfileWithSchool(r.Context(), userID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BuildUserProfileResponse(r.Context(), deps, u2, school))
	}
}

// GuruProfilePassword changes password for guru/teaching staff.
// PUT /api/v1/guru/profile/password
func GuruProfilePassword(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}

		var req struct {
			CurrentPassword string `json:"currentPassword"`
			NewPassword     string `json:"newPassword"`
			ConfirmPassword string `json:"confirmPassword"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "validation_error", "invalid body")
			return
		}
		if req.CurrentPassword == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
			writeError(w, http.StatusBadRequest, "validation_error", "currentPassword/newPassword/confirmPassword required")
			return
		}
		if len(req.NewPassword) < 6 {
			writeError(w, http.StatusBadRequest, "validation_error", "newPassword must be at least 6 characters")
			return
		}
		if req.NewPassword != req.ConfirmPassword {
			writeError(w, http.StatusBadRequest, "validation_error", "confirmPassword does not match")
			return
		}

		if err := deps.AuthService.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			if err == service.ErrInvalidCreds {
				writeError(w, http.StatusUnauthorized, "invalid_current_password", "current password is incorrect")
				return
			}
			writeInternalError(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "password updated"})
	}
}

// TrainerStatus returns paid_slots, registered_students_count, and optional students. GET /api/v1/trainer/status
func TrainerStatus(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		includeStudents := r.URL.Query().Get("students") != ""
		paidSlots, registeredCount, students, err := deps.TrainerService.Status(r.Context(), userID, includeStudents)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		resp := dto.TrainerStatusResponse{
			PaidSlots:              paidSlots,
			RegisteredStudentsCount: registeredCount,
		}
		if includeStudents {
			resp.Students = make([]dto.UserInfo, 0, len(students))
			for _, u := range students {
				resp.Students = append(resp.Students, userToUserInfo(u))
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// TrainerPay adds paid slots after payment confirmation. POST /api/v1/trainer/pay
func TrainerPay(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		var req dto.TrainerPayRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Permintaan tidak valid.")
			return
		}
		if req.Quantity <= 0 {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Jumlah harus positif.")
			return
		}
		if err := deps.TrainerService.Pay(r.Context(), userID, req.Quantity); err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "slots updated"})
	}
}

// TrainerCreateStudent creates a student and links to trainer. POST /api/v1/trainer/students
func TrainerCreateStudent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		var req dto.TrainerCreateStudentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Permintaan tidak valid.")
			return
		}
		if req.Email == "" || req.Password == "" || req.Name == "" {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Nama, email, dan password wajib diisi.")
			return
		}
		u, err := deps.TrainerService.CreateStudent(r.Context(), userID, req.Name, req.Email, req.Password)
		if err != nil {
			if err == service.ErrNoSlotsAvailable {
				writeError(w, http.StatusForbidden, "NO_SLOTS_AVAILABLE", "Tidak ada slot berbayar untuk menambah siswa.")
				return
			}
			if err == service.ErrEmailExists {
				writeError(w, http.StatusConflict, "EMAIL_EXISTS", "Email sudah terdaftar.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"user": userAuthMap(r.Context(), deps.RoleRepo, u),
		})
	}
}

// TrainerStudentsList lists students linked to current trainer/guru.
// GET /api/v1/trainer/students
func TrainerStudentsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		_, _, students, err := deps.TrainerService.Status(r.Context(), userID, true)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		out := make([]dto.TrainerStudentItem, 0, len(students))
		for _, u := range students {
			out = append(out, dto.TrainerStudentItem{
				ID:    u.ID,
				Name:  u.Name,
				Email: u.Email,
				Role:  u.Role,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}
}

// TrainerStudentGet returns detail a single student for current trainer.
// GET /api/v1/trainer/students/:studentId
func TrainerStudentGet(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		studentID := chi.URLParam(r, "studentId")
		if studentID == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "studentId wajib diisi.")
			return
		}
		_, _, students, err := deps.TrainerService.Status(r.Context(), userID, true)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		for _, u := range students {
			if u.ID == studentID {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(dto.TrainerStudentItem{
					ID:    u.ID,
					Name:  u.Name,
					Email: u.Email,
					Role:  u.Role,
				})
				return
			}
		}
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Siswa tidak ditemukan.")
	}
}

// TrainerStudentUpdate updates a student's name/email.
// PUT /api/v1/trainer/students/:studentId
func TrainerStudentUpdate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		studentID := chi.URLParam(r, "studentId")
		if studentID == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "studentId wajib diisi.")
			return
		}
		var req dto.TrainerStudentUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Permintaan tidak valid.")
			return
		}

		// Ensure student belongs to this trainer by checking in the list.
		_, _, students, err := deps.TrainerService.Status(r.Context(), userID, true)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		found := false
		for _, u := range students {
			if u.ID == studentID {
				found = true
				break
			}
		}
		if !found {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Siswa tidak ditemukan.")
			return
		}

		u, err := deps.UserRepo.FindByID(r.Context(), studentID)
		if err != nil {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Siswa tidak ditemukan.")
			return
		}

		if req.Name != "" {
			u.Name = req.Name
		}
		if req.Email != "" {
			existing, err := deps.UserRepo.FindByEmail(r.Context(), req.Email)
			if err == nil && existing.ID != u.ID {
				writeError(w, http.StatusConflict, "EMAIL_EXISTS", "Email sudah digunakan.")
				return
			}
			u.Email = req.Email
		}

		if err := deps.UserRepo.Update(r.Context(), u); err != nil {
			writeErrorFromUserRepoUpdate(w, r, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.TrainerStudentItem{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Role:  u.Role,
		})
	}
}

// TrainerTryoutList returns list of tryouts for trainer/guru (untuk pilih tryout sebelum lihat analisis).
// GET /api/v1/trainer/tryouts
func TrainerTryoutList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.AdminService.ListTryouts(r.Context())
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		out := make([]dto.TryoutResponse, len(list))
		for i := range list {
			out[i] = tryoutToDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// TrainerTryoutAnalysis exposes per-question analysis for trainers/guru using admin service.
// GET /api/v1/trainer/tryouts/{tryoutId}/analysis
func TrainerTryoutAnalysis(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		analysis, err := deps.AdminService.GetTryoutAnalysis(r.Context(), tryoutID)
		if err != nil {
			if err == service.ErrNotFound {
				writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(analysis)
	}
}

// TrainerTryoutStudents exposes per-student list for trainers/guru.
// GET /api/v1/trainer/tryouts/{tryoutId}/students
func TrainerTryoutStudents(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		list, err := deps.AdminService.ListTryoutStudents(r.Context(), tryoutID)
		if err != nil {
			if err == service.ErrNotFound {
				writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}
}

// TrainerAttemptAIAnalysis exposes AI-based analysis per attempt for trainers/guru.
// GET atau POST /api/v1/trainer/tryouts/{tryoutId}/attempts/{attemptId}/ai-analysis
func TrainerAttemptAIAnalysis(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		analysis, err := deps.AdminService.GetAttemptAIAnalysis(r.Context(), tryoutID, attemptID)
		if err != nil {
			if err == service.ErrNotFound {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Data tidak ditemukan.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(analysis)
	}
}

func userToUserInfo(u domain.User) dto.UserInfo {
	return dto.UserInfo{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role}
}
