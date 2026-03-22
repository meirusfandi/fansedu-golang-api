package service

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/json"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/mail"
)

var (
	ErrCourseNotFound  = errors.New("course not found")
	ErrOrderNotFound   = errors.New("order not found")
	ErrOrderNotPending = errors.New("order is not pending")
	ErrPromoInvalid    = errors.New("kode promo tidak valid")
	ErrPromoExpired    = errors.New("kode promo sudah kadaluarsa")
	ErrPromoMaxUses    = errors.New("kode promo sudah mencapai batas penggunaan")
)

type CheckoutInitiateResult struct {
	OrderID          string
	UserID           string
	TotalPrice       int    // rupiah
	NormalPrice      int    // rupiah
	PromoCode        string
	Discount         int    // rupiah
	DiscountPercent  float64
	FinalPrice       int    // rupiah
	ConfirmationCode string
	IsNewUser        bool
	CourseTitle      string
	Price            int    // rupiah (harga course)
	IsCollective     bool
	Quantity         int
	UnitPrice        int
	Subtotal         int
	UniqueCode       int
	Students         []domain.OrderStudent
}

type CheckoutService interface {
	Initiate(ctx context.Context, courseSlug, email, name, promoCode, optionalLoggedInUserID, roleHint, buyerRole string, quantity int, students []domain.OrderStudent) (*CheckoutInitiateResult, error)
	CreatePaymentSession(ctx context.Context, orderID, paymentMethod string) (paymentURL string, err error)
	HandlePaymentWebhook(ctx context.Context, orderID string) error
	SubmitPaymentProof(ctx context.Context, orderID, proofURL, senderAccountNo, senderName string) error
	VerifyOrder(ctx context.Context, orderID string) error
}

type checkoutService struct {
	courseRepo     courseRepoForCheckout
	userRepo       userRepoForCheckout
	orderRepo      orderRepoForCheckout
	orderItemRepo  orderItemRepoForCheckout
	paymentRepo    paymentRepoForCheckout
	enrollmentRepo enrollmentRepoForCheckout
	promoRepo      promoRepoForCheckout
	mailer         mail.Mailer
	inviteRepo     inviteRepoForCheckout
	appURL         string
}

type inviteRepoForCheckout interface {
	Create(ctx context.Context, inv domain.UserInvite) (domain.UserInvite, error)
	GetByOrderID(ctx context.Context, orderID string) (domain.UserInvite, error)
}

type promoRepoForCheckout interface {
	GetByCode(ctx context.Context, code string) (domain.PromoCode, error)
	IncrementUsedCount(ctx context.Context, id string) error
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
	UpdateStatus(ctx context.Context, id, status string) error
	UpdatePaymentProof(ctx context.Context, orderID, proofURL, senderAccountNo, senderName string) error
}

type orderItemRepoForCheckout interface {
	Create(ctx context.Context, oi domain.OrderItem) (domain.OrderItem, error)
	ListByOrderID(ctx context.Context, orderID string) ([]domain.OrderItem, error)
}

type paymentRepoForCheckout interface {
	Create(ctx context.Context, p domain.Payment) (domain.Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error)
	Update(ctx context.Context, p domain.Payment) error
}

type enrollmentRepoForCheckout interface {
	GetByUserAndCourse(ctx context.Context, userID, courseID string) (domain.CourseEnrollment, error)
	Create(ctx context.Context, e domain.CourseEnrollment) (domain.CourseEnrollment, error)
}

