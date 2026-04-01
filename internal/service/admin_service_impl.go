package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"
	"github.com/meirusfandi/fansedu-golang-api/internal/ai"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type adminService struct {
	userRepo interface {
		List(ctx context.Context, role string) ([]domain.User, error)
		FindByID(ctx context.Context, id string) (domain.User, error)
		Create(ctx context.Context, u domain.User) (domain.User, error)
		Update(ctx context.Context, u domain.User) error
		Count(ctx context.Context) (int, error)
	}
	tryoutRepo interface {
		Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
		GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
		List(ctx context.Context) ([]domain.TryoutSession, error)
		ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
		Update(ctx context.Context, t domain.TryoutSession) error
		Delete(ctx context.Context, id string) error
	}
	questionRepo interface {
		Create(ctx context.Context, q domain.Question) (domain.Question, error)
		GetByID(ctx context.Context, id string) (domain.Question, error)
		ListByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Question, error)
		Update(ctx context.Context, q domain.Question) error
		Delete(ctx context.Context, id string) error
	}
	courseRepo interface {
		Create(ctx context.Context, c domain.Course) (domain.Course, error)
		GetByID(ctx context.Context, id string) (domain.Course, error)
		List(ctx context.Context) ([]domain.Course, error)
		Update(ctx context.Context, c domain.Course) error
		Count(ctx context.Context) (int, error)
	}
	enrollmentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
		Count(ctx context.Context) (int, error)
		CountEnrolledInMonth(ctx context.Context, year, month int) (int, error)
	}
	courseContentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseContent, error)
		GetByID(ctx context.Context, id string) (domain.CourseContent, error)
		Create(ctx context.Context, c domain.CourseContent) (domain.CourseContent, error)
		Update(ctx context.Context, c domain.CourseContent) error
		Delete(ctx context.Context, id string) error
	}
	paymentRepo interface {
		Create(ctx context.Context, p domain.Payment) (domain.Payment, error)
		List(ctx context.Context, limit int) ([]domain.Payment, error)
		GetByID(ctx context.Context, id string) (domain.Payment, error)
		Update(ctx context.Context, p domain.Payment) error
		CountPaidInMonth(ctx context.Context, year, month int) (int, error)
		TotalAmountPaidInMonth(ctx context.Context, year, month int) (int64, error)
	}
	certificateRepo interface {
		Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error)
	}
	attemptRepo interface {
		ListByUserID(ctx context.Context, userID string) ([]domain.Attempt, error)
		GetByID(ctx context.Context, id string) (domain.Attempt, error)
		ListSubmittedByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Attempt, error)
		ParticipantsCountByTryout(ctx context.Context, tryoutSessionID string) (int, error)
		Update(ctx context.Context, a domain.Attempt) error
	}
	attemptAnswerRepo interface {
		ListByQuestionFromSubmittedAttempts(ctx context.Context, tryoutSessionID, questionID string) ([]domain.AttemptAnswer, error)
		ListByTryoutFromSubmittedAttempts(ctx context.Context, tryoutSessionID string) ([]domain.AttemptAnswer, error)
		ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error)
		GetByAttemptAndQuestion(ctx context.Context, attemptID, questionID string) (domain.AttemptAnswer, error)
		SetAnswerGrading(ctx context.Context, attemptID, questionID string, isCorrect *bool) error
		UpdateAnswerReview(ctx context.Context, attemptID, questionID string, reviewerComment *string, manualScore *float64, reviewedByUserID string) error
		EnsureAnswerRowForReview(ctx context.Context, attemptID, questionID string) error
		ClearManualGradingForAttempt(ctx context.Context, attemptID string, clearReviewMeta bool) error
	}
	feedbackRepo interface {
		GetByAttemptID(ctx context.Context, attemptID string) (domain.AttemptFeedback, error)
		Create(ctx context.Context, f domain.AttemptFeedback) (domain.AttemptFeedback, error)
		Update(ctx context.Context, f domain.AttemptFeedback) error
	}
	feedbackGenerator ai.FeedbackGenerator
	userCount  func(ctx context.Context) (int, error)
	attemptAvg func(ctx context.Context) (float64, error)
	certCount  func(ctx context.Context) (int, error)
}

