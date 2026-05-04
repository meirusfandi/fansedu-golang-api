package service

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/mail"
	"github.com/meirusfandi/fansedu-golang-api/internal/paymentgateway"
)

var (
	ErrCourseNotFound          = errors.New("course not found")
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderNotPending         = errors.New("order is not pending")
	ErrPromoInvalid            = errors.New("kode promo tidak valid")
	ErrPromoExpired            = errors.New("kode promo sudah kadaluarsa")
	ErrPromoMaxUses            = errors.New("kode promo sudah mencapai batas penggunaan")
	ErrPackageNoLinkedCourses  = errors.New("paket belum dihubungkan ke kelas manapun")
	ErrManualOrderNoCourses    = errors.New("minimal satu courseId diperlukan")
	ErrManualOrderUserNotFound = errors.New("pengguna tidak ditemukan")
)

type CheckoutInitiateResult struct {
	OrderID          string
	UserID           string
	TotalPrice       int // rupiah
	NormalPrice      int // rupiah
	PromoCode        string
	Discount         int // rupiah
	DiscountPercent  float64
	FinalPrice       int // rupiah
	ConfirmationCode string
	IsNewUser        bool
	CourseTitle      string
	Price            int // rupiah (harga course)
	IsCollective     bool
	Quantity         int
	UnitPrice        int
	Subtotal         int
	UniqueCode       int
	Students         []domain.OrderStudent
}

type CheckoutService interface {
	Initiate(ctx context.Context, courseSlug, email, name, promoCode, optionalLoggedInUserID, roleHint, buyerRole string, quantity int, students []domain.OrderStudent) (*CheckoutInitiateResult, error)
	// InitiatePackage checkout berdasarkan slug paket landing; satu pembayaran meng-enroll ke semua kelas di package_courses.
	InitiatePackage(ctx context.Context, packageSlug, email, name, promoCode, optionalLoggedInUserID, roleHint, buyerRole string, quantity int, students []domain.OrderStudent) (*CheckoutInitiateResult, error)
	CreatePaymentSession(ctx context.Context, orderID, paymentMethod string) (paymentURL string, err error)
	HandlePaymentWebhook(ctx context.Context, orderID string) error
	SubmitPaymentProof(ctx context.Context, orderID, proofURL, senderAccountNo, senderName string, proofAt *time.Time) error
	VerifyOrder(ctx context.Context, orderID string, purchasedAt *time.Time) error
	// CreateAdminManualOrder membuat order pending + order_items untuk input admin (tanpa checkout publik).
	CreateAdminManualOrder(ctx context.Context, userID string, courseIDs []string, totalPrice *int) (domain.Order, error)
}

type checkoutService struct {
	courseRepo     courseRepoForCheckout
	landingRepo    landingRepoForCheckout
	userRepo       userRepoForCheckout
	orderRepo      orderRepoForCheckout
	orderItemRepo  orderItemRepoForCheckout
	paymentRepo    paymentRepoForCheckout
	enrollmentRepo enrollmentRepoForCheckout
	promoRepo      promoRepoForCheckout
	mailer         mail.Mailer
	inviteRepo     inviteRepoForCheckout
	appURL         string
	midtrans       *paymentgateway.MidtransClient
}

type landingRepoForCheckout interface {
	GetBySlug(ctx context.Context, slug string) (domain.LandingPackage, error)
	GetByID(ctx context.Context, id string) (domain.LandingPackage, error)
	ListLinkedCourses(ctx context.Context, packageID string) ([]domain.PackageLinkedCourse, error)
}

type inviteRepoForCheckout interface {
	Create(ctx context.Context, inv domain.UserInvite) (domain.UserInvite, error)
	GetByOrderID(ctx context.Context, orderID string) (domain.UserInvite, error)
}

type promoRepoForCheckout interface {
	GetByCode(ctx context.Context, code string) (domain.PromoCode, error)
	IncrementUsedCount(ctx context.Context, id string) error
	HasUnusedClaim(ctx context.Context, userID, promoCodeID string) (bool, error)
	MarkClaimUsedForOrder(ctx context.Context, userID, promoCodeID, orderID string) error
}

