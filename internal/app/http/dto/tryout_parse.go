package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
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

// UnmarshalTryoutFloat64 parses JSON number, int, or numeric string (max_score dari form/JS).
func UnmarshalTryoutFloat64(raw json.RawMessage) (float64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, fmt.Errorf("null or empty float")
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return f, nil
	}
	var i int
	if err := json.Unmarshal(raw, &i); err == nil {
		return float64(i), nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseFloat(s, 64)
}

// PickQuestionJSONField picks first existing key (camelCase, snake_case, alias).
func PickQuestionJSONField(m map[string]json.RawMessage, keys ...string) (json.RawMessage, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v, true
		}
	}
	return nil, false
}

// fillQuestionCreateFromMap mengisi QuestionCreateRequest dari objek JSON (campuran penamaan field).
func fillQuestionCreateFromMap(m map[string]json.RawMessage, r *QuestionCreateRequest) error {
	if raw, ok := PickQuestionJSONField(m, "tryoutSessionId", "tryout_session_id"); ok {
		if err := json.Unmarshal(raw, &r.TryoutSessionID); err != nil {
			return fmt.Errorf("tryoutSessionId: %w", err)
		}
	}
	if raw, ok := PickQuestionJSONField(m, "sortOrder", "sort_order", "order"); ok {
		n, err := UnmarshalTryoutInt(raw)
		if err != nil {
			return fmt.Errorf("sortOrder: %w", err)
		}
		r.SortOrder = n
	}
	if raw, ok := PickQuestionJSONField(m, "type", "questionType", "question_type"); ok {
		if err := json.Unmarshal(raw, &r.Type); err != nil {
			return fmt.Errorf("type: %w", err)
		}
	}
	if raw, ok := PickQuestionJSONField(m, "body", "stem", "question", "questionText", "question_text"); ok {
		if err := json.Unmarshal(raw, &r.Body); err != nil {
			return fmt.Errorf("body: %w", err)
		}
	}
	if raw, ok := PickQuestionJSONField(m, "imageUrl", "image_url"); ok {
		if string(raw) == "null" {
			r.ImageURL = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("imageUrl: %w", err)
			}
			r.ImageURL = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "imageUrls", "image_urls"); ok {
		if string(raw) == "null" {
			r.ImageURLs = nil
		} else {
			var arr []string
			if err := json.Unmarshal(raw, &arr); err != nil {
				return fmt.Errorf("imageUrls: %w", err)
			}
			r.ImageURLs = arr
		}
	}
	if raw, ok := PickQuestionJSONField(m, "options", "choices"); ok {
		var v interface{}
		if err := json.Unmarshal(raw, &v); err != nil {
			return fmt.Errorf("options: %w", err)
		}
		r.Options = v
	}
	if raw, ok := PickQuestionJSONField(m, "maxScore", "max_score", "points"); ok {
		f, err := UnmarshalTryoutFloat64(raw)
		if err != nil {
			return fmt.Errorf("maxScore: %w", err)
		}
		r.MaxScore = f
	}
	if raw, ok := PickQuestionJSONField(m, "moduleId", "module_id"); ok {
		if string(raw) == "null" {
			r.ModuleID = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("moduleId: %w", err)
			}
			r.ModuleID = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "moduleTitle", "module_title"); ok {
		if string(raw) == "null" {
			r.ModuleTitle = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("moduleTitle: %w", err)
			}
			r.ModuleTitle = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "bidang"); ok {
		if string(raw) == "null" {
			r.Bidang = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("bidang: %w", err)
			}
			r.Bidang = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "tags"); ok {
		if string(raw) == "null" {
			r.Tags = nil
		} else {
			var arr []string
			if err := json.Unmarshal(raw, &arr); err != nil {
				return fmt.Errorf("tags: %w", err)
			}
			r.Tags = arr
		}
	}
	if raw, ok := PickQuestionJSONField(m, "correctOption", "correct_option"); ok {
		if string(raw) == "null" {
			r.CorrectOption = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("correctOption: %w", err)
			}
			r.CorrectOption = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "correctText", "correct_text"); ok {
		if string(raw) == "null" {
			r.CorrectText = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("correctText: %w", err)
			}
			r.CorrectText = &s
		}
	}
	return nil
}

