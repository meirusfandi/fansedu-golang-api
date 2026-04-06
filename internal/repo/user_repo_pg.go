package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) UserRepo {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	id := uuid.New().String()
	var passHash interface{} = u.PasswordHash
	if u.PasswordHash == "" {
		passHash = nil
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, name, role, avatar_url, school_id, subject_id, level_id, email_verified, email_verified_at, must_set_password, phone, whatsapp)
		VALUES ($1::uuid, $2, $3, $4, $5::user_role, $6, $7::uuid, $8::uuid, $9::uuid, $10, $11, $12, $13, $14)
		RETURNING id, email, password_hash, name, role, avatar_url, school_id, subject_id, level_id,
		          email_verified, email_verified_at, must_set_password,
		          phone, whatsapp, class_level, city, province, gender, birth_date, bio, parent_name, parent_phone, instagram,
		          created_at, updated_at
	`, id, u.Email, passHash, u.Name, u.Role, u.AvatarURL, u.SchoolID, u.SubjectID, u.LevelID, u.EmailVerified, u.EmailVerifiedAt, u.MustSetPassword, u.Phone, u.Whatsapp)
	var out domain.User
	var avatarURL, schoolID, subjectID, levelID, pass *string
	var emailVerifiedAt *time.Time
	var emailVerified, mustSetPassword bool
	var phone, whatsapp, classLevel, city, province, gender *string
	var bio, parentName, parentPhone, instagram *string
	var birthDate *time.Time
	err := row.Scan(
		&out.ID, &out.Email, &pass, &out.Name, &out.Role,
		&avatarURL, &schoolID, &subjectID, &levelID,
		&emailVerified, &emailVerifiedAt, &mustSetPassword,
		&phone, &whatsapp, &classLevel, &city, &province, &gender, &birthDate, &bio, &parentName, &parentPhone, &instagram,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}
	if pass != nil {
		out.PasswordHash = *pass
	}
	out.AvatarURL = avatarURL
	out.SchoolID = schoolID
	out.SubjectID = subjectID
	out.LevelID = levelID
	out.EmailVerified = emailVerified
	out.EmailVerifiedAt = emailVerifiedAt
	out.MustSetPassword = mustSetPassword
	out.Phone = phone
	out.Whatsapp = whatsapp
	out.ClassLevel = classLevel
	out.City = city
	out.Province = province
	out.Gender = gender
	out.BirthDate = birthDate
	out.Bio = bio
	out.ParentName = parentName
	out.ParentPhone = parentPhone
	out.Instagram = instagram
	return out, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, avatar_url, school_id, subject_id, level_id,
		       email_verified, email_verified_at, must_set_password,
		       phone, whatsapp, class_level, city, province, gender, birth_date, bio, parent_name, parent_phone, instagram,
		       created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	var out domain.User
	var avatarURL, schoolID, subjectID, levelID, pass *string
	var emailVerifiedAt *time.Time
	var emailVerified, mustSetPassword bool
	var phone, whatsapp, classLevel, city, province, gender *string
	var bio, parentName, parentPhone, instagram *string
	var birthDate *time.Time
	err := row.Scan(
		&out.ID, &out.Email, &pass, &out.Name, &out.Role,
		&avatarURL, &schoolID, &subjectID, &levelID,
		&emailVerified, &emailVerifiedAt, &mustSetPassword,
		&phone, &whatsapp, &classLevel, &city, &province, &gender, &birthDate, &bio, &parentName, &parentPhone, &instagram,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}
	if pass != nil {
		out.PasswordHash = *pass
	}
	out.AvatarURL = avatarURL
	out.SchoolID = schoolID
	out.SubjectID = subjectID
	out.LevelID = levelID
	out.EmailVerified = emailVerified
	out.EmailVerifiedAt = emailVerifiedAt
	out.MustSetPassword = mustSetPassword
	out.Phone = phone
	out.Whatsapp = whatsapp
	out.ClassLevel = classLevel
	out.City = city
	out.Province = province
	out.Gender = gender
	out.BirthDate = birthDate
	out.Bio = bio
	out.ParentName = parentName
	out.ParentPhone = parentPhone
	out.Instagram = instagram
	return out, nil
}

func (r *userRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, avatar_url, school_id, subject_id, level_id,
		       email_verified, email_verified_at, must_set_password,
		       phone, whatsapp, class_level, city, province, gender, birth_date, bio, parent_name, parent_phone, instagram,
		       created_at, updated_at
		FROM users WHERE id = $1::uuid
	`, id)
	var out domain.User
	var avatarURL, schoolID, subjectID, levelID, pass *string
	var emailVerifiedAt *time.Time
	var emailVerified, mustSetPassword bool
	var phone, whatsapp, classLevel, city, province, gender *string
	var bio, parentName, parentPhone, instagram *string
	var birthDate *time.Time
	err := row.Scan(
		&out.ID, &out.Email, &pass, &out.Name, &out.Role,
		&avatarURL, &schoolID, &subjectID, &levelID,
		&emailVerified, &emailVerifiedAt, &mustSetPassword,
		&phone, &whatsapp, &classLevel, &city, &province, &gender, &birthDate, &bio, &parentName, &parentPhone, &instagram,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}
	if pass != nil {
		out.PasswordHash = *pass
	}
	out.AvatarURL = avatarURL
	out.SchoolID = schoolID
	out.SubjectID = subjectID
	out.LevelID = levelID
	out.EmailVerified = emailVerified
	out.EmailVerifiedAt = emailVerifiedAt
	out.MustSetPassword = mustSetPassword
	out.Phone = phone
	out.Whatsapp = whatsapp
	out.ClassLevel = classLevel
	out.City = city
	out.Province = province
	out.Gender = gender
	out.BirthDate = birthDate
	out.Bio = bio
	out.ParentName = parentName
	out.ParentPhone = parentPhone
	out.Instagram = instagram
	return out, nil
}

