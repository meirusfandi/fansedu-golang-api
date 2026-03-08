package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	httpapi "github.com/meirusfandi/fansedu-golang-api/internal/app/http"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/handlers"
	"github.com/meirusfandi/fansedu-golang-api/internal/ai"
	"github.com/meirusfandi/fansedu-golang-api/internal/config"
	"github.com/meirusfandi/fansedu-golang-api/internal/db"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

func main() {
	envFlag := flag.String("env", "", "environment: dev|development (local), prod|production (server). Overrides ENV.")
	flag.Parse()

	if v := strings.TrimSpace(*envFlag); v != "" {
		switch strings.ToLower(v) {
		case "dev", "development":
			os.Setenv("ENV", config.EnvDevelopment)
		case "prod", "production":
			os.Setenv("ENV", config.EnvProduction)
		default:
			log.Fatalf("invalid -env=%q: use dev|development or prod|production", v)
		}
	}

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
		deps := buildDeps(pool, []byte(cfg.JWTSecret), cfg.OpenAIAPIKey)
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

func buildDeps(pool *pgxpool.Pool, jwtSecret []byte, openAIAPIKey string) *handlers.Deps {
	userRepo := repo.NewUserRepo(pool)
	tryoutRepo := repo.NewTryoutRepo(pool)
	questionRepo := repo.NewQuestionRepo(pool)
	attemptRepo := repo.NewAttemptRepo(pool)
	attemptAnswerRepo := repo.NewAttemptAnswerRepo(pool)
	feedbackRepo := repo.NewFeedbackRepo(pool)
	courseRepo := repo.NewCourseRepo(pool)
	enrollmentRepo := repo.NewEnrollmentRepo(pool)
	certificateRepo := repo.NewCertificateRepo(pool)
	courseContentRepo := repo.NewCourseContentRepo(pool)
	paymentRepo := repo.NewPaymentRepo(pool)
	roleRepo := repo.NewRoleRepo(pool)
	schoolRepo := repo.NewSchoolRepo(pool)
	settingRepo := repo.NewSettingRepo(pool)
	eventRepo := repo.NewEventRepo(pool)
	subjectRepo := repo.NewSubjectRepo(pool)
	levelRepo := repo.NewLevelRepo(pool)
	tryoutRegistrationRepo := repo.NewTryoutRegistrationRepo(pool)
	trainerRepo := repo.NewTrainerRepo(pool)
	notificationRepo := repo.NewNotificationRepo(pool)
	courseMessageRepo := repo.NewCourseMessageRepo(pool)
	courseDiscussionRepo := repo.NewCourseDiscussionRepo(pool)
	courseDiscussionReplyRepo := repo.NewCourseDiscussionReplyRepo(pool)

	var feedbackGen ai.FeedbackGenerator
	if openAIAPIKey != "" {
		feedbackGen = ai.NewOpenAIFeedbackGenerator(openAIAPIKey)
	} else {
		feedbackGen = ai.NewFallbackFeedbackGenerator()
	}

	authService := service.NewAuthService(userRepo, jwtSecret)
	tryoutService := service.NewTryoutService(tryoutRepo, tryoutRegistrationRepo)
	attemptService := service.NewAttemptService(attemptRepo, attemptAnswerRepo, feedbackRepo, questionRepo, tryoutRepo, feedbackGen)
	dashboardService := service.NewDashboardService(userRepo, attemptRepo, tryoutRepo, feedbackRepo, questionRepo, attemptAnswerRepo)
	adminService := service.NewAdminService(
		userRepo,
		tryoutRepo, questionRepo, courseRepo, enrollmentRepo,
		courseContentRepo, paymentRepo,
		certificateRepo,
		attemptRepo,
		attemptAnswerRepo,
		func(ctx context.Context) (int, error) { return userRepo.CountByRole(ctx, "student") },
		attemptRepo.AvgScoreSubmitted,
		certificateRepo.Count,
	)
	courseService := service.NewCourseService(courseRepo, enrollmentRepo)
	trainerService := service.NewTrainerService(userRepo, trainerRepo)

	return &handlers.Deps{
		DB:                 pool,
		JWTSecret:          jwtSecret,
		AuthService:        authService,
		TryoutService:      tryoutService,
		AttemptService:     attemptService,
		DashboardService:   dashboardService,
		AdminService:       adminService,
		CourseService:      courseService,
		TrainerService:     trainerService,
		UserRepo:           userRepo,
		QuestionRepo:       questionRepo,
		AttemptAnswerRepo:  attemptAnswerRepo,
		CertificateRepo:          certificateRepo,
		RoleRepo:                 roleRepo,
		SchoolRepo:               schoolRepo,
		SettingRepo:              settingRepo,
		EventRepo:                eventRepo,
		SubjectRepo:              subjectRepo,
		LevelRepo:                levelRepo,
		TryoutRegistrationRepo:   tryoutRegistrationRepo,
		EnrollmentRepo:           enrollmentRepo,
		CourseRepo:               courseRepo,
		PaymentRepo:              paymentRepo,
		NotificationRepo:        notificationRepo,
		CourseMessageRepo:        courseMessageRepo,
		CourseDiscussionRepo:      courseDiscussionRepo,
		CourseDiscussionReplyRepo: courseDiscussionReplyRepo,
	}
}
