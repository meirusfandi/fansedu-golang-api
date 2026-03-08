package dto

// TrainerStatusResponse for GET /api/v1/trainer/status
type TrainerStatusResponse struct {
	PaidSlots             int        `json:"paid_slots"`
	RegisteredStudentsCount int      `json:"registered_students_count"`
	Students              []UserInfo `json:"students,omitempty"`
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
	Name   string          `json:"name,omitempty"`
	Email  string          `json:"email,omitempty"`
	School *SchoolProfile  `json:"school,omitempty"`
}

// SchoolProfile objek sekolah untuk response (nama field dinormalisasi untuk frontend)
type SchoolProfile struct {
	ID             string `json:"id"`
	NamaSekolah    string `json:"nama_sekolah"`
	NPSN           string `json:"npsn"`
	KabupatenKota  string `json:"kabupaten_kota"`
	Alamat         string `json:"alamat"`
	Telepon        string `json:"telepon"`
}

// TrainerProfileUpdateRequest for PUT /api/v1/trainer/profile
type TrainerProfileUpdateRequest struct {
	Name     string  `json:"name"`
	SchoolID *string `json:"school_id,omitempty"` // opsional: kaitkan guru ke sekolah; kosong/null = lepas
}
