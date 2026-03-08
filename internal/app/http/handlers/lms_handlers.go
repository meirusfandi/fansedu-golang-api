package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
			AmountCents  int     `json:"amount_cents"`
			Type         string  `json:"type"`
			ReferenceID  *string `json:"reference_id"`
			Description  *string `json:"description"`
			ProofURL     *string `json:"proof_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.AmountCents <= 0 {
			http.Error(w, "amount_cents must be positive", http.StatusBadRequest)
			return
		}
		ptype := req.Type
		if ptype == "" {
			ptype = domain.PaymentTypeCoursePurchase
		}
		p := domain.Payment{
			UserID:      userID,
			AmountCents: req.AmountCents,
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
		ID:           p.ID,
		UserID:       p.UserID,
		AmountCents:  p.AmountCents,
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

// StudentCoursesList returns courses the student is enrolled in. GET /api/v1/student/courses
func StudentCoursesList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		enrollments, err := deps.EnrollmentRepo.ListByUserID(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.CourseItem, 0, len(enrollments))
		for _, e := range enrollments {
			c, err := deps.CourseRepo.GetByID(r.Context(), e.CourseID)
			if err != nil {
				continue
			}
			var desc string
			if c.Description != nil {
				desc = *c.Description
			}
			out = append(out, dto.CourseItem{
				ID:          c.ID,
				Title:       c.Title,
				Description: desc,
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
