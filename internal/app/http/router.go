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

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", handlers.Health())
		if deps == nil {
			return
		}

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

		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Auth(deps.JWTSecret))
			r.Use(middleware.AdminOnly())
			r.Get("/overview", handlers.AdminOverview(deps))
			r.Post("/tryouts", handlers.AdminCreateTryout(deps))
			r.Put("/tryouts/{tryoutId}", handlers.AdminUpdateTryout(deps))
			r.Delete("/tryouts/{tryoutId}", handlers.AdminDeleteTryout(deps))
			r.Post("/tryouts/{tryoutId}/questions", handlers.AdminCreateQuestion(deps))
			r.Put("/questions/{questionId}", handlers.AdminUpdateQuestion(deps))
			r.Delete("/questions/{questionId}", handlers.AdminDeleteQuestion(deps))
			r.Post("/courses", handlers.AdminCreateCourse(deps))
			r.Get("/courses/{courseId}/enrollments", handlers.AdminListEnrollments(deps))
			r.Post("/certificates", handlers.AdminIssueCertificate(deps))
		})
	})

	return r
}
