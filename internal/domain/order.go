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
	Quantity         int    // jumlah item/siswa
	UnitPrice        int    // harga per item setelah promo (tanpa unique code)
	Subtotal         int    // UnitPrice * Quantity (tanpa unique code)
	UniqueCode       int    // kode unik nominal transfer, 1x per order
	IsCollective     bool   // pembelian kolektif (guru/instructor)
	StudentsJSON     []byte // JSON metadata siswa kolektif
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

type OrderStudent struct {
	Name   string  `json:"name,omitempty"`
	Email  string  `json:"email,omitempty"`
	UserID *string `json:"user_id,omitempty"`
}

type OrderItem struct {
	ID        string
	OrderID   string
	CourseID  string
	Price     int // rupiah
	CreatedAt time.Time
}
