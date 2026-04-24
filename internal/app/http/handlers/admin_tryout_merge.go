package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// mergeTryoutSessionFromJSON updates only fields present in the JSON body (camelCase or snake_case).
// Mencegah PUT partial dari admin UI menimpa data dengan nilai nol.
func mergeTryoutSessionFromJSON(data []byte, t *domain.TryoutSession) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if raw, ok := dto.PickTryoutJSONField(m, "title", "title"); ok {
		if err := json.Unmarshal(raw, &t.Title); err != nil {
			return fmt.Errorf("title: %w", err)
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "shortTitle", "short_title"); ok {
		if string(raw) == "null" {
			t.ShortTitle = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("shortTitle: %w", err)
			}
			t.ShortTitle = &s
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "description", "description"); ok {
		if string(raw) == "null" {
			t.Description = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("description: %w", err)
			}
			t.Description = &s
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "durationMinutes", "duration_minutes"); ok {
		n, err := dto.UnmarshalTryoutInt(raw)
		if err != nil {
			return fmt.Errorf("durationMinutes: %w", err)
		}
		t.DurationMinutes = n
	}
	if raw, ok := dto.PickTryoutJSONField(m, "questionsCount", "questions_count"); ok {
		n, err := dto.UnmarshalTryoutInt(raw)
		if err != nil {
			return fmt.Errorf("questionsCount: %w", err)
		}
		t.QuestionsCount = n
	}
	if raw, ok := dto.PickTryoutJSONField(m, "level", "level"); ok {
		if err := json.Unmarshal(raw, &t.Level); err != nil {
			return fmt.Errorf("level: %w", err)
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "subject", "subject"); ok {
		if string(raw) == "null" {
			t.Subject = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("subject: %w", err)
			}
			s = strings.TrimSpace(s)
			if s == "" {
				t.Subject = nil
			} else {
				t.Subject = &s
			}
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "schoolLevel", "school_level"); ok {
		if string(raw) == "null" {
			t.SchoolLevel = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("schoolLevel: %w", err)
			}
			s = strings.TrimSpace(s)
			if s == "" {
				t.SchoolLevel = nil
			} else {
				t.SchoolLevel = &s
			}
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "subjectId", "subject_id"); ok {
		if string(raw) == "null" {
			t.SubjectID = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("subjectId: %w", err)
			}
			s = strings.TrimSpace(s)
			if s == "" {
				t.SubjectID = nil
			} else {
				t.SubjectID = &s
			}
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "levelId", "level_id"); ok {
		if string(raw) == "null" {
			t.LevelID = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("levelId: %w", err)
			}
			s = strings.TrimSpace(s)
			if s == "" {
				t.LevelID = nil
			} else {
				t.LevelID = &s
			}
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "opensAt", "opens_at"); ok {
		tt, err := dto.UnmarshalTryoutTimeJSON(raw)
		if err != nil {
			return fmt.Errorf("opensAt: %w", err)
		}
		t.OpensAt = tt
	}
	if raw, ok := dto.PickTryoutJSONField(m, "closesAt", "closes_at"); ok {
		tt, err := dto.UnmarshalTryoutTimeJSON(raw)
		if err != nil {
			return fmt.Errorf("closesAt: %w", err)
		}
		t.ClosesAt = tt
	}
	if raw, ok := dto.PickTryoutJSONField(m, "maxParticipants", "max_participants"); ok {
		if string(raw) == "null" {
			t.MaxParticipants = nil
		} else {
			n, err := dto.UnmarshalTryoutInt(raw)
			if err != nil {
				return fmt.Errorf("maxParticipants: %w", err)
			}
			t.MaxParticipants = &n
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "status", "status"); ok {
		if err := json.Unmarshal(raw, &t.Status); err != nil {
			return fmt.Errorf("status: %w", err)
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "gradingMode", "grading_mode"); ok {
		if err := json.Unmarshal(raw, &t.GradingMode); err != nil {
			return fmt.Errorf("gradingMode: %w", err)
		}
	}
	return nil
}
