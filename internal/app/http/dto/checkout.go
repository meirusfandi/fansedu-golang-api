package dto

// CheckoutInitiateRequest for POST /checkout/initiate
type CheckoutInitiateRequest struct {
	CourseSlug string `json:"courseSlug"`
	Email      string `json:"email"`
	Name       string `json:"name"`
}

// CheckoutInitiateResponse returned after initiate
type CheckoutInitiateResponse struct {
	OrderID    string `json:"orderId"`
	UserID     string `json:"userId"`
	TotalPrice int    `json:"totalPrice"`
	IsNewUser  bool   `json:"isNewUser"`
}

// CheckoutPaymentSessionRequest for POST /checkout/payment-session
type CheckoutPaymentSessionRequest struct {
	OrderID       string `json:"orderId"`
	PaymentMethod string `json:"paymentMethod"`
}

// CheckoutPaymentSessionResponse with URL to redirect user
type CheckoutPaymentSessionResponse struct {
	PaymentURL string `json:"paymentUrl"`
}
