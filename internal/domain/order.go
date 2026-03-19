package domain

import "time"

const (
	OrderStatusPending = "pending"
	OrderStatusPaid    = "paid"
	OrderStatusFailed  = "failed"
)

type Order struct {
	ID               string
	UserID           string
	Status           string
	TotalPrice       int    // final amount (rupiah)
	NormalPrice      int    // harga normal sebelum promo (rupiah)
	PromoCode        *string
	Discount         int    // potongan (rupiah)
	DiscountPercent  *float64
	ConfirmationCode   *string // 3 digit unik untuk konfirmasi pembayaran
	PaymentMethod      *string
	PaymentReference   *string
	PaymentProofURL    *string
	PaymentProofAt     *time.Time
	SenderAccountNo    *string
	SenderName         *string
	RoleHint           *string // student|instructor for auto-create user
	BuyerEmail         *string // email pembeli untuk guest checkout
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type OrderItem struct {
	ID        string
	OrderID   string
	CourseID  string
	Price     int // rupiah
	CreatedAt time.Time
}
