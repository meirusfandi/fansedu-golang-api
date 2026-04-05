package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

func promoToAdminVoucherResp(p domain.PromoCode) dto.AdminVoucherResponse {
	var vu *string
	if p.ValidUntil != nil {
		s := p.ValidUntil.UTC().Format(time.RFC3339)
		vu = &s
	}
	return dto.AdminVoucherResponse{
		ID:                p.ID,
		Code:              p.Code,
		DiscountType:      p.DiscountType,
		DiscountValue:     p.DiscountValue,
		ValidFrom:         p.ValidFrom.UTC().Format(time.RFC3339),
		ValidUntil:        vu,
		MaxUses:           p.MaxUses,
		UsedCount:         p.UsedCount,
		IsActive:          p.IsActive,
		RequiresClaim:     p.RequiresClaim,
		AppliesToCourses:  p.AppliesToCourses,
		AppliesToPackages: p.AppliesToPackages,
		CreatedAt:         p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// AdminListVouchers GET /api/v1/admin/vouchers
func AdminListVouchers(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.PromoRepo.List(r.Context())
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		out := make([]dto.AdminVoucherResponse, 0, len(list))
		for _, p := range list {
			out = append(out, promoToAdminVoucherResp(p))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}
}

// AdminGetVoucher GET /api/v1/admin/vouchers/{voucherId}
func AdminGetVoucher(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(chi.URLParam(r, "voucherId"))
		if _, err := uuid.Parse(id); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "id voucher tidak valid")
			return
		}
		p, err := deps.PromoRepo.GetByID(r.Context(), id)
		if err != nil {
			if err == repo.ErrPromoNotFound {
				writeError(w, http.StatusNotFound, "not_found", "voucher tidak ditemukan")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(promoToAdminVoucherResp(p))
	}
}

// AdminCreateVoucher POST /api/v1/admin/vouchers
func AdminCreateVoucher(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body dto.AdminVoucherCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "body tidak valid")
			return
		}
		code := strings.TrimSpace(body.Code)
		if code == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "code wajib diisi")
			return
		}
		dt := strings.ToLower(strings.TrimSpace(body.DiscountType))
		if dt != domain.PromoDiscountTypePercent && dt != domain.PromoDiscountTypeFixed {
			writeError(w, http.StatusBadRequest, "bad_request", "discountType harus percent atau fixed")
			return
		}
		if body.DiscountValue < 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "discountValue tidak valid")
			return
		}
		if dt == domain.PromoDiscountTypePercent && body.DiscountValue > 100 {
			writeError(w, http.StatusBadRequest, "bad_request", "diskon persen maksimal 100")
			return
		}
		var validFrom time.Time
		if body.ValidFrom != nil && strings.TrimSpace(*body.ValidFrom) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*body.ValidFrom))
			if err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "validFrom harus RFC3339")
				return
			}
			validFrom = t
		} else {
			validFrom = time.Now()
		}
		var validUntil *time.Time
		if body.ValidUntil != nil && strings.TrimSpace(*body.ValidUntil) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*body.ValidUntil))
			if err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "validUntil harus RFC3339")
				return
			}
			validUntil = &t
		}
		maxUses := body.MaxUses
		if maxUses != nil && *maxUses < 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "maxUses tidak valid")
			return
		}
		isActive := true
		if body.IsActive != nil {
			isActive = *body.IsActive
		}
		requiresClaim := false
		if body.RequiresClaim != nil {
			requiresClaim = *body.RequiresClaim
		}
		appliesCourses := true
		if body.AppliesToCourses != nil {
			appliesCourses = *body.AppliesToCourses
		}
		appliesPackages := true
		if body.AppliesToPackages != nil {
			appliesPackages = *body.AppliesToPackages
		}
		if !appliesCourses && !appliesPackages {
			writeError(w, http.StatusBadRequest, "bad_request", "minimal salah satu appliesToCourses atau appliesToPackages harus true")
			return
		}
		p := domain.PromoCode{
			Code:              code,
			DiscountType:      dt,
			DiscountValue:     body.DiscountValue,
			ValidFrom:         validFrom,
			ValidUntil:        validUntil,
			MaxUses:           maxUses,
			IsActive:          isActive,
			RequiresClaim:     requiresClaim,
			AppliesToCourses:  appliesCourses,
			AppliesToPackages: appliesPackages,
		}
		created, err := deps.PromoRepo.Create(r.Context(), p)
		if err != nil {
			if errors.Is(err, repo.ErrPromoCodeDuplicate) {
				writeError(w, http.StatusConflict, "duplicate_code", "kode voucher sudah dipakai")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(promoToAdminVoucherResp(created))
	}
}

