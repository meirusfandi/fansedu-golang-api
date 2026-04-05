package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type ctxKeyUserID struct{}
type ctxKeyRole struct{}

func Auth(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				WriteJSONError(w, http.StatusUnauthorized, "unauthorized", "Authorization header required")
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return jwtSecret, nil
			})
			if err != nil || token == nil || !token.Valid {
				WriteJSONError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				WriteJSONError(w, http.StatusUnauthorized, "unauthorized", "Invalid token claims")
				return
			}

			userID := jwtStringClaim(claims, "sub")
			role := jwtStringClaim(claims, "role")
			ctx := context.WithValue(r.Context(), ctxKeyUserID{}, userID)
			ctx = context.WithValue(ctx, ctxKeyRole{}, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminOnly() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := GetRole(r.Context())
			if !isAdminRole(role) {
				WriteJSONError(w, http.StatusForbidden, "forbidden", "Admin access required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isAdminRole(role string) bool {
	role = strings.TrimSpace(role)
	switch role {
	case domain.UserRoleAdmin,
		domain.UserRoleSuperAdmin,
		domain.UserRoleFinanceAdmin,
		domain.UserRoleAcademicAdmin,
		domain.UserRoleContentAdmin:
		return true
	default:
		return false
	}
}

var rolePermissions = map[string]map[string]struct{}{
	domain.UserRoleAdmin: {
		"*": {},
	},
	domain.UserRoleSuperAdmin: {
		"*": {},
	},
	domain.UserRoleFinanceAdmin: {
		"admin.overview.read": {},
		"payments.manage":     {},
		"orders.verify":       {},
		"vouchers.manage":     {},
		"reports.read":        {},
		"analytics.read":      {},
		"admin.audit.read":    {},
		"errors.read":         {},
	},
	domain.UserRoleAcademicAdmin: {
		"admin.overview.read": {},
		"users.manage":        {},
		"courses.manage":      {},
		"tryouts.manage":      {},
		"certificates.issue":  {},
		"reports.read":        {},
		"analytics.read":      {},
		"master-data.manage":  {},
		"admin.audit.read":    {},
		"errors.read":         {},
		"errors.manage":       {},
		"vouchers.manage":     {},
	},
	domain.UserRoleContentAdmin: {
		"admin.overview.read": {},
		"courses.manage":      {},
		"tryouts.manage":      {},
		"landing.manage":      {},
		"master-data.manage":  {},
		"admin.audit.read":    {},
		"errors.read":         {},
		"vouchers.manage":     {},
	},
}

func HasPermission(role, permission string) bool {
	perms, ok := rolePermissions[strings.TrimSpace(role)]
	if !ok {
		return false
	}
	if _, ok := perms["*"]; ok {
		return true
	}
	_, ok = perms[permission]
	return ok
}

func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := GetRole(r.Context())
			if !HasPermission(role, permission) {
				WriteJSONError(w, http.StatusForbidden, "forbidden", "Insufficient permission")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type AdminAuditLogger func(ctx context.Context, userID, role, method, path string, statusCode int, duration time.Duration, requestID string) error

type responseRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

func AdminAuditLog(logger AdminAuditLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rr, r)
			if logger == nil {
				return
			}
			userID, _ := GetUserID(r.Context())
			role, _ := GetRole(r.Context())
			requestID, _ := GetRequestID(r.Context())
			_ = logger(r.Context(), userID, role, r.Method, r.URL.Path, rr.status, time.Since(start), requestID)
		})
	}
}

// TrainerOnly restricts access to teaching staff (guru, legacy instructor enum, trainer).
func TrainerOnly() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := GetRole(r.Context())
			if !domain.IsTeachingStaffRoleCode(role) {
				WriteJSONError(w, http.StatusForbidden, "forbidden", "Trainer or guru access required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetUserID(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeyUserID{})
	s, ok := v.(string)
	return s, ok
}

func GetRole(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeyRole{})
	s, ok := v.(string)
	return s, ok
}

// OptionalAuth parses JWT when Authorization header is present and sets user/role in context.
// Does not return 401 when header is missing — so checkout can work for both guest and logged-in users.
func OptionalAuth(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return jwtSecret, nil
			})
			if err != nil || token == nil || !token.Valid {
				next.ServeHTTP(w, r)
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			userID := jwtStringClaim(claims, "sub")
			role := jwtStringClaim(claims, "role")
			ctx := context.WithValue(r.Context(), ctxKeyUserID{}, userID)
			ctx = context.WithValue(ctx, ctxKeyRole{}, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func jwtStringClaim(claims jwt.MapClaims, key string) string {
	v, ok := claims[key]
	if !ok || v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%.0f", x))
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

func bearerToken(v string) (string, bool) {
	if v == "" {
		return "", false
	}
	parts := strings.SplitN(v, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

// UserRepoForMiddleware adalah interface minimal untuk middleware lookup user
type UserRepoForMiddleware interface {
	FindByID(ctx context.Context, id string) (interface{ GetMustSetPassword() bool }, error)
}

// PasswordSetupGuard blocks protected endpoints when must_set_password=true.
// Allowlisted paths (like /auth/me, /auth/set-password, /auth/logout) should NOT use this middleware.
// Use this middleware AFTER Auth middleware.
func PasswordSetupGuard(userFinder func(ctx context.Context, id string) (bool, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			if !ok || userID == "" {
				next.ServeHTTP(w, r)
				return
			}

			mustSetPassword, err := userFinder(r.Context(), userID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if mustSetPassword {
				WriteJSONError(w, http.StatusForbidden, "PASSWORD_SETUP_REQUIRED", "Silakan atur password terlebih dahulu.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

