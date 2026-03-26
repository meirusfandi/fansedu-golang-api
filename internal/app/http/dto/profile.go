package dto

// UserProfileResponse: GET /auth/me, /student/profile, /trainer/profile — seluruh field JSON camelCase.
type UserProfileResponse struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Email           string         `json:"email"`
	Role            string         `json:"role"`
	RoleSlug        string         `json:"roleSlug,omitempty"`
	RoleCode        string         `json:"roleCode,omitempty"`
	MustSetPassword bool           `json:"mustSetPassword"`
	EmailVerified   bool           `json:"emailVerified"`
	AvatarURL       *string        `json:"avatarUrl,omitempty"`
	Phone           *string        `json:"phone,omitempty"`
	Whatsapp        *string        `json:"whatsapp,omitempty"`
	ClassLevel      *string        `json:"classLevel,omitempty"`
	City            *string        `json:"city,omitempty"`
	Province        *string        `json:"province,omitempty"`
	Gender          *string        `json:"gender,omitempty"`
	BirthDate       *string        `json:"birthDate,omitempty"` // YYYY-MM-DD
	Bio             *string        `json:"bio,omitempty"`
	ParentName      *string        `json:"parentName,omitempty"`
	ParentPhone     *string        `json:"parentPhone,omitempty"`
	Instagram       *string        `json:"instagram,omitempty"`
	SchoolID        *string        `json:"schoolId,omitempty"`
	SchoolName      string         `json:"schoolName,omitempty"`
	SubjectID       *string        `json:"subjectId,omitempty"`
	School          *SchoolProfile `json:"school,omitempty"`
}

// UserProfileUpdateRequest: PUT profil — camelCase sama dengan response.
type UserProfileUpdateRequest struct {
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Phone       *string `json:"phone,omitempty"`
	Whatsapp    *string `json:"whatsapp,omitempty"`
	ClassLevel  *string `json:"classLevel,omitempty"`
	City        *string `json:"city,omitempty"`
	Province    *string `json:"province,omitempty"`
	Gender      *string `json:"gender,omitempty"`
	BirthDate   *string `json:"birthDate,omitempty"` // YYYY-MM-DD
	Bio         *string `json:"bio,omitempty"`
	ParentName  *string `json:"parentName,omitempty"`
	ParentPhone *string `json:"parentPhone,omitempty"`
	Instagram   *string `json:"instagram,omitempty"`
	SchoolID    *string `json:"schoolId,omitempty"`
	SchoolName  *string `json:"schoolName,omitempty"`
	SubjectID   *string `json:"subjectId,omitempty"`
}

// TrainerProfileUpdateRequest alias agar kontrak API trainer tetap sama.
type TrainerProfileUpdateRequest = UserProfileUpdateRequest
