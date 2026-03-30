package ai

import (
	"context"
	"fmt"
	"strings"
)

// FallbackFeedbackGenerator feedback rule-based tanpa AI (jika AI tidak tersedia)
type FallbackFeedbackGenerator struct{}

func NewFallbackFeedbackGenerator() FeedbackGenerator {
	return &FallbackFeedbackGenerator{}
}

func (g *FallbackFeedbackGenerator) Generate(ctx context.Context, req FeedbackRequest) (*GeneratedFeedback, error) {
	pct := 0.0
	if req.MaxScore > 0 {
		pct = req.Score / req.MaxScore * 100
	}
	summary := fmt.Sprintf("Tryout selesai. Skor Anda %.2f dari %.2f.", req.Score, req.MaxScore)
	recap := fmt.Sprintf("Skor: %.2f dari %.2f (%.1f%%). Lanjutkan berlatih untuk hasil lebih baik.", req.Score, req.MaxScore, pct)
	if s := strings.TrimSpace(req.OverallNarrative); s != "" {
		recap = fmt.Sprintf("Skor: %.2f dari %.2f (%.1f%%). %s", req.Score, req.MaxScore, pct, s)
	}
	strength := []string{}
	improvement := []string{"Perbanyak latihan soal", "Review materi yang kurang dikuasai"}
	if pct >= 70 {
		strength = append(strength, "Performa baik")
	}
	if pct < 50 {
		improvement = append(improvement, "Fokus pada dasar materi")
	}
	recommendation := "Lanjutkan berlatih dan ulangi materi yang masih lemah."
	return &GeneratedFeedback{
		Summary:          summary,
		Recap:            recap,
		StrengthAreas:    strength,
		ImprovementAreas: improvement,
		Recommendation:   recommendation,
	}, nil
}

// Ensure FallbackFeedbackGenerator implements FeedbackGenerator
var _ FeedbackGenerator = (*FallbackFeedbackGenerator)(nil)
