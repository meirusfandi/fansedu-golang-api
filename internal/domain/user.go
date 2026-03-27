package domain

import (
	"strings"
	"time"
)

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

// IsStudentRoleCode reports whether code is the student enum (JWT / users.role).
func IsStudentRoleCode(code string) bool {
	return strings.TrimSpace(code) == UserRoleStudent
}

// IsTeachingStaffRoleCode: guru, instructor (enum legacy), atau trainer — akses portal pengajar.
func IsTeachingStaffRoleCode(code string) bool {
	switch strings.TrimSpace(code) {
	case UserRoleGuru, UserRoleTrainer, "instructor":
		return true
	default:
		return false
	}
}

// DisplayRoleForAPI maps stored user_role ke label response JSON; guru dan legacy enum instructor ditampilkan sebagai "guru".
func DisplayRoleForAPI(code string) string {
	c := strings.TrimSpace(code)
	if c == UserRoleGuru || c == "instructor" {
		return UserRoleGuru
	}
	return c
}

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
