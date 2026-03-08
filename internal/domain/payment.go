package domain

import "time"

const (
	PaymentStatusPending  = "pending"
	PaymentStatusPaid    = "paid"
	PaymentStatusFailed  = "failed"
	PaymentStatusRefunded = "refunded"

	PaymentTypeCoursePurchase = "course_purchase"
	PaymentTypeSubscription   = "subscription"
	PaymentTypeTryout         = "tryout"
	PaymentTypeOther          = "other"
)

type Payment struct {
	ID             string
	UserID         string
	AmountCents    int
	Currency       string
	Status         string
	Type           string
	ReferenceID    *string
	Description    *string
	ProofURL       *string
	ConfirmedBy    *string
	ConfirmedAt    *time.Time
	RejectionNote  *string
	PaidAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
