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
	"github.com/redis/go-redis/v9"

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

	rdb, err := config.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	if rdb != nil {
		defer func() { _ = rdb.Close() }()
		if err := config.PingRedis(ctx, rdb); err != nil {
			log.Printf("warning: redis ping failed (%v) — geo & leaderboard cache may not work", err)
		} else {
			log.Printf("redis: connected")
		}
	}

	deps := buildDeps(pool, cfg, rdb)
	router := httpapi.NewRouter(deps)

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

func buildDeps(pool *pgxpool.Pool, cfg config.Config, rdb *redis.Client) *handlers.Deps {
	ttl := time.Duration(cfg.GeoCacheTTLSeconds) * time.Second
	geoSvc := service.NewGeoService(rdb, cfg.GeoUpstreamBaseURL, ttl)
	jwtSecret := []byte(cfg.JWTSecret)
	openAIAPIKey := cfg.OpenAIAPIKey
	appURL := cfg.AppURL
	adminPasswordBypassKey := cfg.AdminPasswordBypassKey
	migrateBypassKey := cfg.MigrateBypassKey

	if pool == nil {
		return &handlers.Deps{
			Redis:                  rdb,
			SchoolListCacheTTL:     time.Duration(cfg.SchoolListCacheSeconds) * time.Second,
			PackagesListCacheTTL:   time.Duration(cfg.PackagesListCacheSeconds) * time.Second,
			GeoService:             geoSvc,
			JWTSecret:              jwtSecret,
			AdminPasswordBypassKey: adminPasswordBypassKey,
			MigrateBypassKey:       migrateBypassKey,
		}
	}

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
	appErrLogRepo := repo.NewApplicationErrorLogRepo(pool)
	landingPackageRepo := repo.NewLandingPackageRepoPg(pool)
	courseAdminLinkRepo := repo.NewCourseAdminLinkRepo(pool)
	var mailer mail.Mailer = mail.NewLogMailer()
	if cfg.SMTPPassword != "" {
		smtpUser := cfg.SMTPUser
		if smtpUser == "" {
			smtpUser = cfg.SMTPFrom
		}
		m, err := mail.NewSMTPMailer(mail.SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			User:     smtpUser,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
		})
		if err != nil {
			log.Printf("warning: smtp mailer: %v — using log mailer", err)
		} else {
			mailer = m
			log.Printf("mail: SMTP enabled host=%s port=%d from=%s", cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)
		}
	} else if cfg.IsProduction() {
		log.Printf("warning: BREVO_SMTP_KEY/SMTP_PASSWORD not set — transactional email hanya di-log (LogMailer)")
	}
	if appURL == "" {
		appURL = "http://localhost:5173"
	}
	checkoutService := service.NewCheckoutService(courseRepo, landingPackageRepo, userRepo, orderRepo, orderItemRepo, paymentRepo, enrollmentRepo, promoRepo, mailer, userInviteRepo, appURL)

	return &handlers.Deps{
		DB:                      pool,
		Redis:                   rdb,
		SchoolListCacheTTL:      time.Duration(cfg.SchoolListCacheSeconds) * time.Second,
		PackagesListCacheTTL:    time.Duration(cfg.PackagesListCacheSeconds) * time.Second,
		JWTSecret:               jwtSecret,
		AdminPasswordBypassKey: adminPasswordBypassKey,
		MigrateBypassKey:       migrateBypassKey,
		GeoService:             geoSvc,
		AuthService:            authService,
		TryoutService:          tryoutService,
		AttemptService:         attemptService,
		DashboardService:       dashboardService,
		AdminService:           adminService,
		CourseService:          courseService,
		TrainerService:         trainerService,
		CheckoutService:        checkoutService,
		UserRepo:               userRepo,
		QuestionRepo:           questionRepo,
		AttemptAnswerRepo:      attemptAnswerRepo,
		CertificateRepo:        certificateRepo,
		RoleRepo:               roleRepo,
		SchoolRepo:             schoolRepo,
		SettingRepo:            settingRepo,
		EventRepo:              eventRepo,
		SubjectRepo:            subjectRepo,
		LevelRepo:              levelRepo,
		LandingPackageRepo:     landingPackageRepo,
		CourseAdminLinkRepo:    courseAdminLinkRepo,
		TryoutRegistrationRepo: tryoutRegistrationRepo,
		EnrollmentRepo:         enrollmentRepo,
		CourseRepo:             courseRepo,
		CourseContentRepo:      courseContentRepo,
		PaymentRepo:            paymentRepo,
		OrderRepo:              orderRepo,
		OrderItemRepo:          orderItemRepo,
		PromoRepo:              promoRepo,
		AnalyticsRepo:          analyticsRepo,
		ApplicationErrorLogRepo: appErrLogRepo,
		NotificationRepo:       notificationRepo,
		TrainerRepo:            trainerRepo,
		CourseMessageRepo:      courseMessageRepo,
		CourseDiscussionRepo:      courseDiscussionRepo,
		CourseDiscussionReplyRepo: courseDiscussionReplyRepo,
	}
}
