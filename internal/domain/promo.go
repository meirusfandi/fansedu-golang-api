package domain

import "time"

const (
	PromoDiscountTypePercent = "percent"
	PromoDiscountTypeFixed  = "fixed"
)

type PromoCode struct {
	ID                 string
	Code               string
	DiscountType       string
	DiscountValue      int
	ValidFrom          time.Time
	ValidUntil         *time.Time
	MaxUses            *int
	UsedCount          int
	IsActive           bool
	RequiresClaim      bool
	AppliesToCourses   bool
	AppliesToPackages  bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// VoucherClaim baris klaim user ke voucher (untuk requires_claim).
type VoucherClaim struct {
	ID          string
	UserID      string
	PromoCodeID string
	ClaimedAt   time.Time
	UsedAt      *time.Time
	OrderID     *string
}