func NewAdminService(
	userRepo interface {
		List(ctx context.Context, role string) ([]domain.User, error)
		FindByID(ctx context.Context, id string) (domain.User, error)
		Create(ctx context.Context, u domain.User) (domain.User, error)
		Update(ctx context.Context, u domain.User) error
		Count(ctx context.Context) (int, error)
	},
	tryoutRepo interface {
		Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
		GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
		List(ctx context.Context) ([]domain.TryoutSession, error)
		ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
		Update(ctx context.Context, t domain.TryoutSession) error
		Delete(ctx context.Context, id string) error
	},
	questionRepo interface {
		Create(ctx context.Context, q domain.Question) (domain.Question, error)
		GetByID(ctx context.Context, id string) (domain.Question, error)
		ListByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Question, error)
		Update(ctx context.Context, q domain.Question) error
		Delete(ctx context.Context, id string) error
	},
	courseRepo interface {
		Create(ctx context.Context, c domain.Course) (domain.Course, error)
		GetByID(ctx context.Context, id string) (domain.Course, error)
		List(ctx context.Context) ([]domain.Course, error)
		Update(ctx context.Context, c domain.Course) error
		Count(ctx context.Context) (int, error)
	},
	enrollmentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
		Count(ctx context.Context) (int, error)
		CountEnrolledInMonth(ctx context.Context, year, month int) (int, error)
	},
	courseContentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseContent, error)
		GetByID(ctx context.Context, id string) (domain.CourseContent, error)
		Create(ctx context.Context, c domain.CourseContent) (domain.CourseContent, error)
		Update(ctx context.Context, c domain.CourseContent) error
		Delete(ctx context.Context, id string) error
	},
	paymentRepo interface {
		Create(ctx context.Context, p domain.Payment) (domain.Payment, error)
		List(ctx context.Context, limit int) ([]domain.Payment, error)
		GetByID(ctx context.Context, id string) (domain.Payment, error)
		Update(ctx context.Context, p domain.Payment) error
		CountPaidInMonth(ctx context.Context, year, month int) (int, error)
		TotalAmountPaidInMonth(ctx context.Context, year, month int) (int64, error)
	},
	certificateRepo interface {
		Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error)
	},
	attemptRepo interface {
		ListByUserID(ctx context.Context, userID string) ([]domain.Attempt, error)
		GetByID(ctx context.Context, id string) (domain.Attempt, error)
		ListSubmittedByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Attempt, error)
		ParticipantsCountByTryout(ctx context.Context, tryoutSessionID string) (int, error)
		Update(ctx context.Context, a domain.Attempt) error
	},
	attemptAnswerRepo interface {
		ListByQuestionFromSubmittedAttempts(ctx context.Context, tryoutSessionID, questionID string) ([]domain.AttemptAnswer, error)
		ListByTryoutFromSubmittedAttempts(ctx context.Context, tryoutSessionID string) ([]domain.AttemptAnswer, error)
		ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error)
		GetByAttemptAndQuestion(ctx context.Context, attemptID, questionID string) (domain.AttemptAnswer, error)
		SetAnswerGrading(ctx context.Context, attemptID, questionID string, isCorrect *bool) error
		UpdateAnswerReview(ctx context.Context, attemptID, questionID string, reviewerComment *string, manualScore *float64, reviewedByUserID string) error
		EnsureAnswerRowForReview(ctx context.Context, attemptID, questionID string) error
		ClearManualGradingForAttempt(ctx context.Context, attemptID string, clearReviewMeta bool) error
	},
	feedbackRepo interface {
		GetByAttemptID(ctx context.Context, attemptID string) (domain.AttemptFeedback, error)
		Create(ctx context.Context, f domain.AttemptFeedback) (domain.AttemptFeedback, error)
		Update(ctx context.Context, f domain.AttemptFeedback) error
	},
	feedbackGenerator ai.FeedbackGenerator,
	userCount func(ctx context.Context) (int, error),
	attemptAvg func(ctx context.Context) (float64, error),
	certCount func(ctx context.Context) (int, error),
) AdminService {
	return &adminService{
		userRepo:         userRepo,
		tryoutRepo:       tryoutRepo,
		questionRepo:     questionRepo,
		courseRepo:       courseRepo,
		enrollmentRepo:   enrollmentRepo,
		courseContentRepo: courseContentRepo,
		paymentRepo:      paymentRepo,
		certificateRepo:  certificateRepo,
		attemptRepo:      attemptRepo,
		attemptAnswerRepo: attemptAnswerRepo,
		feedbackRepo:     feedbackRepo,
		feedbackGenerator: feedbackGenerator,
		userCount:        userCount,
		attemptAvg:       attemptAvg,
		certCount:        certCount,
	}
}

func (s *adminService) Overview(ctx context.Context) (*AdminOverview, error) {
	students, err := s.userCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("count students: %w", err)
	}
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	open, err := s.tryoutRepo.ListOpen(ctx, time.Now())
	if err != nil {
		return nil, fmt.Errorf("list open tryouts: %w", err)
	}
	totalCourses, err := s.courseRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count courses: %w", err)
	}
	totalEnrollments, err := s.enrollmentRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count enrollments: %w", err)
	}
	avg, err := s.attemptAvg(ctx)
	if err != nil {
		return nil, fmt.Errorf("avg attempt score: %w", err)
	}
	certs, err := s.certCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("count certificates: %w", err)
	}
	return &AdminOverview{
		TotalStudents:     students,
		TotalUsers:        totalUsers,
		ActiveTryouts:     len(open),
		TotalCourses:      totalCourses,
		TotalEnrollments:  totalEnrollments,
		AvgScore:          avg,
		TotalCertificates: certs,
	}, nil
}

