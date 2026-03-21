package domain

import "time"

const (
	UserRoleAdmin        = "admin"
	UserRoleSuperAdmin   = "super_admin"
	UserRoleFinanceAdmin = "finance_admin"
	UserRoleAcademicAdmin = "academic_admin"
	UserRoleContentAdmin = "content_admin"
	UserRoleStudent      = "student"
	UserRoleGuru         = "guru"
	UserRoleTrainer      = "trainer" // nanti dibuat oleh admin
)

type User struct {
	ID              string
	Email           string
	PasswordHash    string
	Name            string
	Role            string
	AvatarURL       *string
	SchoolID        *string
	SubjectID       *string
	Phone           *string
	Whatsapp        *string
	ClassLevel      *string
	City            *string
	Province        *string
	Gender          *string
	BirthDate       *time.Time
	Bio             *string
	ParentName      *string
	ParentPhone     *string
	Instagram       *string
	EmailVerified   bool
	EmailVerifiedAt *time.Time
	MustSetPassword bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
