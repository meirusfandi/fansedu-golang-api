package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

func AdminOverview(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overview, err := deps.AdminService.Overview(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.AdminOverviewResponse{
			TotalStudents:     overview.TotalStudents,
			TotalUsers:        overview.TotalUsers,
			ActiveTryouts:     overview.ActiveTryouts,
			TotalCourses:      overview.TotalCourses,
			TotalEnrollments:  overview.TotalEnrollments,
			AvgScore:          overview.AvgScore,
			TotalCertificates: overview.TotalCertificates,
		})
	}
}

func AdminListUsers(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := r.URL.Query().Get("role")
		list, err := deps.AdminService.ListUsers(r.Context(), role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.UserListResponse, len(list))
		for i := range list {
			out[i] = dto.UserListResponse{
				ID:        list[i].ID,
				Email:     list[i].Email,
				Name:      list[i].Name,
				Role:      list[i].Role,
				AvatarURL: list[i].AvatarURL,
				CreatedAt: list[i].CreatedAt.Format(time.RFC3339),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminGetUser(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "userId")
		u, err := deps.AdminService.GetUser(r.Context(), id)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.UserDetailResponse{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Role:      u.Role,
			AvatarURL: u.AvatarURL,
			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		})
	}
}

func AdminCreateUser(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.UserCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Password == "" || req.Name == "" {
			http.Error(w, "email, password, name required", http.StatusBadRequest)
			return
		}
		role := req.Role
		if role == "" {
			role = domain.UserRoleStudent
		}
		if role != domain.UserRoleStudent && role != domain.UserRoleAdmin {
			role = domain.UserRoleStudent
		}
		u := domain.User{Email: req.Email, Name: req.Name, Role: role, AvatarURL: req.AvatarURL}
		created, err := deps.AdminService.CreateUser(r.Context(), u, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.UserDetailResponse{
			ID:        created.ID,
			Email:     created.Email,
			Name:      created.Name,
			Role:      created.Role,
			AvatarURL: created.AvatarURL,
			CreatedAt: created.CreatedAt.Format(time.RFC3339),
			UpdatedAt: created.UpdatedAt.Format(time.RFC3339),
		})
	}
}