func (s *adminService) ListUsers(ctx context.Context, role string) ([]domain.User, error) {
	return s.userRepo.List(ctx, role)
}

func (s *adminService) GetUser(ctx context.Context, id string) (domain.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *adminService) CreateUser(ctx context.Context, u domain.User, password string) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	u.PasswordHash = string(hash)
	if u.Role == "" {
		u.Role = domain.UserRoleStudent
	}
	return s.userRepo.Create(ctx, u)
}

func (s *adminService) UpdateUser(ctx context.Context, u domain.User, newPassword *string) error {
	if newPassword != nil && *newPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hash)
	}
	return s.userRepo.Update(ctx, u)
}

func (s *adminService) ListCourses(ctx context.Context) ([]domain.Course, error) {
	return s.courseRepo.List(ctx)
}

func (s *adminService) GetCourseByID(ctx context.Context, id string) (domain.Course, error) {
	return s.courseRepo.GetByID(ctx, id)
}

func (s *adminService) UpdateCourse(ctx context.Context, c domain.Course) error {
	return s.courseRepo.Update(ctx, c)
}

func (s *adminService) ListTryouts(ctx context.Context) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.List(ctx)
}

func (s *adminService) CreateTryout(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error) {
	return s.tryoutRepo.Create(ctx, t)
}

func (s *adminService) UpdateTryout(ctx context.Context, t domain.TryoutSession) error {
	return s.tryoutRepo.Update(ctx, t)
}

func (s *adminService) DeleteTryout(ctx context.Context, id string) error {
	return s.tryoutRepo.Delete(ctx, id)
}

func (s *adminService) ListQuestions(ctx context.Context, tryoutID string) ([]domain.Question, error) {
	return s.questionRepo.ListByTryoutSessionID(ctx, tryoutID)
}

func (s *adminService) GetQuestion(ctx context.Context, questionID string) (domain.Question, error) {
	return s.questionRepo.GetByID(ctx, questionID)
}

func (s *adminService) CreateQuestion(ctx context.Context, q domain.Question) (domain.Question, error) {
	return s.questionRepo.Create(ctx, q)
}

func (s *adminService) UpdateQuestion(ctx context.Context, q domain.Question) error {
	return s.questionRepo.Update(ctx, q)
}

func (s *adminService) DeleteQuestion(ctx context.Context, id string) error {
	return s.questionRepo.Delete(ctx, id)
}

func (s *adminService) CreateCourse(ctx context.Context, c domain.Course) (domain.Course, error) {
	return s.courseRepo.Create(ctx, c)
}

func (s *adminService) ListEnrollmentsByCourse(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error) {
	return s.enrollmentRepo.ListByCourseID(ctx, courseID)
}

func (s *adminService) IssueCertificate(ctx context.Context, c domain.Certificate) (domain.Certificate, error) {
	return s.certificateRepo.Create(ctx, c)
}

func (s *adminService) ListCourseContents(ctx context.Context, courseID string) ([]domain.CourseContent, error) {
	return s.courseContentRepo.ListByCourseID(ctx, courseID)
}

func (s *adminService) GetCourseContent(ctx context.Context, id string) (domain.CourseContent, error) {
	return s.courseContentRepo.GetByID(ctx, id)
}

func (s *adminService) CreateCourseContent(ctx context.Context, c domain.CourseContent) (domain.CourseContent, error) {
	return s.courseContentRepo.Create(ctx, c)
}

func (s *adminService) UpdateCourseContent(ctx context.Context, c domain.CourseContent) error {
	return s.courseContentRepo.Update(ctx, c)
}

func (s *adminService) DeleteCourseContent(ctx context.Context, id string) error {
	return s.courseContentRepo.Delete(ctx, id)
}

func (s *adminService) ListPayments(ctx context.Context, limit int) ([]domain.Payment, error) {
	return s.paymentRepo.List(ctx, limit)
}

func (s *adminService) CreatePayment(ctx context.Context, p domain.Payment) (domain.Payment, error) {
	return s.paymentRepo.Create(ctx, p)
}

func (s *adminService) ConfirmPayment(ctx context.Context, paymentID string, confirmed bool, adminID string, rejectionNote *string) error {
	p, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}
	now := time.Now()
	if confirmed {
		p.Status = domain.PaymentStatusPaid
		p.PaidAt = &now
		p.ConfirmedBy = &adminID
		p.ConfirmedAt = &now
		p.RejectionNote = nil
	} else {
		p.Status = domain.PaymentStatusFailed
		p.ConfirmedBy = &adminID
		p.ConfirmedAt = &now
		p.RejectionNote = rejectionNote
	}
	return s.paymentRepo.Update(ctx, p)
}

