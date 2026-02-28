package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "github.com/meirusfandi/fansedu-golang-api/internal/app/http"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/handlers"
	"github.com/meirusfandi/fansedu-golang-api/internal/config"
	"github.com/meirusfandi/fansedu-golang-api/internal/db"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config.LoadEnvFile()
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		var err error
		pool, err = db.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("db connect: %v", err)
		}
		defer pool.Close()
	} else if cfg.IsProduction() {
		log.Fatal("production: DATABASE_URL is required")
	}

	var router http.Handler
	if pool == nil {
		router = httpapi.NewRouter(nil)
	} else {
		deps := buildDeps(pool, []byte(cfg.JWTSecret))
		router = httpapi.NewRouter(deps)
	}

	srv := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("http listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Printf("shutdown complete")
}

func buildDeps(pool *pgxpool.Pool, jwtSecret []byte) *handlers.Deps {
	userRepo := repo.NewUserRepo(pool)
	tryoutRepo := repo.NewTryoutRepo(pool)
	questionRepo := repo.NewQuestionRepo(pool)
	attemptRepo := repo.NewAttemptRepo(pool)
	attemptAnswerRepo := repo.NewAttemptAnswerRepo(pool)
	feedbackRepo := repo.NewFeedbackRepo(pool)
	courseRepo := repo.NewCourseRepo(pool)
	enrollmentRepo := repo.NewEnrollmentRepo(pool)
	certificateRepo := repo.NewCertificateRepo(pool)

	authService := service.NewAuthService(userRepo, jwtSecret)
	tryoutService := service.NewTryoutService(tryoutRepo)
	attemptService := service.NewAttemptService(attemptRepo, attemptAnswerRepo, feedbackRepo, questionRepo, tryoutRepo)
	dashboardService := service.NewDashboardService(attemptRepo, tryoutRepo, feedbackRepo)
	adminService := service.NewAdminService(
		tryoutRepo, questionRepo, courseRepo, enrollmentRepo, certificateRepo,
		func(ctx context.Context) (int, error) { return userRepo.CountByRole(ctx, "student") },
		attemptRepo.AvgScoreSubmitted,
		certificateRepo.Count,
	)
	courseService := service.NewCourseService(courseRepo, enrollmentRepo)

	return &handlers.Deps{
		DB:                 pool,
		JWTSecret:          jwtSecret,
		AuthService:        authService,
		TryoutService:      tryoutService,
		AttemptService:     attemptService,
		DashboardService:   dashboardService,
		AdminService:       adminService,
		CourseService:      courseService,
		QuestionRepo:       questionRepo,
		AttemptAnswerRepo:  attemptAnswerRepo,
		CertificateRepo:    certificateRepo,
	}
}
