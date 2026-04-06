package repo

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/usecase/questiongen"
)

type QuestionGenRepo interface {
	questiongen.Repository
}

type questionGenRepo struct {
	pool *pgxpool.Pool
}

func NewQuestionGenRepo(pool *pgxpool.Pool) QuestionGenRepo {
	return &questionGenRepo{pool: pool}
}

func (r *questionGenRepo) ListQuestions(ctx context.Context, req questiongen.ListQuestionsRequest) ([]questiongen.Question, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 500 {
		limit = 500
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, subject, grade, topic, difficulty, question_text, choices_json, correct_answer,
		       explanation, concept_tags, estimated_time_sec
		FROM ai_questions
		WHERE is_active = TRUE
		  AND ($1 = '' OR subject = $1)
		  AND ($2 = '' OR grade = $2)
		  AND ($3 = '' OR topic = $3)
		  AND ($4 = '' OR difficulty = $4)
		ORDER BY updated_at DESC, id DESC
		LIMIT $5
	`, strings.TrimSpace(req.Subject), strings.TrimSpace(req.Grade), strings.TrimSpace(req.Topic), strings.TrimSpace(req.Difficulty), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]questiongen.Question, 0, limit)
	for rows.Next() {
		var q questiongen.Question
		var choices []byte
		var tags []byte
		if err := rows.Scan(&q.ID, &q.Subject, &q.Grade, &q.Topic, &q.Difficulty, &q.QuestionText, &choices, &q.CorrectAnswer, &q.Explanation, &tags, &q.EstimatedSec); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(choices, &q.Choices)
		_ = json.Unmarshal(tags, &q.ConceptTags)
		out = append(out, q)
	}
	return out, rows.Err()
}

func (r *questionGenRepo) GetQuestionByID(ctx context.Context, id string) (questiongen.Question, error) {
	var q questiongen.Question
	var choices []byte
	var tags []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, subject, grade, topic, difficulty, question_text, choices_json, correct_answer,
		       explanation, concept_tags, estimated_time_sec
		FROM ai_questions
		WHERE id = $1::uuid AND is_active = TRUE
	`, id).Scan(&q.ID, &q.Subject, &q.Grade, &q.Topic, &q.Difficulty, &q.QuestionText, &choices, &q.CorrectAnswer, &q.Explanation, &tags, &q.EstimatedSec)
	if err != nil {
		return questiongen.Question{}, err
	}
	_ = json.Unmarshal(choices, &q.Choices)
	_ = json.Unmarshal(tags, &q.ConceptTags)
	return q, nil
}

func (r *questionGenRepo) RecordSubmission(ctx context.Context, userID string, req questiongen.SubmitAnswerRequest, isCorrect bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ai_submissions (id, user_id, question_id, answer, is_correct, time_spent_ms)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6)
	`, uuid.NewString(), userID, req.QuestionID, req.Answer, isCorrect, req.TimeSpentMs)
	return err
}

func (r *questionGenRepo) GetAnalysis(ctx context.Context, userID string, topic, grade string) (questiongen.Analysis, error) {
	var a questiongen.Analysis
	var weakTopic string
	err := r.pool.QueryRow(ctx, `
		WITH base AS (
			SELECT q.topic, s.is_correct, s.time_spent_ms
			FROM ai_submissions s
			INNER JOIN ai_questions q ON q.id = s.question_id
			WHERE s.user_id = $1::uuid
			  AND ($2 = '' OR q.topic = $2)
			  AND ($3 = '' OR q.grade = $3)
		),
		agg AS (
			SELECT
				COUNT(*)::int AS total_attempts,
				COUNT(*) FILTER (WHERE is_correct)::int AS correct_attempts,
				COALESCE(AVG(time_spent_ms), 0)::bigint AS avg_time_ms
			FROM base
		),
		weak AS (
			SELECT topic
			FROM base
			GROUP BY topic
			ORDER BY (COUNT(*) FILTER (WHERE is_correct)::float / NULLIF(COUNT(*),0)) ASC, COUNT(*) DESC
			LIMIT 1
		)
		SELECT
			COALESCE((correct_attempts::float * 100.0) / NULLIF(total_attempts, 0), 0),
			total_attempts,
			correct_attempts,
			avg_time_ms,
			COALESCE((SELECT topic FROM weak), '')
		FROM agg
	`, userID, topic, grade).Scan(&a.AccuracyPercent, &a.TotalAttempts, &a.CorrectAttempts, &a.AvgTimeMs, &weakTopic)
	if err != nil {
		if err == pgx.ErrNoRows {
			return questiongen.Analysis{}, nil
		}
		return questiongen.Analysis{}, err
	}
	a.WeakTopic = weakTopic
	return a, nil
}

func (r *questionGenRepo) GetRanking(ctx context.Context, limit int) ([]questiongen.RankingItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			s.user_id::text,
			COUNT(*) FILTER (WHERE s.is_correct)::int AS score,
			COALESCE((COUNT(*) FILTER (WHERE s.is_correct)::float * 100.0) / NULLIF(COUNT(*), 0), 0) AS accuracy_pct
		FROM ai_submissions s
		GROUP BY s.user_id
		ORDER BY score DESC, accuracy_pct DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []questiongen.RankingItem
	for rows.Next() {
		var it questiongen.RankingItem
		if err := rows.Scan(&it.UserID, &it.Score, &it.AccuracyPct); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *questionGenRepo) CreateSubscription(ctx context.Context, userID string, req questiongen.CreateSubscriptionRequest) (questiongen.Subscription, error) {
	var startAt time.Time
	if req.StartAt != nil {
		startAt = *req.StartAt
	} else {
		startAt = time.Now()
	}
	var s questiongen.Subscription
	s.ID = uuid.NewString()
	s.UserID = userID
	s.PlanCode = req.PlanCode
	s.Status = "active"
	s.StartAt = startAt
	s.EndAt = req.EndAt
	err := r.pool.QueryRow(ctx, `
		INSERT INTO ai_subscriptions (id, user_id, plan_code, status, start_at, end_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)
		RETURNING created_at
	`, s.ID, s.UserID, s.PlanCode, s.Status, s.StartAt, s.EndAt).Scan(&s.CreatedAt)
	if err != nil {
		return questiongen.Subscription{}, err
	}
	return s, nil
}