type courseRepoForCheckout interface {
	GetBySlug(ctx context.Context, slug string) (domain.Course, error)
	GetByID(ctx context.Context, id string) (domain.Course, error)
}

type userRepoForCheckout interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
	Create(ctx context.Context, u domain.User) (domain.User, error)
}

type orderRepoForCheckout interface {
	Create(ctx context.Context, o domain.Order) (domain.Order, error)
	GetByID(ctx context.Context, id string) (domain.Order, error)
	GetPendingByUserAndCourse(ctx context.Context, userID, courseID string) (domain.Order, bool, error)
	GetPendingByUserAndPackage(ctx context.Context, userID, packageID string) (domain.Order, bool, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdatePaymentProof(ctx context.Context, orderID, proofURL, senderAccountNo, senderName string, proofAt *time.Time) error
	UpdateOrderCreatedAt(ctx context.Context, orderID string, createdAt time.Time) error
	UpdatePaymentProofAtOnly(ctx context.Context, orderID string, proofAt time.Time) error
}

type orderItemRepoForCheckout interface {
	Create(ctx context.Context, oi domain.OrderItem) (domain.OrderItem, error)
	ListByOrderID(ctx context.Context, orderID string) ([]domain.OrderItem, error)
}

type paymentRepoForCheckout interface {
	Create(ctx context.Context, p domain.Payment) (domain.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error)
	Update(ctx context.Context, p domain.Payment) error
	MarkPaidByOrderID(ctx context.Context, orderID string, paidAt time.Time) error
	BackdateByOrderID(ctx context.Context, orderID string, t time.Time) error
}

type enrollmentRepoForCheckout interface {
	GetByUserAndCourse(ctx context.Context, userID, courseID string) (domain.CourseEnrollment, error)
	Create(ctx context.Context, e domain.CourseEnrollment) (domain.CourseEnrollment, error)
}

func NewCheckoutService(
	courseRepo courseRepoForCheckout,
	landingRepo landingRepoForCheckout,
	userRepo userRepoForCheckout,
	orderRepo orderRepoForCheckout,
	orderItemRepo orderItemRepoForCheckout,
	paymentRepo paymentRepoForCheckout,
	enrollmentRepo enrollmentRepoForCheckout,
	promoRepo promoRepoForCheckout,
	mailer mail.Mailer,
	inviteRepo inviteRepoForCheckout,
	appURL string,
	midtrans *paymentgateway.MidtransClient,
) CheckoutService {
	return &checkoutService{
		courseRepo:     courseRepo,
		landingRepo:    landingRepo,
		userRepo:       userRepo,
		orderRepo:      orderRepo,
		orderItemRepo:  orderItemRepo,
		paymentRepo:    paymentRepo,
		enrollmentRepo: enrollmentRepo,
		promoRepo:      promoRepo,
		mailer:         mailer,
		inviteRepo:     inviteRepo,
		appURL:         appURL,
		midtrans:       midtrans,
	}
}

func (s *checkoutService) Initiate(ctx context.Context, courseSlug, email, name, promoCode, optionalLoggedInUserID, roleHint, buyerRole string, quantity int, students []domain.OrderStudent) (*CheckoutInitiateResult, error) {
	course, err := s.courseRepo.GetBySlug(ctx, courseSlug)
	if err != nil {
		return nil, ErrCourseNotFound
	}
	if course.Slug == nil || *course.Slug == "" {
		return nil, ErrCourseNotFound
	}

	var user domain.User
	isNewUser := false
	if optionalLoggedInUserID != "" {
		user, err = s.userRepo.FindByID(ctx, optionalLoggedInUserID)
		if err != nil {
			return nil, err
		}
	} else {
		user, err = s.userRepo.FindByEmail(ctx, email)
		if err != nil {
			now := time.Now()
			role := domain.UserRoleStudent
			if roleHint == "instructor" || roleHint == "guru" {
				role = domain.UserRoleGuru
			}
			user, err = s.userRepo.Create(ctx, domain.User{
				Email:           email,
				Name:            name,
				PasswordHash:    "",
				Role:            role,
				EmailVerified:   true,
				EmailVerifiedAt: &now,
				MustSetPassword: true,
			})
			if err != nil {
				return nil, err
			}
			isNewUser = true
		}
	}

	effectiveBuyerRole := normalizeBuyerRole(buyerRole, roleHint, user.Role)
	isCollective := effectiveBuyerRole == domain.UserRoleGuru && quantity > 1
	if effectiveBuyerRole == domain.UserRoleStudent {
		quantity = 1
		isCollective = false
	}
	if quantity <= 0 {
		quantity = 1
	}

	existingOrder, found, err := s.orderRepo.GetPendingByUserAndCourse(ctx, user.ID, course.ID)
	if err != nil {
		return nil, err
	}
	if found {
		if existingOrder.Quantity == quantity && existingOrder.IsCollective == isCollective {
			dp := 0.0
			if existingOrder.DiscountPercent != nil {
				dp = *existingOrder.DiscountPercent
			}
			promoStr := ""
			if existingOrder.PromoCode != nil {
				promoStr = *existingOrder.PromoCode
			}
			confStr := ""
			if existingOrder.ConfirmationCode != nil {
				confStr = *existingOrder.ConfirmationCode
			}
			existingStudents := parseOrderStudents(existingOrder.StudentsJSON)
			return &CheckoutInitiateResult{
				OrderID:          existingOrder.ID,
				UserID:           existingOrder.UserID,
				TotalPrice:       existingOrder.TotalPrice,
				NormalPrice:      existingOrder.NormalPrice,
				PromoCode:        promoStr,
				Discount:         existingOrder.Discount,
				DiscountPercent:  dp,
				FinalPrice:       existingOrder.TotalPrice,
				ConfirmationCode: confStr,
				IsNewUser:        isNewUser,
				CourseTitle:      course.Title,
				Price:            course.Price,
				IsCollective:     existingOrder.IsCollective,
				Quantity:         existingOrder.Quantity,
				UnitPrice:        existingOrder.UnitPrice,
				Subtotal:         existingOrder.Subtotal,
				UniqueCode:       existingOrder.UniqueCode,
				Students:         existingStudents,
			}, nil
		}
	}

	// Semua harga dalam rupiah. Promo fixed: DiscountValue = rupiah.
	normalRupiah := course.Price * quantity
	discountRupiah, discountPercent, appliedPromoCode, appliedPromo, err := s.applyPromoDiscount(ctx, promoCode, user.ID, normalRupiah, true)
	if err != nil {
		return nil, err
	}

	finalRupiah := normalRupiah - discountRupiah
	if finalRupiah < 0 {
		finalRupiah = 0
	}
	confirmationCode := generateConfirmationCode()
	uniqueCode := 0
	_, _ = fmt.Sscanf(confirmationCode, "%d", &uniqueCode)
	totalWithUniqueCode := finalRupiah + uniqueCode
	unitPrice := finalRupiah / quantity
	if unitPrice <= 0 {
		unitPrice = course.Price
	}
	studentsJSON, _ := json.Marshal(students)

	var promoCodePtr *string
	if appliedPromoCode != "" {
		promoCodePtr = &appliedPromoCode
	}
	var roleHintPtr *string
	if roleHint != "" {
		roleHintPtr = &roleHint
	}
	order, err := s.orderRepo.Create(ctx, domain.Order{
		UserID:           user.ID,
		Status:           domain.OrderStatusPending,
		TotalPrice:       totalWithUniqueCode,
		NormalPrice:      normalRupiah,
		Quantity:         quantity,
		UnitPrice:        unitPrice,
		Subtotal:         finalRupiah,
		UniqueCode:       uniqueCode,
		IsCollective:     isCollective,
		StudentsJSON:     studentsJSON,
		PromoCode:        promoCodePtr,
		Discount:         discountRupiah,
		DiscountPercent:  discountPercent,
		ConfirmationCode: &confirmationCode,
		RoleHint:         roleHintPtr,
		BuyerEmail:       &email,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.orderItemRepo.Create(ctx, domain.OrderItem{
		OrderID:  order.ID,
		CourseID: course.ID,
		Price:    normalRupiah,
	})
	if err != nil {
		return nil, err
	}

	s.incrementPromoIfNeededOnInitiate(ctx, appliedPromoCode, appliedPromo)

	dp := 0.0
	if discountPercent != nil {
		dp = *discountPercent
	}

	// Invite link untuk user baru (guest)
	registerLink := ""
	if isNewUser && s.inviteRepo != nil {
		token := generateInviteToken()
		orderIDStr := order.ID
		inv := domain.UserInvite{
			UserID:    user.ID,
			OrderID:   &orderIDStr,
			Email:     email,
			Name:      name,
			Token:     token,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		if _, err := s.inviteRepo.Create(ctx, inv); err == nil {
			registerLink = s.appURL + "/#/register?token=" + url.QueryEscape(token) + "&email=" + url.QueryEscape(email)
		}
	}

	// Email konfirmasi pesanan
	if s.mailer != nil {
		body := mail.OrderConfirmationBody(order.ID, course.Title, finalRupiah, confirmationCode, registerLink)
		_ = s.mailer.Send(email, "Konfirmasi Pesanan - "+course.Title, body)
	}

	return &CheckoutInitiateResult{
		OrderID:          order.ID,
		UserID:           user.ID,
		TotalPrice:       totalWithUniqueCode,
		NormalPrice:      normalRupiah,
		PromoCode:        appliedPromoCode,
		Discount:         discountRupiah,
		DiscountPercent:  dp,
		FinalPrice:       totalWithUniqueCode,
		ConfirmationCode: confirmationCode,
		IsNewUser:        isNewUser,
		CourseTitle:      course.Title,
		Price:            course.Price,
		IsCollective:     isCollective,
		Quantity:         quantity,
		UnitPrice:        unitPrice,
		Subtotal:         finalRupiah,
		UniqueCode:       uniqueCode,
		Students:         students,
	}, nil
}

func (s *checkoutService) InitiatePackage(ctx context.Context, packageSlug, email, name, promoCode, optionalLoggedInUserID, roleHint, buyerRole string, quantity int, students []domain.OrderStudent) (*CheckoutInitiateResult, error) {
	if s.landingRepo == nil {
		return nil, ErrCourseNotFound
	}
	pkg, err := s.landingRepo.GetBySlug(ctx, packageSlug)
	if err != nil {
		return nil, ErrCourseNotFound
	}
	linked, err := s.landingRepo.ListLinkedCourses(ctx, pkg.ID)
	if err != nil {
		return nil, err
	}
	if len(linked) == 0 {
		return nil, ErrPackageNoLinkedCourses
	}

	var user domain.User
	isNewUser := false
	if optionalLoggedInUserID != "" {
		user, err = s.userRepo.FindByID(ctx, optionalLoggedInUserID)
		if err != nil {
			return nil, err
		}
	} else {
		user, err = s.userRepo.FindByEmail(ctx, email)
		if err != nil {
			now := time.Now()
			role := domain.UserRoleStudent
			if roleHint == "instructor" || roleHint == "guru" {
				role = domain.UserRoleGuru
			}
			user, err = s.userRepo.Create(ctx, domain.User{
				Email:           email,
				Name:            name,
				PasswordHash:    "",
				Role:            role,
				EmailVerified:   true,
				EmailVerifiedAt: &now,
				MustSetPassword: true,
			})
			if err != nil {
				return nil, err
			}
			isNewUser = true
		}
	}

	effectiveBuyerRole := normalizeBuyerRole(buyerRole, roleHint, user.Role)
	isCollective := effectiveBuyerRole == domain.UserRoleGuru && quantity > 1
	if effectiveBuyerRole == domain.UserRoleStudent {
		quantity = 1
		isCollective = false
	}
	if quantity <= 0 {
		quantity = 1
	}

	unitPrice := domain.LandingPackagePriceRupiah(pkg)

	existingOrder, found, err := s.orderRepo.GetPendingByUserAndPackage(ctx, user.ID, pkg.ID)
	if err != nil {
		return nil, err
	}
	if found {
		if existingOrder.Quantity == quantity && existingOrder.IsCollective == isCollective {
			dp := 0.0
			if existingOrder.DiscountPercent != nil {
				dp = *existingOrder.DiscountPercent
			}
			promoStr := ""
			if existingOrder.PromoCode != nil {
				promoStr = *existingOrder.PromoCode
			}
			confStr := ""
			if existingOrder.ConfirmationCode != nil {
				confStr = *existingOrder.ConfirmationCode
			}
			existingStudents := parseOrderStudents(existingOrder.StudentsJSON)
			return &CheckoutInitiateResult{
				OrderID:          existingOrder.ID,
				UserID:           existingOrder.UserID,
				TotalPrice:       existingOrder.TotalPrice,
				NormalPrice:      existingOrder.NormalPrice,
				PromoCode:        promoStr,
				Discount:         existingOrder.Discount,
				DiscountPercent:  dp,
				FinalPrice:       existingOrder.TotalPrice,
				ConfirmationCode: confStr,
				IsNewUser:        isNewUser,
				CourseTitle:      pkg.Name,
				Price:            unitPrice,
				IsCollective:     existingOrder.IsCollective,
				Quantity:         existingOrder.Quantity,
				UnitPrice:        existingOrder.UnitPrice,
				Subtotal:         existingOrder.Subtotal,
				UniqueCode:       existingOrder.UniqueCode,
				Students:         existingStudents,
			}, nil
		}
	}

	normalRupiah := unitPrice * quantity
	discountRupiah, discountPercent, appliedPromoCode, appliedPromoPkg, err := s.applyPromoDiscount(ctx, promoCode, user.ID, normalRupiah, false)
	if err != nil {
		return nil, err
	}

	finalRupiah := normalRupiah - discountRupiah
	if finalRupiah < 0 {
		finalRupiah = 0
	}
	confirmationCode := generateConfirmationCode()
	uniqueCode := 0
	_, _ = fmt.Sscanf(confirmationCode, "%d", &uniqueCode)
	totalWithUniqueCode := finalRupiah + uniqueCode
	unitPriceAfter := finalRupiah / quantity
	if unitPriceAfter <= 0 {
		unitPriceAfter = unitPrice
	}
	studentsJSON, _ := json.Marshal(students)

	var promoCodePtr *string
	if appliedPromoCode != "" {
		promoCodePtr = &appliedPromoCode
	}
	var roleHintPtr *string
	if roleHint != "" {
		roleHintPtr = &roleHint
	}
	pkgID := pkg.ID
	order, err := s.orderRepo.Create(ctx, domain.Order{
		UserID:           user.ID,
		Status:           domain.OrderStatusPending,
		TotalPrice:       totalWithUniqueCode,
		NormalPrice:      normalRupiah,
		Quantity:         quantity,
		UnitPrice:        unitPriceAfter,
		Subtotal:         finalRupiah,
		UniqueCode:       uniqueCode,
		IsCollective:     isCollective,
		StudentsJSON:     studentsJSON,
		PromoCode:        promoCodePtr,
		Discount:         discountRupiah,
		DiscountPercent:  discountPercent,
		ConfirmationCode: &confirmationCode,
		RoleHint:         roleHintPtr,
		BuyerEmail:       &email,
		PackageID:        &pkgID,
	})
	if err != nil {
		return nil, err
	}

	for i, lc := range linked {
		linePrice := 0
		if i == 0 {
			linePrice = normalRupiah
		}
		_, err = s.orderItemRepo.Create(ctx, domain.OrderItem{
			OrderID:  order.ID,
			CourseID: lc.ID,
			Price:    linePrice,
		})
		if err != nil {
			return nil, err
		}
	}

	s.incrementPromoIfNeededOnInitiate(ctx, appliedPromoCode, appliedPromoPkg)

	dp := 0.0
	if discountPercent != nil {
		dp = *discountPercent
	}

	registerLink := ""
	if isNewUser && s.inviteRepo != nil {
		token := generateInviteToken()
		orderIDStr := order.ID
		inv := domain.UserInvite{
			UserID:    user.ID,
			OrderID:   &orderIDStr,
			Email:     email,
			Name:      name,
			Token:     token,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}
		if _, err := s.inviteRepo.Create(ctx, inv); err == nil {
			registerLink = s.appURL + "/#/register?token=" + url.QueryEscape(token) + "&email=" + url.QueryEscape(email)
		}
	}

	if s.mailer != nil {
		body := mail.OrderConfirmationBody(order.ID, pkg.Name, finalRupiah, confirmationCode, registerLink)
		_ = s.mailer.Send(email, "Konfirmasi Pesanan - "+pkg.Name, body)
	}

	return &CheckoutInitiateResult{
		OrderID:          order.ID,
		UserID:           user.ID,
		TotalPrice:       totalWithUniqueCode,
		NormalPrice:      normalRupiah,
		PromoCode:        appliedPromoCode,
		Discount:         discountRupiah,
		DiscountPercent:  dp,
		FinalPrice:       totalWithUniqueCode,
		ConfirmationCode: confirmationCode,
		IsNewUser:        isNewUser,
		CourseTitle:      pkg.Name,
		Price:            unitPrice,
		IsCollective:     isCollective,
		Quantity:         quantity,
		UnitPrice:        unitPriceAfter,
		Subtotal:         finalRupiah,
		UniqueCode:       uniqueCode,
		Students:         students,
	}, nil
}

func parseOrderStudents(raw []byte) []domain.OrderStudent {
	if len(raw) == 0 {
		return nil
	}
	var out []domain.OrderStudent
	_ = json.Unmarshal(raw, &out)
	return out
}

func normalizeBuyerRole(buyerRole, roleHint, userRole string) string {
	if strings.TrimSpace(userRole) != "" {
		if userRole == domain.UserRoleGuru || userRole == "instructor" {
			return domain.UserRoleGuru
		}
		return domain.UserRoleStudent
	}
	for _, v := range []string{buyerRole, roleHint} {
		n := strings.ToLower(strings.TrimSpace(v))
		if n == "guru" || n == "instructor" {
			return domain.UserRoleGuru
		}
		if n == "student" || n == "siswa" {
			return domain.UserRoleStudent
		}
	}
	return domain.UserRoleStudent
}

func generateConfirmationCode() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := 100 + rng.Intn(900)
	return fmt.Sprintf("%03d", n)
}

func generateInviteToken() string {
	b := make([]byte, 24)
	_, _ = cryptorand.Read(b)
	return hex.EncodeToString(b)
}

func (s *checkoutService) CreatePaymentSession(ctx context.Context, orderID, paymentMethod string) (string, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return "", ErrOrderNotFound
	}
	if order.Status != domain.OrderStatusPending {
		return "", ErrOrderNotPending
	}

	ref := orderID
	_, err = s.paymentRepo.Create(ctx, domain.Payment{
		UserID:      order.UserID,
		OrderID:     &orderID,
		Amount:      order.TotalPrice,
		Currency:    "IDR",
		Status:      domain.PaymentStatusPending,
		Type:        domain.PaymentTypeCoursePurchase,
		Gateway:     &paymentMethod,
		ReferenceID: &ref,
	})
	if err != nil {
		return "", err
	}

	// If Midtrans configured, create real Snap transaction.
	if s.midtrans != nil && s.midtrans.Enabled() {
		user, _ := s.userRepo.FindByID(ctx, order.UserID)
		var mreq paymentgateway.CreateSnapRequest
		mreq.OrderID = orderID
		mreq.Amount = order.TotalPrice
		mreq.Customer.FirstName = user.Name
		mreq.Customer.Email = user.Email
		mres, err := s.midtrans.CreateSnapTransaction(ctx, mreq)
		if err != nil {
			return "", err
		}
		return mres.RedirectURL, nil
	}
	// Fallback local placeholder for dev-only without gateway keys.
	return "/checkout/pay?order_id=" + orderID + "&method=" + paymentMethod, nil
}

func (s *checkoutService) HandlePaymentWebhook(ctx context.Context, orderID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := s.paymentRepo.MarkPaidByOrderID(ctx, orderID, time.Now()); err != nil {
		return err
	}

	if order.Status != domain.OrderStatusPaid {
		if err := s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusPaid); err != nil {
			return err
		}

		order.Status = domain.OrderStatusPaid
		s.finalizePromoAfterPayment(ctx, order)
	}

	items, err := s.orderItemRepo.ListByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := s.enrollOrderTargets(ctx, order, items, time.Now()); err != nil {
		return err
	}

	return nil
}

// SubmitPaymentProof menyimpan bukti transfer dan mengirim email "Bukti Diterima".
func (s *checkoutService) SubmitPaymentProof(ctx context.Context, orderID, proofURL, senderAccountNo, senderName string, proofAt *time.Time) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.Status != domain.OrderStatusPending {
		return ErrOrderNotPending
	}
	if err := s.orderRepo.UpdatePaymentProof(ctx, orderID, proofURL, senderAccountNo, senderName, proofAt); err != nil {
		return err
	}
	if p, err := s.paymentRepo.GetByOrderID(ctx, orderID); err == nil {
		p.ProofURL = &proofURL
		if p.Gateway == nil {
			gw := "bank_transfer"
			p.Gateway = &gw
		}
		if err := s.paymentRepo.Update(ctx, p); err != nil {
			return err
		}
	} else if errors.Is(err, pgx.ErrNoRows) {
		ref := orderID
		gw := "bank_transfer"
		if _, err := s.paymentRepo.Create(ctx, domain.Payment{
			UserID:      order.UserID,
			OrderID:     &orderID,
			Amount:      order.TotalPrice,
			Currency:    "IDR",
			Status:      domain.PaymentStatusPending,
			Type:        domain.PaymentTypeCoursePurchase,
			Gateway:     &gw,
			ReferenceID: &ref,
			ProofURL:    &proofURL,
		}); err != nil {
			return err
		}
	} else {
		return err
	}
	items, _ := s.orderItemRepo.ListByOrderID(ctx, orderID)
	programName := s.orderProgramTitle(ctx, order, items)
	user, _ := s.userRepo.FindByID(ctx, order.UserID)
	if s.mailer != nil && user.Email != "" {
		body := mail.PaymentProofReceivedBody(orderID, programName)
		_ = s.mailer.Send(user.Email, "Bukti Pembayaran Diterima - "+programName, body)
	}
	return nil
}

