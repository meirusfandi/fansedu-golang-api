package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type TrainerService interface {
	// Status returns paid_slots, registered_students_count, and optional list of students.
	Status(ctx context.Context, trainerID string, includeStudents bool) (paidSlots, registeredCount int, students []domain.User, err error)
	// Pay adds quantity to trainer's paid_slots (after payment confirmation).
	Pay(ctx context.Context, trainerID string, quantity int) error
	// CreateStudent creates a student user and links to trainer; only if registeredCount < paidSlots.
	CreateStudent(ctx context.Context, trainerID string, name, email, password string) (domain.User, error)
}
