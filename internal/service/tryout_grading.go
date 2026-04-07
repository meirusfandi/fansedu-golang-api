package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// TryoutSubmitAnalysis grading + agregat modul + analisis keseluruhan (setelah submit / GET detail).
type TryoutSubmitAnalysis struct {
	Review  []QuestionReviewOutcome
	Modules []ModuleAnalysisAgg
	Overall TryoutOverallAnalysis
}

// TryoutOverallAnalysis ringkasan performa berdasarkan tipe soal + narasi untuk siswa.
type TryoutOverallAnalysis struct {
	TotalQuestions  int
	AnsweredCount   int
	UnansweredCount int
	CorrectCount    int // isCorrect == true
	WrongCount      int // isCorrect == false
	UnscoredCount   int // isCorrect == nil
	ScorePercent    float64
	ScoreGot        float64
	MaxScore        float64
	ByQuestionType  []QuestionTypeOverallStat
	Summary         string // narasi keseluruhan (Bahasa Indonesia)
}

// QuestionTypeOverallStat agregat per jenis soal (PG, benar-salah, isian).
type QuestionTypeOverallStat struct {
	Type         string
	Label        string
	Total        int
	Correct      int
	Wrong        int
	Unscored     int
	ScoreGot     float64
	MaxScore     float64
}

