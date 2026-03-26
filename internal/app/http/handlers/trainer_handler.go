package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

var trainerSchoolSlugClean = regexp.MustCompile(`[^a-z0-9-]+`)

func slugFromSchoolName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	s = trainerSchoolSlugClean.ReplaceAllString(s, "")
	return s
}

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
		resp.Phone = u.Phone
		resp.Whatsapp = u.Whatsapp
		resp.ClassLevel = u.ClassLevel
		resp.City = u.City
		resp.Province = u.Province
		resp.Gender = u.Gender
		if u.BirthDate != nil {
			s := u.BirthDate.UTC().Format("2006-01-02")
			resp.BirthDate = &s
		}
		resp.Bio = u.Bio
		resp.ParentName = u.ParentName
		resp.ParentPhone = u.ParentPhone
		resp.Instagram = u.Instagram
		resp.SchoolID = u.SchoolID
		resp.SubjectID = u.SubjectID
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
		if req.Email != "" {
			existing, err := deps.UserRepo.FindByEmail(r.Context(), req.Email)
			if err == nil && existing.ID != u.ID {
				http.Error(w, "email already in use", http.StatusConflict)
				return
			}
			u.Email = req.Email
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
		if req.SchoolName != nil {
			schoolName := strings.TrimSpace(*req.SchoolName)
			if schoolName != "" {
				slug := slugFromSchoolName(schoolName)
				if slug == "" {
					http.Error(w, "invalid school_name", http.StatusBadRequest)
					return
				}
				school, err := deps.SchoolRepo.GetBySlug(r.Context(), slug)
				if err != nil {
					created, createErr := deps.SchoolRepo.Create(r.Context(), domain.School{
						Name: schoolName,
						Slug: slug,
					})
					if createErr == nil {
						school = created
					} else {
						// kemungkinan race slug duplicate: coba get ulang by slug
						existing, getErr := deps.SchoolRepo.GetBySlug(r.Context(), slug)
						if getErr != nil {
							http.Error(w, "failed to link school", http.StatusInternalServerError)
							return
						}
						school = existing
					}
				}
				u.SchoolID = &school.ID
			}
		}
		if req.SubjectID != nil {
			subjectID := strings.TrimSpace(*req.SubjectID)
			if subjectID == "" {
				u.SubjectID = nil
			} else {
				if _, err := deps.SubjectRepo.GetByID(r.Context(), subjectID); err != nil {
					http.Error(w, "subject not found", http.StatusBadRequest)
					return
				}
				u.SubjectID = &subjectID
			}
		}

		if req.Phone != nil {
			u.Phone = req.Phone
		}
		if req.Whatsapp != nil {
			u.Whatsapp = req.Whatsapp
		}
		if req.ClassLevel != nil {
			u.ClassLevel = req.ClassLevel
		}
		if req.City != nil {
			u.City = req.City
		}
		if req.Province != nil {
			u.Province = req.Province
		}
		if req.Gender != nil {
			u.Gender = req.Gender
		}
		if req.BirthDate != nil {
			b := strings.TrimSpace(*req.BirthDate)
			if b == "" {
				u.BirthDate = nil
			} else {
				parsed, err := time.Parse("2006-01-02", b)
				if err != nil {
					http.Error(w, "invalid birthDate format; expected YYYY-MM-DD", http.StatusBadRequest)
					return
				}
				u.BirthDate = &parsed
			}
		}
		if req.Bio != nil {
			u.Bio = req.Bio
		}
		if req.ParentName != nil {
			u.ParentName = req.ParentName
		}
		if req.ParentPhone != nil {
			u.ParentPhone = req.ParentPhone
		}
		if req.Instagram != nil {
			u.Instagram = req.Instagram
		}
		if err := deps.UserRepo.Update(r.Context(), u); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var schoolProfile *dto.SchoolProfile
		if u.SchoolID != nil && *u.SchoolID != "" {
			if school, err := deps.SchoolRepo.GetByID(r.Context(), *u.SchoolID); err == nil {
				schoolProfile = schoolToProfile(school)
			}
		}
		var birthDateStr *string
		if u.BirthDate != nil {
			s := u.BirthDate.UTC().Format("2006-01-02")
			birthDateStr = &s
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.TrainerProfileResponse{
			Name:        u.Name,
			Email:       u.Email,
			Phone:       u.Phone,
			Whatsapp:    u.Whatsapp,
			ClassLevel:  u.ClassLevel,
			City:        u.City,
			Province:    u.Province,
			Gender:      u.Gender,
			BirthDate:  birthDateStr,
			Bio:         u.Bio,
			ParentName: u.ParentName,
			ParentPhone: u.ParentPhone,
			Instagram:  u.Instagram,
			SchoolID:   u.SchoolID,
			SubjectID:  u.SubjectID,
			School:      schoolProfile,
		})
	}
}

// InstructorProfilePassword changes password for instructor.
// PUT /api/v1/instructor/profile/password
func InstructorProfilePassword(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
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
			writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "password updated"})
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
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		_, _, students, err := deps.TrainerService.Status(r.Context(), userID, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		studentID := chi.URLParam(r, "studentId")
		if studentID == "" {
			http.Error(w, "studentId required", http.StatusBadRequest)
			return
		}
		_, _, students, err := deps.TrainerService.Status(r.Context(), userID, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "student not found", http.StatusNotFound)
	}
}

// TrainerStudentUpdate updates a student's name/email.
// PUT /api/v1/trainer/students/:studentId
func TrainerStudentUpdate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		studentID := chi.URLParam(r, "studentId")
		if studentID == "" {
			http.Error(w, "studentId required", http.StatusBadRequest)
			return
		}
		var req dto.TrainerStudentUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// Ensure student belongs to this trainer by checking in the list.
		_, _, students, err := deps.TrainerService.Status(r.Context(), userID, true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}

		u, err := deps.UserRepo.FindByID(r.Context(), studentID)
		if err != nil {
			http.Error(w, "student not found", http.StatusNotFound)
			return
		}

		if req.Name != "" {
			u.Name = req.Name
		}
		if req.Email != "" {
			existing, err := deps.UserRepo.FindByEmail(r.Context(), req.Email)
			if err == nil && existing.ID != u.ID {
				http.Error(w, "email already in use", http.StatusConflict)
				return
			}
			u.Email = req.Email
		}

		if err := deps.UserRepo.Update(r.Context(), u); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