// VerifyOrder dipanggil admin: set order paid, enroll user, kirim email "Pembayaran Terverifikasi".
// purchasedAt opsional: menyelaraskan tanggal pembelian (order.created_at, enrollment.enrolled_at, payments terkait order).
func (s *checkoutService) VerifyOrder(ctx context.Context, orderID string, purchasedAt *time.Time) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrOrderNotFound
	}
	if err := s.paymentRepo.MarkPaidByOrderID(ctx, orderID, time.Now()); err != nil {
		return err
	}
	newlyPaid := false
	if order.Status != domain.OrderStatusPaid {
		if err := s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusPaid); err != nil {
			return err
		}
		order.Status = domain.OrderStatusPaid
		s.finalizePromoAfterPayment(ctx, order)
		newlyPaid = true
	}

	items, err := s.orderItemRepo.ListByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	enrolledAt := time.Now()
	if purchasedAt != nil {
		enrolledAt = *purchasedAt
	}
	if err := s.enrollOrderTargets(ctx, order, items, enrolledAt); err != nil {
		return err
	}
	if purchasedAt != nil {
		_ = s.orderRepo.UpdateOrderCreatedAt(ctx, orderID, *purchasedAt)
		_ = s.paymentRepo.BackdateByOrderID(ctx, orderID, *purchasedAt)
	}
	if newlyPaid {
		programName := s.orderProgramTitle(ctx, order, items)
		user, _ := s.userRepo.FindByID(ctx, order.UserID)
		registerLink := ""
		if s.inviteRepo != nil && user.Email != "" {
			inv, err := s.inviteRepo.GetByOrderID(ctx, orderID)
			if err == nil && inv.UsedAt == nil {
				registerLink = s.appURL + "/#/register?token=" + inv.Token + "&email=" + url.QueryEscape(user.Email)
			}
		}
		if s.mailer != nil && user.Email != "" {
			body := mail.PaymentVerifiedBody(programName, registerLink)
			_ = s.mailer.Send(user.Email, "Pembayaran Terverifikasi - "+programName, body)
		}
	}
	return nil
}

