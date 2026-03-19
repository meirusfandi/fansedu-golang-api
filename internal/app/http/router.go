package httpapi

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/handlers"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func NewRouter(deps *handlers.Deps) http.Handler {
	r := chi.NewRouter()

	// CORS: default allow frontend (Vite http://localhost:5173) and all; set CORS_ORIGINS to restrict.
	r.Use(middleware.CORS(getEnv("CORS_ORIGINS", "http://localhost:5173,*")))
	r.Use(middleware.RequestID())
	r.Use(middleware.Recover())
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger())

	registerV1 := func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if deps == nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					_, _ = w.Write([]byte(`{"error":"Service unavailable: database not configured. Set DATABASE_URL in .env or .env.dev."}`))
					return
				}
				next.ServeHTTP(w, req)
			})
		})
		r.Get("/health", handlers.Health())
		r.Get("/dashboard", handlers.DashboardGeneral(deps))

		r.Get("/roles", handlers.ListRoles(deps))
		r.Get("/schools", handlers.ListSchools(deps))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.AuthRegister(deps))
			r.Post("/register-with-invite", handlers.AuthRegisterWithInvite(deps))
			r.Post("/login", handlers.AuthLogin(deps))
			r.Post("/verify-email", handlers.AuthVerifyEmail(deps))
			r.Post("/resend-verification", handlers.AuthResendVerification(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Get("/me", handlers.AuthMe(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/logout", handlers.AuthLogout(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/set-password", handlers.AuthSetPassword(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/change-password", handlers.AuthChangePassword(deps))
			r.Post("/forgot-password", handlers.AuthForgotPassword(deps))
			r.Post("/reset-password", handlers.AuthResetPassword(deps))
		})

		r.Route("/checkout", func(r chi.Router) {
			r.With(middleware.OptionalAuth(deps.JWTSecret)).Post("/initiate", handlers.CheckoutInitiate(deps))
			r.Post("/payment-session", handlers.CheckoutPaymentSession(deps))
			r.Post("/payment-session/", handlers.CheckoutPaymentSession(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/orders/{orderId}/payment-proof", handlers.CheckoutPaymentProof(deps))
			r.Post("/orders/{orderId}/complete-purchase-auth", handlers.CompletePurchaseAuth(deps))
		})
		r.Post("/analytics/pageview", handlers.AnalyticsTrackPageview(deps))
		r.Post("/webhook/payment", handlers.PaymentWebhook(deps))

		r.Route("/programs", func(r chi.Router) {
			r.Get("/", handlers.ProgramsList(deps))
			r.Get("/{slug}", handlers.ProgramBySlug(deps))
		})

		r.Get("/packages", handlers.PackagesListLanding(deps))

		r.Route("/tryouts", func(r chi.Router) {
			r.Get("/open", handlers.TryoutListOpen(deps))
			r.Get("/{tryoutId}", handlers.TryoutGetByID(deps))
			r.Get("/{tryoutId}/leaderboard", handlers.TryoutLeaderboard(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/{tryoutId}/register", handlers.TryoutRegister(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/{tryoutId}/start", handlers.TryoutStart(deps))
		})

		r.Route("/attempts", func(r chi.Router) {
			r.With(middleware.Auth(deps.JWTSecret)).Get("/{attemptId}/questions", handlers.AttemptGetQuestions(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Put("/{attemptId}/answers/{questionId}", handlers.AttemptPutAnswer(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/{attemptId}/submit", handlers.AttemptSubmit(deps))
		})

		// Helper untuk PasswordSetupGuard
		passwordGuard := middleware.PasswordSetupGuard(func(ctx context.Context, id string) (bool, error) {
			u, err := deps.UserRepo.FindByID(ctx, id)
			if err != nil {
				return false, err
			}
			return u.MustSetPassword, nil
		})

		r.Route("/student", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Use(passwordGuard)
			r.Get("/dashboard", handlers.DashboardStudent(deps))
			r.Get("/profile", handlers.StudentProfileGet(deps))
			r.Put("/profile", handlers.StudentProfileUpdate(deps))
			r.Get("/courses", handlers.StudentCoursesList(deps))
			r.Get("/courses/by-subject", handlers.StudentCoursesBySubject(deps))
			r.Get("/transactions", handlers.StudentTransactionsList(deps))
			r.Get("/payments", handlers.StudentPaymentsList(deps))
			r.Get("/tryouts", handlers.StudentTryoutList(deps))
			r.Get("/tryouts/open", handlers.StudentTryoutListOpen(deps))
			r.Get("/tryouts/{tryoutId}", handlers.StudentTryoutGetByID(deps))
			r.Get("/attempts", handlers.AttemptListByUser(deps))
			r.Get("/attempts/{attemptId}", handlers.AttemptGetByID(deps))
			r.Get("/certificates", handlers.CertificateList(deps))
		})

		r.Route("/courses", func(r chi.Router) {
			r.Get("/", handlers.CourseList(deps))
			r.Get("/slug/{slug}", handlers.CourseGetBySlug(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/{courseId}/enroll", handlers.CourseEnroll(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Get("/{courseId}/messages", handlers.CourseMessagesList(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/{courseId}/messages", handlers.CourseMessageCreate(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Get("/{courseId}/discussions", handlers.CourseDiscussionsList(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/{courseId}/discussions", handlers.CourseDiscussionCreate(deps))
		})

		r.Route("/discussions", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Get("/{discussionId}", handlers.DiscussionGet(deps))
			r.Get("/{discussionId}/replies", handlers.DiscussionRepliesList(deps))
			r.Post("/{discussionId}/replies", handlers.DiscussionReplyCreate(deps))
		})

		r.Route("/notifications", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Get("/", handlers.NotificationsList(deps))
			r.Patch("/{notificationId}/read", handlers.NotificationMarkRead(deps))
		})

		r.Route("/payments", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Get("/", handlers.PaymentListMine(deps))
			r.Post("/", handlers.PaymentCreate(deps))
		})

		r.Route("/trainer", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Use(passwordGuard)
			r.Use(middleware.TrainerOnly())
			r.Get("/profile", handlers.TrainerProfileGet(deps))
			r.Put("/profile", handlers.TrainerProfileUpdate(deps))
			r.Get("/courses", handlers.TrainerCoursesList(deps))
			r.Post("/courses", handlers.TrainerCourseCreate(deps))
			r.Get("/status", handlers.TrainerStatus(deps))
			r.Post("/pay", handlers.TrainerPay(deps))
			r.Post("/students", handlers.TrainerCreateStudent(deps))
			r.Route("/tryouts", func(r chi.Router) {
				r.Get("/", handlers.TrainerTryoutList(deps))
				r.Get("/{tryoutId}/analysis", handlers.TrainerTryoutAnalysis(deps))
				r.Get("/{tryoutId}/students", handlers.TrainerTryoutStudents(deps))
				r.Get("/{tryoutId}/attempts/{attemptId}/ai-analysis", handlers.TrainerAttemptAIAnalysis(deps))
			})
		})

		r.Route("/instructor", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Use(passwordGuard)
			r.Use(middleware.TrainerOnly())
			r.Get("/courses", handlers.InstructorCoursesList(deps))
			r.Get("/students", handlers.InstructorStudentsList(deps))
			r.Get("/earnings", handlers.InstructorEarningsList(deps))
			r.Route("/tryouts", func(r chi.Router) {
				r.Get("/", handlers.TrainerTryoutList(deps))
				r.Get("/{tryoutId}/analysis", handlers.TrainerTryoutAnalysis(deps))
				r.Get("/{tryoutId}/students", handlers.TrainerTryoutStudents(deps))
				r.Get("/{tryoutId}/attempts/{attemptId}/ai-analysis", handlers.TrainerAttemptAIAnalysis(deps))
			})
		})

		r.Route("/levels", func(r chi.Router) {
			r.Get("/", handlers.AdminListLevels(deps))
			r.Get("/{levelId}", handlers.LevelWithSubjects(deps))
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Use(middleware.AdminOnly())
			r.Get("/overview", handlers.AdminOverview(deps))
			r.Get("/users", handlers.AdminListUsers(deps))
			r.Post("/users", handlers.AdminCreateUser(deps))
			r.Get("/users/{userId}", handlers.AdminGetUser(deps))
			r.Put("/users/{userId}", handlers.AdminUpdateUser(deps))
			r.Get("/courses", handlers.AdminListCourses(deps))
			r.Get("/courses/{courseId}", handlers.AdminGetCourse(deps))
			r.Post("/courses", handlers.AdminCreateCourse(deps))
			r.Put("/courses/{courseId}", handlers.AdminUpdateCourse(deps))
			r.Get("/courses/{courseId}/enrollments", handlers.AdminListEnrollments(deps))
			r.Get("/courses/{courseId}/contents", handlers.AdminListCourseContents(deps))
			r.Post("/courses/{courseId}/contents", handlers.AdminCreateCourseContent(deps))
			r.Put("/courses/{courseId}/contents/{contentId}", handlers.AdminUpdateCourseContent(deps))
			r.Delete("/courses/{courseId}/contents/{contentId}", handlers.AdminDeleteCourseContent(deps))
			r.Get("/payments", handlers.AdminListPayments(deps))
			r.Post("/payments", handlers.AdminCreatePayment(deps))
			r.Put("/payments/{paymentId}", handlers.AdminConfirmPayment(deps))
			r.Put("/orders/{orderId}/verify", handlers.AdminVerifyOrder(deps))
			r.Get("/analytics/summary", handlers.AdminAnalyticsSummary(deps))
			r.Get("/analytics/visitors", handlers.AdminAnalyticsVisitors(deps))
			r.Get("/reports/monthly", handlers.AdminReportMonthly(deps))
			r.Get("/reports/courses/{courseId}", handlers.AdminCourseReport(deps))
			r.Get("/roles", handlers.AdminListRoles(deps))
			r.Post("/roles", handlers.AdminCreateRole(deps))
			r.Get("/roles/{id}", handlers.AdminGetRole(deps))
			r.Put("/roles/{id}", handlers.AdminUpdateRole(deps))
			r.Delete("/roles/{id}", handlers.AdminDeleteRole(deps))
			r.Get("/schools", handlers.AdminListSchools(deps))
			r.Post("/schools", handlers.AdminCreateSchool(deps))
			r.Get("/schools/{id}", handlers.AdminGetSchool(deps))
			r.Put("/schools/{id}", handlers.AdminUpdateSchool(deps))
			r.Delete("/schools/{id}", handlers.AdminDeleteSchool(deps))
			// Alias for frontend master-data UI: /api/v1/admin/master-data/sekolah
			r.Route("/master-data", func(r chi.Router) {
				r.Route("/sekolah", func(r chi.Router) {
					r.Get("/", handlers.AdminListSchools(deps))
					r.Post("/", handlers.AdminCreateSchool(deps))
					r.Get("/{id}", handlers.AdminGetSchool(deps))
					r.Put("/{id}", handlers.AdminUpdateSchool(deps))
					r.Delete("/{id}", handlers.AdminDeleteSchool(deps))
				})
			})
			r.Get("/settings", handlers.AdminListSettings(deps))
			r.Post("/settings", handlers.AdminCreateSetting(deps))
			r.Get("/settings/{id}", handlers.AdminGetSetting(deps))
			r.Put("/settings/{id}", handlers.AdminUpdateSetting(deps))
			r.Delete("/settings/{id}", handlers.AdminDeleteSetting(deps))
			r.Get("/events", handlers.AdminListEvents(deps))
			r.Post("/events", handlers.AdminCreateEvent(deps))
			r.Get("/events/{id}", handlers.AdminGetEvent(deps))
			r.Put("/events/{id}", handlers.AdminUpdateEvent(deps))
			r.Delete("/events/{id}", handlers.AdminDeleteEvent(deps))
			r.Get("/subjects", handlers.AdminListSubjects(deps))
			r.Post("/subjects", handlers.AdminCreateSubject(deps))
			r.Get("/subjects/{id}", handlers.AdminGetSubject(deps))
			r.Put("/subjects/{id}", handlers.AdminUpdateSubject(deps))
			r.Delete("/subjects/{id}", handlers.AdminDeleteSubject(deps))
			r.Route("/levels", func(r chi.Router) {
				r.Get("/", handlers.AdminListLevels(deps))
				r.Post("/", handlers.AdminCreateLevel(deps))
				r.Get("/{id}", handlers.AdminGetLevel(deps))
				r.Get("/{id}/subjects", handlers.LevelWithSubjects(deps))
				r.Put("/{id}", handlers.AdminUpdateLevel(deps))
				r.Delete("/{id}", handlers.AdminDeleteLevel(deps))
			})
			r.Route("/tryouts", func(r chi.Router) {
				r.Get("/", handlers.AdminListTryouts(deps))
				r.Post("/", handlers.AdminCreateTryout(deps))
				r.Route("/{tryoutId}", func(r chi.Router) {
					r.Put("/", handlers.AdminUpdateTryout(deps))
					r.Delete("/", handlers.AdminDeleteTryout(deps))
					r.Get("/analysis", handlers.AdminGetTryoutAnalysis(deps))
					r.Get("/students", handlers.AdminListTryoutStudents(deps))
					r.Get("/questions", handlers.AdminListQuestions(deps))
					r.Get("/questions/stats", handlers.AdminGetTryoutQuestionStatsBulk(deps))
					r.Post("/questions", handlers.AdminCreateQuestion(deps))
					r.Get("/questions/{questionId}", handlers.AdminGetQuestion(deps))
					r.Put("/questions/{questionId}", handlers.AdminUpdateQuestion(deps))
					r.Delete("/questions/{questionId}", handlers.AdminDeleteQuestion(deps))
					r.Get("/questions/{questionId}/stats", handlers.AdminGetQuestionStats(deps))
					r.Get("/attempts/{attemptId}/ai-analysis", handlers.AdminGetAttemptAIAnalysis(deps))
				})
			})
			r.Post("/certificates", handlers.AdminIssueCertificate(deps))
		})
	}

	r.Route("/api/v1", registerV1)

	return r
}
