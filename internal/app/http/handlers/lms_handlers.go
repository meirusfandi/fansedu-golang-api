package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// ListRoles returns all roles (public, for registration dropdown). GET /api/v1/roles
func ListRoles(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.RoleRepo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.RoleItem, len(list))
		for i := range list {
			out[i] = dto.RoleItem{ID: list[i].ID, Name: list[i].Name, Slug: list[i].Slug}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// ListSchools returns all schools (public, for profile dropdown). GET /api/v1/schools
func ListSchools(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.SchoolRepo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.SchoolItem, len(list))
		for i := range list {
			out[i] = dto.SchoolItem{ID: list[i].ID, Name: list[i].Name, Slug: list[i].Slug}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// AuthChangePassword changes password for authenticated user. POST /api/v1/auth/change-password
func AuthChangePassword(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var req struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.NewPassword == "" || len(req.NewPassword) < 6 {
			http.Error(w, "new_password must be at least 6 characters", http.StatusBadRequest)
			return
		}
		if err := deps.AuthService.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
			if err == service.ErrInvalidCreds {
				http.Error(w, "current password is incorrect", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "password updated"})
	}
}

// NotificationsList returns current user's notifications. GET /api/v1/notifications
func NotificationsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 50
		}
		list, err := deps.NotificationRepo.ListByUserID(r.Context(), userID, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.NotificationItem, len(list))
		for i := range list {
			var readAt string
			if list[i].ReadAt != nil {
				readAt = list[i].ReadAt.Format("2006-01-02T15:04:05Z07:00")
			}
			out[i] = dto.NotificationItem{
				ID:        list[i].ID,
				Title:     list[i].Title,
				Body:      list[i].Body,
				Type:      list[i].Type,
				ReadAt:    readAt,
				CreatedAt: list[i].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// NotificationMarkRead marks a notification as read. PATCH /api/v1/notifications/:notificationId/read
func NotificationMarkRead(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "notificationId")
		if err := deps.NotificationRepo.MarkRead(r.Context(), id, userID); err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "marked as read"})
	}
}

