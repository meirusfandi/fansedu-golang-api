package service

import (
	"encoding/json"
	"sort"
	"strings"
	"unicode"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// TryoutSubmitAnalysis grading + agregat modul untuk respons HTTP setelah submit.
type TryoutSubmitAnalysis struct {
	Review  []QuestionReviewOutcome
	Modules []ModuleAnalysisAgg
}

// QuestionReviewOutcome satu baris review setelah submit (untuk FE analisis modul).
type QuestionReviewOutcome struct {
	QuestionID   string
	ScoreGot     float64
	MaxScore     float64
	IsCorrect    *bool
	ModuleKey    string
	ModuleLabel  string
	ModuleID     *string
	ModuleTitle  *string
	Bidang       *string
	Tags         []string
}

// ModuleAnalysisAgg agregat per modul/topik.
type ModuleAnalysisAgg struct {
	ModuleKey     string
	ModuleLabel   string
	QuestionCount int
	CorrectCount  int
	WrongCount    int
	UnscoredCount int
}

func boolPtr(b bool) *bool { return &b }

// GradeTryoutAttempt menghitung skor, benar/salah per soal, dan agregat modul.
func GradeTryoutAttempt(questions []domain.Question, answers []domain.AttemptAnswer) (totalScore, maxScore float64, outcomes []QuestionReviewOutcome, aggs []ModuleAnalysisAgg) {
	answerMap := make(map[string]domain.AttemptAnswer)
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}
	modMap := make(map[string]*ModuleAnalysisAgg)

	for _, q := range questions {
		maxScore += q.MaxScore
		ans, has := answerMap[q.ID]
		var ansPtr *domain.AttemptAnswer
		if has {
			ansPtr = &ans
		}
		score, isCorrect := gradeQuestion(q, ansPtr)
		totalScore += score

		k, lbl := moduleKeyLabel(q)
		tags := parseTags(q.Tags)
		outcomes = append(outcomes, QuestionReviewOutcome{
			QuestionID:  q.ID,
			ScoreGot:    score,
			MaxScore:    q.MaxScore,
			IsCorrect:   isCorrect,
			ModuleKey:   k,
			ModuleLabel: lbl,
			ModuleID:    q.ModuleID,
			ModuleTitle: q.ModuleTitle,
			Bidang:      q.Bidang,
			Tags:        tags,
		})

		if modMap[k] == nil {
			modMap[k] = &ModuleAnalysisAgg{ModuleKey: k, ModuleLabel: lbl}
		}
		m := modMap[k]
		m.QuestionCount++
		if isCorrect == nil {
			m.UnscoredCount++
		} else if *isCorrect {
			m.CorrectCount++
		} else {
			m.WrongCount++
		}
	}

	for _, m := range modMap {
		aggs = append(aggs, *m)
	}
	sort.Slice(aggs, func(i, j int) bool {
		if aggs[i].ModuleLabel == aggs[j].ModuleLabel {
			return aggs[i].ModuleKey < aggs[j].ModuleKey
		}
		return aggs[i].ModuleLabel < aggs[j].ModuleLabel
	})
	return totalScore, maxScore, outcomes, aggs
}

func parseTags(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil
	}
	return arr
}

func moduleKeyLabel(q domain.Question) (key, label string) {
	if q.ModuleID != nil {
		if s := strings.TrimSpace(*q.ModuleID); s != "" {
			label = s
			if q.ModuleTitle != nil && strings.TrimSpace(*q.ModuleTitle) != "" {
				label = strings.TrimSpace(*q.ModuleTitle)
			}
			return s, label
		}
	}
	if q.ModuleTitle != nil {
		if s := strings.TrimSpace(*q.ModuleTitle); s != "" {
			return slugify(s), s
		}
	}
	if q.Bidang != nil {
		if s := strings.TrimSpace(*q.Bidang); s != "" {
			return slugify(s), s
		}
	}
	return "general", "Umum"
}

func slugify(s string) string {
	var b strings.Builder
	lastUnd := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUnd = false
			continue
		}
		if !lastUnd && b.Len() > 0 {
			b.WriteByte('_')
			lastUnd = true
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "general"
	}
	return out
}

func gradeQuestion(q domain.Question, ans *domain.AttemptAnswer) (score float64, isCorrect *bool) {
	if ans == nil {
		switch q.Type {
		case domain.QuestionTypeMultipleChoice, domain.QuestionTypeTrueFalse:
			if q.CorrectOption != nil && strings.TrimSpace(*q.CorrectOption) != "" {
				return 0, boolPtr(false)
			}
			return 0, nil
		case domain.QuestionTypeShort:
			if q.CorrectText != nil && strings.TrimSpace(*q.CorrectText) != "" {
				return 0, boolPtr(false)
			}
			return 0, nil
		default:
			return 0, nil
		}
	}

	switch q.Type {
	case domain.QuestionTypeMultipleChoice, domain.QuestionTypeTrueFalse:
		if q.CorrectOption != nil && strings.TrimSpace(*q.CorrectOption) != "" {
			if ans.SelectedOption == nil || strings.TrimSpace(*ans.SelectedOption) == "" {
				return 0, boolPtr(false)
			}
			ok := strings.EqualFold(strings.TrimSpace(*ans.SelectedOption), strings.TrimSpace(*q.CorrectOption))
			if ok {
				return q.MaxScore, boolPtr(true)
			}
			return 0, boolPtr(false)
		}
		// Tanpa kunci: perilaku lama (ada pilihan = dapat penuh), tidak tahu benar/salah.
		if ans.SelectedOption != nil && strings.TrimSpace(*ans.SelectedOption) != "" {
			return q.MaxScore, nil
		}
		return 0, boolPtr(false)

	case domain.QuestionTypeShort:
		if q.CorrectText != nil && strings.TrimSpace(*q.CorrectText) != "" {
			if ans.AnswerText == nil || strings.TrimSpace(*ans.AnswerText) == "" {
				return 0, boolPtr(false)
			}
			ok := strings.EqualFold(strings.TrimSpace(*ans.AnswerText), strings.TrimSpace(*q.CorrectText))
			if ok {
				return q.MaxScore, boolPtr(true)
			}
			return 0, boolPtr(false)
		}
		if ans.AnswerText != nil && strings.TrimSpace(*ans.AnswerText) != "" {
			return q.MaxScore * 0.5, nil
		}
		return 0, boolPtr(false)

	default:
		if ans.SelectedOption != nil && strings.TrimSpace(*ans.SelectedOption) != "" {
			return q.MaxScore, nil
		}
		if ans.AnswerText != nil && strings.TrimSpace(*ans.AnswerText) != "" {
			return q.MaxScore * 0.5, nil
		}
		return 0, nil
	}
}
