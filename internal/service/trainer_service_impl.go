package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

var (
	ErrNoSlotsAvailable = errors.New("no paid slots available to register more students")
)

type trainerService struct {
	userRepo    interface {
		Create(ctx context.Context, u domain.User) (domain.User, error)
		FindByEmail(ctx context.Context, email string) (domain.User, error)
	}
	trainerRepo repo.TrainerRepo
	bcryptCost  int
}

func NewTrainerService(
	userRepo interface {
		Create(ctx context.Context, u domain.User) (domain.User, error)
		FindByEmail(ctx context.Context, email string) (domain.User, error)
	},
	trainerRepo repo.TrainerRepo,
) TrainerService {
	return &trainerService{
		userRepo:    userRepo,
		trainerRepo: trainerRepo,
		bcryptCost:  bcrypt.DefaultCost,
	}
}

func (s *trainerService) Status(ctx context.Context, trainerID string, includeStudents bool) (paidSlots, registeredCount int, students []domain.User, err error) {
	paidSlots, err = s.trainerRepo.GetOrCreateSlots(ctx, trainerID)
	if err != nil {
		return 0, 0, nil, err
	}
	registeredCount, err = s.trainerRepo.CountStudents(ctx, trainerID)
	if err != nil {
		return 0, 0, nil, err
	}
	if includeStudents {
		students, err = s.trainerRepo.ListStudents(ctx, trainerID)
		if err != nil {
			return 0, 0, nil, err
		}
	}
	return paidSlots, registeredCount, students, nil
}

func (s *trainerService) Pay(ctx context.Context, trainerID string, quantity int) error {
	if quantity <= 0 {
		return nil
	}
	return s.trainerRepo.AddSlots(ctx, trainerID, quantity)
}

func (s *trainerService) CreateStudent(ctx context.Context, trainerID string, name, email, password string) (domain.User, error) {
	paidSlots, err := s.trainerRepo.GetOrCreateSlots(ctx, trainerID)
	if err != nil {
		return domain.User{}, err
	}
	count, err := s.trainerRepo.CountStudents(ctx, trainerID)
	if err != nil {
		return domain.User{}, err
	}
	if count >= paidSlots {
		return domain.User{}, ErrNoSlotsAvailable
	}
	_, err = s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return domain.User{}, ErrEmailExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return domain.User{}, err
	}
	u := domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         domain.UserRoleStudent,
	}
	u, err = s.userRepo.Create(ctx, u)
	if err != nil {
		return domain.User{}, err
	}
	if err := s.trainerRepo.LinkStudent(ctx, trainerID, u.ID); err != nil {
		return domain.User{}, err
	}
	return u, nil
}