func (s *adminService) ReportMonthly(ctx context.Context, year, month int) (*MonthlyReport, error) {
	enrollments, err := s.enrollmentRepo.CountEnrolledInMonth(ctx, year, month)
	if err != nil {
		return nil, fmt.Errorf("count monthly enrollments: %w", err)
	}
	paymentsCount, err := s.paymentRepo.CountPaidInMonth(ctx, year, month)
	if err != nil {
		return nil, fmt.Errorf("count monthly paid payments: %w", err)
	}
	revenue, err := s.paymentRepo.TotalAmountPaidInMonth(ctx, year, month)
	if err != nil {
		return nil, fmt.Errorf("sum monthly revenue: %w", err)
	}
	return &MonthlyReport{
		Year:           year,
		Month:          month,
		NewEnrollments: enrollments,
		PaymentsCount:  paymentsCount,
		TotalRevenue:   revenue,
	}, nil
}

func (s *adminService) GetQuestionStats(ctx context.Context, tryoutID, questionID string) (*QuestionStats, error) {
	q, err := s.questionRepo.GetByID(ctx, questionID)
	if err != nil {
		return nil, ErrNotFound
	}
	if q.TryoutSessionID != tryoutID {
		return nil, ErrNotFound
	}
	participantsCount, _ := s.attemptRepo.ParticipantsCountByTryout(ctx, tryoutID)
	answers, err := s.attemptAnswerRepo.ListByQuestionFromSubmittedAttempts(ctx, tryoutID, questionID)
	if err != nil {
		return nil, err
	}
	answeredCount := len(answers)
	var correctCount, wrongCount int
	for i := range answers {
		score := ComputeQuestionScore(q, &answers[i])
		if score > 0 {
			correctCount++
		} else {
			wrongCount++
		}
	}
	total := correctCount + wrongCount
	var correctPercent, wrongPercent float64
	if total > 0 {
		correctPercent = float64(correctCount) / float64(total) * 100
		wrongPercent = float64(wrongCount) / float64(total) * 100
	}
	return &QuestionStats{
		ParticipantsCount: participantsCount,
		AnsweredCount:     answeredCount,
		CorrectCount:      correctCount,
		WrongCount:        wrongCount,
		CorrectPercent:    roundTwo(correctPercent),
		WrongPercent:      roundTwo(wrongPercent),
	}, nil
}

func (s *adminService) GetTryoutQuestionStatsBulk(ctx context.Context, tryoutID string) (*QuestionStatsBulk, error) {
	if _, err := s.tryoutRepo.GetByID(ctx, tryoutID); err != nil {
		return nil, ErrNotFound
	}
	questions, err := s.questionRepo.ListByTryoutSessionID(ctx, tryoutID)
	if err != nil {
		return nil, err
	}
	participantsCount, _ := s.attemptRepo.ParticipantsCountByTryout(ctx, tryoutID)
	allAnswers, err := s.attemptAnswerRepo.ListByTryoutFromSubmittedAttempts(ctx, tryoutID)
	if err != nil {
		return nil, err
	}
	byQuestion := make(map[string][]domain.AttemptAnswer)
	for _, a := range allAnswers {
		byQuestion[a.QuestionID] = append(byQuestion[a.QuestionID], a)
	}
	out := make([]QuestionStatsItem, 0, len(questions))
	for _, q := range questions {
		answers := byQuestion[q.ID]
		answeredCount := len(answers)
		var correctCount, wrongCount int
		for i := range answers {
			score := ComputeQuestionScore(q, &answers[i])
			if score > 0 {
				correctCount++
			} else {
				wrongCount++
			}
		}
		total := correctCount + wrongCount
		var correctPercent, wrongPercent float64
		if total > 0 {
			correctPercent = float64(correctCount) / float64(total) * 100
			wrongPercent = float64(wrongCount) / float64(total) * 100
		}
		out = append(out, QuestionStatsItem{
			QuestionID:     q.ID,
			AnsweredCount:  answeredCount,
			CorrectCount:   correctCount,
			WrongCount:     wrongCount,
			CorrectPercent: roundTwo(correctPercent),
			WrongPercent:   roundTwo(wrongPercent),
		})
	}
	return &QuestionStatsBulk{
		ParticipantsCount: participantsCount,
		Questions:         out,
	}, nil
}