// QuestionReviewOutcome satu baris review setelah submit (untuk FE analisis modul + jawaban siswa).
type QuestionReviewOutcome struct {
	QuestionID       string
	SortOrder        int
	QuestionType     string
	QuestionTypeLabel string // label manusiawi sesuai tipe
	QuestionBody     string
	AnswerText       *string
	SelectedOption   *string
	CorrectOption    *string
	CorrectText      *string
	ScoreGot         float64
	MaxScore         float64
	IsCorrect        *bool
	ModuleKey        string
	ModuleLabel      string
	ModuleID         *string
	ModuleTitle      *string
	Bidang           *string
	Tags             []string
	AnalysisSummary  string // satu kalimat status
	AnalysisDetail   string // penjelasan per soal (tipe + jawaban + skor)
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

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// resolvedCorrectOption: kolom correct_option, atau opsi yang ditandai benar di JSON options (FE/admin).
func resolvedCorrectOption(q domain.Question) string {
	if q.CorrectOption != nil {
		if s := strings.TrimSpace(*q.CorrectOption); s != "" {
			return s
		}
	}
	return correctOptionFromOptionsJSON(q.Options)
}

func resolvedCorrectText(q domain.Question) string {
	if q.CorrectText != nil {
		if s := strings.TrimSpace(*q.CorrectText); s != "" {
			return s
		}
	}
	return ""
}

func normalizeAnswerText(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

// splitCorrectShortAnswers beberapa jawaban benar untuk isian: dipisah "|" di correct_text.
func splitCorrectShortAnswers(key string) []string {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	parts := strings.Split(key, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func shortAnswersMatch(studentAnswer, correctKey string) bool {
	g := normalizeAnswerText(studentAnswer)
	if g == "" {
		return false
	}
	for _, want := range splitCorrectShortAnswers(correctKey) {
		if strings.EqualFold(g, normalizeAnswerText(want)) {
			return true
		}
	}
	return false
}

// correctOptionFromOptionsJSON mendukung array objek, mis. {"key":"A","correct":true} atau isCorrect / is_correct.
func correctOptionFromOptionsJSON(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal(raw, &arr); err != nil {
		return ""
	}
	for _, o := range arr {
		truthy := false
		switch v := o["correct"].(type) {
		case bool:
			truthy = v
		}
		switch v := o["isCorrect"].(type) {
		case bool:
			truthy = truthy || v
		}
		switch v := o["is_correct"].(type) {
		case bool:
			truthy = truthy || v
		}
		switch v := o["isTrue"].(type) {
		case bool:
			truthy = truthy || v
		}
		if !truthy {
			continue
		}
		for _, k := range []string{"key", "value", "id", "option", "label"} {
			switch v := o[k].(type) {
			case string:
				if s := strings.TrimSpace(v); s != "" {
					return s
				}
			case float64:
				// JSON number untuk index / kode
				s := strings.TrimSpace(jsonNumberString(v))
				if s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func jsonNumberString(n float64) string {
	// bilangan bulat tanpa desimal .0
	if n == float64(int64(n)) {
		return fmt.Sprintf("%.0f", n)
	}
	return fmt.Sprintf("%g", n)
}

func reviewAnalysisSummary(hasAnswer bool, isCorrect *bool) string {
	if !hasAnswer {
		return "Soal tidak dijawab."
	}
	if isCorrect != nil {
		if *isCorrect {
			return "Jawaban Anda benar."
		}
		return "Jawaban Anda kurang tepat."
	}
	return "Jawaban tercatat; penilaian otomatis untuk soal ini belum tersedia."
}

func humanQuestionTypeLabel(qType string) string {
	switch qType {
	case domain.QuestionTypeMultipleChoice:
		return "Pilihan ganda"
	case domain.QuestionTypeTrueFalse:
		return "Benar / Salah"
	case domain.QuestionTypeShort:
		return "Isian singkat"
	default:
		if strings.TrimSpace(qType) == "" {
			return "Soal"
		}
		return "Soal (" + qType + ")"
	}
}

func buildPerQuestionAnalysis(q domain.Question, hasAns bool, ans *domain.AttemptAnswer, isCorrect *bool, scoreGot, maxScore float64) string {
	typeLabel := humanQuestionTypeLabel(q.Type)
	keyOpt := resolvedCorrectOption(q)
	keyTxt := resolvedCorrectText(q)

	switch q.Type {
	case domain.QuestionTypeMultipleChoice, domain.QuestionTypeTrueFalse:
		if !hasAns || ans == nil || ans.SelectedOption == nil || strings.TrimSpace(*ans.SelectedOption) == "" {
			return fmt.Sprintf("%s: tidak ada pilihan yang dikirim. Skor %.0f / %.0f.", typeLabel, scoreGot, maxScore)
		}
		sel := strings.TrimSpace(*ans.SelectedOption)
		if isCorrect != nil && *isCorrect {
			return fmt.Sprintf("%s: pilihan %q sesuai kunci. Skor %.0f / %.0f.", typeLabel, sel, scoreGot, maxScore)
		}
		if isCorrect != nil && !*isCorrect {
			if keyOpt != "" {
				return fmt.Sprintf("%s: Anda memilih %q; kunci %q. Skor %.0f / %.0f.", typeLabel, sel, keyOpt, scoreGot, maxScore)
			}
			return fmt.Sprintf("%s: pilihan %q dinilai salah atau skor 0. Skor %.0f / %.0f.", typeLabel, sel, scoreGot, maxScore)
		}
		return fmt.Sprintf("%s: pilihan %q tercatat; penilaian otomatis belum tersedia (pastikan kunci soal lengkap di bank soal).", typeLabel, sel)

	case domain.QuestionTypeShort:
		if !hasAns || ans == nil || ans.AnswerText == nil || strings.TrimSpace(*ans.AnswerText) == "" {
			return fmt.Sprintf("%s: tidak ada jawaban teks. Skor %.0f / %.0f.", typeLabel, scoreGot, maxScore)
		}
		userAns := strings.TrimSpace(*ans.AnswerText)
		if len(userAns) > 160 {
			userAns = userAns[:157] + "…"
		}
		if isCorrect != nil && *isCorrect {
			return fmt.Sprintf("%s: jawaban teks sesuai kunci. Skor %.0f / %.0f.", typeLabel, scoreGot, maxScore)
		}
		if isCorrect != nil && !*isCorrect {
			if keyTxt != "" {
				return fmt.Sprintf("%s: jawaban Anda tidak sama dengan kunci referensi. Skor %.0f / %.0f. (Cuplikan jawaban: %q)", typeLabel, scoreGot, maxScore, userAns)
			}
			return fmt.Sprintf("%s: jawaban tercatat %q; skor %.0f / %.0f.", typeLabel, userAns, scoreGot, maxScore)
		}
		return fmt.Sprintf("%s: jawaban %q tercatat; menunggu penilaian otomatis atau kunci isian.", typeLabel, userAns)

	default:
		var parts []string
		if hasAns && ans != nil {
			if ans.SelectedOption != nil && strings.TrimSpace(*ans.SelectedOption) != "" {
				parts = append(parts, fmt.Sprintf("pilihan %q", strings.TrimSpace(*ans.SelectedOption)))
			}
			if ans.AnswerText != nil && strings.TrimSpace(*ans.AnswerText) != "" {
				parts = append(parts, "jawaban teks tercatat")
			}
		}
		if len(parts) == 0 {
			return fmt.Sprintf("%s: tidak dijawab. Skor %.0f / %.0f.", typeLabel, scoreGot, maxScore)
		}
		return fmt.Sprintf("%s: %s. Skor %.0f / %.0f.", typeLabel, strings.Join(parts, ", "), scoreGot, maxScore)
	}
}

func outcomeHasAnswer(o QuestionReviewOutcome) bool {
	if o.SelectedOption != nil && strings.TrimSpace(*o.SelectedOption) != "" {
		return true
	}
	if o.AnswerText != nil && strings.TrimSpace(*o.AnswerText) != "" {
		return true
	}
	return false
}

func buildOverallTryoutAnalysis(outcomes []QuestionReviewOutcome, scoreGot, maxScore float64) TryoutOverallAnalysis {
	pct := 0.0
	if maxScore > 0 {
		pct = scoreGot / maxScore * 100
	}
	statMap := make(map[string]*QuestionTypeOverallStat)
	var answered, correctN, wrongN, unscoredN int
	for _, o := range outcomes {
		if outcomeHasAnswer(o) {
			answered++
		}
		if o.IsCorrect != nil {
			if *o.IsCorrect {
				correctN++
			} else {
				wrongN++
			}
		} else {
			unscoredN++
		}
		st := statMap[o.QuestionType]
		if st == nil {
			st = &QuestionTypeOverallStat{
				Type:  o.QuestionType,
				Label: humanQuestionTypeLabel(o.QuestionType),
			}
			statMap[o.QuestionType] = st
		}
		st.Total++
		st.ScoreGot += o.ScoreGot
		st.MaxScore += o.MaxScore
		if o.IsCorrect != nil {
			if *o.IsCorrect {
				st.Correct++
			} else {
				st.Wrong++
			}
		} else {
			st.Unscored++
		}
	}
	byType := make([]QuestionTypeOverallStat, 0, len(statMap))
	for _, st := range statMap {
		byType = append(byType, *st)
	}
	sort.Slice(byType, func(i, j int) bool {
		if byType[i].Label == byType[j].Label {
			return byType[i].Type < byType[j].Type
		}
		return byType[i].Label < byType[j].Label
	})
	n := len(outcomes)
	unanswered := n - answered
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total %d soal. Skor %.2f dari %.2f (%.1f%%). ", n, scoreGot, maxScore, pct))
	sb.WriteString(fmt.Sprintf("Dinilai otomatis: %d benar, %d salah; %d belum dinilai otomatis; %d tanpa jawaban. ",
		correctN, wrongN, unscoredN, unanswered))
	if len(byType) > 0 {
		sb.WriteString("Per jenis: ")
		for i, st := range byType {
			if i > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(fmt.Sprintf("%s: %d soal (benar %d, salah %d, belum dinilai %d; skor %.0f/%.0f).",
				st.Label, st.Total, st.Correct, st.Wrong, st.Unscored, st.ScoreGot, st.MaxScore))
		}
		sb.WriteString(" ")
	}
	if pct >= 75 {
		sb.WriteString("Secara keseluruhan hasil Anda kuat; pertahankan dan perdalam materi yang masih kurang.")
	} else if pct >= 50 {
		sb.WriteString("Hasil cukup; fokus ulang pada soal yang salah dan topik yang masih lemah.")
	} else {
		sb.WriteString("Masih banyak ruang untuk naik; ulangi konsep dasar dan latihan bertahap per tipe soal.")
	}
	return TryoutOverallAnalysis{
		TotalQuestions:   n,
		AnsweredCount:    answered,
		UnansweredCount:  unanswered,
		CorrectCount:     correctN,
		WrongCount:       wrongN,
		UnscoredCount:    unscoredN,
		ScorePercent:     pct,
		ScoreGot:         scoreGot,
		MaxScore:         maxScore,
		ByQuestionType:   byType,
		Summary:          sb.String(),
	}
}

// QuestionMissingAutoGradingKey true jika mode auto tidak bisa menilai soal (kunci PG/BS atau isian belum ada).
func QuestionMissingAutoGradingKey(q domain.Question) bool {
	switch q.Type {
	case domain.QuestionTypeMultipleChoice, domain.QuestionTypeTrueFalse:
		return resolvedCorrectOption(q) == ""
	case domain.QuestionTypeShort:
		return resolvedCorrectText(q) == ""
	default:
		return true
	}
}

func manualReviewAnalysisSummary(hasAnswer bool, score float64, isCorrect *bool) string {
	if !hasAnswer {
		return "Soal tidak dijawab."
	}
	if isCorrect == nil && score == 0 {
		return "Jawaban tercatat; menunggu penilaian pengajar."
	}
	if isCorrect != nil && *isCorrect {
		return "Jawaban dinilai benar."
	}
	if isCorrect != nil && !*isCorrect {
		return "Jawaban dinilai belum memenuhi skor penuh."
	}
	return "Jawaban telah dinilai."
}

func manualBuildPerQuestionAnalysis(q domain.Question, hasAns bool, ans *domain.AttemptAnswer, score float64, maxScr float64, isCorrect *bool) string {
	typeLabel := humanQuestionTypeLabel(q.Type)
	if !hasAns || ans == nil {
		return fmt.Sprintf("%s: tidak dijawab.", typeLabel)
	}
	if isCorrect == nil && score == 0 {
		return fmt.Sprintf("%s: jawaban tercatat; menunggu penilaian pengajar.", typeLabel)
	}
	if maxScr > 0 && score >= maxScr {
		return fmt.Sprintf("%s: skor penuh (%.0f / %.0f).", typeLabel, score, maxScr)
	}
	return fmt.Sprintf("%s: skor %.0f / %.0f.", typeLabel, score, maxScr)
}

// GradeTryoutAttemptWithMode: manual = hanya skor dari manual_score reviewer; kunci tidak ditampilkan di outcome.
func GradeTryoutAttemptWithMode(questions []domain.Question, answers []domain.AttemptAnswer, manualTryout bool) (totalScore, maxScore float64, outcomes []QuestionReviewOutcome, aggs []ModuleAnalysisAgg, overall TryoutOverallAnalysis) {
	answerMap := make(map[string]domain.AttemptAnswer)
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}
	modMap := make(map[string]*ModuleAnalysisAgg)
	grader := gradeQuestion
	if manualTryout {
		grader = gradeQuestionManualOnly
	}

	for _, q := range questions {
		maxScore += q.MaxScore
		ans, has := answerMap[q.ID]
		var ansPtr *domain.AttemptAnswer
		if has {
			ansPtr = &ans
		}
		score, isCorrect := grader(q, ansPtr)
		totalScore += score

		k, lbl := moduleKeyLabel(q)
		tags := parseTags(q.Tags)
		out := QuestionReviewOutcome{
			QuestionID:   q.ID,
			SortOrder:    q.SortOrder,
			QuestionType: q.Type,
			QuestionBody: q.Body,
			ScoreGot:     score,
			MaxScore:     q.MaxScore,
			IsCorrect:    isCorrect,
			ModuleKey:    k,
			ModuleLabel:  lbl,
			ModuleID:     q.ModuleID,
			ModuleTitle:  q.ModuleTitle,
			Bidang:       q.Bidang,
			Tags:         tags,
		}
		if has {
			if ans.AnswerText != nil {
				v := *ans.AnswerText
				out.AnswerText = &v
			}
			if ans.SelectedOption != nil {
				v := *ans.SelectedOption
				out.SelectedOption = &v
			}
		}
		if !manualTryout {
			out.CorrectOption = strPtr(resolvedCorrectOption(q))
			out.CorrectText = strPtr(resolvedCorrectText(q))
		}
		out.QuestionTypeLabel = humanQuestionTypeLabel(q.Type)
		if manualTryout {
			out.AnalysisSummary = manualReviewAnalysisSummary(has, score, isCorrect)
			out.AnalysisDetail = manualBuildPerQuestionAnalysis(q, has, ansPtr, score, q.MaxScore, isCorrect)
		} else {
			out.AnalysisSummary = reviewAnalysisSummary(has, isCorrect)
			out.AnalysisDetail = buildPerQuestionAnalysis(q, has, ansPtr, isCorrect, score, q.MaxScore)
		}
		outcomes = append(outcomes, out)

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
	overall = buildOverallTryoutAnalysis(outcomes, totalScore, maxScore)
	if manualTryout {
		overall.Summary = "Tryout ini dinilai manual. " + overall.Summary
	}
	return totalScore, maxScore, outcomes, aggs, overall
}

// GradeTryoutAttempt menghitung skor, benar/salah per soal, agregat modul, dan analisis keseluruhan (mode auto).
func GradeTryoutAttempt(questions []domain.Question, answers []domain.AttemptAnswer) (totalScore, maxScore float64, outcomes []QuestionReviewOutcome, aggs []ModuleAnalysisAgg, overall TryoutOverallAnalysis) {
	return GradeTryoutAttemptWithMode(questions, answers, false)
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

// ClampManualScoreToQuestionMax membatasi skor manual reviewer agar konsisten dengan gradeQuestion (0 … q.MaxScore).
func ClampManualScoreToQuestionMax(m float64, maxScore float64) float64 {
	if m < 0 {
		return 0
	}
	if m > maxScore {
		return maxScore
	}
	return m
}

// gradeQuestionManualOnly hanya memakai manual_score (jika ada); tidak memakai kunci PG/isian untuk otomatis.
func gradeQuestionManualOnly(q domain.Question, ans *domain.AttemptAnswer) (score float64, isCorrect *bool) {
	if ans != nil && ans.ManualScore != nil {
		return gradeQuestion(q, ans)
	}
	if ans == nil {
		return 0, nil
	}
	return 0, nil
}

func gradeQuestion(q domain.Question, ans *domain.AttemptAnswer) (score float64, isCorrect *bool) {
	if ans == nil {
		switch q.Type {
		case domain.QuestionTypeMultipleChoice, domain.QuestionTypeTrueFalse:
			if resolvedCorrectOption(q) != "" {
				return 0, boolPtr(false)
			}
			return 0, nil
		case domain.QuestionTypeShort:
			if resolvedCorrectText(q) != "" {
				return 0, boolPtr(false)
			}
			return 0, nil
		default:
			return 0, nil
		}
	}

	if ans.ManualScore != nil {
		m := ClampManualScoreToQuestionMax(*ans.ManualScore, q.MaxScore)
		var ic *bool
		if q.MaxScore <= 0 {
			if m <= 0 {
				ic = boolPtr(false)
			}
		} else if m >= q.MaxScore {
			ic = boolPtr(true)
		} else if m <= 0 {
			ic = boolPtr(false)
		}
		return m, ic
	}

	switch q.Type {
	case domain.QuestionTypeMultipleChoice, domain.QuestionTypeTrueFalse:
		key := resolvedCorrectOption(q)
		if key != "" {
			if ans.SelectedOption == nil || strings.TrimSpace(*ans.SelectedOption) == "" {
				return 0, boolPtr(false)
			}
			ok := strings.EqualFold(strings.TrimSpace(*ans.SelectedOption), key)
			if ok {
				return q.MaxScore, boolPtr(true)
			}
			return 0, boolPtr(false)
		}
		// Tanpa kunci PG/TF: tidak memakai skor parsial — tidak ada jawaban otomatis; skor 0 sampai admin set kunci.
		if ans.SelectedOption != nil && strings.TrimSpace(*ans.SelectedOption) != "" {
			return 0, nil
		}
		return 0, boolPtr(false)

	case domain.QuestionTypeShort:
		want := resolvedCorrectText(q)
		if want != "" {
			if ans.AnswerText == nil || strings.TrimSpace(*ans.AnswerText) == "" {
				return 0, boolPtr(false)
			}
			if shortAnswersMatch(*ans.AnswerText, want) {
				return q.MaxScore, boolPtr(true)
			}
			return 0, boolPtr(false)
		}
		// Isian tanpa kunci: tidak ada setengah poin; belum dinilai otomatis.
		if ans.AnswerText != nil && strings.TrimSpace(*ans.AnswerText) != "" {
			return 0, nil
		}
		return 0, boolPtr(false)

	default:
		if ans.SelectedOption != nil && strings.TrimSpace(*ans.SelectedOption) != "" {
			return 0, nil
		}
		if ans.AnswerText != nil && strings.TrimSpace(*ans.AnswerText) != "" {
			return 0, nil
		}
		return 0, boolPtr(false)
	}
}

// autoGradeQuestion skor otomatis saja (abaikan manual_score), untuk tampilan review admin/trainer.
func autoGradeQuestion(q domain.Question, ans *domain.AttemptAnswer) (float64, *bool) {
	if ans == nil {
		return gradeQuestion(q, nil)
	}
	tmp := *ans
	tmp.ManualScore = nil
	return gradeQuestion(q, &tmp)
}