// PaymentListMine returns current user's payments. GET /api/v1/payments
func PaymentListMine(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 50
		}
		list, err := deps.PaymentRepo.ListByUserID(r.Context(), userID, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.UserPaymentResponse, len(list))
		for i := range list {
			out[i] = paymentToResponse(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// PaymentCreate creates a payment (e.g. upload proof). POST /api/v1/payments
func PaymentCreate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var req struct {
			Amount      int     `json:"amount"`
			Type        string  `json:"type"`
			ReferenceID *string `json:"reference_id"`
			Description *string `json:"description"`
			ProofURL    *string `json:"proof_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Amount <= 0 {
			http.Error(w, "amount must be positive", http.StatusBadRequest)
			return
		}
		ptype := req.Type
		if ptype == "" {
			ptype = domain.PaymentTypeCoursePurchase
		}
		p := domain.Payment{
			UserID: userID,
			Amount: req.Amount,
			Currency:    "IDR",
			Status:      domain.PaymentStatusPending,
			Type:        ptype,
			ReferenceID: req.ReferenceID,
			Description: req.Description,
			ProofURL:    req.ProofURL,
		}
		created, err := deps.PaymentRepo.Create(r.Context(), p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(paymentToResponse(created))
	}
}

func paymentToResponse(p domain.Payment) dto.UserPaymentResponse {
	var refID, proofURL, paidAt string
	if p.ReferenceID != nil {
		refID = *p.ReferenceID
	}
	if p.ProofURL != nil {
		proofURL = *p.ProofURL
	}
	if p.PaidAt != nil {
		paidAt = p.PaidAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return dto.UserPaymentResponse{
		ID:       p.ID,
		UserID:   p.UserID,
		Amount:   p.Amount,
		Currency:     p.Currency,
		Status:       p.Status,
		Type:         p.Type,
		ReferenceID:  refID,
		Description:  p.Description,
		ProofURL:     proofURL,
		PaidAt:       paidAt,
		CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CourseMessagesList returns chat messages for a course (enrolled users). GET /api/v1/courses/:courseId/messages
func CourseMessagesList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		courseID := chi.URLParam(r, "courseId")
		_, err := deps.EnrollmentRepo.GetByUserAndCourse(r.Context(), userID, courseID)
		if err != nil {
			http.Error(w, "forbidden: not enrolled", http.StatusForbidden)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		list, err := deps.CourseMessageRepo.ListByCourseID(r.Context(), courseID, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.CourseMessageItem, len(list))
		for i := range list {
			out[i] = dto.CourseMessageItem{
				ID:        list[i].ID,
				UserID:    list[i].UserID,
				Message:   list[i].Message,
				CreatedAt: list[i].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// CourseMessageCreate posts a chat message. POST /api/v1/courses/:courseId/messages
func CourseMessageCreate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		courseID := chi.URLParam(r, "courseId")
		_, err := deps.EnrollmentRepo.GetByUserAndCourse(r.Context(), userID, courseID)
		if err != nil {
			http.Error(w, "forbidden: not enrolled", http.StatusForbidden)
			return
		}
		var req struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
			http.Error(w, "message required", http.StatusBadRequest)
			return
		}
		m := domain.CourseMessage{CourseID: courseID, UserID: userID, Message: req.Message}
		created, err := deps.CourseMessageRepo.Create(r.Context(), m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.CourseMessageItem{
			ID:        created.ID,
			UserID:    created.UserID,
			Message:   created.Message,
			CreatedAt: created.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

// CourseDiscussionsList returns discussion threads for a course. GET /api/v1/courses/:courseId/discussions
func CourseDiscussionsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		courseID := chi.URLParam(r, "courseId")
		_, err := deps.EnrollmentRepo.GetByUserAndCourse(r.Context(), userID, courseID)
		if err != nil {
			http.Error(w, "forbidden: not enrolled", http.StatusForbidden)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 50
		}
		list, err := deps.CourseDiscussionRepo.ListByCourseID(r.Context(), courseID, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.DiscussionItem, len(list))
		for i := range list {
			out[i] = dto.DiscussionItem{
				ID:        list[i].ID,
				CourseID:  list[i].CourseID,
				UserID:    list[i].UserID,
				Title:     list[i].Title,
				Body:      list[i].Body,
				CreatedAt: list[i].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// CourseDiscussionCreate creates a discussion thread. POST /api/v1/courses/:courseId/discussions
func CourseDiscussionCreate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		courseID := chi.URLParam(r, "courseId")
		_, err := deps.EnrollmentRepo.GetByUserAndCourse(r.Context(), userID, courseID)
		if err != nil {
			http.Error(w, "forbidden: not enrolled", http.StatusForbidden)
			return
		}
		var req struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Title == "" || req.Body == "" {
			http.Error(w, "title and body required", http.StatusBadRequest)
			return
		}
		d := domain.CourseDiscussion{CourseID: courseID, UserID: userID, Title: req.Title, Body: req.Body}
		created, err := deps.CourseDiscussionRepo.Create(r.Context(), d)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.DiscussionItem{
			ID:        created.ID,
			CourseID:  created.CourseID,
			UserID:    created.UserID,
			Title:     created.Title,
			Body:      created.Body,
			CreatedAt: created.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

// DiscussionGet returns one discussion. GET /api/v1/discussions/:id
func DiscussionGet(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "discussionId")
		d, err := deps.CourseDiscussionRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, err = deps.EnrollmentRepo.GetByUserAndCourse(r.Context(), userID, d.CourseID)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.DiscussionItem{
			ID:        d.ID,
			CourseID:  d.CourseID,
			UserID:    d.UserID,
			Title:     d.Title,
			Body:      d.Body,
			CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

// DiscussionRepliesList returns replies for a discussion. GET /api/v1/discussions/:id/replies
func DiscussionRepliesList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "discussionId")
		d, err := deps.CourseDiscussionRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, err = deps.EnrollmentRepo.GetByUserAndCourse(r.Context(), userID, d.CourseID)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 100
		}
		list, err := deps.CourseDiscussionReplyRepo.ListByDiscussionID(r.Context(), id, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.DiscussionReplyItem, len(list))
		for i := range list {
			out[i] = dto.DiscussionReplyItem{
				ID:            list[i].ID,
				DiscussionID:  list[i].DiscussionID,
				UserID:        list[i].UserID,
				Body:          list[i].Body,
				CreatedAt:     list[i].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// DiscussionReplyCreate adds a reply. POST /api/v1/discussions/:id/replies
func DiscussionReplyCreate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		id := chi.URLParam(r, "discussionId")
		d, err := deps.CourseDiscussionRepo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, err = deps.EnrollmentRepo.GetByUserAndCourse(r.Context(), userID, d.CourseID)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var req struct {
			Body string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Body == "" {
			http.Error(w, "body required", http.StatusBadRequest)
			return
		}
		reply := domain.CourseDiscussionReply{DiscussionID: id, UserID: userID, Body: req.Body}
		created, err := deps.CourseDiscussionReplyRepo.Create(r.Context(), reply)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.DiscussionReplyItem{
			ID:            created.ID,
			DiscussionID:  created.DiscussionID,
			UserID:        created.UserID,
			Body:          created.Body,
			CreatedAt:     created.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

// TrainerCoursesList returns courses created by the trainer. GET /api/v1/trainer/courses
func TrainerCoursesList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		list, err := deps.CourseRepo.ListByCreatedBy(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.CourseItem, len(list))
		for i := range list {
			var desc string
			if list[i].Description != nil {
				desc = *list[i].Description
			}
			out[i] = dto.CourseItem{
				ID:          list[i].ID,
				Title:       list[i].Title,
				Description: desc,
				CreatedBy:   list[i].CreatedBy,
				CreatedAt:   list[i].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// TrainerCourseCreate creates a course (trainer). POST /api/v1/trainer/courses
func TrainerCourseCreate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var req struct {
			Title       string  `json:"title"`
			Description *string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Title == "" {
			http.Error(w, "title required", http.StatusBadRequest)
			return
		}
		c := domain.Course{Title: req.Title, Description: req.Description, CreatedBy: &userID}
		created, err := deps.CourseRepo.Create(r.Context(), c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var desc string
		if created.Description != nil {
			desc = *created.Description
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.CourseItem{
			ID:          created.ID,
			Title:       created.Title,
			Description: desc,
			CreatedBy:   created.CreatedBy,
			CreatedAt:   created.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

// StudentCoursesList returns courses the student is enrolled in (LMS shape). GET /api/v1/student/courses
func StudentCoursesList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 50 {
			limit = 10
		}
		search := r.URL.Query().Get("search")
		progressStatus := r.URL.Query().Get("progressStatus")
		if progressStatus != "" && progressStatus != "in-progress" && progressStatus != "completed" {
			writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid progressStatus")
			return
		}

		items, total, err := deps.EnrollmentRepo.ListCoursesByUserWithFilters(r.Context(), userID, search, progressStatus, page, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}

		data := make([]dto.StudentCourseItem, 0, len(items))
		for _, row := range items {
			enrolledAt := row.EnrolledAt.Format("2006-01-02T15:04:05Z07:00")
			data = append(data, dto.StudentCourseItem{
				ID:              row.EnrollmentID,
				Program:         dto.StudentCourseProgram{ID: row.CourseID, Slug: row.CourseSlug, Title: row.CourseTitle, Thumbnail: row.CourseThumbnail},
				ProgressPercent: enrollmentProgressPercent(row.EnrollmentStatus),
				EnrolledAt:      enrolledAt,
				LastAccessedAt:  enrolledAt,
			})
		}
		totalPages := int((total + limit - 1) / limit)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.StudentCoursesResponse{
			Data:       data,
			Total:      total,
			Page:       page,
			TotalPages: totalPages,
		})
	}
}

// StudentCoursesBySubject returns courses filtered by the logged-in student's subject. GET /api/v1/student/courses/by-subject
func StudentCoursesBySubject(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		u, err := deps.UserRepo.FindByID(r.Context(), userID)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		list, err := deps.CourseRepo.ListBySubjectID(r.Context(), u.SubjectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.CourseItem, 0, len(list))
		for _, c := range list {
			var desc string
			if c.Description != nil {
				desc = *c.Description
			}
			out = append(out, dto.CourseItem{
				ID:          c.ID,
				Title:       c.Title,
				Description: desc,
				SubjectID:   c.SubjectID,
				CreatedBy:   c.CreatedBy,
				CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// StudentPaymentsList is an alias for PaymentListMine (student sees own payments). GET /api/v1/student/payments
func StudentPaymentsList(deps *Deps) http.HandlerFunc {
	return PaymentListMine(deps)
}

// StudentProfileGet returns current student profile (same shape as GET /auth/me). GET /api/v1/student/profile
func StudentProfileGet(deps *Deps) http.HandlerFunc {
	return AuthMe(deps)
}

// StudentProfileUpdate updates current user name/email. PUT /api/v1/student/profile
func StudentProfileUpdate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		var req struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		u, err := deps.UserRepo.FindByID(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		if req.Name != "" {
			u.Name = req.Name
		}
		if req.Email != "" {
			existing, err := deps.UserRepo.FindByEmail(r.Context(), req.Email)
			if err == nil && existing.ID != userID {
				writeError(w, http.StatusConflict, "conflict", "email already in use")
				return
			}
			u.Email = req.Email
		}
		if err := deps.UserRepo.Update(r.Context(), u); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		role := u.Role
		if role == "guru" {
			role = "instructor"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.AuthUserResponse{
			ID:              u.ID,
			Name:            u.Name,
			Email:           u.Email,
			Role:            role,
			MustSetPassword: u.MustSetPassword,
		})
	}
}

// StudentTransactionsList returns the current user's transactions (orders). GET /api/v1/student/transactions
func StudentTransactionsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 50 {
			limit = 10
		}
		search := r.URL.Query().Get("search")
		status := r.URL.Query().Get("status") // pending|paid
		if status != "" && status != domain.OrderStatusPaid && status != domain.OrderStatusPending {
			writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid status")
			return
		}

		var (
			orders []domain.Order
			total  int
		)

		orders, total, err := deps.OrderRepo.ListByUserIDWithFilters(r.Context(), userID, status, search, page, limit)
		if err != nil {
			// Fallback: keep endpoint working even if server-side query fails.
			// We still filter/paginate in memory, then return the same contract.
			log.Printf("ListByUserIDWithFilters failed: %v", err)

			allOrders, err2 := deps.OrderRepo.ListByUserID(r.Context(), userID)
			if err2 != nil {
				writeError(w, http.StatusInternalServerError, "server_error", err2.Error())
				return
			}

			q := strings.TrimSpace(strings.ToLower(search))
			filtered := make([]domain.Order, 0, len(allOrders))
			for _, o := range allOrders {
				if status != "" && o.Status != status {
					continue
				}
				if q == "" {
					filtered = append(filtered, o)
					continue
				}

				// Match by orderId first.
				if strings.Contains(strings.ToLower(o.ID), q) {
					filtered = append(filtered, o)
					continue
				}

				// Match by program title(s) in order items.
				items, _ := deps.OrderItemRepo.ListByOrderID(r.Context(), o.ID)
				matched := false
				for _, item := range items {
					c, _ := deps.CourseRepo.GetByID(r.Context(), item.CourseID)
					if c.Title != "" && strings.Contains(strings.ToLower(c.Title), q) {
						matched = true
						break
					}
				}
				if matched {
					filtered = append(filtered, o)
				}
			}

			total = len(filtered)
			start := (page - 1) * limit
			if start >= total {
				orders = []domain.Order{}
			} else {
				end := start + limit
				if end > total {
					end = total
				}
				orders = filtered[start:end]
			}
		}

		data := make([]dto.StudentTransactionItem, 0, len(orders))
		for _, o := range orders {
			items, _ := deps.OrderItemRepo.ListByOrderID(r.Context(), o.ID)
			programs := make([]dto.StudentTransactionProgram, 0, len(items))
			for _, item := range items {
				c, _ := deps.CourseRepo.GetByID(r.Context(), item.CourseID)
				programs = append(programs, dto.StudentTransactionProgram{Title: c.Title})
			}
			paidAt := o.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
			if o.Status == domain.OrderStatusPaid {
				paidAt = o.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
			}
			discountPct := 0.0
			if o.DiscountPercent != nil {
				discountPct = *o.DiscountPercent
			}
			promoCode := ""
			if o.PromoCode != nil {
				promoCode = *o.PromoCode
			}
			confCode := ""
			if o.ConfirmationCode != nil {
				confCode = *o.ConfirmationCode
			}
			data = append(data, dto.StudentTransactionItem{
				ID:               o.ID,
				OrderID:          o.ID,
				Status:           o.Status,
				Total:            o.TotalPrice,
				NormalPrice:      o.NormalPrice,
				PromoCode:        promoCode,
				Discount:         o.Discount,
				DiscountPercent:  discountPct,
				ConfirmationCode: confCode,
				Programs:         programs,
				PaidAt:           paidAt,
			})
		}
		totalPages := int((total + limit - 1) / limit)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.StudentTransactionsResponse{
			Data:       data,
			Total:      total,
			Page:       page,
			TotalPages: totalPages,
		})
	}
}

// InstructorCoursesList returns courses taught by the current instructor. GET /api/v1/instructor/courses
func InstructorCoursesList(deps *Deps) http.HandlerFunc {
	return TrainerCoursesList(deps)
}

// InstructorStudentsList returns students enrolled in the instructor's courses. GET /api/v1/instructor/students
func InstructorStudentsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		courses, err := deps.CourseRepo.ListByCreatedBy(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		var data []dto.InstructorStudentItem
		for _, c := range courses {
			enrollments, _ := deps.EnrollmentRepo.ListByCourseID(r.Context(), c.ID)
			for _, e := range enrollments {
				u, err := deps.UserRepo.FindByID(r.Context(), e.UserID)
				if err != nil {
					continue
				}
				data = append(data, dto.InstructorStudentItem{
					UserID:          u.ID,
					Name:            u.Name,
					Email:           u.Email,
					ProgramTitle:    c.Title,
					ProgressPercent: enrollmentProgressPercent(e.Status),
				})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.InstructorStudentsResponse{Data: data})
	}
}

func enrollmentProgressPercent(status string) int {
	switch status {
	case domain.EnrollmentStatusCompleted:
		return 100
	case domain.EnrollmentStatusInProgress:
		return 50
	case domain.EnrollmentStatusEnrolled:
		return 0
	default:
		return 0
	}
}

// InstructorEarningsList returns earnings summary per period. GET /api/v1/instructor/earnings (stub)
func InstructorEarningsList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := middleware.GetUserID(r.Context()); !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.InstructorEarningsResponse{Data: []dto.InstructorEarningItem{}})
	}
}

// PackagesListLanding returns packages for landing page "Program yang Sedang Dibuka". GET /api/v1/packages
// Response: array of objects with snake_case (id, name, slug, price_early_bird, price_normal, ...).
func PackagesListLanding(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var list []domain.LandingPackage
		if deps.LandingPackageRepo != nil {
			var err error
			list, err = deps.LandingPackageRepo.List(r.Context())
			if err != nil {
				writeError(w, http.StatusInternalServerError, "server_error", err.Error())
				return
			}
		}
		if list == nil {
			list = []domain.LandingPackage{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}
}

// formatRupiah formats amount (dalam rupiah) ke string "RpXXX.XXX".
func formatRupiah(rupiah int) string {
	if rupiah < 0 {
		return "Rp0"
	}
	s := strconv.Itoa(rupiah)
	var b []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b = append(b, '.')
		}
		b = append(b, byte(c))
	}
	return "Rp" + string(b)
}