func (s *adminService) GetTryoutAnalysis(ctx context.Context, tryoutID string) (*TryoutAnalysis, error) {
	t, err := s.tryoutRepo.GetByID(ctx, tryoutID)
	if err != nil {
		return nil, ErrNotFound
	}
	questions, err := s.questionRepo.ListByTryoutSessionID(ctx, tryoutID)
	if err != nil {
		return nil, err
	}
	participantsCount, _ := s.attemptRepo.ParticipantsCountByTryout(ctx, tryoutID)
	allAnswers, err := s.attemptAnswerRepo.ListByTryoutFromSubmittedAttempts(ctx, tryoutID)
	if err != nil {
		return nil, err
	}
	byQuestion := make(map[string][]domain.AttemptAnswer)
	for _, a := range allAnswers {
		byQuestion[a.QuestionID] = append(byQuestion[a.QuestionID], a)
	}
	out := make([]TryoutAnalysisQuestion, 0, len(questions))
	for _, q := range questions {
		answers := byQuestion[q.ID]
		answeredCount := len(answers)
		unansweredCount := participantsCount - answeredCount
		if unansweredCount < 0 {
			unansweredCount = 0
		}
		var correctCount, wrongCount int
		optionDist := make(map[string]int)
		for i := range answers {
			score := ComputeQuestionScore(q, &answers[i])
			if score > 0 {
				correctCount++
			} else {
				wrongCount++
			}
			if answers[i].SelectedOption != nil && *answers[i].SelectedOption != "" {
				optionDist[*answers[i].SelectedOption]++
			}
		}
		total := correctCount + wrongCount
		var correctPercent, wrongPercent float64
		if total > 0 {
			correctPercent = float64(correctCount) / float64(total) * 100
			wrongPercent = float64(wrongCount) / float64(total) * 100
		}
		out = append(out, TryoutAnalysisQuestion{
			QuestionNumber:      q.SortOrder,
			QuestionID:          q.ID,
			QuestionType:        q.Type,
			AnsweredCount:      answeredCount,
			UnansweredCount:    unansweredCount,
			CorrectCount:       correctCount,
			WrongCount:         wrongCount,
			CorrectPercent:     roundTwo(correctPercent),
			WrongPercent:       roundTwo(wrongPercent),
			OptionDistribution: optionDist,
		})
	}
	return &TryoutAnalysis{
		TryoutID:          tryoutID,
		TryoutTitle:       t.Title,
		ParticipantsCount: participantsCount,
		Questions:         out,
	}, nil
}

func (s *adminService) ListTryoutStudents(ctx context.Context, tryoutID string) ([]TryoutStudentItem, error) {
	if _, err := s.tryoutRepo.GetByID(ctx, tryoutID); err != nil {
		return nil, ErrNotFound
	}
	attempts, err := s.attemptRepo.ListSubmittedByTryoutSessionID(ctx, tryoutID)
	if err != nil {
		return nil, err
	}
	out := make([]TryoutStudentItem, 0, len(attempts))
	for _, a := range attempts {
		u, err := s.userRepo.FindByID(ctx, a.UserID)
		if err != nil {
			continue
		}
		var submittedAt *string
		if a.SubmittedAt != nil {
			s := a.SubmittedAt.Format(time.RFC3339)
			submittedAt = &s
		}
		out = append(out, TryoutStudentItem{
			UserID:      u.ID,
			UserName:    u.Name,
			UserEmail:   u.Email,
			AttemptID:   a.ID,
			Score:       a.Score,
			MaxScore:    a.MaxScore,
			Percentile:  a.Percentile,
			SubmittedAt: submittedAt,
		})
	}
	return out, nil
}

