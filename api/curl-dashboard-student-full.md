# Curl — Dashboard Siswa (lengkap dengan detail penilaian & rekomendasi)

Base URL: `http://localhost:8080/api/v1`

---

## Request

```bash
# 1. Login dulu untuk dapat token (ganti email/password sesuai user siswa)
curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"ahmad@example.com","password":"password123"}'

# 2. Simpan token dari response, lalu panggil dashboard
export TOKEN="<paste_token_dari_response_login>"

curl -s -X GET "http://localhost:8080/api/v1/student/dashboard" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

Satu baris (ganti `$TOKEN`):

```bash
curl -s -X GET "http://localhost:8080/api/v1/student/dashboard" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Response (200 OK) — contoh lengkap

Termasuk **detail penilaian attempt** (`answer_breakdown`), **rekomendasi belajar**, dan **bagian yang perlu ditingkatkan** (`improvement_areas` / `strength_areas`).

```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Ahmad Siswa",
    "email": "ahmad@example.com",
    "role": "student",
    "avatar_url": null,
    "school_id": "660e8400-e29b-41d4-a716-446655440001",
    "subject_id": "770e8400-e29b-41d4-a716-446655440002",
    "school_name": "SMAN 1 Solok",
    "subject_name": "Matematika"
  },
  "summary": {
    "total_attempts": 3,
    "avg_score": 72.5,
    "avg_percentile": 68
  },
  "open_tryouts": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "title": "Tryout Matematika SMA",
      "short_title": "Tryout MTK",
      "description": "Tryout persiapan OSN",
      "duration_minutes": 90,
      "questions_count": 20,
      "level": "medium",
      "subject_id": "770e8400-e29b-41d4-a716-446655440002",
      "opens_at": "2025-02-01T08:00:00Z",
      "closes_at": "2025-02-28T22:00:00Z",
      "max_participants": null,
      "status": "open"
    }
  ],
  "recent_attempts": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440004",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "tryout_session_id": "880e8400-e29b-41d4-a716-446655440003",
      "started_at": "2025-02-15T10:00:00Z",
      "submitted_at": "2025-02-15T11:25:00Z",
      "status": "submitted",
      "score": 75.5,
      "max_score": 100,
      "percentile": 72,
      "time_seconds_spent": 5100
    }
  ],
  "strength_areas": [
    "Pilihan Ganda"
  ],
  "improvement_areas": [
    "Isian Singkat"
  ],
  "recommendation": "Skor keseluruhan: 75.5% dari total. Fokus perbaiki: Isian Singkat. Rekomendasi: perbanyak latihan soal tipe tersebut dan ulangi materi terkait.",
  "learning_evaluation": {
    "attempt_id": "990e8400-e29b-41d4-a716-446655440004",
    "answer_breakdown": [
      {
        "question_id": "q-uuid-1",
        "question_type": "multiple_choice",
        "max_score": 5,
        "score_got": 5,
        "status": "correct"
      },
      {
        "question_id": "q-uuid-2",
        "question_type": "short",
        "max_score": 5,
        "score_got": 2.5,
        "status": "partial"
      },
      {
        "question_id": "q-uuid-3",
        "question_type": "true_false",
        "max_score": 5,
        "score_got": 0,
        "status": "wrong"
      }
    ],
    "strength_areas": [
      "Pilihan Ganda"
    ],
    "improvement_areas": [
      "Isian Singkat"
    ],
    "recommendation": "Skor keseluruhan: 75.5% dari total. Fokus perbaiki: Isian Singkat. Rekomendasi: perbanyak latihan soal tipe tersebut dan ulangi materi terkait."
  }
}
```

---

## Penjelasan field yang diminta

| Field | Keterangan |
|-------|-------------|
| **user** | Data user siswa (nama, email, sekolah, bidang/subject). |
| **summary** | Ringkasan: total attempt, rata-rata skor, rata-rata percentile. |
| **open_tryouts** | Daftar tryout yang sedang buka. |
| **recent_attempts** | Riwayat attempt terbaru (id, skor, status, waktu). |
| **strength_areas** | Area kuat (dari analisis jawaban, mis. "Pilihan Ganda"). |
| **improvement_areas** | Bagian yang perlu ditingkatkan (mis. "Isian Singkat"). |
| **recommendation** | Rekomendasi belajar (teks) untuk siswa. |
| **learning_evaluation** | Detail penilaian per attempt terakhir yang sudah submit. |
| **learning_evaluation.attempt_id** | ID attempt yang dianalisis. |
| **learning_evaluation.answer_breakdown** | Detail per soal: `question_id`, `question_type`, `max_score`, `score_got`, `status` (`correct` / `partial` / `wrong` / `unanswered`). |
| **learning_evaluation.strength_areas** | Area kuat dari attempt tersebut. |
| **learning_evaluation.improvement_areas** | Bagian yang perlu dikembangkan dari attempt tersebut. |
| **learning_evaluation.recommendation** | Rekomendasi belajar berdasarkan attempt tersebut. |

**Catatan:** `learning_evaluation` hanya ada jika siswa punya minimal satu attempt dengan status `submitted`. Jika belum pernah submit tryout, field ini tidak muncul (atau `null`).
