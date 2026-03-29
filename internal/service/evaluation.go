package service

import (
	"fmt"
	"strings"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// QuestionScoreDetail satu soal dalam detail penilaian attempt
type QuestionScoreDetail struct {
	QuestionID   string  `json:"questionId"`
	QuestionType string  `json:"questionType"`
	MaxScore     float64 `json:"maxScore"`
	ScoreGot     float64 `json:"scoreGot"`
	Status       string  `json:"status"`
}

// AttemptEvaluation hasil analisis attempt + rekomendasi (untuk dashboard)
type AttemptEvaluation struct {
	AttemptID        string                `json:"attemptId,omitempty"`
	AnswerBreakdown  []QuestionScoreDetail `json:"answerBreakdown"`
	StrengthAreas    []string              `json:"strengthAreas"`
	ImprovementAreas []string              `json:"improvementAreas"`
	Recommendation   string                `json:"recommendation"`
}

// ComputeQuestionScore mengembalikan score yang didapat untuk satu soal (selaras dengan penilaian submit tryout).
func ComputeQuestionScore(q domain.Question, ans *domain.AttemptAnswer) float64 {
	s, _ := gradeQuestion(q, ans)
	return s
}

// EvaluateAttemptAnswers membangun detail penilaian dan rekomendasi dari questions + answers (rule-based).
func EvaluateAttemptAnswers(questions []domain.Question, answers []domain.AttemptAnswer) AttemptEvaluation {
	answerMap := make(map[string]domain.AttemptAnswer)
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}
	var breakdown []QuestionScoreDetail
	typeStats := make(map[string]struct{ total, correct, partial, wrong int })
	for _, q := range questions {
		ans, hasAns := answerMap[q.ID]
		var ansPtr *domain.AttemptAnswer
		if hasAns {
			ansPtr = &ans
		}
		got, ic := gradeQuestion(q, ansPtr)
		status := "unanswered"
		if hasAns {
			if ic != nil {
				if *ic {
					status = "correct"
				} else {
					status = "wrong"
				}
			} else {
				if got >= q.MaxScore && q.MaxScore > 0 {
					status = "correct"
				} else if got > 0 {
					status = "partial"
				} else {
					status = "unscored"
				}
			}
		}
		breakdown = append(breakdown, QuestionScoreDetail{
			QuestionID:   q.ID,
			QuestionType: q.Type,
			MaxScore:     q.MaxScore,
			ScoreGot:     got,
			Status:       status,
		})
		t := typeStats[q.Type]
		t.total++
		switch status {
		case "correct":
			t.correct++
		case "partial", "unscored":
			t.partial++
		case "wrong", "unanswered":
			t.wrong++
		}
		typeStats[q.Type] = t
	}
	strength, improvement, rec := buildRecommendation(typeStats, breakdown)
	return AttemptEvaluation{
		AnswerBreakdown:  breakdown,
		StrengthAreas:    strength,
		ImprovementAreas: improvement,
		Recommendation:   rec,
	}
}

func buildRecommendation(typeStats map[string]struct{ total, correct, partial, wrong int }, breakdown []QuestionScoreDetail) (strength, improvement []string, recommendation string) {
	typeLabels := map[string]string{
		domain.QuestionTypeMultipleChoice: "Pilihan Ganda",
		domain.QuestionTypeTrueFalse:      "Benar/Salah",
		domain.QuestionTypeShort:         "Isian Singkat",
	}
	for qType, label := range typeLabels {
		s := typeStats[qType]
		if s.total == 0 {
			continue
		}
		acc := float64(s.correct) / float64(s.total)
		if acc >= 0.8 {
			strength = append(strength, label)
		} else if acc < 0.5 || s.wrong+s.partial > 0 {
			improvement = append(improvement, label)
		}
	}
	var totalScore, maxScore float64
	for _, b := range breakdown {
		totalScore += b.ScoreGot
		maxScore += b.MaxScore
	}
	pct := 0.0
	if maxScore > 0 {
		pct = totalScore / maxScore * 100
	}
	if len(improvement) == 0 && len(strength) > 0 {
		recommendation = "Hasil tryout Anda baik. Pertahankan dengan tetap berlatih, terutama pada area kekuatan Anda."
	} else if len(improvement) > 0 {
		recParts := []string{fmt.Sprintf("Skor keseluruhan: %.1f%% dari total.", pct)}
		recParts = append(recParts, "Fokus perbaiki: "+strings.Join(improvement, ", ")+".")
		recParts = append(recParts, "Rekomendasi: perbanyak latihan soal tipe tersebut dan ulangi materi terkait.")
		recommendation = strings.Join(recParts, " ")
	} else {
		recommendation = fmt.Sprintf("Skor: %.1f%%. Lanjutkan berlatih dan coba tryout lain untuk mengukur perkembangan.", pct)
	}
	return strength, improvement, recommendation
}