func (s *adminService) GetAttemptAIAnalysis(ctx context.Context, tryoutID, attemptID string) (*AttemptAIAnalysisResponse, error) {
	attempt, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return nil, ErrNotFound
	}
	if attempt.TryoutSessionID != tryoutID || attempt.Status != domain.AttemptStatusSubmitted {
		return nil, ErrNotFound
	}
	questions, err := s.questionRepo.ListByTryoutSessionID(ctx, tryoutID)
	if err != nil {
		return nil, err
	}
	answers, err := s.attemptAnswerRepo.ListByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	score, maxScore := 0.0, 0.0
	if attempt.Score != nil {
		score = *attempt.Score
	}
	if attempt.MaxScore != nil {
		maxScore = *attempt.MaxScore
	}

	fb, err := s.feedbackRepo.GetByAttemptID(ctx, attemptID)
	if err == nil && (fb.Summary != nil && *fb.Summary != "" || fb.Recap != nil && *fb.Recap != "") {
		return feedbackToAIAnalysisResponse(attemptID, fb)
	}
	_, _, _, _, overall := GradeTryoutAttempt(questions, answers)
	gen, err := s.feedbackGenerator.Generate(ctx, ai.FeedbackRequest{
		Questions:        questions,
		Answers:          answers,
		Score:            score,
		MaxScore:         maxScore,
		OverallNarrative: overall.Summary,
	})
	if err != nil {
		return &AttemptAIAnalysisResponse{
			AttemptID:        attemptID,
			Summary:          "Analisis AI tidak tersedia.",
			Recap:            "",
			StrengthAreas:    nil,
			ImprovementAreas: nil,
			Recommendation:   "Coba lagi nanti atau periksa konfigurasi AI.",
		}, nil
	}
	strengthJSON, _ := json.Marshal(gen.StrengthAreas)
	improvementJSON, _ := json.Marshal(gen.ImprovementAreas)
	_, _ = s.feedbackRepo.Create(ctx, domain.AttemptFeedback{
		AttemptID:         attemptID,
		Summary:           &gen.Summary,
		Recap:             &gen.Recap,
		StrengthAreas:     strengthJSON,
		ImprovementAreas:  improvementJSON,
		RecommendationText: &gen.Recommendation,
	})
	return &AttemptAIAnalysisResponse{
		AttemptID:        attemptID,
		Summary:          gen.Summary,
		Recap:            gen.Recap,
		StrengthAreas:    gen.StrengthAreas,
		ImprovementAreas: gen.ImprovementAreas,
		Recommendation:   gen.Recommendation,
	}, nil
}

func feedbackToAIAnalysisResponse(attemptID string, fb domain.AttemptFeedback) (*AttemptAIAnalysisResponse, error) {
	resp := AttemptAIAnalysisResponse{AttemptID: attemptID}
	if fb.Summary != nil {
		resp.Summary = *fb.Summary
	}
	if fb.Recap != nil {
		resp.Recap = *fb.Recap
	}
	if fb.RecommendationText != nil {
		resp.Recommendation = *fb.RecommendationText
	}
	_ = json.Unmarshal(fb.StrengthAreas, &resp.StrengthAreas)
	_ = json.Unmarshal(fb.ImprovementAreas, &resp.ImprovementAreas)
	return &resp, nil
}

