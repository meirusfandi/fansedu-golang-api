package dto

// VoucherClaimRequest POST /api/v1/vouchers/claim
type VoucherClaimRequest struct {
	Code string `json:"code"`
}

// MyVoucherItem voucher yang sudah diklaim, belum dipakai checkout.
type MyVoucherItem struct {
	ClaimID       string  `json:"claimId"`
	PromoID       string  `json:"promoId"`
	Code          string  `json:"code"`
	DiscountType  string  `json:"discountType"`
	DiscountValue int     `json:"discountValue"`
	ValidUntil    *string `json:"validUntil,omitempty"`
}
