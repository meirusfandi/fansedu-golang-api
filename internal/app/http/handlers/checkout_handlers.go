package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/paymentgateway"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// CheckoutInitiate creates or finds user and creates pending order. POST /api/v1/checkout/initiate
func CheckoutInitiate(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ProgramID     string                    `json:"programId"`
			ProgramSlug   string                    `json:"programSlug"`
			CourseSlug    string                    `json:"courseSlug"`
			Name          string                    `json:"name"`
			Email         string                    `json:"email"`
			PromoCode     string                    `json:"promoCode"`
			NormalPrice   int                       `json:"normalPrice"`
			Price         int                       `json:"price"`
			FinalPrice    int                       `json:"finalPrice"`
			ExpectedTotal int                       `json:"expectedTotal"`
			RoleHint      string                    `json:"roleHint"`
			BuyerRole     string                    `json:"buyerRole"`
			Quantity      int                       `json:"quantity"`
			Students      []dto.CheckoutStudentItem `json:"students"`
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
			writeError(w, http.StatusBadRequest, "bad_request", "programId, programSlug, or courseSlug required")
			return
		}

		// Harga dari request sebagai fallback jika DB tidak punya
		reqNormalPrice := req.NormalPrice
		reqPrice := req.Price
		if reqPrice == 0 {
			reqPrice = req.FinalPrice
		}
		if reqPrice == 0 {
			reqPrice = req.ExpectedTotal
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
				priceRupiah := domain.LandingPackagePriceRupiah(pkg)
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
		roleHint := req.RoleHint
		buyerRole := req.BuyerRole
		students := make([]domain.OrderStudent, 0, len(req.Students))
		for _, s := range req.Students {
			students = append(students, domain.OrderStudent{
				Name:   strings.TrimSpace(s.Name),
				Email:  strings.TrimSpace(s.Email),
				UserID: s.UserID,
			})
		}
		var result *service.CheckoutInitiateResult
		var err error
		packageTried := false
		if deps.LandingPackageRepo != nil {
			if pkg, pkgErr := deps.LandingPackageRepo.GetBySlug(r.Context(), courseSlug); pkgErr == nil {
				if linked, lerr := deps.LandingPackageRepo.ListLinkedCourses(r.Context(), pkg.ID); lerr == nil && len(linked) > 0 {
					packageTried = true
					result, err = deps.CheckoutService.InitiatePackage(
						r.Context(),
						courseSlug,
						req.Email,
						req.Name,
						strings.TrimSpace(req.PromoCode),
						loggedInUserID,
						roleHint,
						buyerRole,
						req.Quantity,
						students,
					)
				}
			}
		}
		if packageTried && err != nil {
			if err == service.ErrCourseNotFound {
				writeError(w, http.StatusNotFound, "not_found", "program not found for slug "+courseSlug)
				return
			}
			if writeCheckoutPromoError(w, err) {
				return
			}
			if err == service.ErrPackageNoLinkedCourses {
				writeError(w, http.StatusBadRequest, "package_no_classes", "paket belum dihubungkan ke kelas")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		if result == nil {
			result, err = deps.CheckoutService.Initiate(
				r.Context(),
				courseSlug,
				req.Email,
				req.Name,
				strings.TrimSpace(req.PromoCode),
				loggedInUserID,
				roleHint,
				buyerRole,
				req.Quantity,
				students,
			)
			if err != nil {
				// If course still not found, try one more time to create it lazily
				// from landing packages (or using request price), then retry Initiate once.
				if err == service.ErrCourseNotFound {
					priceForCreate := reqNormalPrice
					titleForCreate := courseSlug
					var descForCreate *string
					if deps.LandingPackageRepo != nil {
						if pkg, pkgErr := deps.LandingPackageRepo.GetBySlug(r.Context(), courseSlug); pkgErr == nil {
							pkgPrice := domain.LandingPackagePriceRupiah(pkg)
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
							if retryResult, retryErr := deps.CheckoutService.Initiate(
								r.Context(),
								courseSlug,
								req.Email,
								req.Name,
								strings.TrimSpace(req.PromoCode),
								loggedInUserID,
								roleHint,
								buyerRole,
								req.Quantity,
								students,
							); retryErr == nil {
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
					if writeCheckoutPromoError(w, err) {
						return
					}
					writeInternalError(w, r, err)
					return
				}
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
			IsCollective:    result.IsCollective,
			Quantity:        result.Quantity,
			UnitPrice:       result.UnitPrice,
			Subtotal:        result.Subtotal,
			UniqueCode:      result.UniqueCode,
			Students:        toCheckoutStudentItems(result.Students),
		})
	}
}

func toCheckoutStudentItems(students []domain.OrderStudent) []dto.CheckoutStudentItem {
	if len(students) == 0 {
		return nil
	}
	out := make([]dto.CheckoutStudentItem, 0, len(students))
	for _, s := range students {
		out = append(out, dto.CheckoutStudentItem{
			Name:   s.Name,
			Email:  s.Email,
			UserID: s.UserID,
		})
	}
	return out
}

// CheckoutPaymentSession creates gateway session and returns payment URL. POST /api/v1/checkout/payment-session
func CheckoutPaymentSession(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			CheckoutID    string `json:"checkoutId"`
			OrderID       string `json:"orderId"`
			PaymentMethod string `json:"paymentMethod"`
			PromoCode     string `json:"promoCode"`
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
			writeInternalError(w, r, err)
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

// parseWebhookGrossAmountIDR mengurai gross_amount Midtrans (mis. "150000.00") ke rupiah integer.
func parseWebhookGrossAmountIDR(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return 0, false
	}
	return int(f + 0.5), true
}

func paymentWebhookSecretOK(r *http.Request, secret string) bool {
	if strings.TrimSpace(secret) == "" {
		return false
	}
	h := strings.TrimSpace(r.Header.Get("X-Payment-Webhook-Secret"))
	if h != "" && subtle.ConstantTimeCompare([]byte(h), []byte(secret)) == 1 {
		return true
	}
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if strings.HasPrefix(auth, prefix) {
		tok := strings.TrimSpace(auth[len(prefix):])
		if tok != "" && subtle.ConstantTimeCompare([]byte(tok), []byte(secret)) == 1 {
			return true
		}
	}
	return false
}

// PaymentWebhook handles gateway webhook (e.g. Midtrans). POST /api/v1/webhook/payment
// Body dari Midtrans mengikuti format resmi mereka (snake_case); verifikasi signature + nominal.
func PaymentWebhook(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			OrderID string `json:"orderId"`
			// Field Midtrans (nama key sesuai dokumentasi Midtrans, bukan konvensi API kita)
			MidOrderID        string `json:"order_id"`
			TransactionStatus string `json:"transaction_status"`
			FraudStatus       string `json:"fraud_status"`
			StatusCode        string `json:"status_code"`
			GrossAmount       string `json:"gross_amount"`
			SignatureKey      string `json:"signature_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Body permintaan tidak valid.")
			return
		}
		orderID := strings.TrimSpace(body.OrderID)
		if orderID == "" {
			orderID = strings.TrimSpace(body.MidOrderID)
		}
		if orderID == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "orderId wajib diisi.")
			return
		}

		isMidtrans := strings.TrimSpace(body.TransactionStatus) != "" ||
			strings.TrimSpace(body.SignatureKey) != "" ||
			(strings.TrimSpace(body.MidOrderID) != "" && strings.TrimSpace(body.StatusCode) != "")

		if isMidtrans {
			if strings.TrimSpace(deps.MidtransServerKey) == "" {
				writeError(w, http.StatusServiceUnavailable, "payment_gateway_unconfigured", "MIDTRANS_SERVER_KEY belum di-set; tidak bisa memverifikasi notifikasi Midtrans.")
				return
			}
			midOID := strings.TrimSpace(body.MidOrderID)
			if midOID == "" {
				midOID = orderID
			}
			ok := paymentgateway.VerifyMidtransSignature(
				midOID,
				body.StatusCode,
				body.GrossAmount,
				body.SignatureKey,
				deps.MidtransServerKey,
			)
			if !ok {
				writeError(w, http.StatusUnauthorized, "invalid_signature", "signature Midtrans tidak valid")
				return
			}
			tx := strings.ToLower(strings.TrimSpace(body.TransactionStatus))
			if tx != "settlement" && tx != "capture" {
				w.WriteHeader(http.StatusOK)
				return
			}
			if tx == "capture" && strings.ToLower(strings.TrimSpace(body.FraudStatus)) != "accept" {
				w.WriteHeader(http.StatusOK)
				return
			}
			amt, okAmt := parseWebhookGrossAmountIDR(body.GrossAmount)
			if !okAmt {
				writeError(w, http.StatusBadRequest, "gross_amount_invalid", "gross_amount Midtrans tidak valid atau kosong")
				return
			}
			order, err := deps.OrderRepo.GetByID(r.Context(), orderID)
			if err != nil {
				writeError(w, http.StatusNotFound, "order_not_found", "order tidak ditemukan")
				return
			}
			if order.TotalPrice != amt {
				writeError(w, http.StatusBadRequest, "amount_mismatch", "gross_amount tidak sama dengan total order")
				return
			}
		} else {
			if !paymentWebhookSecretOK(r, deps.PaymentWebhookSecret) {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Header X-Payment-Webhook-Secret atau Authorization: Bearer <PAYMENT_WEBHOOK_SECRET> wajib cocok dengan konfigurasi server")
				return
			}
		}

		if err := deps.CheckoutService.HandlePaymentWebhook(r.Context(), orderID); err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// UploadDir untuk bukti pembayaran (relatif dari working directory).
const paymentProofUploadDir = "uploads/payment-proofs"

const maxPaymentProofBytes = 5 << 20 // 5 MiB

var allowedPaymentProofContentTypes = map[string]struct{}{
	"image/jpeg": {}, "image/png": {}, "image/webp": {}, "application/pdf": {},
}

// CheckoutPaymentProof menerima upload bukti transfer. POST /api/v1/checkout/orders/:orderId/payment-proof
// FormData: proof (file), amount, senderAccountNo, senderName
func CheckoutPaymentProof(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "orderId")
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

// CompletePurchaseAuth bootstrap auth setelah purchase valid.
// POST /api/v1/checkout/orders/:orderId/complete-purchase-auth
// Jika user belum ada, akan auto-create dengan mustSetPassword=true.
// Hanya bisa dipanggil jika order status sudah valid (paid atau proof submitted).
func CompletePurchaseAuth(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := chi.URLParam(r, "orderId")
		if orderID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "orderId required")
			return
		}

		var req dto.CompletePurchaseAuthRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		order, err := deps.OrderRepo.GetByID(r.Context(), orderID)
		if err != nil {
			writeError(w, http.StatusNotFound, "order_not_found", "order tidak ditemukan")
			return
		}

		// Validasi status purchase - harus paid atau ada payment proof
		validStatus := order.Status == domain.OrderStatusPaid || order.PaymentProofURL != nil
		if !validStatus {
			writeError(w, http.StatusConflict, "purchase_not_completed", "order belum dibayar atau bukti pembayaran belum diupload")
			return
		}

		// Cari atau buat user
		var user domain.User
		var isNewUser bool

		// Coba cari user by order.UserID
		user, err = deps.UserRepo.FindByID(r.Context(), order.UserID)
		if err != nil {
			// User tidak ditemukan, coba cari by email dari order.BuyerEmail
			email := ""
			if order.BuyerEmail != nil {
				email = *order.BuyerEmail
			}
			if email == "" {
				writeError(w, http.StatusBadRequest, "bad_request", "order tidak memiliki email pembeli")
				return
			}
			user, err = deps.UserRepo.FindByEmail(r.Context(), email)
			if err != nil {
				// Auto-create user baru
				rh := strings.TrimSpace(req.RoleHint)
				var reqHint *string
				if rh != "" {
					reqHint = &rh
				}
				role, rerr := resolveCheckoutUserRoleCode(r.Context(), deps.RoleRepo, reqHint, order.RoleHint)
				if rerr != nil {
					writeInternalError(w, r, rerr)
					return
				}
				now := time.Now()
				user, err = deps.UserRepo.Create(r.Context(), domain.User{
					Email:           email,
					Name:            email, // Sementara pakai email sebagai nama
					PasswordHash:    "",
					Role:            role,
					EmailVerified:   true,
					EmailVerifiedAt: &now,
					MustSetPassword: true,
				})
				if err != nil {
					writeInternalError(w, r, err)
					return
				}
				isNewUser = true
			}
		}

		// Jika user existing dan belum punya password, set mustSetPassword
		if !isNewUser && user.PasswordHash == "" {
			user.MustSetPassword = true
			if err := deps.UserRepo.Update(r.Context(), user); err != nil {
				writeErrorFromUserRepoUpdate(w, r, err)
				return
			}
		}

		// Generate JWT token (bootstrap token untuk complete-purchase-auth)
		token := generateBootstrapToken(deps.JWTSecret, user.ID, user.Role)

		nextAction := ""
		if user.MustSetPassword {
			nextAction = "SET_PASSWORD"
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.CompletePurchaseAuthResponse{
			Token:           token,
			User:            userAuthMap(r.Context(), deps.RoleRepo, user),
			MustSetPassword: user.MustSetPassword,
			NextAction:      nextAction,
		})
	}
}

func writeCheckoutPromoError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, service.ErrPromoInvalid):
		writeError(w, http.StatusBadRequest, "invalid_promo", "kode promo tidak valid")
	case errors.Is(err, service.ErrPromoExpired):
		writeError(w, http.StatusBadRequest, "promo_expired", "kode promo sudah kadaluarsa")
	case errors.Is(err, service.ErrPromoMaxUses):
		writeError(w, http.StatusBadRequest, "promo_max_uses", "kode promo sudah mencapai batas penggunaan")
	case errors.Is(err, service.ErrPromoInactive):
		writeError(w, http.StatusBadRequest, "promo_inactive", "kode promo tidak aktif")
	case errors.Is(err, service.ErrPromoWrongScope):
		writeError(w, http.StatusBadRequest, "promo_wrong_scope", "kode promo tidak berlaku untuk jenis pembelian ini")
	case errors.Is(err, service.ErrPromoRequiresClaim):
		writeError(w, http.StatusBadRequest, "voucher_requires_claim", "voucher harus diklaim ke akun Anda terlebih dahulu")
	default:
		return false
	}
	return true
}

func generateBootstrapToken(jwtSecret []byte, userID, role string) string {
	role = strings.TrimSpace(role)
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := t.SignedString(jwtSecret)
	return tokenStr
}
