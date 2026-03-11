package repo

import (
	"context"
	"errors"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var ErrPackageNotFound = errors.New("package not found")

// LandingPackageRepo lists packages for the landing page.
type LandingPackageRepo interface {
	List(ctx context.Context) ([]domain.LandingPackage, error)
	GetBySlug(ctx context.Context, slug string) (domain.LandingPackage, error)
}

// NewLandingPackageRepoStub returns a stub that returns empty list (no DB table required).
// Replace with a real implementation when packages table exists.
func NewLandingPackageRepoStub() LandingPackageRepo {
	return &landingPackageRepoStub{}
}

type landingPackageRepoStub struct{}

func (s *landingPackageRepoStub) List(ctx context.Context) ([]domain.LandingPackage, error) {
	return nil, nil
}

func (s *landingPackageRepoStub) GetBySlug(ctx context.Context, slug string) (domain.LandingPackage, error) {
	return domain.LandingPackage{}, ErrPackageNotFound
}