// fillQuestionUpdateFromMap mengisi QuestionUpdateRequest hanya untuk key yang ada di JSON.
func fillQuestionUpdateFromMap(m map[string]json.RawMessage, r *QuestionUpdateRequest) error {
	if raw, ok := PickQuestionJSONField(m, "sortOrder", "sort_order", "order"); ok {
		if string(raw) == "null" {
			r.SortOrder = nil
		} else {
			n, err := UnmarshalTryoutInt(raw)
			if err != nil {
				return fmt.Errorf("sortOrder: %w", err)
			}
			r.SortOrder = &n
		}
	}
	if raw, ok := PickQuestionJSONField(m, "type", "questionType", "question_type"); ok {
		if string(raw) == "null" {
			r.Type = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("type: %w", err)
			}
			r.Type = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "body", "stem", "question", "questionText", "question_text"); ok {
		if string(raw) == "null" {
			r.Body = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("body: %w", err)
			}
			r.Body = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "imageUrl", "image_url"); ok {
		if string(raw) == "null" {
			r.ImageURL = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("imageUrl: %w", err)
			}
			r.ImageURL = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "imageUrls", "image_urls"); ok {
		if string(raw) == "null" {
			r.ImageURLs = nil
		} else {
			var arr []string
			if err := json.Unmarshal(raw, &arr); err != nil {
				return fmt.Errorf("imageUrls: %w", err)
			}
			r.ImageURLs = &arr
		}
	}
	if raw, ok := PickQuestionJSONField(m, "options", "choices"); ok {
		if string(raw) == "null" {
			r.Options = nil
		} else {
			var v interface{}
			if err := json.Unmarshal(raw, &v); err != nil {
				return fmt.Errorf("options: %w", err)
			}
			r.Options = &v
		}
	}
	if raw, ok := PickQuestionJSONField(m, "maxScore", "max_score", "points"); ok {
		if string(raw) == "null" {
			r.MaxScore = nil
		} else {
			f, err := UnmarshalTryoutFloat64(raw)
			if err != nil {
				return fmt.Errorf("maxScore: %w", err)
			}
			r.MaxScore = &f
		}
	}
	if raw, ok := PickQuestionJSONField(m, "moduleId", "module_id"); ok {
		if string(raw) == "null" {
			r.ModuleID = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("moduleId: %w", err)
			}
			r.ModuleID = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "moduleTitle", "module_title"); ok {
		if string(raw) == "null" {
			r.ModuleTitle = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("moduleTitle: %w", err)
			}
			r.ModuleTitle = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "bidang"); ok {
		if string(raw) == "null" {
			r.Bidang = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("bidang: %w", err)
			}
			r.Bidang = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "tags"); ok {
		if string(raw) == "null" {
			r.Tags = nil
		} else {
			var arr []string
			if err := json.Unmarshal(raw, &arr); err != nil {
				return fmt.Errorf("tags: %w", err)
			}
			r.Tags = &arr
		}
	}
	if raw, ok := PickQuestionJSONField(m, "correctOption", "correct_option"); ok {
		if string(raw) == "null" {
			r.CorrectOption = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("correctOption: %w", err)
			}
			r.CorrectOption = &s
		}
	}
	if raw, ok := PickQuestionJSONField(m, "correctText", "correct_text"); ok {
		if string(raw) == "null" {
			r.CorrectText = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("correctText: %w", err)
			}
			r.CorrectText = &s
		}
	}
	return nil
}

// UnmarshalJSON untuk POST soal admin — terima camelCase dan snake_case.
func (r *QuestionCreateRequest) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	return fillQuestionCreateFromMap(m, r)
}

// UnmarshalJSON untuk PUT soal admin — hanya field yang dikirim di body yang diisi.
func (r *QuestionUpdateRequest) UnmarshalJSON(data []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	return fillQuestionUpdateFromMap(m, r)
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
	if raw, ok := PickTryoutJSONField(m, "subject", "subject"); ok {
		if string(raw) == "null" {
			r.Subject = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("subject: %w", err)
			}
			s = strings.TrimSpace(s)
			if s == "" {
				r.Subject = nil
			} else {
				r.Subject = &s
			}
		}
	}
	if raw, ok := PickTryoutJSONField(m, "schoolLevel", "school_level"); ok {
		if string(raw) == "null" {
			r.SchoolLevel = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("schoolLevel: %w", err)
			}
			s = strings.TrimSpace(s)
			if s == "" {
				r.SchoolLevel = nil
			} else {
				r.SchoolLevel = &s
			}
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
	if raw, ok := PickTryoutJSONField(m, "levelId", "level_id"); ok {
		if string(raw) == "null" {
			r.LevelID = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return fmt.Errorf("levelId: %w", err)
			}
			s = strings.TrimSpace(s)
			if s == "" {
				r.LevelID = nil
			} else {
				r.LevelID = &s
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
