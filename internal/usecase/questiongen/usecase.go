package questiongen

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Question struct {
	ID            string   `json:"id"`
	Subject       string   `json:"subject"`
	Grade         string   `json:"grade"`
	Topic         string   `json:"topic"`
	Difficulty    string   `json:"difficulty"`
	QuestionText  string   `json:"questionText"`
	Choices       []string `json:"choices,omitempty"`
	CorrectAnswer string   `json:"correctAnswer"`
	Explanation   string   `json:"explanation"`
	SolutionSteps []string `json:"solutionSteps"`
	ConceptTags   []string `json:"conceptTags"`
	EstimatedSec  int      `json:"estimatedSec"`
}

type Subscription struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userId"`
	PlanCode  string     `json:"planCode"`
	Status    string     `json:"status"`
	StartAt   time.Time  `json:"startAt"`
	EndAt     *time.Time `json:"endAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type SubmitAnswerResult struct {
	QuestionID   string `json:"questionId"`
	IsCorrect    bool   `json:"isCorrect"`
	CorrectAnswer string `json:"correctAnswer"`
	Explanation  string `json:"explanation"`
}

type Analysis struct {
	AccuracyPercent float64 `json:"accuracyPercent"`
	TotalAttempts   int     `json:"totalAttempts"`
	CorrectAttempts int     `json:"correctAttempts"`
	AvgTimeMs       int64   `json:"avgTimeMs"`
	WeakTopic       string  `json:"weakTopic"`
	Recommendations []Question `json:"recommendations"`
}

type RankingItem struct {
	UserID      string  `json:"userId"`
	Score       int     `json:"score"`
	AccuracyPct float64 `json:"accuracyPct"`
}

type GenerateQuestionsRequest struct {
	Subject    string `json:"subject"`
	Grade      string `json:"grade"`
	Topic      string `json:"topic"`
	Difficulty string `json:"difficulty"`
	Count      int    `json:"count"`
}

type SubmitAnswerRequest struct {
	QuestionID string `json:"questionId"`
	Answer     string `json:"answer"`
	TimeSpentMs int64 `json:"timeSpentMs"`
}

type ListQuestionsRequest struct {
	Subject    string
	Grade      string
	Topic      string
	Difficulty string
	Limit      int
}

type CreateSubscriptionRequest struct {
	PlanCode string     `json:"planCode"`
	StartAt  *time.Time `json:"startAt,omitempty"`
	EndAt    *time.Time `json:"endAt,omitempty"`
}

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
)

type Repository interface {
	ListQuestions(ctx context.Context, req ListQuestionsRequest) ([]Question, error)
	GetQuestionByID(ctx context.Context, id string) (Question, error)
	RecordSubmission(ctx context.Context, userID string, req SubmitAnswerRequest, isCorrect bool) error
	GetAnalysis(ctx context.Context, userID string, topic, grade string) (Analysis, error)
	GetRanking(ctx context.Context, limit int) ([]RankingItem, error)
	CreateSubscription(ctx context.Context, userID string, req CreateSubscriptionRequest) (Subscription, error)
}

type Usecase struct {
	repo Repository
}

func New(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) GenerateQuestions(ctx context.Context, req GenerateQuestionsRequest) ([]Question, error) {
	req.Subject = strings.TrimSpace(strings.ToLower(req.Subject))
	req.Grade = strings.TrimSpace(strings.ToLower(req.Grade))
	req.Topic = strings.TrimSpace(strings.ToLower(req.Topic))
	req.Difficulty = strings.TrimSpace(strings.ToLower(req.Difficulty))
	if req.Subject == "" || req.Grade == "" || req.Topic == "" || req.Count <= 0 {
		return nil, ErrInvalidInput
	}
	if req.Subject != "math" && req.Subject != "informatics" {
		return nil, ErrInvalidInput
	}
	if req.Grade != "sd" && req.Grade != "smp" && req.Grade != "sma" {
		return nil, ErrInvalidInput
	}
	if req.Difficulty != "" &&
		req.Difficulty != "easy" &&
		req.Difficulty != "medium" &&
		req.Difficulty != "hard" &&
		req.Difficulty != "olympiad" {
		return nil, ErrInvalidInput
	}
	if req.Count > 100 {
		req.Count = 100
	}
	bank, err := u.repo.ListQuestions(ctx, ListQuestionsRequest{
		Subject:    req.Subject,
		Grade:      req.Grade,
		Topic:      req.Topic,
		Difficulty: req.Difficulty,
		// Fetch kandidat lebih banyak untuk variasi, lalu shuffle di memory.
		Limit: req.Count * 4,
	})
	if err != nil {
		return nil, err
	}
	bank = enrichQuestionOutput(bank)
	if len(bank) >= req.Count {
		return shuffleQuestions(bank)[:req.Count], nil
	}
	// Hybrid fallback: generate variasi berbasis template rule ketika bank belum cukup.
	need := req.Count - len(bank)
	variants := generateTemplateVariants(req, need)
	out := append(bank, variants...)
	return shuffleQuestions(out), nil
}

func (u *Usecase) SubmitAnswer(ctx context.Context, userID string, req SubmitAnswerRequest) (SubmitAnswerResult, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(req.QuestionID) == "" {
		return SubmitAnswerResult{}, ErrInvalidInput
	}
	q, err := u.repo.GetQuestionByID(ctx, req.QuestionID)
	if err != nil {
		return SubmitAnswerResult{}, err
	}
	isCorrect := strings.EqualFold(strings.TrimSpace(req.Answer), strings.TrimSpace(q.CorrectAnswer))
	if err := u.repo.RecordSubmission(ctx, userID, req, isCorrect); err != nil {
		return SubmitAnswerResult{}, err
	}
	return SubmitAnswerResult{
		QuestionID:    q.ID,
		IsCorrect:     isCorrect,
		CorrectAnswer: q.CorrectAnswer,
		Explanation:   q.Explanation,
	}, nil
}

func (u *Usecase) Analysis(ctx context.Context, userID, topic, grade string) (Analysis, error) {
	if strings.TrimSpace(userID) == "" {
		return Analysis{}, ErrInvalidInput
	}
	a, err := u.repo.GetAnalysis(ctx, userID, strings.TrimSpace(strings.ToLower(topic)), strings.TrimSpace(strings.ToLower(grade)))
	if err != nil {
		return Analysis{}, err
	}
	targetTopic := topic
	if strings.TrimSpace(targetTopic) == "" {
		targetTopic = a.WeakTopic
	}
	if strings.TrimSpace(targetTopic) != "" {
		reco, err := u.repo.ListQuestions(ctx, ListQuestionsRequest{
			Topic:      strings.TrimSpace(strings.ToLower(targetTopic)),
			Grade:      strings.TrimSpace(strings.ToLower(grade)),
			Difficulty: "medium",
			Limit:      5,
		})
		if err == nil {
			a.Recommendations = enrichQuestionOutput(reco)
		}
	}
	return a, nil
}

func (u *Usecase) Ranking(ctx context.Context, limit int) ([]RankingItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return u.repo.GetRanking(ctx, limit)
}

func (u *Usecase) Questions(ctx context.Context, req ListQuestionsRequest) ([]Question, error) {
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	list, err := u.repo.ListQuestions(ctx, req)
	if err != nil {
		return nil, err
	}
	return enrichQuestionOutput(list), nil
}

func (u *Usecase) Subscribe(ctx context.Context, userID string, req CreateSubscriptionRequest) (Subscription, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(req.PlanCode) == "" {
		return Subscription{}, ErrInvalidInput
	}
	req.PlanCode = strings.ToLower(strings.TrimSpace(req.PlanCode))
	return u.repo.CreateSubscription(ctx, userID, req)
}

func enrichQuestionOutput(in []Question) []Question {
	out := make([]Question, 0, len(in))
	for _, q := range in {
		if len(q.SolutionSteps) == 0 && strings.TrimSpace(q.Explanation) != "" {
			parts := strings.Split(q.Explanation, "\n")
			steps := make([]string, 0, len(parts))
			for _, p := range parts {
				s := strings.TrimSpace(p)
				if s != "" {
					steps = append(steps, s)
				}
			}
			if len(steps) == 0 {
				steps = []string{q.Explanation}
			}
			q.SolutionSteps = steps
		}
		out = append(out, q)
	}
	return out
}

func shuffleQuestions(in []Question) []Question {
	out := make([]Question, len(in))
	copy(out, in)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func generateTemplateVariants(req GenerateQuestionsRequest, n int) []Question {
	if n <= 0 {
		return nil
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	levelSec, trickTag := difficultyScale(req.Difficulty)
	out := make([]Question, 0, n)
	for i := 0; i < n; i++ {
		seed := r.Intn(80) + 20
		stem := templateStem(req.Subject, req.Topic, req.Grade, seed, req.Difficulty, i+1)
		q := Question{
			ID:            "",
			Subject:       req.Subject,
			Grade:         req.Grade,
			Topic:         req.Topic,
			Difficulty:    req.Difficulty,
			QuestionText:  stem,
			Choices:       []string{"A", "B", "C", "D"},
			CorrectAnswer: "B",
			Explanation: strings.Join([]string{
				"Identifikasi informasi penting pada soal cerita.",
				"Modelkan ke bentuk " + req.Topic + " dan susun langkah penyelesaian.",
				"Lakukan perhitungan hati-hati untuk menghindari jebakan umum.",
				"Verifikasi hasil dengan substitusi balik.",
			}, "\n"),
			SolutionSteps: []string{
				"Identifikasi informasi penting pada soal cerita.",
				"Modelkan ke bentuk " + req.Topic + " dan susun langkah penyelesaian.",
				"Lakukan perhitungan hati-hati untuk menghindari jebakan umum.",
				"Verifikasi hasil dengan substitusi balik.",
			},
			ConceptTags:  []string{req.Topic, req.Difficulty, trickTag, "olympiad-style"},
			EstimatedSec: levelSec + r.Intn(120),
		}
		out = append(out, q)
	}
	return out
}

func difficultyScale(d string) (estimatedSec int, trickTag string) {
	switch strings.ToLower(strings.TrimSpace(d)) {
	case "easy":
		return 180, "basic-reasoning"
	case "medium":
		return 300, "multi-step"
	case "hard":
		return 480, "non-obvious"
	case "olympiad":
		return 720, "insight-required"
	default:
		return 300, "multi-step"
	}
}

func templateStem(subject, topic, grade string, seed int, difficulty string, no int) string {
	if strings.EqualFold(strings.TrimSpace(subject), "informatics") {
		return fmt.Sprintf("Pada simulasi olimpiade informatika (%s-%s), terdapat graf dengan %d simpul. Tentukan strategi %s paling efisien untuk kasus ke-%d.", grade, difficulty, seed, topic, no)
	}
	// Default math: wajib narasi cerita.
	return fmt.Sprintf("Di desa Fansedu, panitia olimpiade %s menyiapkan tantangan %s untuk kelas %s. Dengan data awal %d, tentukan hasil akhir yang benar pada variasi ke-%d.", topic, difficulty, strings.ToUpper(grade), seed, no)
}

// ParseClassLevel helper untuk validasi cepat (opsional dipakai caller).
func ParseClassLevel(v string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(v))
}

