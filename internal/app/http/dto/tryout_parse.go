package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PickTryoutJSONField returns the raw value for camelCase or snake_case key (admin / docs mix).
func PickTryoutJSONField(m map[string]json.RawMessage, camel, snake string) (json.RawMessage, bool) {
	if v, ok := m[camel]; ok {
		return v, true
	}
	if v, ok := m[snake]; ok {
		return v, true
	}
	return nil, false
}

// UnmarshalTryoutTimeJSON parses opens_at / closes_at from JSON (string or RFC3339 time).
func UnmarshalTryoutTimeJSON(raw json.RawMessage) (time.Time, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return time.Time{}, fmt.Errorf("null or empty time")
	}
	var tt time.Time
	if err := json.Unmarshal(raw, &tt); err == nil && !tt.IsZero() {
		return tt, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return time.Time{}, err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, lay := range layouts {
		if t, err := time.Parse(lay, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time: %q", s)
}

// UnmarshalTryoutInt unmarshals JSON number as int or float (nilai dari JS/JSON umumnya aman).
func UnmarshalTryoutInt(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, fmt.Errorf("null or empty int")
	}
	var i int
	if err := json.Unmarshal(raw, &i); err == nil {
		return i, nil
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0, err
	}
	return int(f), nil
}

// fillTryoutCreateFromMap fills TryoutCreateRequest from a JSON object (camelCase and/or snake_case keys).
func fillTryoutCreateFromMap(m map[string]json.RawMessage, r *TryoutCreateRequest) error {
	if raw, ok := PickTryoutJSONField(m, "title", "title"); ok {
		if err := json.Unmarshal(raw, &r.Title); err != nil {
			return fmt.Errorf("title: %w", err)
		}
	}
	if raw, ok := PickTryoutJSONField(m, "shortTitle", "short_title"); ok {
		if string(raw) == "null" {
			r.ShortTitle = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("shortTitle: %w", err)
			}
			r.ShortTitle = &s
		}
	}
	if raw, ok := PickTryoutJSONField(m, "description", "description"); ok {
		if string(raw) == "null" {
			r.Description = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("description: %w", err)
			}
			r.Description = &s
		}
	}
	if raw, ok := PickTryoutJSONField(m, "durationMinutes", "duration_minutes"); ok {
		n, err := UnmarshalTryoutInt(raw)
		if err != nil {
			return fmt.Errorf("durationMinutes: %w", err)
		}
		r.DurationMinutes = n
	}
	if raw, ok := PickTryoutJSONField(m, "questionsCount", "questions_count"); ok {
		n, err := UnmarshalTryoutInt(raw)
		if err != nil {
			return fmt.Errorf("questionsCount: %w", err)
		}
		r.QuestionsCount = n
	}
	if raw, ok := PickTryoutJSONField(m, "level", "level"); ok {
		if err := json.Unmarshal(raw, &r.Level); err != nil {
			return fmt.Errorf("level: %w", err)
		}
	}
	if raw, ok := PickTryoutJSONField(m, "subjectId", "subject_id"); ok {
		if string(raw) == "null" {
			r.SubjectID = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("subjectId: %w", err)
			}
			s = strings.TrimSpace(s)
			if s == "" {
				r.SubjectID = nil
			} else {
				r.SubjectID = &s
			}
		}
	}
	if raw, ok := PickTryoutJSONField(m, "opensAt", "opens_at"); ok {
		tt, err := UnmarshalTryoutTimeJSON(raw)
		if err != nil {
			return fmt.Errorf("opensAt: %w", err)
		}
		r.OpensAt = tt
	}
	if raw, ok := PickTryoutJSONField(m, "closesAt", "closes_at"); ok {
		tt, err := UnmarshalTryoutTimeJSON(raw)
		if err != nil {
			return fmt.Errorf("closesAt: %w", err)
		}
		r.ClosesAt = tt
	}
	if raw, ok := PickTryoutJSONField(m, "maxParticipants", "max_participants"); ok {
		if string(raw) == "null" {
			r.MaxParticipants = nil
		} else {
			n, err := UnmarshalTryoutInt(raw)
			if err != nil {
				return fmt.Errorf("maxParticipants: %w", err)
			}
			r.MaxParticipants = &n
		}
	}
	if raw, ok := PickTryoutJSONField(m, "status", "status"); ok {
		if err := json.Unmarshal(raw, &r.Status); err != nil {
			return fmt.Errorf("status: %w", err)
		}
	}
	return nil
}

// UnmarshalJSON accepts camelCase (frontend) and snake_case (dokumentasi / tooling lama).
func (r *TryoutCreateRequest) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	return fillTryoutCreateFromMap(m, r)
}
