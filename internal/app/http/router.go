package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/handlers"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
)

func NewRouter(deps *handlers.Deps) http.Handler {
	r := chi.NewRouter()

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

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.AuthRegister(deps))
			r.Post("/login", handlers.AuthLogin(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/logout", handlers.AuthLogout(deps))
			r.Post("/forgot-password", handlers.AuthForgotPassword(deps))
			r.Post("/reset-password", handlers.AuthResetPassword(deps))
		})

		r.Route("/tryouts", func(r chi.Router) {
			r.Get("/open", handlers.TryoutListOpen(deps))
			r.Get("/{tryoutId}", handlers.TryoutGetByID(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/{tryoutId}/start", handlers.TryoutStart(deps))
		})

		r.Route("/attempts", func(r chi.Router) {
			r.With(middleware.Auth(deps.JWTSecret)).Get("/{attemptId}/questions", handlers.AttemptGetQuestions(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Put("/{attemptId}/answers/{questionId}", handlers.AttemptPutAnswer(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/{attemptId}/submit", handlers.AttemptSubmit(deps))
		})

		r.Route("/student", func(r chi.Router) {
			r.With(middleware.Auth(deps.JWTSecret)).Get("/dashboard", handlers.DashboardStudent(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Get("/attempts", handlers.AttemptListByUser(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Get("/attempts/{attemptId}", handlers.AttemptGetByID(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Get("/certificates", handlers.CertificateList(deps))
		})

		r.Route("/courses", func(r chi.Router) {
			r.Get("/", handlers.CourseList(deps))
			r.With(middleware.Auth(deps.JWTSecret)).Post("/{courseId}/enroll", handlers.CourseEnroll(deps))
		})

		r.Get("/levels", handlers.AdminListLevels(deps))
		r.Get("/levels/{id}", handlers.LevelWithSubjects(deps))

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
			r.Get("/reports/monthly", handlers.AdminReportMonthly(deps))
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
			r.Post("/tryouts", handlers.AdminCreateTryout(deps))
			r.Put("/tryouts/{tryoutId}", handlers.AdminUpdateTryout(deps))
			r.Delete("/tryouts/{tryoutId}", handlers.AdminDeleteTryout(deps))
			r.Post("/tryouts/{tryoutId}/questions", handlers.AdminCreateQuestion(deps))
			r.Put("/questions/{questionId}", handlers.AdminUpdateQuestion(deps))
			r.Delete("/questions/{questionId}", handlers.AdminDeleteQuestion(deps))
			r.Post("/certificates", handlers.AdminIssueCertificate(deps))
		})
	}

	r.Route("/api/v1", registerV1)

	return r
}
