package repo

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// LandingPackageRepo lists packages for the landing page.
type LandingPackageRepo interface {
	List(ctx context.Context) ([]domain.LandingPackage, error)
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
