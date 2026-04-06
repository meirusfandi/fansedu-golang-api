# AI Question Generator - Fansedu

Dokumen ini merangkum arsitektur production-ready untuk generator soal olimpiade (Math SD/SMP, Informatika SMA).

## 1) Struktur data soal

Entity utama:

- `subject`: `math | informatics`
- `grade`: `sd | smp | sma`
- `topic`: contoh `dp`, `graph`, `aritmatika`
- `difficulty`: `easy | medium | hard | olympiad`
- `question_text`
- `choices_json` (untuk MCQ)
- `correct_answer`
- `explanation` + `solution_steps`
- `concept_tags`
- `estimated_time_sec`

Implementasi tabel ada di:

- `internal/db/migrations/055_ai_question_engine.sql`

## 2) Algoritma generator (Hybrid Rule-based + AI)

Pipeline:

1. Ambil bank soal sesuai filter (`subject/grade/topic/difficulty`)
2. Random + shuffle untuk variasi
3. Jika stok kurang:
   - generate variasi rule-based template (fallback deterministic)
   - (opsional tahap lanjut) enrich via LLM untuk narasi dan distractor
4. Difficulty scaling:
   - easy: langkah pendek, waktu estimasi rendah
   - medium: multi-step
   - hard: non-obvious
   - olympiad: insight-required

Implementasi usecase:

- `internal/usecase/questiongen/usecase.go`

## 3) Pseudocode

```text
generate(req):
  validate(req)
  bank = repo.list(req.filters)
  bank = enrich(bank) // solution_steps, tag, normalisasi output
  if len(bank) >= req.count:
    return shuffle(bank)[0:req.count]

  need = req.count - len(bank)
  variants = template.generate(req, need)  // rule-based fallback
  all = shuffle(bank + variants)
  return all[0:req.count]
```

## 4) Rekomendasi berdasarkan kesalahan user

Saat `GET /analysis`:

- hitung `accuracy`, `avg_time`, `weak_topic`
- otomatis ambil `recommendations` dari topik lemah
- default difficulty `medium` untuk remedial sebelum naik level

## 5) Endpoint

- `POST /api/v1/generate-questions`
- `POST /api/v1/submit-answer`
- `GET /api/v1/analysis`
- `GET /api/v1/ranking`
- `GET /api/v1/questions`
- `POST /api/v1/subscription`

## 6) Strategi scaling 1000+ user

- PostgreSQL index pada filter utama (`subject, grade, topic, difficulty`)
- Redis untuk cache hasil query populer + leaderboard
- Asynchronous worker untuk AI enrichment batch
- Rate limiting + retry policy untuk provider LLM
- Pre-generate bank soal untuk topik populer
- Monitoring: p95 latency, cache hit ratio, DB slow query, error rate

