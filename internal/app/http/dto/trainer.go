package dto

// TrainerStatusResponse for GET /api/v1/trainer/status
type TrainerStatusResponse struct {
	PaidSlots               int        `json:"paidSlots"`
	RegisteredStudentsCount int        `json:"registeredStudentsCount"`
	Students                []UserInfo `json:"students,omitempty"`
}

// UserInfo minimal user for list responses (no password)
type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// TrainerPayRequest for POST /api/v1/trainer/pay
type TrainerPayRequest struct {
	Quantity int `json:"quantity"`
}

// TrainerCreateStudentRequest for POST /api/v1/trainer/students
type TrainerCreateStudentRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TrainerProfileResponse for GET /api/v1/trainer/profile
type TrainerProfileResponse struct {
	Name   string         `json:"name,omitempty"`
	Email  string         `json:"email,omitempty"`
	Phone  *string        `json:"phone,omitempty"`
	Whatsapp *string      `json:"whatsapp,omitempty"`
	ClassLevel *string   `json:"classLevel,omitempty"`
	City   *string        `json:"city,omitempty"`
	Province *string      `json:"province,omitempty"`
	Gender *string        `json:"gender,omitempty"`
	BirthDate *string    `json:"birthDate,omitempty"` // YYYY-MM-DD
	Bio    *string        `json:"bio,omitempty"`
	ParentName *string    `json:"parentName,omitempty"`
	ParentPhone *string  `json:"parentPhone,omitempty"`
	Instagram *string    `json:"instagram,omitempty"`
	SchoolID  *string         `json:"schoolId,omitempty"`
	SubjectID *string         `json:"subjectId,omitempty"`
	School    *SchoolProfile `json:"school,omitempty"`
}

// SchoolProfile nested school; seluruh key camelCase.
type SchoolProfile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NPSN        string `json:"npsn,omitempty"`
	RegencyCity string `json:"regencyCity,omitempty"`
	Address     string `json:"address,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

// TrainerStudentItem minimal student data for trainer screens.
type TrainerStudentItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type TrainerStudentUpdateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
