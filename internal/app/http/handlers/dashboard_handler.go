package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// DashboardGeneral returns public/general dashboard stats (no auth). GET /api/v1/dashboard
func DashboardGeneral(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overview, _ := deps.AdminService.Overview(r.Context())
		levels, _ := deps.LevelRepo.List(r.Context())
		subjects, _ := deps.SubjectRepo.List(r.Context())
		schools, _ := deps.SchoolRepo.List(r.Context())
		siteName := "FansEdu LMS"
		if s, err := deps.SettingRepo.GetByKey(r.Context(), "site_name"); err == nil && s.Value != nil {
			siteName = *s.Value
		}
		out := dto.GeneralDashboardResponse{
			SiteName:      siteName,
			TotalLevels:   len(levels),
			TotalSubjects: len(subjects),
			TotalSchools:  len(schools),
		}
		if overview != nil {
			out.OpenTryouts = overview.ActiveTryouts
			out.TotalCourses = overview.TotalCourses
			out.TotalStudents = overview.TotalStudents
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func DashboardStudent(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		u, err := deps.UserRepo.FindByID(r.Context(), userID)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		now := time.Now().UTC()

		// Courses count + recent courses
		recentRows, coursesTotal, err := deps.EnrollmentRepo.ListCoursesByUserWithFilters(r.Context(), userID, "", "", 1, 10)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		recentCourses := make([]dto.StudentCourseItem, 0, len(recentRows))
		for _, row := range recentRows {
			enrolledAt := row.EnrolledAt.UTC().Format("2006-01-02T15:04:05Z07:00")
			recentCourses = append(recentCourses, dto.StudentCourseItem{
				ID:              row.EnrollmentID,
				Program:         dto.StudentCourseProgram{ID: row.CourseID, Slug: row.CourseSlug, Title: row.CourseTitle, Thumbnail: row.CourseThumbnail},
				ProgressPercent: enrollmentProgressPercent(row.EnrollmentStatus),
				EnrolledAt:      enrolledAt,
				LastAccessedAt:  enrolledAt,
			})
		}
		if len(recentCourses) > 5 {
			recentCourses = recentCourses[:5]
		}

		// Tryout summary
		attempts, _ := deps.AttemptService.ListByUser(r.Context(), userID)
		submittedAttempts := make([]domain.Attempt, 0)
		for _, a := range attempts {
			if a.Status == domain.AttemptStatusSubmitted {
				submittedAttempts = append(submittedAttempts, a)
			}
		}

		attemptedCount := len(attempts)
		completedCount := len(submittedAttempts)

		avgScore := 0.0
		bestScore := 0.0
		if completedCount > 0 {
			var sum float64
			var n int
			for _, a := range submittedAttempts {
				if a.Score != nil {
					sum += *a.Score
					if n == 0 || *a.Score > bestScore {
						bestScore = *a.Score
					}
					n++
				}
			}
			if n > 0 {
				avgScore = sum / float64(n)
			}
		}

		registeredCount := 0
		registeredCount, _ = deps.TryoutRegistrationRepo.CountRegisteredForStudent(r.Context(), userID, u.SubjectID)

		// upcoming count + streak days
		tryouts, _ := deps.TryoutService.ListForStudent(r.Context(), u.SubjectID, u.LevelID)
		upcomingCount := 0
		for _, t := range tryouts {
			if t.Status == domain.TryoutStatusOpen && t.OpensAt.After(now) {
				upcomingCount++
			}
		}

		daySet := map[string]struct{}{}
		lastAttemptAt := time.Time{}
		lastAttemptAtSet := false
		for _, a := range submittedAttempts {
			if a.SubmittedAt == nil {
				continue
			}
			dayKey := a.SubmittedAt.UTC().Format("2006-01-02")
			daySet[dayKey] = struct{}{}
			if !lastAttemptAtSet || a.SubmittedAt.After(lastAttemptAt) {
				lastAttemptAt = *a.SubmittedAt
				lastAttemptAtSet = true
			}
		}
		streakDays := 0
		if lastAttemptAtSet {
			current := lastAttemptAt
			for {
				key := current.UTC().Format("2006-01-02")
				if _, ok := daySet[key]; !ok {
					break
				}
				streakDays++
				current = current.Add(-24 * time.Hour)
			}
		}
		lastAttemptAtStr := ""
		if lastAttemptAtSet {
			lastAttemptAtStr = lastAttemptAt.UTC().Format(time.RFC3339)
		}

		// weekly target
		weekAgo := now.Add(-7 * 24 * time.Hour)
		enrollments, _ := deps.EnrollmentRepo.ListByUserID(r.Context(), userID)
		completedLessons := 0
		for _, e := range enrollments {
			if e.Status == domain.EnrollmentStatusCompleted && e.CompletedAt != nil && e.CompletedAt.After(weekAgo) {
				completedLessons++
			}
		}
		completedTryouts := 0
		for _, a := range submittedAttempts {
			if a.SubmittedAt != nil && a.SubmittedAt.After(weekAgo) {
				completedTryouts++
			}
		}

		out := dto.StudentDashboardResponse{
			CoursesCount:  coursesTotal,
			RecentCourses: recentCourses,
			TryoutSummary: dto.StudentTryoutSummary{
				AttemptedCount:  attemptedCount,
				CompletedCount:  completedCount,
				RegisteredCount: registeredCount,
				AverageScore:    avgScore,
				BestScore:       bestScore,
				UpcomingCount:   upcomingCount,
				StreakDays:      streakDays,
				LastAttemptAt:   lastAttemptAtStr,
			},
			WeeklyTarget: dto.StudentWeeklyTarget{
				TargetLessons:    8,
				TargetTryouts:    1,
				CompletedLessons: completedLessons,
				CompletedTryouts: completedTryouts,
			},
			Badges: []dto.StudentBadge{},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// GET /api/v1/student/next-actions
func StudentNextActions(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		u, err := deps.UserRepo.FindByID(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}

		actions := make([]dto.StudentNextAction, 0, 3)

		// payment pending
		pendingOrders, _, _ := deps.OrderRepo.ListByUserIDWithFilters(r.Context(), userID, domain.OrderStatusPending, "", 1, 1)
		if len(pendingOrders) > 0 {
			actions = append(actions, dto.StudentNextAction{
				ID:          "action-payment-pending",
				Type:        "payment_pending",
				Title:       "Pembayaran tertunda",
				Description: "Silakan lengkapi bukti pembayaran untuk memverifikasi order kamu.",
				Href:        "#/student/transactions",
				Priority:    1,
			})
		}

		// continue course
		if len(actions) == 0 {
			inProgRows, _, _ := deps.EnrollmentRepo.ListCoursesByUserWithFilters(r.Context(), userID, "", "in-progress", 1, 1)
			if len(inProgRows) > 0 {
				row := inProgRows[0]
				actions = append(actions, dto.StudentNextAction{
					ID:              "action-continue-course",
					Type:            "continue_course",
					Title:           "Lanjutkan " + row.CourseTitle,
					Description:     "Progress kamu " + strconv.Itoa(enrollmentProgressPercent(row.EnrollmentStatus)) + "%",
					Href:            "#/student/courses/" + row.CourseSlug,
					Priority:        1,
				})
			}
			// fallback: if user only has "enrolled" (0%) courses, show that too.
			if len(actions) == 0 {
				firstRows, _, _ := deps.EnrollmentRepo.ListCoursesByUserWithFilters(r.Context(), userID, "", "", 1, 1)
				if len(firstRows) > 0 {
					row := firstRows[0]
					if enrollmentProgressPercent(row.EnrollmentStatus) < 100 {
						actions = append(actions, dto.StudentNextAction{
							ID:              "action-continue-course",
							Type:            "continue_course",
							Title:           "Lanjutkan " + row.CourseTitle,
							Description:     "Progress kamu " + strconv.Itoa(enrollmentProgressPercent(row.EnrollmentStatus)) + "%",
							Href:            "#/student/courses/" + row.CourseSlug,
							Priority:        1,
						})
					}
				}
			}
		}

		// start tryout
		if len(actions) == 0 {
			openTryouts, _ := deps.TryoutService.ListOpenForStudent(r.Context(), u.SubjectID, u.LevelID)
			attempts, _ := deps.AttemptService.ListByUser(r.Context(), userID)
			attempted := map[string]struct{}{}
			for _, a := range attempts {
				attempted[a.TryoutSessionID] = struct{}{}
			}
			for _, t := range openTryouts {
				if _, ok := attempted[t.ID]; ok {
					continue
				}
				isRegistered, _ := deps.TryoutRegistrationRepo.IsRegistered(r.Context(), userID, t.ID)
				if !isRegistered {
					continue
				}
				actions = append(actions, dto.StudentNextAction{
					ID:          "action-start-tryout-" + t.ID,
					Type:        "start_tryout",
					Title:       "Mulai " + t.Title,
					Description: "Ujian tryout dimulai sekarang.",
					Href:        "#/student/tryouts/" + t.ID,
					Priority:    1,
				})
				break
			}
		}

		if len(actions) == 0 {
			actions = append(actions, dto.StudentNextAction{
				ID:          "action-custom",
				Type:        "custom",
				Title:       "Mulai dari sini",
				Description: "Pilih tryout atau lanjutkan kelas yang tersedia.",
				Href:        "#/student/courses",
				Priority:    1,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.StudentNextActionsResponse{Data: actions})
	}
}
