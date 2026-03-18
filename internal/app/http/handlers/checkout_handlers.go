package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// CheckoutInitiate creates or finds user and creates pending order. POST /api/v1/checkout/initiate
func CheckoutInitiate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ProgramID     string `json:"programId"`
			ProgramSlug   string `json:"programSlug"`
			CourseSlug    string `json:"course_slug"`
			Name          string `json:"name"`
			Email         string `json:"email"`
			PromoCode     string `json:"promoCode"`
			NormalPrice   int    `json:"normalPrice"`
			NormalPriceL  int    `json:"normal_price"`
			Price         int    `json:"price"`
			FinalPrice    int    `json:"finalPrice"`
			FinalPriceL   int    `json:"final_price"`
			ExpectedTotal int    `json:"expectedTotal"`
			ExpectedTotalL int   `json:"expected_total"`
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

		// Harga dari request sebagai fallback jika DB tidak punya
		reqNormalPrice := req.NormalPrice
		if reqNormalPrice == 0 {
			reqNormalPrice = req.NormalPriceL
		}
		reqPrice := req.Price
		if reqPrice == 0 {
			reqPrice = req.FinalPrice
		}
		if reqPrice == 0 {
			reqPrice = req.FinalPriceL
		}
		if reqPrice == 0 {
			reqPrice = req.ExpectedTotal
		}
		if reqPrice == 0 {
			reqPrice = req.ExpectedTotalL
		}
		if reqNormalPrice == 0 && reqPrice > 0 {
			reqNormalPrice = reqPrice
		}

		// Ensure a backing course exists for this slug. Many landing programs only
		// exist in the packages table; in that case, lazily create a course so
		// checkout can proceed instead of returning 404.
		existingCourse, courseErr := deps.CourseRepo.GetBySlug(r.Context(), courseSlug)
		needsUpdate := courseErr != nil || existingCourse.Price == 0

		if needsUpdate && deps.LandingPackageRepo != nil {
			if pkg, pkgErr := deps.LandingPackageRepo.GetBySlug(r.Context(), courseSlug); pkgErr == nil {
				priceRupiah := packagePriceRupiah(pkg)
				// Jika package tidak punya harga, gunakan dari request
				if priceRupiah == 0 {
					priceRupiah = reqNormalPrice
				}
				desc := pkg.ShortDescription
				slug := pkg.Slug
				if courseErr != nil {
					// Course tidak ada, buat baru
					_, _ = deps.CourseRepo.Create(r.Context(), domain.Course{
						Title:       pkg.Name,
						Slug:        &slug,
						Description: desc,
						Price:       priceRupiah,
					})
				} else if existingCourse.Price == 0 && priceRupiah > 0 {
					// Course ada tapi price = 0, update harganya
					existingCourse.Price = priceRupiah
					_ = deps.CourseRepo.Update(r.Context(), existingCourse)
				}
			} else if courseErr != nil && reqNormalPrice > 0 {
				// Package tidak ditemukan tapi ada harga dari request, buat course baru
				_, _ = deps.CourseRepo.Create(r.Context(), domain.Course{
					Title:       courseSlug,
					Slug:        &courseSlug,
					Price:       reqNormalPrice,
				})
			}
		} else if courseErr != nil && reqNormalPrice > 0 {
			// Tidak ada package repo, tapi ada harga dari request
			_, _ = deps.CourseRepo.Create(r.Context(), domain.Course{
				Title: courseSlug,
				Slug:  &courseSlug,
				Price: reqNormalPrice,
			})
		}

		loggedInUserID, _ := middleware.GetUserID(r.Context())
		result, err := deps.CheckoutService.Initiate(r.Context(), courseSlug, req.Email, req.Name, strings.TrimSpace(req.PromoCode), loggedInUserID)
		if err != nil {
			// If course still not found, try one more time to create it lazily
			// from landing packages (or using request price), then retry Initiate once.
			if err == service.ErrCourseNotFound {
				priceForCreate := reqNormalPrice
				titleForCreate := courseSlug
				var descForCreate *string
				if deps.LandingPackageRepo != nil {
					if pkg, pkgErr := deps.LandingPackageRepo.GetBySlug(r.Context(), courseSlug); pkgErr == nil {
						pkgPrice := packagePriceRupiah(pkg)
						if pkgPrice > 0 {
							priceForCreate = pkgPrice
						}
						titleForCreate = pkg.Name
						descForCreate = pkg.ShortDescription
					}
				}
				if priceForCreate > 0 {
					if _, createErr := deps.CourseRepo.Create(r.Context(), domain.Course{
						Title:       titleForCreate,
						Slug:        &courseSlug,
						Description: descForCreate,
						Price:       priceForCreate,
					}); createErr == nil {
						// Retry once after successful create
						if retryResult, retryErr := deps.CheckoutService.Initiate(r.Context(), courseSlug, req.Email, req.Name, strings.TrimSpace(req.PromoCode), loggedInUserID); retryErr == nil {
							result = retryResult
							err = nil
						} else {
							err = retryErr
						}
					}
				}
			}
			if err != nil {
				if err == service.ErrCourseNotFound {
					writeError(w, http.StatusNotFound, "not_found", "program not found for slug "+courseSlug)
					return
				}
				if err == service.ErrPromoInvalid {
					writeError(w, http.StatusBadRequest, "invalid_promo", "kode promo tidak valid")
					return
				}
				if err == service.ErrPromoExpired {
					writeError(w, http.StatusBadRequest, "promo_expired", "kode promo sudah kadaluarsa")
					return
				}
				if err == service.ErrPromoMaxUses {
					writeError(w, http.StatusBadRequest, "promo_max_uses", "kode promo sudah mencapai batas penggunaan")
					return
				}
				writeError(w, http.StatusInternalServerError, "server_error", err.Error())
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.CheckoutInitiateResponseLMS{
			CheckoutID:       result.OrderID,
			OrderID:          result.OrderID,
			Total:            result.TotalPrice,
			Program:          dto.CheckoutProgramInfo{
				Title:        result.CourseTitle,
				PriceDisplay: formatRupiah(result.NormalPrice),
			},
			NormalPrice:      result.NormalPrice,
			PromoCode:        result.PromoCode,
			Discount:         result.Discount,
			DiscountPercent: result.DiscountPercent,
			FinalPrice:      result.FinalPrice,
			ConfirmationCode: result.ConfirmationCode,
		})
	}
}

