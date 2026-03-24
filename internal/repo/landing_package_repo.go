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
	GetByID(ctx context.Context, id string) (domain.LandingPackage, error)
	ListLinkedCourses(ctx context.Context, packageID string) ([]domain.PackageLinkedCourse, error)
	ReplaceLinkedCourses(ctx context.Context, packageID string, courseIDs []string) error
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

func (s *landingPackageRepoStub) GetByID(ctx context.Context, id string) (domain.LandingPackage, error) {
	return domain.LandingPackage{}, ErrPackageNotFound
}

func (s *landingPackageRepoStub) ListLinkedCourses(ctx context.Context, packageID string) ([]domain.PackageLinkedCourse, error) {
	return nil, nil
}

func (s *landingPackageRepoStub) ReplaceLinkedCourses(ctx context.Context, packageID string, courseIDs []string) error {
	return nil
}
