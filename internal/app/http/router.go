package httpapi

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

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

// allowMinimalAPI allows /health and /geo without PostgreSQL (e.g. geo + Redis only).
func allowMinimalAPI(path string) bool {
	p := strings.TrimSuffix(path, "/")
	if p == "" {
		p = path
	}
	if p == "/api/v1/health" {
		return true
	}
	if strings.HasPrefix(p, "/api/v1/geo") {
		return true
	}
	return false
}

func NewRouter(deps *handlers.Deps) http.Handler {
	var appErrInserter middleware.ApplicationErrorLogInserter
	if deps != nil {
		appErrInserter = deps.ApplicationErrorLogRepo
	}
	r := chi.NewRouter()

	// CORS: default allow frontend (Vite http://localhost:5173) and all; set CORS_ORIGINS to restrict.
	r.Use(middleware.CORS(getEnv("CORS_ORIGINS", "http://localhost:5173,*")))
	r.Use(middleware.RequestID())
	r.Use(middleware.Recover(appErrInserter))
	// Redirect trailing slashes so both "/x" and "/x/" can match.
	r.Use(chimw.RedirectSlashes)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger())

	registerV1 := func(r chi.Router) {
		r.Use(middleware.ErrorResponseLogger(appErrInserter))
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if deps != nil && deps.DB != nil {
					next.ServeHTTP(w, req)
					return
				}
				if deps != nil && allowMinimalAPI(req.URL.Path) {
					next.ServeHTTP(w, req)
					return
				}
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"error":{"code":"SERVICE_UNAVAILABLE","message":"Layanan sementara tidak tersedia."}}`))
			})
		})
		r.Get("/health", handlers.Health())
		r.Get("/geo/provinces", handlers.GeoProvinces(deps))
		r.Get("/geo/regencies/{provinceId}", handlers.GeoRegencies(deps))
		r.Get("/dashboard", handlers.DashboardGeneral(deps))

		r.Get("/roles", handlers.ListRoles(deps))
		r.Get("/schools", handlers.ListSchools(deps))
		r.Get("/schools/{schoolId}", handlers.SchoolGetPublic(deps))
		r.With(middleware.Auth(deps.JWTSecret)).Post("/schools", handlers.SchoolCreateByUser(deps))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.AuthRegister(deps))
			r.Post("/register-with-invite", handlers.AuthRegisterWithInvite(deps))
			r.Post("/login", handlers.AuthLogin(deps))
			r.Post("/hash-password", handlers.AdminGeneratePasswordHash(deps))
			r.Post("/admin/password-bypass", handlers.AuthAdminPasswordBypass(deps))
			r.Post("/admin/run-migrate", handlers.AuthRunMigrateBypass(deps))
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
		r.Post("/analytics/events", handlers.AnalyticsTrackEvent(deps))
		r.Post("/webhook/payment", handlers.PaymentWebhook(deps))

		r.Route("/programs", func(r chi.Router) {
			r.Get("/", handlers.ProgramsList(deps))
			r.Get("/{slug}", handlers.ProgramBySlug(deps))
		})

		r.Get("/packages", handlers.PackagesListLanding(deps))

		r.Route("/tryouts", func(r chi.Router) {
			r.Get("/", handlers.TryoutList(deps))
			r.Get("/open", handlers.TryoutListOpen(deps))
			r.Get("/{tryoutId}", handlers.TryoutGetByID(deps))
			r.Get("/{tryoutId}/leaderboard/top", handlers.TryoutLeaderboardTop(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Get("/{tryoutId}/leaderboard/rank", handlers.TryoutLeaderboardMyRank(deps))
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
			return deps.UserRepo.MustSetPasswordByID(ctx, id)
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
			r.Get("/tryouts/{tryoutId}/status", handlers.StudentTryoutStatus(deps))
			r.Get("/tryouts/history", handlers.StudentTryoutHistory(deps))
			r.Get("/tryouts/{tryoutId}/attempts/{attemptId}/paper", handlers.StudentTryoutAttemptPaper(deps))
			r.Get("/tryouts/{tryoutId}", handlers.StudentTryoutGetByID(deps))
			r.Get("/next-actions", handlers.StudentNextActions(deps))
			r.Post("/tryouts/{tryoutId}/register", handlers.StudentTryoutRegister(deps))
			r.Post("/tryouts/{tryoutId}/start", handlers.StudentTryoutStart(deps))
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
			r.Get("/students", handlers.TrainerStudentsList(deps))
			r.Get("/students/{studentId}", handlers.TrainerStudentGet(deps))
			r.Put("/students/{studentId}", handlers.TrainerStudentUpdate(deps))
			r.Route("/tryouts", func(r chi.Router) {
				r.Get("/", handlers.TrainerTryoutList(deps))
				r.Get("/{tryoutId}/paper", handlers.TrainerGuruTryoutPaperGet(deps))
				r.Put("/{tryoutId}/paper", handlers.TrainerGuruTryoutPaperPut(deps))
				r.Get("/{tryoutId}/analysis", handlers.TrainerTryoutAnalysis(deps))
				r.Get("/{tryoutId}/students", handlers.TrainerTryoutStudents(deps))
				r.Post("/{tryoutId}/auto-grade-submitted", handlers.TrainerPostTryoutAutoGradeSubmitted(deps))
				r.Get("/{tryoutId}/attempts/{attemptId}/ai-analysis", handlers.TrainerAttemptAIAnalysis(deps))
				r.Get("/{tryoutId}/attempts/{attemptId}/review", handlers.TrainerGetAttemptReview(deps))
				r.Post("/{tryoutId}/attempts/{attemptId}/auto-grade", handlers.TrainerPostAttemptAutoGrade(deps))
				r.Put("/{tryoutId}/attempts/{attemptId}/answers/{questionId}/review", handlers.TrainerPutAttemptAnswerReview(deps))
			})
		})

		r.Route("/guru", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Use(passwordGuard)
			r.Use(middleware.TrainerOnly())
			r.Get("/profile", handlers.TrainerProfileGet(deps))
			r.Put("/profile", handlers.TrainerProfileUpdate(deps))
			r.Put("/profile/password", handlers.GuruProfilePassword(deps))
			r.Get("/courses", handlers.GuruCoursesList(deps))
			r.Get("/students", handlers.GuruStudentsList(deps))
			r.Get("/earnings", handlers.GuruEarningsList(deps))
			r.Route("/tryouts", func(r chi.Router) {
				r.Get("/", handlers.TrainerTryoutList(deps))
				r.Get("/{tryoutId}/paper", handlers.TrainerGuruTryoutPaperGet(deps))
				r.Put("/{tryoutId}/paper", handlers.TrainerGuruTryoutPaperPut(deps))
				r.Get("/{tryoutId}/analysis", handlers.TrainerTryoutAnalysis(deps))
				r.Get("/{tryoutId}/students", handlers.TrainerTryoutStudents(deps))
				r.Post("/{tryoutId}/auto-grade-submitted", handlers.TrainerPostTryoutAutoGradeSubmitted(deps))
				r.Get("/{tryoutId}/attempts/{attemptId}/ai-analysis", handlers.TrainerAttemptAIAnalysis(deps))
				r.Get("/{tryoutId}/attempts/{attemptId}/review", handlers.TrainerGetAttemptReview(deps))
				r.Post("/{tryoutId}/attempts/{attemptId}/auto-grade", handlers.TrainerPostAttemptAutoGrade(deps))
				r.Put("/{tryoutId}/attempts/{attemptId}/answers/{questionId}/review", handlers.TrainerPutAttemptAnswerReview(deps))
			})
		})

		r.Route("/levels", func(r chi.Router) {
			r.Get("/", handlers.AdminListLevels(deps))
			r.Get("/{levelId}", handlers.LevelWithSubjects(deps))
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Use(middleware.AdminOnly())
			r.Use(middleware.AdminAuditLog(func(ctx context.Context, userID, role, method, path string, statusCode int, duration time.Duration, requestID string) error {
				if deps.DB == nil || userID == "" {
					return nil
				}
				_, err := deps.DB.Exec(ctx, `
					INSERT INTO admin_audit_logs (
						admin_user_id, role, method, path, status_code, duration_ms, request_id, created_at
					) VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, NOW())
				`, userID, role, method, path, statusCode, int(duration.Milliseconds()), requestID)
				return err
			}))
			r.With(middleware.RequirePermission("admin.overview.read")).Get("/overview", handlers.AdminOverview(deps))
			r.With(middleware.RequirePermission("users.manage")).Get("/users", handlers.AdminListUsers(deps))
			r.With(middleware.RequirePermission("users.manage")).Post("/users", handlers.AdminCreateUser(deps))
			r.With(middleware.RequirePermission("users.manage")).Get("/users/{userId}", handlers.AdminGetUser(deps))
			r.With(middleware.RequirePermission("users.manage")).Put("/users/{userId}", handlers.AdminUpdateUser(deps))
			r.With(middleware.RequirePermission("users.manage")).Post("/tools/hash-password", handlers.AdminGeneratePasswordHash(deps))
			r.With(middleware.RequirePermission("courses.manage")).Get("/courses", handlers.AdminListCourses(deps))
			r.With(middleware.RequirePermission("courses.manage")).Get("/courses/{courseId}", handlers.AdminGetCourse(deps))
			r.With(middleware.RequirePermission("courses.manage")).Post("/courses", handlers.AdminCreateCourse(deps))
			r.With(middleware.RequirePermission("courses.manage")).Put("/courses/{courseId}", handlers.AdminUpdateCourse(deps))
			r.With(middleware.RequirePermission("courses.manage")).Get("/courses/{courseId}/enrollments", handlers.AdminListEnrollments(deps))
			r.With(middleware.RequirePermission("courses.manage")).Get("/courses/{courseId}/contents", handlers.AdminListCourseContents(deps))
			r.With(middleware.RequirePermission("courses.manage")).Post("/courses/{courseId}/contents", handlers.AdminCreateCourseContent(deps))
			r.With(middleware.RequirePermission("courses.manage")).Put("/courses/{courseId}/contents/{contentId}", handlers.AdminUpdateCourseContent(deps))
			r.With(middleware.RequirePermission("courses.manage")).Delete("/courses/{courseId}/contents/{contentId}", handlers.AdminDeleteCourseContent(deps))
			r.With(middleware.RequirePermission("courses.manage")).Get("/courses/{courseId}/manage", handlers.AdminCourseManageGet(deps))
			r.With(middleware.RequirePermission("courses.manage")).Put("/courses/{courseId}/linked-packages", handlers.AdminCourseLinkedPackagesPut(deps))
			r.With(middleware.RequirePermission("courses.manage")).Put("/courses/{courseId}/linked-tryouts", handlers.AdminCourseLinkedTryoutsPut(deps))
			r.With(middleware.RequirePermission("payments.manage")).Get("/payments", handlers.AdminListPayments(deps))
			r.With(middleware.RequirePermission("payments.manage")).Get("/transactions/{orderId}", handlers.AdminTransactionDetail(deps))
			r.With(middleware.RequirePermission("payments.manage")).Post("/payments", handlers.AdminCreatePayment(deps))
			r.With(middleware.RequirePermission("payments.manage")).Put("/payments/{paymentId}", handlers.AdminConfirmPayment(deps))
			r.With(middleware.RequirePermission("payments.manage")).Post("/payments/{paymentId}/confirm", handlers.AdminConfirmPaymentByAction(deps))
			r.With(middleware.RequirePermission("payments.manage")).Post("/payments/{paymentId}/reject", handlers.AdminRejectPaymentByAction(deps))
			r.With(middleware.RequirePermission("orders.verify")).Put("/orders/{orderId}/verify", handlers.AdminVerifyOrder(deps))
			r.With(middleware.RequirePermission("analytics.read")).Get("/analytics/summary", handlers.AdminAnalyticsSummary(deps))
			r.With(middleware.RequirePermission("analytics.read")).Get("/analytics/visitors", handlers.AdminAnalyticsVisitors(deps))
			r.With(middleware.RequirePermission("admin.audit.read")).Get("/audit-logs", handlers.AdminAuditLogsList(deps))
			r.Route("/error-logs", func(r chi.Router) {
				r.With(middleware.RequirePermission("errors.read")).Get("/analytics", handlers.AdminErrorLogsAnalytics(deps))
				r.With(middleware.RequirePermission("errors.read")).Get("/", handlers.AdminErrorLogsList(deps))
				r.With(middleware.RequirePermission("errors.read")).Get("/{id}", handlers.AdminErrorLogGet(deps))
				r.With(middleware.RequirePermission("errors.manage")).Patch("/{id}", handlers.AdminErrorLogPatch(deps))
			})
			r.With(middleware.RequirePermission("reports.read")).Get("/reports/monthly", handlers.AdminReportMonthly(deps))
			r.With(middleware.RequirePermission("reports.read")).Get("/reports/courses/{courseId}", handlers.AdminCourseReport(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Get("/roles", handlers.AdminListRoles(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Post("/roles", handlers.AdminCreateRole(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Get("/roles/{id}", handlers.AdminGetRole(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Put("/roles/{id}", handlers.AdminUpdateRole(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Delete("/roles/{id}", handlers.AdminDeleteRole(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Get("/schools", handlers.AdminListSchools(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Post("/schools", handlers.AdminCreateSchool(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Get("/schools/{id}", handlers.AdminGetSchool(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Put("/schools/{id}", handlers.AdminUpdateSchool(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Delete("/schools/{id}", handlers.AdminDeleteSchool(deps))
			// Alias for frontend master-data UI: /api/v1/admin/master-data/sekolah
			r.Route("/master-data", func(r chi.Router) {
				r.Route("/sekolah", func(r chi.Router) {
					r.With(middleware.RequirePermission("master-data.manage")).Get("/", handlers.AdminListSchools(deps))
					r.With(middleware.RequirePermission("master-data.manage")).Post("/", handlers.AdminCreateSchool(deps))
					r.With(middleware.RequirePermission("master-data.manage")).Get("/{id}", handlers.AdminGetSchool(deps))
					r.With(middleware.RequirePermission("master-data.manage")).Put("/{id}", handlers.AdminUpdateSchool(deps))
					r.With(middleware.RequirePermission("master-data.manage")).Delete("/{id}", handlers.AdminDeleteSchool(deps))
				})
			})
			r.With(middleware.RequirePermission("master-data.manage")).Get("/settings", handlers.AdminListSettings(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Post("/settings", handlers.AdminCreateSetting(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Get("/settings/{id}", handlers.AdminGetSetting(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Put("/settings/{id}", handlers.AdminUpdateSetting(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Delete("/settings/{id}", handlers.AdminDeleteSetting(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Get("/events", handlers.AdminListEvents(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Post("/events", handlers.AdminCreateEvent(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Get("/events/{id}", handlers.AdminGetEvent(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Put("/events/{id}", handlers.AdminUpdateEvent(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Delete("/events/{id}", handlers.AdminDeleteEvent(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Get("/subjects", handlers.AdminListSubjects(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Post("/subjects", handlers.AdminCreateSubject(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Get("/subjects/{id}", handlers.AdminGetSubject(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Put("/subjects/{id}", handlers.AdminUpdateSubject(deps))
			r.With(middleware.RequirePermission("master-data.manage")).Delete("/subjects/{id}", handlers.AdminDeleteSubject(deps))
			r.Route("/levels", func(r chi.Router) {
				r.With(middleware.RequirePermission("master-data.manage")).Get("/", handlers.AdminListLevels(deps))
				r.With(middleware.RequirePermission("master-data.manage")).Post("/", handlers.AdminCreateLevel(deps))
				r.With(middleware.RequirePermission("master-data.manage")).Get("/{id}", handlers.AdminGetLevel(deps))
				r.With(middleware.RequirePermission("master-data.manage")).Get("/{id}/subjects", handlers.LevelWithSubjects(deps))
				r.With(middleware.RequirePermission("master-data.manage")).Put("/{id}", handlers.AdminUpdateLevel(deps))
				r.With(middleware.RequirePermission("master-data.manage")).Delete("/{id}", handlers.AdminDeleteLevel(deps))
			})
			r.Route("/tryouts", func(r chi.Router) {
				r.With(middleware.RequirePermission("tryouts.manage")).Get("/", handlers.AdminListTryouts(deps))
				r.With(middleware.RequirePermission("tryouts.manage")).Post("/", handlers.AdminCreateTryout(deps))
				r.Route("/{tryoutId}", func(r chi.Router) {
					r.With(middleware.RequirePermission("tryouts.manage")).Get("/", handlers.AdminGetTryout(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Put("/", handlers.AdminUpdateTryout(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Delete("/", handlers.AdminDeleteTryout(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Get("/analysis", handlers.AdminGetTryoutAnalysis(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Get("/students", handlers.AdminListTryoutStudents(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Post("/auto-grade-submitted", handlers.AdminPostTryoutAutoGradeSubmitted(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Get("/questions", handlers.AdminListQuestions(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Get("/questions/stats", handlers.AdminGetTryoutQuestionStatsBulk(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Post("/questions", handlers.AdminCreateQuestion(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Get("/questions/{questionId}", handlers.AdminGetQuestion(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Put("/questions/{questionId}", handlers.AdminUpdateQuestion(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Delete("/questions/{questionId}", handlers.AdminDeleteQuestion(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Get("/questions/{questionId}/stats", handlers.AdminGetQuestionStats(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Get("/attempts/{attemptId}/ai-analysis", handlers.AdminGetAttemptAIAnalysis(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Get("/attempts/{attemptId}/review", handlers.AdminGetAttemptReview(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Post("/attempts/{attemptId}/auto-grade", handlers.AdminPostAttemptAutoGrade(deps))
					r.With(middleware.RequirePermission("tryouts.manage")).Put("/attempts/{attemptId}/answers/{questionId}/review", handlers.AdminPutAttemptAnswerReview(deps))
				})
			})
			r.With(middleware.RequirePermission("certificates.issue")).Post("/certificates", handlers.AdminIssueCertificate(deps))
			r.Route("/landing", func(r chi.Router) {
				r.With(middleware.RequirePermission("landing.manage")).Get("/site-settings", handlers.AdminLandingSiteSettingsList(deps))
				r.With(middleware.RequirePermission("landing.manage")).Put("/site-settings", handlers.AdminLandingSiteSettingsUpsert(deps))
				r.With(middleware.RequirePermission("landing.manage")).Get("/resources/{resource}", handlers.AdminLandingResourceList(deps))
				r.With(middleware.RequirePermission("landing.manage")).Post("/resources/{resource}", handlers.AdminLandingResourceCreate(deps))
				r.With(middleware.RequirePermission("landing.manage")).Put("/resources/{resource}/{id}", handlers.AdminLandingResourceUpdate(deps))
				r.With(middleware.RequirePermission("landing.manage")).Delete("/resources/{resource}/{id}", handlers.AdminLandingResourceDelete(deps))
				r.With(middleware.RequirePermission("landing.manage")).Get("/packages", handlers.AdminLandingPackagesList(deps))
				r.With(middleware.RequirePermission("landing.manage")).Post("/packages", handlers.AdminLandingPackageCreate(deps))
				r.With(middleware.RequirePermission("landing.manage")).Put("/packages/{id}", handlers.AdminLandingPackageUpdate(deps))
				r.With(middleware.RequirePermission("landing.manage")).Delete("/packages/{id}", handlers.AdminLandingPackageDelete(deps))
			})
		})
	}

	r.Route("/api/v1", registerV1)

	return r
}
