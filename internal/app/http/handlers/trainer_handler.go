package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// TrainerProfileGet returns guru profile: name, email, school.
// Data sekolah yang terhubung dengan akun guru; jika user punya school_id dan sekolah ada, objek school diisi.
// GET /api/v1/trainer/profile
func TrainerProfileGet(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		u, err := deps.UserRepo.FindByID(r.Context(), userID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		resp := dto.TrainerProfileResponse{
			Name:  u.Name,
			Email: u.Email,
		}
		// Objek school diisi jika guru terhubung ke sekolah (school_id); frontend pakai untuk "Detail info sekolah".
		if u.SchoolID != nil && *u.SchoolID != "" {
			school, err := deps.SchoolRepo.GetByID(r.Context(), *u.SchoolID)
			if err == nil {
				resp.School = schoolToProfile(school)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// TrainerProfileUpdate updates guru profile (name, school_id). PUT /api/v1/trainer/profile
func TrainerProfileUpdate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var req dto.TrainerProfileUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		u, err := deps.UserRepo.FindByID(r.Context(), userID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if req.Name != "" {
			u.Name = req.Name
		}
		if req.SchoolID != nil {
			sid := strings.TrimSpace(*req.SchoolID)
			if sid == "" {
				u.SchoolID = nil
			} else {
				if _, err := deps.SchoolRepo.GetByID(r.Context(), sid); err != nil {
					http.Error(w, "school not found", http.StatusBadRequest)
					return
				}
				u.SchoolID = &sid
			}
		}
		if err := deps.UserRepo.Update(r.Context(), u); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "profile updated"})
	}
}

// schoolToProfile maps domain.School to frontend-normalized SchoolProfile (nama_sekolah, alamat, dll).
func schoolToProfile(s domain.School) *dto.SchoolProfile {
	p := &dto.SchoolProfile{
		ID:            s.ID,
		NamaSekolah:   s.Name,
		NPSN:          "",
		KabupatenKota: "",
		Alamat:        "",
		Telepon:       "",
	}
	if s.Address != nil {
		p.Alamat = *s.Address
	}
	return p
}

// TrainerStatus returns paid_slots, registered_students_count, and optional students. GET /api/v1/trainer/status
func TrainerStatus(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		includeStudents := r.URL.Query().Get("students") != ""
		paidSlots, registeredCount, students, err := deps.TrainerService.Status(r.Context(), userID, includeStudents)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var req dto.TrainerPayRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Quantity <= 0 {
			http.Error(w, "quantity must be positive", http.StatusBadRequest)
			return
		}
		if err := deps.TrainerService.Pay(r.Context(), userID, req.Quantity); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var req dto.TrainerCreateStudentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Password == "" || req.Name == "" {
			http.Error(w, "name, email, password required", http.StatusBadRequest)
			return
		}
		u, err := deps.TrainerService.CreateStudent(r.Context(), userID, req.Name, req.Email, req.Password)
		if err != nil {
			if err == service.ErrNoSlotsAvailable {
				http.Error(w, "no paid slots available to register more students", http.StatusForbidden)
				return
			}
			if err == service.ErrEmailExists {
				http.Error(w, "email already registered", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Auto-daftarkan siswa ke tryout open
		_ = deps.TryoutRegistrationRepo.EnsureStudentForAllOpenTryouts(r.Context(), u.ID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"user": userToMap(u),
		})
	}
}

// TrainerTryoutList returns list of tryouts for trainer/guru (untuk pilih tryout sebelum lihat analisis).
// GET /api/v1/trainer/tryouts
func TrainerTryoutList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.AdminService.ListTryouts(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}
}

// TrainerAttemptAIAnalysis exposes AI-based analysis per attempt for trainers/guru.
// GET /api/v1/trainer/tryouts/{tryoutId}/attempts/{attemptId}/ai-analysis
func TrainerAttemptAIAnalysis(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		analysis, err := deps.AdminService.GetAttemptAIAnalysis(r.Context(), tryoutID, attemptID)
		if err != nil {
			if err == service.ErrNotFound {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(analysis)
	}
}

func userToUserInfo(u domain.User) dto.UserInfo {
	return dto.UserInfo{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role}
}
