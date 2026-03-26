package domain

import "time"

type Role struct {
	ID            string
	Name          string
	Slug          string
	UserRoleCode  string // nilai users.role / JWT (label enum user_role); biasanya beda dari slug publik (mis. siswa → student)
	Description   *string
	IconURL       *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
