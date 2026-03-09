package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// CheckoutInitiate creates or finds user and creates pending order. POST /api/v1/checkout/initiate
func CheckoutInitiate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ProgramID   string `json:"programId"`
			ProgramSlug string `json:"programSlug"`
			CourseSlug  string `json:"course_slug"`
			Name        string `json:"name"`
			Email       string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		if req.Email == "" || req.Name == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "name and email required")
			return
		}
		courseSlug := req.ProgramSlug
		if courseSlug == "" && req.ProgramID != "" {
			course, err := deps.CourseRepo.GetByID(r.Context(), req.ProgramID)
			if err != nil {
				writeError(w, http.StatusNotFound, "not_found", "program not found")
				return
			}
			if course.Slug != nil {
				courseSlug = *course.Slug
			}
		}
		if courseSlug == "" {
			courseSlug = req.CourseSlug
		}
		if courseSlug == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "programId, programSlug, or course_slug required")
			return
		}
		result, err := deps.CheckoutService.Initiate(r.Context(), courseSlug, req.Email, req.Name)
		if err != nil {
			if err == service.ErrCourseNotFound {
				writeError(w, http.StatusNotFound, "not_found", "program not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.CheckoutInitiateResponseLMS{
			CheckoutID: result.OrderID,
			OrderID:    result.OrderID,
			Total:      result.TotalPrice / 100,
			Program: dto.CheckoutProgramInfo{
				Title:        result.CourseTitle,
				PriceDisplay: formatRupiah(result.PriceCents),
			},
		})
	}
}

// CheckoutPaymentSession creates gateway session and returns payment URL. POST /api/v1/checkout/payment-session
func CheckoutPaymentSession(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			CheckoutID     string `json:"checkoutId"`
			OrderID        string `json:"order_id"`
			PaymentMethod  string `json:"paymentMethod"`
			PaymentMethodL string `json:"payment_method"`
			PromoCode      string `json:"promoCode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		orderID := req.CheckoutID
		if orderID == "" {
			orderID = req.OrderID
		}
		pm := req.PaymentMethod
		if pm == "" {
			pm = req.PaymentMethodL
		}
		if orderID == "" || pm == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "checkoutId and paymentMethod required")
			return
		}
		paymentURL, err := deps.CheckoutService.CreatePaymentSession(r.Context(), orderID, pm)
		if err != nil {
			if err == service.ErrOrderNotFound {
				writeError(w, http.StatusNotFound, "not_found", "order not found")
				return
			}
			if err == service.ErrOrderNotPending {
				writeError(w, http.StatusBadRequest, "bad_request", "order is not pending")
				return
			}
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		amount := 0
		if order, err := deps.OrderRepo.GetByID(r.Context(), orderID); err == nil {
			amount = order.TotalPriceCents / 100
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.CheckoutPaymentSessionResponseLMS{
			PaymentURL:           paymentURL,
			OrderID:              orderID,
			Expiry:               "",
			VirtualAccountNumber: "",
			Amount:               amount,
		})
	}
}

// PaymentWebhook handles gateway webhook (e.g. Midtrans/Stripe). POST /api/v1/webhook/payment
func PaymentWebhook(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			OrderID string `json:"order_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if body.OrderID == "" {
			http.Error(w, "order_id required", http.StatusBadRequest)
			return
		}
		// TODO: verify gateway signature (Midtrans/Stripe) using header or body
		if err := deps.CheckoutService.HandlePaymentWebhook(r.Context(), body.OrderID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
