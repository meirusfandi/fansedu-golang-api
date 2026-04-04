package handlers

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

type Deps struct {
	DB                     *pgxpool.Pool
	Redis                  *redis.Client
	SchoolListCacheTTL     time.Duration
	PackagesListCacheTTL   time.Duration
	JWTSecret              []byte
	AdminPasswordBypassKey string
	MigrateBypassKey       string

	AuthService      service.AuthService
	TryoutService    service.TryoutService
	AttemptService   service.AttemptService
	DashboardService service.DashboardService
	AdminService          service.AdminService
	CourseService         service.CourseService
	CourseProgramService  service.CourseProgramService
	LearningService       service.LearningService
	TrainerService   service.TrainerService
	CheckoutService  service.CheckoutService
	GeoService       service.GeoService

	UserRepo                 repo.UserRepo
	QuestionRepo             repo.QuestionRepo
	AttemptAnswerRepo        repo.AttemptAnswerRepo
	CertificateRepo          repo.CertificateRepo
	TryoutRegistrationRepo   repo.TryoutRegistrationRepo

	EnrollmentRepo            repo.EnrollmentRepo
	CourseRepo                repo.CourseRepo
	CourseContentRepo         repo.CourseContentRepo
	PaymentRepo               repo.PaymentRepo
	OrderRepo                 repo.OrderRepo
	OrderItemRepo             repo.OrderItemRepo
	PromoRepo                 repo.PromoRepo
	AnalyticsRepo             repo.AnalyticsRepo
	NotificationRepo          repo.NotificationRepo
	TrainerRepo               repo.TrainerRepo
	CourseMessageRepo         repo.CourseMessageRepo
	CourseDiscussionRepo      repo.CourseDiscussionRepo
	CourseDiscussionReplyRepo repo.CourseDiscussionReplyRepo

	RoleRepo    repo.RoleRepo
	SchoolRepo  repo.SchoolRepo
	SettingRepo repo.SettingRepo
	EventRepo   repo.EventRepo
	SubjectRepo       repo.SubjectRepo
	LevelRepo         repo.LevelRepo
	LandingPackageRepo     repo.LandingPackageRepo
	CourseAdminLinkRepo    repo.CourseAdminLinkRepo

	ApplicationErrorLogRepo repo.ApplicationErrorLogRepo
}