// AdminUpdateVoucher PUT /api/v1/admin/vouchers/{voucherId}
func AdminUpdateVoucher(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(chi.URLParam(r, "voucherId"))
		if _, err := uuid.Parse(id); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "id voucher tidak valid")
			return
		}
		cur, err := deps.PromoRepo.GetByID(r.Context(), id)
		if err != nil {
			if err == repo.ErrPromoNotFound {
				writeError(w, http.StatusNotFound, "not_found", "voucher tidak ditemukan")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		var body dto.AdminVoucherUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "body tidak valid")
			return
		}
		if body.Code != nil {
			c := strings.TrimSpace(*body.Code)
			if c == "" {
				writeError(w, http.StatusBadRequest, "bad_request", "code tidak boleh kosong")
				return
			}
			cur.Code = c
		}
		if body.DiscountType != nil {
			dt := strings.ToLower(strings.TrimSpace(*body.DiscountType))
			if dt != domain.PromoDiscountTypePercent && dt != domain.PromoDiscountTypeFixed {
				writeError(w, http.StatusBadRequest, "bad_request", "discountType harus percent atau fixed")
				return
			}
			cur.DiscountType = dt
		}
		if body.DiscountValue != nil {
			if *body.DiscountValue < 0 {
				writeError(w, http.StatusBadRequest, "bad_request", "discountValue tidak valid")
				return
			}
			cur.DiscountValue = *body.DiscountValue
		}
		if cur.DiscountType == domain.PromoDiscountTypePercent && cur.DiscountValue > 100 {
			writeError(w, http.StatusBadRequest, "bad_request", "diskon persen maksimal 100")
			return
		}
		if body.ValidFrom != nil {
			if strings.TrimSpace(*body.ValidFrom) == "" {
				writeError(w, http.StatusBadRequest, "bad_request", "validFrom tidak boleh kosong")
				return
			}
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*body.ValidFrom))
			if err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "validFrom harus RFC3339")
				return
			}
			cur.ValidFrom = t
		}
		if body.ValidUntil != nil {
			if strings.TrimSpace(*body.ValidUntil) == "" {
				cur.ValidUntil = nil
			} else {
				t, err := time.Parse(time.RFC3339, strings.TrimSpace(*body.ValidUntil))
				if err != nil {
					writeError(w, http.StatusBadRequest, "bad_request", "validUntil harus RFC3339")
					return
				}
				cur.ValidUntil = &t
			}
		}
		if body.MaxUses != nil {
			if *body.MaxUses < 0 {
				writeError(w, http.StatusBadRequest, "bad_request", "maxUses tidak valid")
				return
			}
			cur.MaxUses = body.MaxUses
		}
		if body.IsActive != nil {
			cur.IsActive = *body.IsActive
		}
		if body.RequiresClaim != nil {
			cur.RequiresClaim = *body.RequiresClaim
		}
		if body.AppliesToCourses != nil {
			cur.AppliesToCourses = *body.AppliesToCourses
		}
		if body.AppliesToPackages != nil {
			cur.AppliesToPackages = *body.AppliesToPackages
		}
		if !cur.AppliesToCourses && !cur.AppliesToPackages {
			writeError(w, http.StatusBadRequest, "bad_request", "minimal salah satu appliesToCourses atau appliesToPackages harus true")
			return
		}
		err = deps.PromoRepo.Update(r.Context(), cur)
		if err != nil {
			if errors.Is(err, repo.ErrPromoCodeDuplicate) {
				writeError(w, http.StatusConflict, "duplicate_code", "kode voucher sudah dipakai")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		p, err := deps.PromoRepo.GetByID(r.Context(), id)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(promoToAdminVoucherResp(p))
	}
}

// AdminDeleteVoucher DELETE /api/v1/admin/vouchers/{voucherId}
func AdminDeleteVoucher(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(chi.URLParam(r, "voucherId"))
		if _, err := uuid.Parse(id); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "id voucher tidak valid")
			return
		}
		if err := deps.PromoRepo.Delete(r.Context(), id); err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
