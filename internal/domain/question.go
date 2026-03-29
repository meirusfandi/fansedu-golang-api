package domain

import (
	"encoding/json"
	"time"
)

const (
	QuestionTypeShort          = "short"
	QuestionTypeMultipleChoice = "multiple_choice"
	QuestionTypeTrueFalse       = "true_false"
)

type Question struct {
	ID               string
	TryoutSessionID  string
	SortOrder        int
	Type             string
	Body             string          // Teks atau HTML isi soal (boleh berisi HTML dan <img>)
	ImageURL         *string         // URL gambar utama (opsional, backward compatible)
	ImageURLs        json.RawMessage // JSONB array URL gambar: ["url1", "url2"]
	Options          json.RawMessage // JSONB: ["A","B","C","D"] or [{"key":"A","label":"..."}]
	MaxScore         float64
	ModuleID         *string // pengelompokan modul (opsional)
	ModuleTitle      *string
	Bidang           *string
	Tags             json.RawMessage // JSONB array string
	CorrectOption    *string         // kunci PG/benar-salah (tidak dikirim ke siswa saat ujian)
	CorrectText      *string         // kunci isian singkat (opsional)
	CreatedAt        time.Time
}