func (r *userRepo) MustSetPasswordByID(ctx context.Context, id string) (bool, error) {
	var m bool
	err := r.pool.QueryRow(ctx, `SELECT must_set_password FROM users WHERE id = $1::uuid`, id).Scan(&m)
	return m, err
}

func (r *userRepo) FindByIDProfile(ctx context.Context, id string) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, name, role, avatar_url, school_id, subject_id, level_id,
		       email_verified, email_verified_at, must_set_password,
		       phone, whatsapp, class_level, city, province, gender, birth_date, bio, parent_name, parent_phone, instagram,
		       created_at, updated_at
		FROM users WHERE id = $1::uuid
	`, id)
	var out domain.User
	out.PasswordHash = ""
	var avatarURL, schoolID, subjectID, levelID *string
	var emailVerifiedAt *time.Time
	var emailVerified, mustSetPassword bool
	var phone, whatsapp, classLevel, city, province, gender *string
	var bio, parentName, parentPhone, instagram *string
	var birthDate *time.Time
	err := row.Scan(
		&out.ID, &out.Email, &out.Name, &out.Role,
		&avatarURL, &schoolID, &subjectID, &levelID,
		&emailVerified, &emailVerifiedAt, &mustSetPassword,
		&phone, &whatsapp, &classLevel, &city, &province, &gender, &birthDate, &bio, &parentName, &parentPhone, &instagram,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}
	out.AvatarURL = avatarURL
	out.SchoolID = schoolID
	out.SubjectID = subjectID
	out.LevelID = levelID
	out.EmailVerified = emailVerified
	out.EmailVerifiedAt = emailVerifiedAt
	out.MustSetPassword = mustSetPassword
	out.Phone = phone
	out.Whatsapp = whatsapp
	out.ClassLevel = classLevel
	out.City = city
	out.Province = province
	out.Gender = gender
	out.BirthDate = birthDate
	out.Bio = bio
	out.ParentName = parentName
	out.ParentPhone = parentPhone
	out.Instagram = instagram
	return out, nil
}

func (r *userRepo) FindByIDProfileWithSchool(ctx context.Context, id string) (domain.User, *domain.School, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			u.id, u.email, u.name, u.role, u.avatar_url, u.school_id, u.subject_id, u.level_id,
			u.email_verified, u.email_verified_at, u.must_set_password,
			u.phone, u.whatsapp, u.class_level, u.city, u.province, u.gender, u.birth_date, u.bio, u.parent_name, u.parent_phone, u.instagram,
			u.created_at, u.updated_at,
			s.id, s.name, s.slug, s.description, s.address, s.logo_url, s.created_at, s.updated_at
		FROM users u
		LEFT JOIN schools s ON s.id = u.school_id
		WHERE u.id = $1::uuid
	`, id)
	var out domain.User
	out.PasswordHash = ""
	var avatarURL, schoolID, subjectID, levelID *string
	var emailVerifiedAt *time.Time
	var emailVerified, mustSetPassword bool
	var phone, whatsapp, classLevel, city, province, gender *string
	var bio, parentName, parentPhone, instagram *string
	var birthDate *time.Time
	var sid, sname, sslug *string
	var sdesc, saddr, slog *string
	var screated, supdated *time.Time
	err := row.Scan(
		&out.ID, &out.Email, &out.Name, &out.Role,
		&avatarURL, &schoolID, &subjectID, &levelID,
		&emailVerified, &emailVerifiedAt, &mustSetPassword,
		&phone, &whatsapp, &classLevel, &city, &province, &gender, &birthDate, &bio, &parentName, &parentPhone, &instagram,
		&out.CreatedAt, &out.UpdatedAt,
		&sid, &sname, &sslug, &sdesc, &saddr, &slog, &screated, &supdated,
	)
	if err != nil {
		return domain.User{}, nil, err
	}
	out.AvatarURL = avatarURL
	out.SchoolID = schoolID
	out.SubjectID = subjectID
	out.LevelID = levelID
	out.EmailVerified = emailVerified
	out.EmailVerifiedAt = emailVerifiedAt
	out.MustSetPassword = mustSetPassword
	out.Phone = phone
	out.Whatsapp = whatsapp
	out.ClassLevel = classLevel
	out.City = city
	out.Province = province
	out.Gender = gender
	out.BirthDate = birthDate
	out.Bio = bio
	out.ParentName = parentName
	out.ParentPhone = parentPhone
	out.Instagram = instagram
	var school *domain.School
	if sid != nil && *sid != "" && sname != nil {
		school = &domain.School{
			ID:          *sid,
			Name:        *sname,
			Slug:        derefString(sslug),
			Description: sdesc,
			Address:     saddr,
			LogoURL:     slog,
		}
		if screated != nil {
			school.CreatedAt = *screated
		}
		if supdated != nil {
			school.UpdatedAt = *supdated
		}
	}
	return out, school, nil
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (r *userRepo) CountByRole(ctx context.Context, role string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = $1::user_role`, role).Scan(&n)
	return n, err
}

func (r *userRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (r *userRepo) List(ctx context.Context, role string) ([]domain.User, error) {
	query := `SELECT id, email, password_hash, name, role, avatar_url, school_id, subject_id, level_id,
	                 email_verified, email_verified_at, must_set_password,
	                 phone, whatsapp, class_level, city, province, gender, birth_date, bio, parent_name, parent_phone, instagram,
	                 created_at, updated_at
	          FROM users`
	args := []interface{}{}
	if role != "" {
		query += ` WHERE role = $1::user_role`
		args = append(args, role)
	}
	query += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.User
	for rows.Next() {
		var u domain.User
		var avatarURL, schoolID, subjectID, levelID, pass *string
		var emailVerifiedAt *time.Time
		var emailVerified, mustSetPassword bool
		var phone, whatsapp, classLevel, city, province, gender *string
		var bio, parentName, parentPhone, instagram *string
		var birthDate *time.Time
		if err := rows.Scan(
			&u.ID, &u.Email, &pass, &u.Name, &u.Role,
			&avatarURL, &schoolID, &subjectID, &levelID,
			&emailVerified, &emailVerifiedAt, &mustSetPassword,
			&phone, &whatsapp, &classLevel, &city, &province, &gender, &birthDate, &bio, &parentName, &parentPhone, &instagram,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if pass != nil {
			u.PasswordHash = *pass
		}
		u.AvatarURL = avatarURL
		u.SchoolID = schoolID
		u.SubjectID = subjectID
		u.LevelID = levelID
		u.EmailVerified = emailVerified
		u.EmailVerifiedAt = emailVerifiedAt
		u.MustSetPassword = mustSetPassword
		u.Phone = phone
		u.Whatsapp = whatsapp
		u.ClassLevel = classLevel
		u.City = city
		u.Province = province
		u.Gender = gender
		u.BirthDate = birthDate
		u.Bio = bio
		u.ParentName = parentName
		u.ParentPhone = parentPhone
		u.Instagram = instagram
		list = append(list, u)
	}
	return list, rows.Err()
}

func (r *userRepo) Update(ctx context.Context, u domain.User) error {
	// Jangan timpa password_hash dengan string kosong (mis. bila hash tidak ter-load ke struct).
	ct, err := r.pool.Exec(ctx, `
		UPDATE users SET
			name = $2,
			email = $3,
			role = $4::user_role,
			avatar_url = $5,
			school_id = $6::uuid,
			subject_id = $7::uuid,
			level_id = $8::uuid,
			password_hash = CASE
				WHEN NULLIF(trim($9::text), '') IS NULL THEN password_hash
				ELSE $9
			END,
			email_verified = $10,
			email_verified_at = $11,
			must_set_password = $12,
			phone = $13,
			whatsapp = $14,
			class_level = $15,
			city = $16,
			province = $17,
			gender = $18,
			birth_date = $19,
			bio = $20,
			parent_name = $21,
			parent_phone = $22,
			instagram = $23,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, u.ID, u.Name, u.Email, u.Role, u.AvatarURL, u.SchoolID, u.SubjectID, u.LevelID, u.PasswordHash, u.EmailVerified, u.EmailVerifiedAt, u.MustSetPassword,
		u.Phone, u.Whatsapp, u.ClassLevel, u.City, u.Province, u.Gender, u.BirthDate, u.Bio, u.ParentName, u.ParentPhone, u.Instagram)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