func (s *checkoutService) CreateAdminManualOrder(ctx context.Context, userID string, courseIDs []string, totalPrice *int) (domain.Order, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domain.Order{}, ErrManualOrderUserNotFound
	}
	if _, err := s.userRepo.FindByID(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, ErrManualOrderUserNotFound
		}
		return domain.Order{}, err
	}
	seen := make(map[string]struct{})
	var uniq []string
	for _, raw := range courseIDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	if len(uniq) == 0 {
		return domain.Order{}, ErrManualOrderNoCourses
	}
	sum := 0
	prices := make(map[string]int, len(uniq))
	for _, cid := range uniq {
		c, err := s.courseRepo.GetByID(ctx, cid)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.Order{}, ErrCourseNotFound
			}
			return domain.Order{}, err
		}
		prices[cid] = c.Price
		sum += c.Price
	}
	total := sum
	if totalPrice != nil && *totalPrice >= 0 {
		total = *totalPrice
	}
	order, err := s.orderRepo.Create(ctx, domain.Order{
		UserID:      userID,
		Status:      domain.OrderStatusPending,
		TotalPrice:  total,
		NormalPrice: sum,
		Quantity:    1,
	})
	if err != nil {
		return domain.Order{}, err
	}
	for _, cid := range uniq {
		if _, err := s.orderItemRepo.Create(ctx, domain.OrderItem{
			OrderID:  order.ID,
			CourseID: cid,
			Price:    prices[cid],
		}); err != nil {
			return domain.Order{}, err
		}
	}
	return order, nil
}

