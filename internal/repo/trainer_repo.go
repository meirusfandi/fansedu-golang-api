package repo

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type TrainerRepo interface {
	// GetOrCreateSlots returns paid_slots for trainer (0 if no row yet).
	GetOrCreateSlots(ctx context.Context, trainerID string) (paidSlots int, err error)
	// AddSlots adds quantity to trainer's paid_slots (after payment confirmation).
	AddSlots(ctx context.Context, trainerID string, quantity int) error
	// CountStudents returns number of students linked to this trainer.
	CountStudents(ctx context.Context, trainerID string) (int, error)
	// ListStudents returns students linked to this trainer.
	ListStudents(ctx context.Context, trainerID string) ([]domain.User, error)
	// LinkStudent links a student to the trainer (fails if already linked).
	LinkStudent(ctx context.Context, trainerID, studentID string) error
}
