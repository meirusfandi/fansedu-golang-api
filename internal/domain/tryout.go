package domain

import "time"

const (
	TryoutLevelEasy   = "easy"
	TryoutLevelMedium = "medium"
	TryoutLevelHard   = "hard"

	TryoutStatusDraft  = "draft"
	TryoutStatusOpen   = "open"
	TryoutStatusClosed = "closed"

	// TryoutGradingModeAuto: kunci jawaban + penilaian otomatis saat submit.
	TryoutGradingModeAuto = "auto"
	// TryoutGradingModeManual: jawaban disimpan saat submit; skor diisi admin/guru lewat review.
	TryoutGradingModeManual = "manual"
)

type TryoutSession struct {
	ID               string
	Title            string
	ShortTitle       *string
	Description      *string
	DurationMinutes  int
	QuestionsCount   int
	Level            string
	Subject          *string
	SchoolLevel      *string
	SubjectID        *string // bidang: siswa hanya lihat tryout yang subject_id = user.subject_id atau NULL (umum)
	OpensAt          time.Time
	ClosesAt         time.Time
	MaxParticipants  *int
	Status           string
	GradingMode      string // auto | manual
	CreatedBy        *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
