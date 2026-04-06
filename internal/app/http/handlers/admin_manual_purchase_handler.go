package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// AdminCreateManualOrder POST /api/v1/admin/orders/manual — order pending + item kelas (admin input pembayaran manual).
func AdminCreateManualOrder(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AdminManualOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "body tidak valid")
			return
		}
		order, err := deps.CheckoutService.CreateAdminManualOrder(r.Context(), req.UserID, req.CourseIDs, req.TotalPrice)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrManualOrderUserNotFound):
				writeError(w, http.StatusNotFound, "user_not_found", "pengguna tidak ditemukan")
			case errors.Is(err, service.ErrManualOrderNoCourses):
				writeError(w, http.StatusBadRequest, "no_courses", "minimal satu courseId diperlukan")
			case errors.Is(err, service.ErrCourseNotFound):
				writeError(w, http.StatusNotFound, "course_not_found", "satu atau lebih kelas tidak ditemukan")
			default:
				writeInternalError(w, r, err)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"orderId":     order.ID,
			"userId":      order.UserID,
			"status":      order.Status,
			"totalPrice":  order.TotalPrice,
			"normalPrice": order.NormalPrice,
			"createdAt":   order.CreatedAt.Format(time.RFC3339),
		})
	}
}

// AdminOrderPaymentProof POST /api/v1/admin/orders/{orderId}/payment-proof — sama seperti checkout (multipart proof).
func AdminOrderPaymentProof(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := strings.TrimSpace(chi.URLParam(r, "orderId"))
		proofPath, senderAccountNo, senderName, st, code, msg := extractPaymentProofUpload(r, orderID)
		if st != 0 {
			writeError(w, st, code, msg)
			return
		}
		if err := deps.CheckoutService.SubmitPaymentProof(r.Context(), orderID, proofPath, senderAccountNo, senderName, nil); err != nil {
			if err == service.ErrOrderNotFound {
				writeError(w, http.StatusNotFound, "order_not_found", "order tidak ditemukan")
				return
			}
			if err == service.ErrOrderNotPending {
				writeError(w, http.StatusBadRequest, "ORDER_NOT_PENDING", "Order sudah dibayar atau tidak dalam status menunggu.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Bukti pembayaran berhasil diupload"})
	}
}

// AdminPatchOrderPurchaseMeta PATCH /api/v1/admin/orders/{orderId}/purchase-meta — ubah tanggal pembelian / waktu upload bukti di order.
func AdminPatchOrderPurchaseMeta(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := strings.TrimSpace(chi.URLParam(r, "orderId"))
		if _, err := uuid.Parse(orderID); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "orderId tidak valid")
			return
		}
		var req dto.AdminPatchOrderPurchaseMetaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "body tidak valid")
			return
		}
		if (req.PurchasedAt == nil || strings.TrimSpace(*req.PurchasedAt) == "") &&
			(req.PaymentProofAt == nil || strings.TrimSpace(*req.PaymentProofAt) == "") {
			writeError(w, http.StatusBadRequest, "bad_request", "isi purchasedAt dan/atau paymentProofAt")
			return
		}
		if _, err := deps.OrderRepo.GetByID(r.Context(), orderID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusNotFound, "order_not_found", "order tidak ditemukan")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		if req.PurchasedAt != nil && strings.TrimSpace(*req.PurchasedAt) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.PurchasedAt))
			if err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "purchasedAt harus RFC3339")
				return
			}
			if err := deps.OrderRepo.UpdateOrderCreatedAt(r.Context(), orderID, t); err != nil {
				writeInternalError(w, r, err)
				return
			}
			_ = deps.PaymentRepo.BackdateByOrderID(r.Context(), orderID, t)
		}
		if req.PaymentProofAt != nil && strings.TrimSpace(*req.PaymentProofAt) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.PaymentProofAt))
			if err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "paymentProofAt harus RFC3339")
				return
			}
			if err := deps.OrderRepo.UpdatePaymentProofAtOnly(r.Context(), orderID, t); err != nil {
				writeInternalError(w, r, err)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Tanggal pembelian diperbarui"})
	}
}