func AdminUpdateUser(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "userId")
		u, err := deps.AdminService.GetUser(r.Context(), id)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		var req dto.UserUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Name != nil {
			u.Name = *req.Name
		}
		if req.Email != nil {
			u.Email = *req.Email
		}
		if req.Role != nil {
			u.Role = *req.Role
		}
		if req.AvatarURL != nil {
			u.AvatarURL = req.AvatarURL
		}
		if err := deps.AdminService.UpdateUser(r.Context(), u, req.Password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func AdminListTryouts(deps *Deps) http.HandlerFunc {
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

func AdminCreateTryout(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.TryoutCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		t := domain.TryoutSession{
			Title:           req.Title,
			ShortTitle:      req.ShortTitle,
			Description:     req.Description,
			DurationMinutes: req.DurationMinutes,
			QuestionsCount:  req.QuestionsCount,
			Level:           req.Level,
			OpensAt:         req.OpensAt,
			ClosesAt:        req.ClosesAt,
			MaxParticipants: req.MaxParticipants,
			Status:          req.Status,
		}
		if t.Status == "" {
			t.Status = domain.TryoutStatusOpen
		}
		created, err := deps.AdminService.CreateTryout(r.Context(), t)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(tryoutToDTO(created))
	}
}

func AdminUpdateTryout(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "tryoutId")
		var req dto.TryoutCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		t := domain.TryoutSession{
			ID:               id,
			Title:            req.Title,
			ShortTitle:       req.ShortTitle,
			Description:      req.Description,
			DurationMinutes:  req.DurationMinutes,
			QuestionsCount:   req.QuestionsCount,
			Level:            req.Level,
			OpensAt:          req.OpensAt,
			ClosesAt:         req.ClosesAt,
			MaxParticipants:  req.MaxParticipants,
			Status:           req.Status,
		}
		if err := deps.AdminService.UpdateTryout(r.Context(), t); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func AdminDeleteTryout(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "tryoutId")
		if err := deps.AdminService.DeleteTryout(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminListQuestions(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		list, err := deps.AdminService.ListQuestions(r.Context(), tryoutID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.QuestionResponse, len(list))
		for i := range list {
			out[i] = questionToDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminGetQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		questionID := chi.URLParam(r, "questionId")
		q, err := deps.AdminService.GetQuestion(r.Context(), questionID)
		if err != nil {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if q.TryoutSessionID != tryoutID {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(questionToDTO(q))
	}
}

func AdminCreateQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		var req dto.QuestionCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		opts, _ := json.Marshal(req.Options)
		q := domain.Question{
			TryoutSessionID: tryoutID,
			SortOrder:        req.SortOrder,
			Type:             req.Type,
			Body:             req.Body,
			Options:          opts,
			MaxScore:         req.MaxScore,
		}
		if q.MaxScore == 0 {
			q.MaxScore = 1
		}
		created, err := deps.AdminService.CreateQuestion(r.Context(), q)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(questionToDTO(created))
	}
}

func AdminUpdateQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		questionID := chi.URLParam(r, "questionId")
		var req dto.QuestionUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		q, err := deps.QuestionRepo.GetByID(r.Context(), questionID)
		if err != nil {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if q.TryoutSessionID != tryoutID {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if req.SortOrder != nil {
			q.SortOrder = *req.SortOrder
		}
		if req.Type != nil {
			q.Type = *req.Type
		}
		if req.Body != nil {
			q.Body = *req.Body
		}
		if req.Options != nil {
			opts, _ := json.Marshal(req.Options)
			q.Options = opts
		}
		if req.MaxScore != nil {
			q.MaxScore = *req.MaxScore
		}
		if err := deps.AdminService.UpdateQuestion(r.Context(), q); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(questionToDTO(q))
	}
}

func AdminDeleteQuestion(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		questionID := chi.URLParam(r, "questionId")
		q, err := deps.AdminService.GetQuestion(r.Context(), questionID)
		if err != nil {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if q.TryoutSessionID != tryoutID {
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		if err := deps.AdminService.DeleteQuestion(r.Context(), questionID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminListCourses(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.AdminService.ListCourses(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.CourseResponse, len(list))
		for i := range list {
			out[i] = dto.CourseResponse{
				ID:          list[i].ID,
				Title:       list[i].Title,
				Description: list[i].Description,
				CreatedBy:   list[i].CreatedBy,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminGetCourse(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "courseId")
		c, err := deps.AdminService.GetCourseByID(r.Context(), id)
		if err != nil {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.CourseResponse{
			ID:          c.ID,
			Title:       c.Title,
			Description: c.Description,
			CreatedBy:   c.CreatedBy,
		})
	}
}

func AdminCreateCourse(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.CourseCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		createdBy, ok := middleware.GetUserID(r.Context())
		var createdByPtr *string
		if ok && createdBy != "" {
			createdByPtr = &createdBy
		}
		c := domain.Course{
			Title:       req.Title,
			Description: req.Description,
			CreatedBy:   createdByPtr,
		}
		created, err := deps.AdminService.CreateCourse(r.Context(), c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.CourseResponse{
			ID:          created.ID,
			Title:       created.Title,
			Description: created.Description,
			CreatedBy:   created.CreatedBy,
		})
	}
}

func AdminUpdateCourse(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "courseId")
		c, err := deps.AdminService.GetCourseByID(r.Context(), id)
		if err != nil {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		var req dto.CourseCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		c.Title = req.Title
		c.Description = req.Description
		if err := deps.AdminService.UpdateCourse(r.Context(), c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func AdminListEnrollments(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := chi.URLParam(r, "courseId")
		list, err := deps.AdminService.ListEnrollmentsByCourse(r.Context(), courseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.EnrollmentResponse, len(list))
		for i := range list {
			out[i] = dto.EnrollmentResponse{
				ID:         list[i].ID,
				UserID:     list[i].UserID,
				CourseID:   list[i].CourseID,
				Status:     list[i].Status,
				EnrolledAt: list[i].EnrolledAt.Format(time.RFC3339),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminIssueCertificate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.CertificateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		c := domain.Certificate{
			UserID:          req.UserID,
			TryoutSessionID: req.TryoutSessionID,
			CourseID:        req.CourseID,
		}
		created, err := deps.AdminService.IssueCertificate(r.Context(), c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         created.ID,
			"user_id":    created.UserID,
			"issued_at":  created.IssuedAt.Format(time.RFC3339),
		})
	}
}

func courseContentToDTO(c domain.CourseContent) dto.CourseContentResponse {
	var content interface{}
	if len(c.Content) > 0 {
		_ = json.Unmarshal(c.Content, &content)
	}
	return dto.CourseContentResponse{
		ID:          c.ID,
		CourseID:    c.CourseID,
		Title:       c.Title,
		Description: c.Description,
		SortOrder:   c.SortOrder,
		Type:        c.Type,
		Content:     content,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
	}
}

func AdminListCourseContents(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := chi.URLParam(r, "courseId")
		list, err := deps.AdminService.ListCourseContents(r.Context(), courseID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.CourseContentResponse, len(list))
		for i := range list {
			out[i] = courseContentToDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func AdminCreateCourseContent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := chi.URLParam(r, "courseId")
		var req dto.CourseContentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Title == "" || req.Type == "" {
			http.Error(w, "title and type required", http.StatusBadRequest)
			return
		}
		content, _ := json.Marshal(req.Content)
		c := domain.CourseContent{
			CourseID:    courseID,
			Title:       req.Title,
			Description: req.Description,
			SortOrder:   req.SortOrder,
			Type:        req.Type,
			Content:     content,
		}
		created, err := deps.AdminService.CreateCourseContent(r.Context(), c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(courseContentToDTO(created))
	}
}

func AdminUpdateCourseContent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentID := chi.URLParam(r, "contentId")
		c, err := deps.AdminService.GetCourseContent(r.Context(), contentID)
		if err != nil {
			http.Error(w, "content not found", http.StatusNotFound)
			return
		}
		var req dto.CourseContentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.Title != "" {
			c.Title = req.Title
		}
		if req.Description != nil {
			c.Description = req.Description
		}
		// allow 0
		c.SortOrder = req.SortOrder
		if req.Type != "" {
			c.Type = req.Type
		}
		if req.Content != nil {
			content, _ := json.Marshal(req.Content)
			c.Content = content
		}
		if err := deps.AdminService.UpdateCourseContent(r.Context(), c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func AdminDeleteCourseContent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentID := chi.URLParam(r, "contentId")
		if err := deps.AdminService.DeleteCourseContent(r.Context(), contentID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AdminListPayments(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := parseInt(l); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		list, err := deps.AdminService.ListPayments(r.Context(), limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.PaymentResponse, len(list))
		for i := range list {
			var paidAt *string
			if list[i].PaidAt != nil {
				s := list[i].PaidAt.Format(time.RFC3339)
				paidAt = &s
			}
			out[i] = dto.PaymentResponse{
				ID:          list[i].ID,
				UserID:      list[i].UserID,
				AmountCents: list[i].AmountCents,
				Currency:    list[i].Currency,
				Status:      list[i].Status,
				Type:        list[i].Type,
				PaidAt:      paidAt,
				CreatedAt:   list[i].CreatedAt.Format(time.RFC3339),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func AdminCreatePayment(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.PaymentCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if req.UserID == "" || req.AmountCents <= 0 {
			http.Error(w, "user_id and amount_cents required", http.StatusBadRequest)
			return
		}
		status := req.Status
		if status == "" {
			status = domain.PaymentStatusPending
		}
		ptype := req.Type
		if ptype == "" {
			ptype = domain.PaymentTypeCoursePurchase
		}
		var paidAt *time.Time
		if status == domain.PaymentStatusPaid {
			now := time.Now()
			paidAt = &now
		}
		currency := req.Currency
		if currency == "" {
			currency = "IDR"
		}
		p := domain.Payment{
			UserID:      req.UserID,
			AmountCents: req.AmountCents,
			Currency:    currency,
			Status:      status,
			Type:        ptype,
			ReferenceID: req.ReferenceID,
			Description: req.Description,
			PaidAt:      paidAt,
		}
		created, err := deps.AdminService.CreatePayment(r.Context(), p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		var paidAtStr *string
		if created.PaidAt != nil {
			s := created.PaidAt.Format(time.RFC3339)
			paidAtStr = &s
		}
		_ = json.NewEncoder(w).Encode(dto.PaymentResponse{
			ID:          created.ID,
			UserID:      created.UserID,
			AmountCents: created.AmountCents,
			Currency:    created.Currency,
			Status:      created.Status,
			Type:        created.Type,
			PaidAt:      paidAtStr,
			CreatedAt:   created.CreatedAt.Format(time.RFC3339),
		})
	}
}

func AdminReportMonthly(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		year, month := time.Now().Year(), int(time.Now().Month())
		if y := r.URL.Query().Get("year"); y != "" {
			if n, err := parseInt(y); err == nil && n >= 2020 && n <= 2100 {
				year = n
			}
		}
		if m := r.URL.Query().Get("month"); m != "" {
			if n, err := parseInt(m); err == nil && n >= 1 && n <= 12 {
				month = n
			}
		}
		report, err := deps.AdminService.ReportMonthly(r.Context(), year, month)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.MonthlyReportResponse{
			Year:              report.Year,
			Month:             report.Month,
			NewEnrollments:    report.NewEnrollments,
			PaymentsCount:     report.PaymentsCount,
			TotalRevenueCents: report.TotalRevenueCents,
		})
	}
}
