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

Termasuk **detail penilaian attempt** (`answerBreakdown`), **rekomendasi belajar**, dan **bagian yang perlu ditingkatkan** (`improvementAreas` / `strengthAreas`).

```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Ahmad Siswa",
    "email": "ahmad@example.com",
    "role": "student",
    "avatarUrl": null,
    "schoolId": "660e8400-e29b-41d4-a716-446655440001",
    "subjectId": "770e8400-e29b-41d4-a716-446655440002",
    "schoolName": "SMAN 1 Solok",
    "subjectName": "Matematika"
  },
  "summary": {
    "totalAttempts": 3,
    "avgScore": 72.5,
    "avgPercentile": 68
  },
  "openTryouts": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "title": "Tryout Matematika SMA",
      "shortTitle": "Tryout MTK",
      "description": "Tryout persiapan OSN",
      "durationMinutes": 90,
      "questionsCount": 20,
      "level": "medium",
      "subjectId": "770e8400-e29b-41d4-a716-446655440002",
      "opensAt": "2025-02-01T08:00:00Z",
      "closesAt": "2025-02-28T22:00:00Z",
      "maxParticipants": null,
      "status": "open"
    }
  ],
  "recentAttempts": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440004",
      "userId": "550e8400-e29b-41d4-a716-446655440000",
      "tryoutSessionId": "880e8400-e29b-41d4-a716-446655440003",
      "startedAt": "2025-02-15T10:00:00Z",
      "submittedAt": "2025-02-15T11:25:00Z",
      "status": "submitted",
      "score": 75.5,
      "maxScore": 100,
      "percentile": 72,
      "timeSecondsSpent": 5100
    }
  ],
  "strengthAreas": [
    "Pilihan Ganda"
  ],
  "improvementAreas": [
    "Isian Singkat"
  ],
  "recommendation": "Skor keseluruhan: 75.5% dari total. Fokus perbaiki: Isian Singkat. Rekomendasi: perbanyak latihan soal tipe tersebut dan ulangi materi terkait.",
  "learningEvaluation": {
    "attemptId": "990e8400-e29b-41d4-a716-446655440004",
    "answerBreakdown": [
      {
        "questionId": "q-uuid-1",
        "questionType": "multiple_choice",
        "maxScore": 5,
        "scoreGot": 5,
        "status": "correct"
      },
      {
        "questionId": "q-uuid-2",
        "questionType": "short",
        "maxScore": 5,
        "scoreGot": 2.5,
        "status": "partial"
      },
      {
        "questionId": "q-uuid-3",
        "questionType": "true_false",
        "maxScore": 5,
        "scoreGot": 0,
        "status": "wrong"
      }
    ],
    "strengthAreas": [
      "Pilihan Ganda"
    ],
    "improvementAreas": [
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
| **openTryouts** | Daftar tryout yang sedang buka. |
| **recentAttempts** | Riwayat attempt terbaru (id, skor, status, waktu). |
| **strengthAreas** | Area kuat (dari analisis jawaban, mis. "Pilihan Ganda"). |
| **improvementAreas** | Bagian yang perlu ditingkatkan (mis. "Isian Singkat"). |
| **recommendation** | Rekomendasi belajar (teks) untuk siswa. |
| **learningEvaluation** | Detail penilaian per attempt terakhir yang sudah submit. |
| **learningEvaluation.attemptId** | ID attempt yang dianalisis. |
| **learningEvaluation.answerBreakdown** | Detail per soal: `questionId`, `questionType`, `maxScore`, `scoreGot`, `status` (`correct` / `partial` / `wrong` / `unanswered`). |
| **learningEvaluation.strengthAreas** | Area kuat dari attempt tersebut. |
| **learningEvaluation.improvementAreas** | Bagian yang perlu dikembangkan dari attempt tersebut. |
| **learningEvaluation.recommendation** | Rekomendasi belajar berdasarkan attempt tersebut. |

**Catatan:** `learningEvaluation` hanya ada jika siswa punya minimal satu attempt dengan status `submitted`. Jika belum pernah submit tryout, field ini tidak muncul (atau `null`).