func (s *adminService) GetAttemptReview(ctx context.Context, tryoutID, attemptID string) (*AttemptReviewResponse, error) {
	attempt, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return nil, ErrNotFound
	}
	if attempt.TryoutSessionID != tryoutID || attempt.Status != domain.AttemptStatusSubmitted {
		return nil, ErrNotFound
	}
	student, err := s.userRepo.FindByID(ctx, attempt.UserID)
	if err != nil {
		return nil, ErrNotFound
	}
	questions, err := s.questionRepo.ListByTryoutSessionID(ctx, tryoutID)
	if err != nil {
		return nil, err
	}
	answers, err := s.attemptAnswerRepo.ListByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	ansByQ := make(map[string]domain.AttemptAnswer, len(answers))
	for _, a := range answers {
		ansByQ[a.QuestionID] = a
	}
	reviewerNames := make(map[string]string)
	var submittedAt *string
	if attempt.SubmittedAt != nil {
		sa := attempt.SubmittedAt.Format(time.RFC3339)
		submittedAt = &sa
	}
	out := &AttemptReviewResponse{
		AttemptID:   attemptID,
		TryoutID:    tryoutID,
		Status:      attempt.Status,
		SubmittedAt: submittedAt,
		Score:       attempt.Score,
		MaxScore:    attempt.MaxScore,
		Percentile:  attempt.Percentile,
		Student: AttemptReviewStudent{
			UserID: student.ID,
			Name:   student.Name,
			Email:  student.Email,
		},
		Items: make([]AttemptReviewItem, 0, len(questions)),
	}
	for _, q := range questions {
		ans, has := ansByQ[q.ID]
		var ansPtr *domain.AttemptAnswer
		if has {
			a := ans
			ansPtr = &a
		}
		autoScore, autoIC := autoGradeQuestion(q, ansPtr)
		got, ic := gradeQuestion(q, ansPtr)
		var opts any
		if len(q.Options) > 0 {
			_ = json.Unmarshal(q.Options, &opts)
		}
		item := AttemptReviewItem{
			QuestionID:     q.ID,
			SortOrder:      q.SortOrder,
			Type:           q.Type,
			Body:           q.Body,
			MaxScore:       q.MaxScore,
			Options:        opts,
			CorrectOption:  q.CorrectOption,
			CorrectText:    q.CorrectText,
			AutoScore:      autoScore,
			AutoIsCorrect:  autoIC,
			ScoreGot:       got,
			IsCorrect:      ic,
			ManualScore:    nil,
			ReviewerComment: nil,
		}
		if has {
			item.IsMarked = ans.IsMarked
			if ans.AnswerText != nil {
				v := *ans.AnswerText
				item.AnswerText = &v
			}
			if ans.SelectedOption != nil {
				v := *ans.SelectedOption
				item.SelectedOption = &v
			}
			item.ManualScore = ans.ManualScore
			item.ReviewerComment = ans.ReviewerComment
			if ans.ReviewedByUserID != nil && *ans.ReviewedByUserID != "" {
				uid := *ans.ReviewedByUserID
				item.ReviewedByUserID = &uid
				name, ok := reviewerNames[uid]
				if !ok {
					if ru, rerr := s.userRepo.FindByID(ctx, uid); rerr == nil {
						name = ru.Name
						reviewerNames[uid] = name
					}
				}
				if name != "" {
					n := name
					item.ReviewedByName = &n
				}
			}
			if ans.ReviewedAt != nil {
				rs := ans.ReviewedAt.UTC().Format(time.RFC3339)
				item.ReviewedAt = &rs
			}
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}

func (s *adminService) PutAttemptAnswerReview(ctx context.Context, tryoutID, attemptID, questionID, reviewerUserID string, patch AttemptAnswerReviewPatch) (AttemptAnswerReviewResult, error) {
	if !patch.HasReviewerComment && !patch.HasManualScore {
		return AttemptAnswerReviewResult{}, ErrAttemptReviewNoFields
	}
	attempt, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return AttemptAnswerReviewResult{}, ErrNotFound
	}
	if attempt.TryoutSessionID != tryoutID || attempt.Status != domain.AttemptStatusSubmitted {
		return AttemptAnswerReviewResult{}, ErrNotFound
	}
	q, err := s.questionRepo.GetByID(ctx, questionID)
	if err != nil || q.TryoutSessionID != tryoutID {
		return AttemptAnswerReviewResult{}, ErrNotFound
	}
	if err := s.attemptAnswerRepo.EnsureAnswerRowForReview(ctx, attemptID, questionID); err != nil {
		return AttemptAnswerReviewResult{}, err
	}
	cur, err := s.attemptAnswerRepo.GetByAttemptAndQuestion(ctx, attemptID, questionID)
	if err != nil {
		return AttemptAnswerReviewResult{}, err
	}
	newComm := cur.ReviewerComment
	if patch.HasReviewerComment {
		newComm = patch.ReviewerComment
	}
	newManual := cur.ManualScore
	manualClamped := false
	if patch.HasManualScore {
		if patch.ManualScore == nil {
			newManual = nil
		} else {
			raw := *patch.ManualScore
			v := ClampManualScoreToQuestionMax(raw, q.MaxScore)
			manualClamped = math.Abs(raw-v) > 1e-6
			newManual = &v
		}
	}
	if err := s.attemptAnswerRepo.UpdateAnswerReview(ctx, attemptID, questionID, newComm, newManual, reviewerUserID); err != nil {
		return AttemptAnswerReviewResult{}, err
	}
	if err := s.recalculateAttemptTotals(ctx, tryoutID, attemptID); err != nil {
		return AttemptAnswerReviewResult{}, err
	}
	a2, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return AttemptAnswerReviewResult{}, err
	}
	score := 0.0
	if a2.Score != nil {
		score = *a2.Score
	}
	ansAfter, err := s.attemptAnswerRepo.GetByAttemptAndQuestion(ctx, attemptID, questionID)
	if err != nil {
		return AttemptAnswerReviewResult{}, err
	}
	return AttemptAnswerReviewResult{
		StudentUserID:       attempt.UserID,
		AttemptScore:        score,
		QuestionManualScore: ansAfter.ManualScore,
		QuestionMaxScore:    q.MaxScore,
		ManualScoreClamped:  manualClamped,
	}, nil
}

func (s *adminService) AutoGradeAttempt(ctx context.Context, tryoutID, attemptID string, opts AutoGradeAttemptOpts) (string, float64, error) {
	attempt, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return "", 0, ErrNotFound
	}
	if attempt.TryoutSessionID != tryoutID || attempt.Status != domain.AttemptStatusSubmitted {
		return "", 0, ErrNotFound
	}
	if err := s.attemptAnswerRepo.ClearManualGradingForAttempt(ctx, attemptID, opts.ClearReviewerComments); err != nil {
		return "", 0, err
	}
	if err := s.recalculateAttemptTotals(ctx, tryoutID, attemptID); err != nil {
		return "", 0, err
	}
	a2, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return "", 0, err
	}
	score := 0.0
	if a2.Score != nil {
		score = *a2.Score
	}
	return attempt.UserID, score, nil
}

