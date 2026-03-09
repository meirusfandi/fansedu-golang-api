package domain

import "time"

const (
	OrderStatusPending = "pending"
	OrderStatusPaid    = "paid"
	OrderStatusFailed  = "failed"
)

type Order struct {
	ID                string
	UserID            string
	Status            string
	TotalPriceCents   int
	PaymentMethod     *string
	PaymentReference  *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type OrderItem struct {
	ID         string
	OrderID    string
	CourseID   string
	PriceCents int
	CreatedAt  time.Time
}
