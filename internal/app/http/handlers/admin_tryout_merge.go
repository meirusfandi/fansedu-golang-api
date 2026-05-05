package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// mergeTryoutSessionFromJSON updates only fields present in the JSON body (camelCase).
// Mencegah PUT partial dari admin UI menimpa data dengan nilai nol.
func mergeTryoutSessionFromJSON(data []byte, t *domain.TryoutSession) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if raw, ok := dto.PickTryoutJSONField(m, "title"); ok {
		if err := json.Unmarshal(raw, &t.Title); err != nil {
			return fmt.Errorf("title: %w", err)
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "shortTitle"); ok {
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
	if raw, ok := dto.PickTryoutJSONField(m, "description"); ok {
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
	if raw, ok := dto.PickTryoutJSONField(m, "durationMinutes"); ok {
		n, err := dto.UnmarshalTryoutInt(raw)
		if err != nil {
			return fmt.Errorf("durationMinutes: %w", err)
		}
		t.DurationMinutes = n
	}
	if raw, ok := dto.PickTryoutJSONField(m, "questionsCount"); ok {
		n, err := dto.UnmarshalTryoutInt(raw)
		if err != nil {
			return fmt.Errorf("questionsCount: %w", err)
		}
		t.QuestionsCount = n
	}
	if raw, ok := dto.PickTryoutJSONField(m, "level"); ok {
		if err := json.Unmarshal(raw, &t.Level); err != nil {
			return fmt.Errorf("level: %w", err)
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "subject"); ok {
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
	if raw, ok := dto.PickTryoutJSONField(m, "schoolLevel"); ok {
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
	if raw, ok := dto.PickTryoutJSONField(m, "subjectId"); ok {
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
	if raw, ok := dto.PickTryoutJSONField(m, "levelId"); ok {
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
	if raw, ok := dto.PickTryoutJSONField(m, "opensAt"); ok {
		tt, err := dto.UnmarshalTryoutTimeJSON(raw)
		if err != nil {
			return fmt.Errorf("opensAt: %w", err)
		}
		t.OpensAt = tt
	}
	if raw, ok := dto.PickTryoutJSONField(m, "closesAt"); ok {
		tt, err := dto.UnmarshalTryoutTimeJSON(raw)
		if err != nil {
			return fmt.Errorf("closesAt: %w", err)
		}
		t.ClosesAt = tt
	}
	if raw, ok := dto.PickTryoutJSONField(m, "maxParticipants"); ok {
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
	if raw, ok := dto.PickTryoutJSONField(m, "status"); ok {
		if err := json.Unmarshal(raw, &t.Status); err != nil {
			return fmt.Errorf("status: %w", err)
		}
	}
	if raw, ok := dto.PickTryoutJSONField(m, "gradingMode"); ok {
		if err := json.Unmarshal(raw, &t.GradingMode); err != nil {
			return fmt.Errorf("gradingMode: %w", err)
		}
	}
	return nil
}