func NewCheckoutService(
	courseRepo courseRepoForCheckout,
	userRepo userRepoForCheckout,
	orderRepo orderRepoForCheckout,
	orderItemRepo orderItemRepoForCheckout,
	paymentRepo paymentRepoForCheckout,
	enrollmentRepo enrollmentRepoForCheckout,
	promoRepo promoRepoForCheckout,
	mailer mail.Mailer,
	inviteRepo inviteRepoForCheckout,
	appURL string,
) CheckoutService {
	return &checkoutService{
		courseRepo:     courseRepo,
		userRepo:       userRepo,
		orderRepo:      orderRepo,
		orderItemRepo:  orderItemRepo,
		paymentRepo:    paymentRepo,
		enrollmentRepo: enrollmentRepo,
		promoRepo:      promoRepo,
		mailer:         mailer,
		inviteRepo:     inviteRepo,
		appURL:         appURL,
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
	discountRupiah := 0
	var discountPercent *float64
	appliedPromoCode := ""

	if promoCode != "" && s.promoRepo != nil {
		promo, err := s.promoRepo.GetByCode(ctx, promoCode)
		if err != nil {
			return nil, ErrPromoInvalid
		}
		now := time.Now()
		if now.Before(promo.ValidFrom) {
			return nil, ErrPromoInvalid
		}
		if promo.ValidUntil != nil && now.After(*promo.ValidUntil) {
			return nil, ErrPromoExpired
		}
		if promo.MaxUses != nil && promo.UsedCount >= *promo.MaxUses {
			return nil, ErrPromoMaxUses
		}
		switch promo.DiscountType {
		case domain.PromoDiscountTypePercent:
			if promo.DiscountValue <= 0 || promo.DiscountValue > 100 {
				return nil, ErrPromoInvalid
			}
			discountRupiah = normalRupiah * promo.DiscountValue / 100
			p := float64(promo.DiscountValue)
			discountPercent = &p
		case domain.PromoDiscountTypeFixed:
			if promo.DiscountValue <= 0 {
				return nil, ErrPromoInvalid
			}
			discountValRupiah := promo.DiscountValue
			if discountValRupiah >= normalRupiah {
				discountRupiah = normalRupiah
				p := 100.0
				discountPercent = &p
			} else {
				discountRupiah = discountValRupiah
				p := float64(discountValRupiah*100) / float64(normalRupiah)
				discountPercent = &p
			}
		default:
			return nil, ErrPromoInvalid
		}
		appliedPromoCode = promo.Code
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
		OrderID:   order.ID,
		CourseID:  course.ID,
		Price:     normalRupiah,
	})
	if err != nil {
		return nil, err
	}

	if appliedPromoCode != "" && s.promoRepo != nil {
		promo, _ := s.promoRepo.GetByCode(ctx, appliedPromoCode)
		_ = s.promoRepo.IncrementUsedCount(ctx, promo.ID)
	}

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
		UserID:   order.UserID,
		OrderID:  &orderID,
		Amount:   order.TotalPrice,
		Currency: "IDR",
		Status:        domain.PaymentStatusPending,
		Type:          domain.PaymentTypeCoursePurchase,
		Gateway:       &paymentMethod,
		ReferenceID:   &ref,
	})
	if err != nil {
		return "", err
	}

	// Stub: return a placeholder URL. Replace with real Midtrans/Stripe API call.
	return "/checkout/pay?order_id=" + orderID + "&method=" + paymentMethod, nil
}

func (s *checkoutService) HandlePaymentWebhook(ctx context.Context, orderID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status == domain.OrderStatusPaid {
		return nil
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusPaid); err != nil {
		return err
	}

	items, err := s.orderItemRepo.ListByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := s.enrollOrderTargets(ctx, order, items); err != nil {
		return err
	}

	return nil
}

// SubmitPaymentProof menyimpan bukti transfer dan mengirim email "Bukti Diterima".
func (s *checkoutService) SubmitPaymentProof(ctx context.Context, orderID, proofURL, senderAccountNo, senderName string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.Status != domain.OrderStatusPending {
		return ErrOrderNotPending
	}
	if err := s.orderRepo.UpdatePaymentProof(ctx, orderID, proofURL, senderAccountNo, senderName); err != nil {
		return err
	}
	programName := ""
	if items, err := s.orderItemRepo.ListByOrderID(ctx, orderID); err == nil && len(items) > 0 {
		if course, err := s.courseRepo.GetByID(ctx, items[0].CourseID); err == nil {
			programName = course.Title
		}
	}
	if programName == "" {
		programName = "Program"
	}
	user, _ := s.userRepo.FindByID(ctx, order.UserID)
	if s.mailer != nil && user.Email != "" {
		body := mail.PaymentProofReceivedBody(orderID, programName)
		_ = s.mailer.Send(user.Email, "Bukti Pembayaran Diterima - "+programName, body)
	}
	return nil
}

// VerifyOrder dipanggil admin: set order paid, enroll user, kirim email "Pembayaran Terverifikasi".
func (s *checkoutService) VerifyOrder(ctx context.Context, orderID string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.Status == domain.OrderStatusPaid {
		return nil
	}
	if err := s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusPaid); err != nil {
		return err
	}
	items, err := s.orderItemRepo.ListByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := s.enrollOrderTargets(ctx, order, items); err != nil {
		return err
	}
	programName := ""
	if len(items) > 0 {
		if course, err := s.courseRepo.GetByID(ctx, items[0].CourseID); err == nil {
			programName = course.Title
		}
	}
	if programName == "" {
		programName = "Program"
	}
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
	return nil
}

// enrollOrderTargets enrolls buyer (default) or collective students (if provided).
func (s *checkoutService) enrollOrderTargets(ctx context.Context, order domain.Order, items []domain.OrderItem) error {
	targetUserIDs := collectOrderTargetUserIDs(order)
	now := time.Now()
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
				EnrolledAt: now,
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
