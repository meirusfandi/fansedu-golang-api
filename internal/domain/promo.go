package domain

import "time"

const (
	PromoDiscountTypePercent = "percent"
	PromoDiscountTypeFixed  = "fixed"
)

type PromoCode struct {
	ID            string
	Code          string
	DiscountType  string
	DiscountValue int
	ValidFrom     time.Time
	ValidUntil    *time.Time
	MaxUses       *int
	UsedCount     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
