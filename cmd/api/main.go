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
	"github.com/meirusfandi/fansedu-golang-api/internal/mail"
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
		deps := buildDeps(pool, []byte(cfg.JWTSecret), cfg.OpenAIAPIKey, cfg.AppURL, cfg.AdminPasswordBypassKey)
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

func buildDeps(pool *pgxpool.Pool, jwtSecret []byte, openAIAPIKey, appURL, adminPasswordBypassKey string) *handlers.Deps {
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
	emailVerificationTokenRepo := repo.NewEmailVerificationTokenRepo(pool)

	var feedbackGen ai.FeedbackGenerator
	if openAIAPIKey != "" {
		feedbackGen = ai.NewOpenAIFeedbackGenerator(openAIAPIKey)
	} else {
		feedbackGen = ai.NewFallbackFeedbackGenerator()
	}

	userInviteRepo := repo.NewUserInviteRepo(pool)
	authService := service.NewAuthService(userRepo, emailVerificationTokenRepo, userInviteRepo, jwtSecret)
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
		feedbackRepo,
		feedbackGen,
		func(ctx context.Context) (int, error) { return userRepo.CountByRole(ctx, "student") },
		attemptRepo.AvgScoreSubmitted,
		certificateRepo.Count,
	)
	courseService := service.NewCourseService(courseRepo, enrollmentRepo)
	trainerService := service.NewTrainerService(userRepo, trainerRepo)
	orderRepo := repo.NewOrderRepo(pool)
	orderItemRepo := repo.NewOrderItemRepo(pool)
	promoRepo := repo.NewPromoRepo(pool)
	analyticsRepo := repo.NewAnalyticsRepo(pool)
	mailer := mail.NewLogMailer()
	if appURL == "" {
		appURL = "http://localhost:5173"
	}
	checkoutService := service.NewCheckoutService(courseRepo, userRepo, orderRepo, orderItemRepo, paymentRepo, enrollmentRepo, promoRepo, mailer, userInviteRepo, appURL)
	landingPackageRepo := repo.NewLandingPackageRepoPg(pool)

	return &handlers.Deps{
		DB:                 pool,
		JWTSecret:          jwtSecret,
		AdminPasswordBypassKey: adminPasswordBypassKey,
		AuthService:        authService,
		TryoutService:      tryoutService,
		AttemptService:     attemptService,
		DashboardService:   dashboardService,
		AdminService:       adminService,
		CourseService:      courseService,
		TrainerService:     trainerService,
		CheckoutService:    checkoutService,
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
		LandingPackageRepo:       landingPackageRepo,
		TryoutRegistrationRepo:   tryoutRegistrationRepo,
		EnrollmentRepo:           enrollmentRepo,
		CourseRepo:               courseRepo,
		CourseContentRepo:        courseContentRepo,
		PaymentRepo:              paymentRepo,
		OrderRepo:                orderRepo,
		OrderItemRepo:            orderItemRepo,
		PromoRepo:                promoRepo,
		AnalyticsRepo:            analyticsRepo,
		NotificationRepo:        notificationRepo,
		TrainerRepo:             trainerRepo,
		CourseMessageRepo:        courseMessageRepo,
		CourseDiscussionRepo:      courseDiscussionRepo,
		CourseDiscussionReplyRepo: courseDiscussionReplyRepo,
	}
}
