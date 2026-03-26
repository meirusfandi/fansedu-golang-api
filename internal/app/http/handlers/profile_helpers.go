package handlers

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var trainerSchoolSlugClean = regexp.MustCompile(`[^a-z0-9-]+`)

func slugFromSchoolName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	s = trainerSchoolSlugClean.ReplaceAllString(s, "")
	return s
}

type profileApplyError struct {
	status int
	code   string
	msg    string
}

func (e *profileApplyError) Error() string { return e.msg }

func profileApplyErr(status int, code, msg string) error {
	return &profileApplyError{status: status, code: code, msg: msg}
}

func writeErrorFromProfileApply(w http.ResponseWriter, err error) {
	var pe *profileApplyError
	if errors.As(err, &pe) {
		writeError(w, pe.status, pe.code, pe.msg)
		return
	}
	writeError(w, http.StatusInternalServerError, "server_error", err.Error())
}

// SchoolToProfile maps domain.School to SchoolProfile (camelCase JSON).
func SchoolToProfile(s domain.School) *dto.SchoolProfile {
	p := &dto.SchoolProfile{
		ID:          s.ID,
		Name:        s.Name,
		NPSN:        "",
		RegencyCity: "",
		Address:     "",
		Phone:       "",
	}
	if s.Address != nil {
		p.Address = *s.Address
	}
	return p
}

// BuildUserProfileResponse builds full profile JSON (auth/me, student profile, trainer profile).
func BuildUserProfileResponse(ctx context.Context, deps *Deps, u domain.User, school *domain.School) dto.UserProfileResponse {
	out := dto.UserProfileResponse{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Role:            domain.DisplayRoleForAPI(u.Role),
		RoleCode:        u.Role,
		MustSetPassword: u.MustSetPassword,
		EmailVerified:   u.EmailVerified,
		AvatarURL:       u.AvatarURL,
		Phone:           u.Phone,
		Whatsapp:        u.Whatsapp,
		ClassLevel:      u.ClassLevel,
		City:            u.City,
		Province:        u.Province,
		Gender:          u.Gender,
		Bio:             u.Bio,
		ParentName:      u.ParentName,
		ParentPhone:     u.ParentPhone,
		Instagram:       u.Instagram,
		SchoolID:        u.SchoolID,
		SubjectID:       u.SubjectID,
	}
	if deps.RoleRepo != nil {
		if row, err := deps.RoleRepo.GetByUserRoleCode(ctx, u.Role); err == nil {
			out.RoleSlug = row.Slug
		}
	}
	if u.BirthDate != nil {
		s := u.BirthDate.UTC().Format("2006-01-02")
		out.BirthDate = &s
	}
	if school != nil {
		out.School = SchoolToProfile(*school)
		out.SchoolName = school.Name
	} else if deps != nil && deps.SchoolRepo != nil && u.SchoolID != nil {
		sid := strings.TrimSpace(*u.SchoolID)
		if sid != "" {
			if s, err := deps.SchoolRepo.GetByID(ctx, sid); err == nil {
				out.School = SchoolToProfile(s)
				out.SchoolName = s.Name
			}
		}
	}
	return out
}

// ApplyUserProfileUpdate mutates u from req (same rules as trainer profile PUT). u must be loaded with FindByID.
func ApplyUserProfileUpdate(ctx context.Context, deps *Deps, u *domain.User, req *dto.UserProfileUpdateRequest) error {
	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Email != "" {
		existing, err := deps.UserRepo.FindByEmail(ctx, req.Email)
		if err == nil && existing.ID != u.ID {
			return profileApplyErr(http.StatusConflict, "conflict", "email already in use")
		}
		u.Email = req.Email
	}

	if req.SchoolID != nil {
		sid := strings.TrimSpace(*req.SchoolID)
		if sid == "" {
			u.SchoolID = nil
		} else {
			if _, err := deps.SchoolRepo.GetByID(ctx, sid); err != nil {
				return profileApplyErr(http.StatusBadRequest, "bad_request", "school not found")
			}
			u.SchoolID = &sid
		}
	}
	if req.SchoolName != nil {
		schoolName := strings.TrimSpace(*req.SchoolName)
		if schoolName != "" {
			slug := slugFromSchoolName(schoolName)
			if slug == "" {
				return profileApplyErr(http.StatusBadRequest, "bad_request", "invalid school_name")
			}
			school, err := deps.SchoolRepo.GetBySlug(ctx, slug)
			if err != nil {
				created, createErr := deps.SchoolRepo.Create(ctx, domain.School{
					Name: schoolName,
					Slug: slug,
				})
				if createErr == nil {
					school = created
				} else {
					existing, getErr := deps.SchoolRepo.GetBySlug(ctx, slug)
					if getErr != nil {
						return profileApplyErr(http.StatusInternalServerError, "server_error", "failed to link school")
					}
					school = existing
				}
			}
			u.SchoolID = &school.ID
		}
	}
	if req.SubjectID != nil {
		subjectID := strings.TrimSpace(*req.SubjectID)
		if subjectID == "" {
			u.SubjectID = nil
		} else {
			if _, err := deps.SubjectRepo.GetByID(ctx, subjectID); err != nil {
				return profileApplyErr(http.StatusBadRequest, "bad_request", "subject not found")
			}
			u.SubjectID = &subjectID
		}
	}

	if req.Phone != nil {
		u.Phone = req.Phone
	}
	if req.Whatsapp != nil {
		u.Whatsapp = req.Whatsapp
	}
	if req.ClassLevel != nil {
		u.ClassLevel = req.ClassLevel
	}
	if req.City != nil {
		u.City = req.City
	}
	if req.Province != nil {
		u.Province = req.Province
	}
	if req.Gender != nil {
		u.Gender = req.Gender
	}
	if req.BirthDate != nil {
		b := strings.TrimSpace(*req.BirthDate)
		if b == "" {
			u.BirthDate = nil
		} else {
			parsed, err := time.Parse("2006-01-02", b)
			if err != nil {
				return profileApplyErr(http.StatusBadRequest, "validation_error", "invalid birthDate format; expected YYYY-MM-DD")
			}
			u.BirthDate = &parsed
		}
	}
	if req.Bio != nil {
		u.Bio = req.Bio
	}
	if req.ParentName != nil {
		u.ParentName = req.ParentName
	}
	if req.ParentPhone != nil {
		u.ParentPhone = req.ParentPhone
	}
	if req.Instagram != nil {
		u.Instagram = req.Instagram
	}
	return nil
}
