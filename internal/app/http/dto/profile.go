package dto

// UserProfileResponse: GET /auth/me, /student/profile, /trainer/profile, /guru/profile — bentuk JSON sama untuk semua peran.
type UserProfileResponse struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Email           string         `json:"email"`
	Role            string         `json:"role"`
	RoleSlug        string         `json:"roleSlug"`
	RoleCode        string         `json:"roleCode"`
	MustSetPassword bool           `json:"mustSetPassword"`
	EmailVerified   bool           `json:"emailVerified"`
	AvatarURL       *string        `json:"avatarUrl"`
	Phone           *string        `json:"phone"`
	Whatsapp        *string        `json:"whatsapp"`
	ClassLevel      *string        `json:"classLevel"`
	City            *string        `json:"city"`
	Province        *string        `json:"province"`
	Gender          *string        `json:"gender"`
	BirthDate       *string        `json:"birthDate"` // YYYY-MM-DD
	Bio             *string        `json:"bio"`
	ParentName      *string        `json:"parentName"`
	ParentPhone     *string        `json:"parentPhone"`
	Instagram       *string        `json:"instagram"`
	SchoolID        *string        `json:"schoolId"`
	SchoolName      string         `json:"schoolName"`
	SubjectID       *string        `json:"subjectId"`
	LevelID         *string        `json:"levelId"`
	School          *SchoolProfile `json:"school"`
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
	LevelID     *string `json:"levelId,omitempty"`
}

// TrainerProfileUpdateRequest alias agar kontrak API trainer tetap sama.
type TrainerProfileUpdateRequest = UserProfileUpdateRequest