// packagePriceRupiah returns price in rupiah from package (early bird preferred, then normal).
func packagePriceRupiah(pkg domain.LandingPackage) int {
	if pkg.PriceEarlyBird != nil && *pkg.PriceEarlyBird > 0 {
		return int(*pkg.PriceEarlyBird)
	}
	if pkg.PriceNormal != nil {
		return int(*pkg.PriceNormal)
	}
	return 0
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
				writeError(w, http.StatusNotFound, "order_not_found", "order tidak ditemukan; pastikan checkout/initiate sudah dipanggil dan orderId benar")
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
			amount = order.TotalPrice
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

// UploadDir untuk bukti pembayaran (relatif dari working directory).
const paymentProofUploadDir = "uploads/payment-proofs"

// CheckoutPaymentProof menerima upload bukti transfer. POST /api/v1/checkout/orders/:orderId/payment-proof
// FormData: proof (file), amount, senderAccountNo, senderName
func CheckoutPaymentProof(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "orderId")
		if orderID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "orderId required")
			return
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB
			writeError(w, http.StatusBadRequest, "bad_request", "invalid multipart form")
			return
		}
		file, fh, err := r.FormFile("proof")
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "proof file required")
			return
		}
		defer file.Close()
		senderAccountNo := strings.TrimSpace(r.FormValue("senderAccountNo"))
		senderName := strings.TrimSpace(r.FormValue("senderName"))

		filename := "proof.dat"
		if fh != nil && fh.Filename != "" {
			filename = fh.Filename
		}
		safeName := strings.ReplaceAll(filepath.Base(filename), "..", "")
		if safeName == "" {
			safeName = "proof"
		}
		dir := filepath.Join(paymentProofUploadDir, orderID)
		if err := os.MkdirAll(dir, 0755); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", "failed to create upload dir")
			return
		}
		dstPath := filepath.Join(dir, safeName)
		dst, err := os.Create(dstPath)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", "failed to save file")
			return
		}
		_, _ = io.Copy(dst, file)
		dst.Close()
		proofPath := "/" + filepath.ToSlash(dstPath)

		if err := deps.CheckoutService.SubmitPaymentProof(r.Context(), orderID, proofPath, senderAccountNo, senderName); err != nil {
			if err == service.ErrOrderNotFound {
				writeError(w, http.StatusNotFound, "order_not_found", "order tidak ditemukan")
				return
			}
			if err == service.ErrOrderNotPending {
				writeError(w, http.StatusBadRequest, "bad_request", "order sudah dibayar atau tidak pending")
				return
			}
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Bukti pembayaran berhasil diupload"})
	}
}
