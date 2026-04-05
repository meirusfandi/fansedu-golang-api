package service

import (
	"context"
	"errors"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var (
	ErrPromoInactive      = errors.New("kode promo tidak aktif")
	ErrPromoWrongScope    = errors.New("kode promo tidak berlaku untuk jenis pembelian ini")
	ErrPromoRequiresClaim = errors.New("voucher harus diklaim ke akun terlebih dahulu")
)

// promoCheckoutKind: true = checkout satu kelas, false = checkout paket.
func (s *checkoutService) applyPromoDiscount(
	ctx context.Context,
	promoCode string,
	userID string,
	normalRupiah int,
	forCourseCheckout bool,
) (discountRupiah int, discountPercent *float64, appliedCode string, promo domain.PromoCode, err error) {
	if promoCode == "" || s.promoRepo == nil {
		return 0, nil, "", domain.PromoCode{}, nil
	}
	promo, err = s.promoRepo.GetByCode(ctx, promoCode)
	if err != nil {
		return 0, nil, "", domain.PromoCode{}, ErrPromoInvalid
	}
	now := time.Now()
	// Wajib: cek kedaluwarsa sebelum aturan lain (selain keberadaan kode).
	if promo.ValidUntil != nil && now.After(*promo.ValidUntil) {
		return 0, nil, "", promo, ErrPromoExpired
	}
	if now.Before(promo.ValidFrom) {
		return 0, nil, "", promo, ErrPromoInvalid
	}
	if !promo.IsActive {
		return 0, nil, "", promo, ErrPromoInactive
	}
	if promo.MaxUses != nil && promo.UsedCount >= *promo.MaxUses {
		return 0, nil, "", promo, ErrPromoMaxUses
	}
	if forCourseCheckout {
		if !promo.AppliesToCourses {
			return 0, nil, "", promo, ErrPromoWrongScope
		}
	} else {
		if !promo.AppliesToPackages {
			return 0, nil, "", promo, ErrPromoWrongScope
		}
	}
	if promo.RequiresClaim {
		ok, e := s.promoRepo.HasUnusedClaim(ctx, userID, promo.ID)
		if e != nil {
			return 0, nil, "", promo, e
		}
		if !ok {
			return 0, nil, "", promo, ErrPromoRequiresClaim
		}
	}

	switch promo.DiscountType {
	case domain.PromoDiscountTypePercent:
		if promo.DiscountValue <= 0 || promo.DiscountValue > 100 {
			return 0, nil, "", promo, ErrPromoInvalid
		}
		discountRupiah = normalRupiah * promo.DiscountValue / 100
		p := float64(promo.DiscountValue)
		discountPercent = &p
	case domain.PromoDiscountTypeFixed:
		if promo.DiscountValue <= 0 {
			return 0, nil, "", promo, ErrPromoInvalid
		}
		discountValRupiah := promo.DiscountValue
		if discountValRupiah >= normalRupiah {
			discountRupiah = normalRupiah
			p := 100.0
			discountPercent = &p
		} else {
			discountRupiah = discountValRupiah
			p := float64(discountValRupiah*100) / float64(normalRupiah)
			discountPercent = &p
		}
	default:
		return 0, nil, "", promo, ErrPromoInvalid
	}
	return discountRupiah, discountPercent, promo.Code, promo, nil
}

func (s *checkoutService) incrementPromoIfNeededOnInitiate(ctx context.Context, appliedCode string, promo domain.PromoCode) {
	if appliedCode == "" || s.promoRepo == nil {
		return
	}
	if promo.RequiresClaim {
		return
	}
	_ = s.promoRepo.IncrementUsedCount(ctx, promo.ID)
}

func (s *checkoutService) finalizePromoAfterPayment(ctx context.Context, order domain.Order) {
	if order.PromoCode == nil || *order.PromoCode == "" || s.promoRepo == nil {
		return
	}
	promo, err := s.promoRepo.GetByCode(ctx, *order.PromoCode)
	if err != nil {
		return
	}
	if promo.RequiresClaim {
		_ = s.promoRepo.MarkClaimUsedForOrder(ctx, order.UserID, promo.ID, order.ID)
		_ = s.promoRepo.IncrementUsedCount(ctx, promo.ID)
	}
}