func (s *adminService) AutoGradeAllSubmittedAttempts(ctx context.Context, tryoutID string, opts AutoGradeAttemptOpts) (AutoGradeAllSubmittedResult, error) {
	if _, err := s.tryoutRepo.GetByID(ctx, tryoutID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AutoGradeAllSubmittedResult{}, ErrNotFound
		}
		return AutoGradeAllSubmittedResult{}, err
	}
	list, err := s.attemptRepo.ListSubmittedByTryoutSessionID(ctx, tryoutID)
	if err != nil {
		return AutoGradeAllSubmittedResult{}, err
	}
	out := AutoGradeAllSubmittedResult{
		TryoutID: tryoutID,
		Total:    len(list),
		Results:  make([]AutoGradeSubmittedItem, 0, len(list)),
	}
	for _, a := range list {
		uid, score, err := s.AutoGradeAttempt(ctx, tryoutID, a.ID, opts)
		if err != nil {
			out.Failed++
			out.Results = append(out.Results, AutoGradeSubmittedItem{
				AttemptID: a.ID,
				UserID:    a.UserID,
				Ok:        false,
				Error:     err.Error(),
			})
			continue
		}
		out.Succeeded++
		out.Results = append(out.Results, AutoGradeSubmittedItem{
			AttemptID: a.ID,
			UserID:    uid,
			Ok:        true,
			Score:     score,
		})
	}
	return out, nil
}

func (s *adminService) recalculateAttemptTotals(ctx context.Context, tryoutID, attemptID string) error {
	a, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return err
	}
	questions, err := s.questionRepo.ListByTryoutSessionID(ctx, tryoutID)
	if err != nil {
		return err
	}
	answers, err := s.attemptAnswerRepo.ListByAttemptID(ctx, attemptID)
	if err != nil {
		return err
	}
	score, maxScore, outcomes, _, _ := GradeTryoutAttempt(questions, answers)
	for _, o := range outcomes {
		_ = s.attemptAnswerRepo.SetAnswerGrading(ctx, attemptID, o.QuestionID, o.IsCorrect)
	}
	var percentile *float64
	if others, perr := s.attemptRepo.ListSubmittedByTryoutSessionID(ctx, tryoutID); perr == nil {
		scores := make([]float64, 0, len(others))
		for _, o := range others {
			if o.ID == attemptID {
				continue
			}
			if o.Score != nil {
				scores = append(scores, *o.Score)
			}
		}
		scores = append(scores, score)
		percentile = percentileRankPercent(scores, score)
	}
	a.Score = &score
	a.MaxScore = &maxScore
	a.Percentile = percentile
	return s.attemptRepo.Update(ctx, a)
}

func (s *adminService) GetCourseReport(ctx context.Context, courseID string) (*CourseReport, error) {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, ErrNotFound
	}
	enrollments, err := s.enrollmentRepo.ListByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	report := &CourseReport{
		Course: CourseReportCourse{
			ID:          course.ID,
			Title:       course.Title,
			Description: course.Description,
		},
		GeneratedAt: now,
		Students:    make([]CourseReportStudent, 0, len(enrollments)),
	}
	for _, e := range enrollments {
		u, err := s.userRepo.FindByID(ctx, e.UserID)
		if err != nil {
			continue
		}
		attempts, _ := s.attemptRepo.ListByUserID(ctx, e.UserID)
		tryoutScores := make([]CourseReportTryoutScore, 0)
		var lastActivity *time.Time
		for _, a := range attempts {
			if a.Status != domain.AttemptStatusSubmitted {
				continue
			}
			if a.SubmittedAt != nil && (lastActivity == nil || lastActivity.Before(*a.SubmittedAt)) {
				t := *a.SubmittedAt
				lastActivity = &t
			}
			tryoutTitle := ""
			if t, err := s.tryoutRepo.GetByID(ctx, a.TryoutSessionID); err == nil {
				tryoutTitle = t.Title
			}
			tryoutScores = append(tryoutScores, CourseReportTryoutScore{
				TryoutID:    a.TryoutSessionID,
				TryoutTitle: tryoutTitle,
				AttemptID:   a.ID,
				Score:       a.Score,
				MaxScore:    a.MaxScore,
				Percentile:  a.Percentile,
				SubmittedAt: a.SubmittedAt,
			})
		}
		report.Students = append(report.Students, CourseReportStudent{
			StudentID:        u.ID,
			StudentName:      u.Name,
			StudentEmail:     u.Email,
			EnrolledAt:       e.EnrolledAt,
			EnrollmentStatus: e.Status,
			Progress: CourseReportStudentProgress{
				Status:      e.Status,
				CompletedAt: e.CompletedAt,
			},
			TryoutScores: tryoutScores,
			Attendance: CourseReportStudentAttendance{
				TryoutsParticipated: len(tryoutScores),
				LastActivityAt:      lastActivity,
			},
		})
	}
	return report, nil
}

// ErrNotFound is returned when a tryout or question is not found (for 404 responses).
var ErrNotFound = errors.New("not found")

func roundTwo(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
