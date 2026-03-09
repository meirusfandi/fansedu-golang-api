package service

import (
	"context"
	"errors"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var (
	ErrCourseNotFound   = errors.New("course not found")
	ErrOrderNotFound   = errors.New("order not found")
	ErrOrderNotPending = errors.New("order is not pending")
)

type CheckoutInitiateResult struct {
	OrderID    string
	UserID     string
	TotalPrice int
	IsNewUser   bool
	CourseTitle string
	PriceCents  int
}

type CheckoutService interface {
	Initiate(ctx context.Context, courseSlug, email, name string) (*CheckoutInitiateResult, error)
	CreatePaymentSession(ctx context.Context, orderID, paymentMethod string) (paymentURL string, err error)
	HandlePaymentWebhook(ctx context.Context, orderID string) error
}

type checkoutService struct {
	courseRepo     courseRepoForCheckout
	userRepo       userRepoForCheckout
	orderRepo      orderRepoForCheckout
	orderItemRepo  orderItemRepoForCheckout
	paymentRepo    paymentRepoForCheckout
	enrollmentRepo enrollmentRepoForCheckout
}

type courseRepoForCheckout interface {
	GetBySlug(ctx context.Context, slug string) (domain.Course, error)
}

type userRepoForCheckout interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, u domain.User) (domain.User, error)
}

type orderRepoForCheckout interface {
	Create(ctx context.Context, o domain.Order) (domain.Order, error)
	GetByID(ctx context.Context, id string) (domain.Order, error)
	UpdateStatus(ctx context.Context, id, status string) error
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
) CheckoutService {
	return &checkoutService{
		courseRepo:     courseRepo,
		userRepo:       userRepo,
		orderRepo:      orderRepo,
		orderItemRepo:  orderItemRepo,
		paymentRepo:    paymentRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

func (s *checkoutService) Initiate(ctx context.Context, courseSlug, email, name string) (*CheckoutInitiateResult, error) {
	course, err := s.courseRepo.GetBySlug(ctx, courseSlug)
	if err != nil {
		return nil, ErrCourseNotFound
	}
	if course.Slug == nil || *course.Slug == "" {
		return nil, ErrCourseNotFound
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	isNewUser := false
	if err != nil {
		user, err = s.userRepo.Create(ctx, domain.User{
			Email:        email,
			Name:         name,
			PasswordHash: "",
			Role:         domain.UserRoleStudent,
		})
		if err != nil {
			return nil, err
		}
		isNewUser = true
	}

	order, err := s.orderRepo.Create(ctx, domain.Order{
		UserID:          user.ID,
		Status:          domain.OrderStatusPending,
		TotalPriceCents: course.PriceCents,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.orderItemRepo.Create(ctx, domain.OrderItem{
		OrderID:    order.ID,
		CourseID:   course.ID,
		PriceCents: course.PriceCents,
	})
	if err != nil {
		return nil, err
	}

	return &CheckoutInitiateResult{
		OrderID:     order.ID,
		UserID:      user.ID,
		TotalPrice:  order.TotalPriceCents,
		IsNewUser:   isNewUser,
		CourseTitle: course.Title,
		PriceCents:  course.PriceCents,
	}, nil
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
		UserID:        order.UserID,
		OrderID:       &orderID,
		AmountCents:   order.TotalPriceCents,
		Currency:      "IDR",
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

	now := time.Now()
	for _, item := range items {
		_, err = s.enrollmentRepo.GetByUserAndCourse(ctx, order.UserID, item.CourseID)
		if err == nil {
			continue
		}
		_, err = s.enrollmentRepo.Create(ctx, domain.CourseEnrollment{
			UserID:     order.UserID,
			CourseID:   item.CourseID,
			Status:     domain.EnrollmentStatusEnrolled,
			EnrolledAt: now,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
