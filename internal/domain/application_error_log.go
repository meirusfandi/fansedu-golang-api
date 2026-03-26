package domain

import "time"

// Tipe error untuk filter & analitik admin.
const (
	AppErrTypeValidation     = "validation"
	AppErrTypeAuthentication = "authentication"
	AppErrTypeAuthorization = "authorization"
	AppErrTypeNotFound       = "not_found"
	AppErrTypeConflict       = "conflict"
	AppErrTypeRateLimit      = "rate_limit"
	AppErrTypeDatabase       = "database"
	AppErrTypeExternal       = "external_service"
	AppErrTypeInternal       = "internal"
	AppErrTypePanic          = "panic"
	AppErrTypeClient         = "client" // 4xx generik
	AppErrTypeServer         = "server" // 5xx generik
)

type ApplicationErrorLog struct {
	ID           string
	CreatedAt    time.Time
	ErrorType    string
	ErrorCode    *string
	Message      string
	HTTPStatus   int
	Method       string
	Path         string
	QueryString  *string
	UserID       *string
	UserRole     *string
	RequestID    *string
	IPAddress    *string
	UserAgent    *string
	StackTrace   *string
	Meta         map[string]any
	ResolvedAt   *time.Time
	ResolvedBy   *string
	AdminNote    *string
}
