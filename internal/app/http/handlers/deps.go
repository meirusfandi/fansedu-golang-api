package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

type Deps struct {
	DB        *pgxpool.Pool
	JWTSecret []byte

	AuthService      service.AuthService
	TryoutService    service.TryoutService
	AttemptService   service.AttemptService
	DashboardService service.DashboardService
	AdminService     service.AdminService
	CourseService    service.CourseService

	UserRepo                 repo.UserRepo
	QuestionRepo             repo.QuestionRepo
	AttemptAnswerRepo        repo.AttemptAnswerRepo
	CertificateRepo          repo.CertificateRepo
	TryoutRegistrationRepo   repo.TryoutRegistrationRepo

	RoleRepo    repo.RoleRepo
	SchoolRepo  repo.SchoolRepo
	SettingRepo repo.SettingRepo
	EventRepo   repo.EventRepo
	SubjectRepo repo.SubjectRepo
	LevelRepo   repo.LevelRepo
}
