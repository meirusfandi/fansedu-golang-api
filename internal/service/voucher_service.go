package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

var (
	ErrVoucherNotClaimable   = errors.New("kode ini tidak memerlukan klaim; gunakan langsung saat checkout")
	ErrVoucherAlreadyClaimed = errors.New("voucher sudah diklaim di akun Anda")
)

// VoucherService klaim voucher ke akun (requires_claim) dan daftar voucher siap pakai.
type VoucherService interface {
	Claim(ctx context.Context, userID, code string) error
	ListMine(ctx context.Context, userID string) ([]repo.UserVoucherListRow, error)
}

type voucherService struct {
	promos repo.PromoRepo
}

func NewVoucherService(promos repo.PromoRepo) VoucherService {
	return &voucherService{promos: promos}
}

func (s *voucherService) Claim(ctx context.Context, userID, code string) error {
	code = strings.TrimSpace(code)
	if code == "" || s.promos == nil {
		return ErrPromoInvalid
	}
	p, err := s.promos.GetByCode(ctx, code)
	if err != nil {
		return ErrPromoInvalid
	}
	now := time.Now()
	if p.ValidUntil != nil && now.After(*p.ValidUntil) {
		return ErrPromoExpired
	}
	if now.Before(p.ValidFrom) {
		return ErrPromoInvalid
	}
	if !p.IsActive {
		return ErrPromoInactive
	}
	if !p.RequiresClaim {
		return ErrVoucherNotClaimable
	}
	if p.MaxUses != nil && p.UsedCount >= *p.MaxUses {
		return ErrPromoMaxUses
	}
	ok, err := s.promos.HasUnusedClaim(ctx, userID, p.ID)
	if err != nil {
		return err
	}
	if ok {
		return ErrVoucherAlreadyClaimed
	}
	err = s.promos.InsertClaim(ctx, userID, p.ID)
	if err != nil {
		if errors.Is(err, repo.ErrVoucherClaimDuplicate) {
			return ErrVoucherAlreadyClaimed
		}
		return err
	}
	return nil
}

func (s *voucherService) ListMine(ctx context.Context, userID string) ([]repo.UserVoucherListRow, error) {
	if s.promos == nil {
		return nil, nil
	}
	return s.promos.ListMyUnusedVouchers(ctx, userID)
}
