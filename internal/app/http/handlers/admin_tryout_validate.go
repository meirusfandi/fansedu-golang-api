package handlers

import (
	"fmt"
	"strings"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

func normalizeTryoutLevel(level string) string {
	if strings.TrimSpace(level) == "" {
		return domain.TryoutLevelMedium
	}
	return strings.TrimSpace(level)
}

func isValidTryoutLevel(level string) bool {
	switch level {
	case domain.TryoutLevelEasy, domain.TryoutLevelMedium, domain.TryoutLevelHard:
		return true
	default:
		return false
	}
}

func isValidTryoutStatus(status string) bool {
	switch status {
	case domain.TryoutStatusDraft, domain.TryoutStatusOpen, domain.TryoutStatusClosed:
		return true
	default:
		return false
	}
}

func normalizeTryoutGradingMode(s string) string {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case domain.TryoutGradingModeManual:
		return domain.TryoutGradingModeManual
	default:
		return domain.TryoutGradingModeAuto
	}
}

// validateTryoutForCreate memastikan data wajib terisi sebelum INSERT.
func validateTryoutForCreate(t *domain.TryoutSession) error {
	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" {
		return fmt.Errorf("title is required")
	}
	if t.OpensAt.IsZero() || t.ClosesAt.IsZero() {
		return fmt.Errorf("opensAt and closesAt are required")
	}
	if t.ClosesAt.Before(t.OpensAt) {
		return fmt.Errorf("closesAt must be on or after opensAt")
	}
	t.Level = normalizeTryoutLevel(t.Level)
	if !isValidTryoutLevel(t.Level) {
		return fmt.Errorf("invalid level: use easy, medium, or hard")
	}
	if t.Status == "" {
		t.Status = domain.TryoutStatusOpen
	}
	if !isValidTryoutStatus(t.Status) {
		return fmt.Errorf("invalid status: use draft, open, or closed")
	}
	if t.DurationMinutes < 0 || t.QuestionsCount < 0 {
		return fmt.Errorf("durationMinutes and questionsCount must be >= 0")
	}
	if t.MaxParticipants != nil && *t.MaxParticipants < 0 {
		return fmt.Errorf("maxParticipants must be >= 0")
	}
	t.GradingMode = normalizeTryoutGradingMode(t.GradingMode)
	return nil
}

// validateTryoutAfterAdminUpdate memvalidasi state akhir setelah merge PATCH.
func validateTryoutAfterAdminUpdate(t *domain.TryoutSession) error {
	t.Title = strings.TrimSpace(t.Title)
	t.Level = strings.TrimSpace(t.Level)
	t.Status = strings.TrimSpace(t.Status)
	if t.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if t.OpensAt.IsZero() || t.ClosesAt.IsZero() {
		return fmt.Errorf("opensAt and closesAt must be set")
	}
	if t.ClosesAt.Before(t.OpensAt) {
		return fmt.Errorf("closesAt must be on or after opensAt")
	}
	if !isValidTryoutLevel(t.Level) {
		return fmt.Errorf("invalid level: use easy, medium, or hard")
	}
	if !isValidTryoutStatus(t.Status) {
		return fmt.Errorf("invalid status: use draft, open, or closed")
	}
	if t.DurationMinutes < 0 || t.QuestionsCount < 0 {
		return fmt.Errorf("durationMinutes and questionsCount must be >= 0")
	}
	if t.MaxParticipants != nil && *t.MaxParticipants < 0 {
		return fmt.Errorf("maxParticipants must be >= 0")
	}
	t.GradingMode = normalizeTryoutGradingMode(t.GradingMode)
	return nil
}

// restoreTryoutFieldsIfEmptyPatch mengembalikan level/status/title jika body mengirim string kosong (bukan menghapus key).
func restoreTryoutFieldsIfEmptyPatch(t *domain.TryoutSession, orig domain.TryoutSession) {
	if strings.TrimSpace(t.Title) == "" {
		t.Title = orig.Title
	}
	if strings.TrimSpace(t.Level) == "" {
		t.Level = orig.Level
	}
	if strings.TrimSpace(t.Status) == "" {
		t.Status = orig.Status
	}
	if strings.TrimSpace(t.GradingMode) == "" {
		t.GradingMode = orig.GradingMode
	}
	if t.GradingMode == "" {
		t.GradingMode = domain.TryoutGradingModeAuto
	}
}
