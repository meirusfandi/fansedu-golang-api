package dto

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // optional: slug atau user_role_code yang ada di tabel roles; kosong = default dari tabel roles
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User            any    `json:"user"`
	Token           string `json:"token"`
	MustSetPassword bool   `json:"mustSetPassword"`
	NextAction      string `json:"nextAction,omitempty"`
}

type AuthUserResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	RoleSlug        string `json:"roleSlug,omitempty"`
	RoleCode        string `json:"roleCode,omitempty"`
	MustSetPassword bool   `json:"mustSetPassword"`
}

type SetPasswordRequest struct {
	NewPassword string `json:"newPassword"`
}

type SetPasswordResponse struct {
	Message         string `json:"message"`
	MustSetPassword bool   `json:"mustSetPassword"`
}

type CompletePurchaseAuthRequest struct {
	RoleHint string `json:"roleHint,omitempty"`
}

type CompletePurchaseAuthResponse struct {
	Token           string `json:"token"`
	User            any    `json:"user"`
	MustSetPassword bool   `json:"mustSetPassword"`
	NextAction      string `json:"nextAction,omitempty"`
}

type PasswordHashRequest struct {
	Password string `json:"password"`
}

type PasswordHashResponse struct {
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
}

