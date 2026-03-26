package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

var (
	errUnknownRoleSlug  = errors.New("unknown role slug")
	errRolesTableEmpty  = errors.New("roles table is empty")
)

// registerFallbackUserRoleCode dipakai hanya untuk self-register publik bila tidak ada baris cocok di tabel roles
// (mis. migrasi/seed belum jalan), tetap membatasi ke peran non-admin yang valid di enum user_role.
func registerFallbackUserRoleCode(input string) (string, bool) {
	h := strings.ToLower(strings.TrimSpace(input))
	switch h {
	case "student", "siswa":
		return domain.UserRoleStudent, true
	case "guru", "pengajar", "pembimbing":
		return domain.UserRoleGuru, true
	case "instructor":
		return "instructor", true
	case "trainer":
		return domain.UserRoleTrainer, true
	default:
		return "", false
	}
}

func storedUserRoleCodeFromRow(e domain.Role) string {
	code := strings.TrimSpace(e.UserRoleCode)
	if code == "" {
		code = strings.TrimSpace(e.Slug)
	}
	return code
}

// resolveUserRoleCodeForUserTable maps input ke users.role / JWT: cocokkan ke baris roles
// lewat slug ATAU lewat nilai efektif user_role_code (enum), case-insensitive.
func resolveUserRoleCodeForUserTable(ctx context.Context, rr repo.RoleRepo, input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errUnknownRoleSlug
	}
	if row, err := rr.GetBySlug(ctx, input); err == nil {
		return storedUserRoleCodeFromRow(row), nil
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	if row, err := rr.GetByUserRoleCode(ctx, input); err == nil {
		return storedUserRoleCodeFromRow(row), nil
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	return "", errUnknownRoleSlug
}

// defaultUserRoleCode picks default registration / admin user role from tabel roles (siswa lalu student).
func defaultUserRoleCode(ctx context.Context, rr repo.RoleRepo) (string, error) {
	for _, s := range []string{"siswa", "student"} {
		code, err := resolveUserRoleCodeForUserTable(ctx, rr, s)
		if err == nil {
			return code, nil
		}
	}
	list, err := rr.List(ctx)
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", errRolesTableEmpty
	}
	code := strings.TrimSpace(list[0].UserRoleCode)
	if code == "" {
		code = strings.TrimSpace(list[0].Slug)
	}
	return code, nil
}

func isLegacyCheckoutRoleHint(s string) bool {
	h := strings.ToLower(strings.TrimSpace(s))
	return h == "student" || h == "instructor" || h == "guru" || h == "siswa"
}

func normalizeLegacyCheckoutRoleHint(s string) string {
	h := strings.ToLower(strings.TrimSpace(s))
	if h == "instructor" || h == "guru" {
		return domain.UserRoleGuru
	}
	return domain.UserRoleStudent
}

func checkoutLegacyRoleCode(reqHint, orderHint *string) string {
	if reqHint != nil && isLegacyCheckoutRoleHint(*reqHint) {
		return normalizeLegacyCheckoutRoleHint(*reqHint)
	}
	if orderHint != nil && isLegacyCheckoutRoleHint(*orderHint) {
		return normalizeLegacyCheckoutRoleHint(*orderHint)
	}
	return domain.UserRoleStudent
}

// resolveCheckoutUserRoleCode maps hint (slug publik atau label legacy) → users.role, memakai tabel roles bila tersedia.
func resolveCheckoutUserRoleCode(ctx context.Context, rr repo.RoleRepo, reqHint *string, orderHint *string) (string, error) {
	if rr == nil {
		return checkoutLegacyRoleCode(reqHint, orderHint), nil
	}
	try := func(h *string) (string, bool, error) {
		if h == nil {
			return "", false, nil
		}
		s := strings.TrimSpace(*h)
		if s == "" {
			return "", false, nil
		}
		code, err := resolveUserRoleCodeForUserTable(ctx, rr, s)
		if err == nil {
			return code, true, nil
		}
		if errors.Is(err, errUnknownRoleSlug) && isLegacyCheckoutRoleHint(s) {
			return normalizeLegacyCheckoutRoleHint(s), true, nil
		}
		if !errors.Is(err, errUnknownRoleSlug) {
			return "", false, err
		}
		return "", false, nil
	}
	if c, ok, err := try(reqHint); err != nil {
		return "", err
	} else if ok {
		return c, nil
	}
	if c, ok, err := try(orderHint); err != nil {
		return "", err
	} else if ok {
		return c, nil
	}
	code, err := defaultUserRoleCode(ctx, rr)
	if err != nil {
		return domain.UserRoleStudent, nil
	}
	return code, nil
}

// authUserResponse builds /auth/me-style payload; Role = display, RoleCode = enum/JWT, RoleSlug = slug publik (bila ada di DB).
func authUserResponse(ctx context.Context, rr repo.RoleRepo, u domain.User) dto.AuthUserResponse {
	out := dto.AuthUserResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Role:            domain.DisplayRoleForAPI(u.Role),
		RoleCode:        u.Role,
		MustSetPassword: u.MustSetPassword,
	}
	if rr != nil {
		if row, err := rr.GetByUserRoleCode(ctx, u.Role); err == nil {
			out.RoleSlug = row.Slug
		}
	}
	return out
}

// userAuthMap is used for login/register JSON user object.
func userAuthMap(ctx context.Context, rr repo.RoleRepo, u domain.User) map[string]interface{} {
	m := map[string]interface{}{
		"id":              u.ID,
		"name":            u.Name,
		"email":           u.Email,
		"role":            domain.DisplayRoleForAPI(u.Role),
		"roleCode":        u.Role,
		"mustSetPassword": u.MustSetPassword,
	}
	if u.AvatarURL != nil {
		m["avatarUrl"] = *u.AvatarURL
	}
	if rr != nil {
		if row, err := rr.GetByUserRoleCode(ctx, u.Role); err == nil {
			m["roleSlug"] = row.Slug
		}
	}
	return m
}