func (s *checkoutService) orderProgramTitle(ctx context.Context, order domain.Order, items []domain.OrderItem) string {
	if order.PackageID != nil && s.landingRepo != nil {
		if pkg, err := s.landingRepo.GetByID(ctx, *order.PackageID); err == nil && pkg.Name != "" {
			return pkg.Name
		}
	}
	if len(items) > 0 {
		if course, err := s.courseRepo.GetByID(ctx, items[0].CourseID); err == nil {
			return course.Title
		}
	}
	return "Program"
}

// enrollOrderTargets enrolls buyer (default) or collective students (if provided).
func (s *checkoutService) enrollOrderTargets(ctx context.Context, order domain.Order, items []domain.OrderItem, enrolledAt time.Time) error {
	targetUserIDs := collectOrderTargetUserIDs(order)
	for _, targetUserID := range targetUserIDs {
		for _, item := range items {
			_, err := s.enrollmentRepo.GetByUserAndCourse(ctx, targetUserID, item.CourseID)
			if err == nil {
				continue
			}
			if _, err := s.enrollmentRepo.Create(ctx, domain.CourseEnrollment{
				UserID:     targetUserID,
				CourseID:   item.CourseID,
				Status:     domain.EnrollmentStatusEnrolled,
				EnrolledAt: enrolledAt,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func collectOrderTargetUserIDs(order domain.Order) []string {
	if order.IsCollective {
		uniq := map[string]struct{}{}
		out := make([]string, 0)
		students := parseOrderStudents(order.StudentsJSON)
		for _, st := range students {
			if st.UserID == nil {
				continue
			}
			userID := strings.TrimSpace(*st.UserID)
			if userID == "" {
				continue
			}
			if _, ok := uniq[userID]; ok {
				continue
			}
			uniq[userID] = struct{}{}
			out = append(out, userID)
		}
		if len(out) > 0 {
			return out
		}
	}
	return []string{order.UserID}
}
