package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// GeneratedFeedback hasil generate AI untuk attempt feedback
type GeneratedFeedback struct {
	Summary          string   `json:"summary"`
	Recap            string   `json:"recap"`
	StrengthAreas    []string `json:"strength_areas"`
	ImprovementAreas []string `json:"improvement_areas"`
	Recommendation   string   `json:"recommendation"`
}

// FeedbackGenerator interface untuk generate feedback (AI atau fallback)
type FeedbackGenerator interface {
	Generate(ctx context.Context, req FeedbackRequest) (*GeneratedFeedback, error)
}

// FeedbackRequest input untuk generate feedback
type FeedbackRequest struct {
	Questions []domain.Question
	Answers   []domain.AttemptAnswer
	Score     float64
	MaxScore  float64
}

type openAIGenerator struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIFeedbackGenerator membuat generator yang memanggil OpenAI API.
// apiKey: OPENAI_API_KEY. Jika kosong, Generate mengembalikan error (pakai fallback).
func NewOpenAIFeedbackGenerator(apiKey string) FeedbackGenerator {
	return &openAIGenerator{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type openAIChatRequest struct {
	Model    string        `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (g *openAIGenerator) Generate(ctx context.Context, req FeedbackRequest) (*GeneratedFeedback, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("openai: API key not set")
	}
	answerMap := make(map[string]domain.AttemptAnswer)
	for _, a := range req.Answers {
		answerMap[a.QuestionID] = a
	}
	var sb string
	for i, q := range req.Questions {
		ans := answerMap[q.ID]
		sb += fmt.Sprintf("Soal %d (tipe: %s, max_score: %.1f): %s\n", i+1, q.Type, q.MaxScore, truncate(q.Body, 200))
		if ans.AnswerText != nil && *ans.AnswerText != "" {
			sb += fmt.Sprintf("  Jawaban (teks): %s\n", truncate(*ans.AnswerText, 150))
		}
		if ans.SelectedOption != nil && *ans.SelectedOption != "" {
			sb += fmt.Sprintf("  Jawaban (pilihan): %s\n", *ans.SelectedOption)
		}
		if ans.AnswerText == nil && ans.SelectedOption == nil {
			sb += "  (tidak dijawab)\n"
		}
	}
	pct := 0.0
	if req.MaxScore > 0 {
		pct = req.Score / req.MaxScore * 100
	}
	systemPrompt := `Kamu adalah asisten pendidikan. Berikan feedback singkat untuk hasil tryout siswa dalam bahasa Indonesia.
Respon HARUS valid JSON saja, tanpa markdown atau teks lain, dengan format:
{"summary":"...","recap":"...","strength_areas":["...","..."],"improvement_areas":["...","..."],"recommendation":"..."}
- summary: ringkasan 1-2 kalimat hasil tryout
- recap: rangkuman skor dan performa (sebut skor dan persentase)
- strength_areas: array 2-4 area/topik yang dikuasai (string)
- improvement_areas: array 2-4 area yang perlu ditingkatkan (string)
- recommendation: 1-2 kalimat rekomendasi belajar`
	userPrompt := fmt.Sprintf("Skor siswa: %.2f dari %.2f (%.1f%%).\n\nDetail soal dan jawaban:\n%s\n\nBeri feedback JSON.", req.Score, req.MaxScore, pct, sb)

	body, _ := json.Marshal(openAIChatRequest{
		Model: "gpt-4o-mini",
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)
	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai: status %d", resp.StatusCode)
	}
	var chatResp openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}
	if len(chatResp.Choices) == 0 || chatResp.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf("openai: empty response")
	}
	content := chatResp.Choices[0].Message.Content
	content = extractJSON(content)
	var out GeneratedFeedback
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, fmt.Errorf("openai: parse response: %w", err)
	}
	return &out, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func extractJSON(s string) string {
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '{' {
			start = i
			break
		}
	}
	end := len(s)
	for i := len(s) - 1; i >= start; i-- {
		if s[i] == '}' {
			end = i + 1
			break
		}
	}
	return s[start:end]
}
