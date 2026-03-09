package dto

// CheckoutInitiateRequest for POST /checkout/initiate
type CheckoutInitiateRequest struct {
	CourseSlug string `json:"course_slug"`
	Email      string `json:"email"`
	Name       string `json:"name"`
}

// CheckoutInitiateResponse returned after initiate
type CheckoutInitiateResponse struct {
	OrderID    string `json:"order_id"`
	UserID     string `json:"user_id"`
	TotalPrice int    `json:"total_price"`
	IsNewUser  bool   `json:"is_new_user"`
}

// CheckoutPaymentSessionRequest for POST /checkout/payment-session
type CheckoutPaymentSessionRequest struct {
	OrderID        string `json:"order_id"`
	PaymentMethod  string `json:"payment_method"`
}

// CheckoutPaymentSessionResponse with URL to redirect user
type CheckoutPaymentSessionResponse struct {
	PaymentURL string `json:"payment_url"`
}
