package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
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
		INSERT INTO users (id, email, password_hash, name, role, avatar_url, school_id, subject_id, email_verified, email_verified_at, must_set_password)
		VALUES ($1::uuid, $2, $3, $4, $5::user_role, $6, $7::uuid, $8::uuid, $9, $10, $11)
		RETURNING id, email, password_hash, name, role, avatar_url, school_id, subject_id,
		          email_verified, email_verified_at, must_set_password,
		          phone, whatsapp, class_level, city, province, gender, birth_date, bio, parent_name, parent_phone, instagram,
		          created_at, updated_at
	`, id, u.Email, passHash, u.Name, u.Role, u.AvatarURL, u.SchoolID, u.SubjectID, u.EmailVerified, u.EmailVerifiedAt, u.MustSetPassword)
	var out domain.User
	var avatarURL, schoolID, subjectID, pass *string
	var emailVerifiedAt *time.Time
	var emailVerified, mustSetPassword bool
	var phone, whatsapp, classLevel, city, province, gender *string
	var bio, parentName, parentPhone, instagram *string
	var birthDate *time.Time
	err := row.Scan(
		&out.ID, &out.Email, &pass, &out.Name, &out.Role,
		&avatarURL, &schoolID, &subjectID,
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
		SELECT id, email, password_hash, name, role, avatar_url, school_id, subject_id,
		       email_verified, email_verified_at, must_set_password,
		       phone, whatsapp, class_level, city, province, gender, birth_date, bio, parent_name, parent_phone, instagram,
		       created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	var out domain.User
	var avatarURL, schoolID, subjectID, pass *string
	var emailVerifiedAt *time.Time
	var emailVerified, mustSetPassword bool
	var phone, whatsapp, classLevel, city, province, gender *string
	var bio, parentName, parentPhone, instagram *string
	var birthDate *time.Time
	err := row.Scan(
		&out.ID, &out.Email, &pass, &out.Name, &out.Role,
		&avatarURL, &schoolID, &subjectID,
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
		SELECT id, email, password_hash, name, role, avatar_url, school_id, subject_id,
		       email_verified, email_verified_at, must_set_password,
		       phone, whatsapp, class_level, city, province, gender, birth_date, bio, parent_name, parent_phone, instagram,
		       created_at, updated_at
		FROM users WHERE id = $1::uuid
	`, id)
	var out domain.User
	var avatarURL, schoolID, subjectID, pass *string
	var emailVerifiedAt *time.Time
	var emailVerified, mustSetPassword bool
	var phone, whatsapp, classLevel, city, province, gender *string
	var bio, parentName, parentPhone, instagram *string
	var birthDate *time.Time
	err := row.Scan(
		&out.ID, &out.Email, &pass, &out.Name, &out.Role,
		&avatarURL, &schoolID, &subjectID,
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
	query := `SELECT id, email, password_hash, name, role, avatar_url, school_id, subject_id,
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
		var avatarURL, schoolID, subjectID, pass *string
		var emailVerifiedAt *time.Time
		var emailVerified, mustSetPassword bool
		var phone, whatsapp, classLevel, city, province, gender *string
		var bio, parentName, parentPhone, instagram *string
		var birthDate *time.Time
		if err := rows.Scan(
			&u.ID, &u.Email, &pass, &u.Name, &u.Role,
			&avatarURL, &schoolID, &subjectID,
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
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET
			name = $2,
			email = $3,
			role = $4::user_role,
			avatar_url = $5,
			school_id = $6::uuid,
			subject_id = $7::uuid,
			password_hash = $8,
			email_verified = $9,
			email_verified_at = $10,
			must_set_password = $11,
			phone = $12,
			whatsapp = $13,
			class_level = $14,
			city = $15,
			province = $16,
			gender = $17,
			birth_date = $18,
			bio = $19,
			parent_name = $20,
			parent_phone = $21,
			instagram = $22,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, u.ID, u.Name, u.Email, u.Role, u.AvatarURL, u.SchoolID, u.SubjectID, u.PasswordHash, u.EmailVerified, u.EmailVerifiedAt, u.MustSetPassword,
		u.Phone, u.Whatsapp, u.ClassLevel, u.City, u.Province, u.Gender, u.BirthDate, u.Bio, u.ParentName, u.ParentPhone, u.Instagram)
	return err
}
