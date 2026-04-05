package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// VoucherClaim POST /api/v1/vouchers/claim
func VoucherClaim(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "login diperlukan")
			return
		}
		if deps.VoucherService == nil {
			writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Layanan sementara tidak tersedia.")
			return
		}
		var body dto.VoucherClaimRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "body tidak valid")
			return
		}
		code := strings.TrimSpace(body.Code)
		if code == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "code wajib diisi")
			return
		}
		err := deps.VoucherService.Claim(r.Context(), userID, code)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrPromoInvalid):
				writeError(w, http.StatusBadRequest, "invalid_voucher", "kode voucher tidak valid atau belum berlaku")
			case errors.Is(err, service.ErrPromoExpired):
				writeError(w, http.StatusBadRequest, "voucher_expired", "kode voucher sudah kadaluarsa")
			case errors.Is(err, service.ErrPromoInactive):
				writeError(w, http.StatusBadRequest, "voucher_inactive", "voucher tidak aktif")
			case errors.Is(err, service.ErrPromoMaxUses):
				writeError(w, http.StatusBadRequest, "voucher_max_uses", "kuota voucher habis")
			case errors.Is(err, service.ErrVoucherNotClaimable):
				writeError(w, http.StatusBadRequest, "voucher_not_claimable", "kode ini tidak perlu diklaim; gunakan langsung saat checkout")
			case errors.Is(err, service.ErrVoucherAlreadyClaimed):
				writeError(w, http.StatusConflict, "voucher_already_claimed", "voucher sudah ada di akun Anda")
			default:
				writeInternalError(w, r, err)
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// VoucherListMine GET /api/v1/vouchers/mine
func VoucherListMine(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "login diperlukan")
			return
		}
		if deps.VoucherService == nil {
			writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Layanan sementara tidak tersedia.")
			return
		}
		rows, err := deps.VoucherService.ListMine(r.Context(), userID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		out := make([]dto.MyVoucherItem, 0, len(rows))
		for _, row := range rows {
			it := dto.MyVoucherItem{
				ClaimID:       row.ClaimID,
				PromoID:       row.PromoID,
				Code:          row.Code,
				DiscountType:  row.DiscountType,
				DiscountValue: row.DiscountValue,
			}
			if row.ValidUntil != nil {
				s := row.ValidUntil.UTC().Format(time.RFC3339)
				it.ValidUntil = &s
			}
			out = append(out, it)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}
}
