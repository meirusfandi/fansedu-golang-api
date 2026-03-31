# Curl — Dashboard (Request & Response)

Base URL: `http://localhost:8080/api/v1`

---

## 1. Dashboard Umum (tanpa auth)

**Method:** `GET`  
**Request body:** Tidak ada (GET tidak punya body)

### Request

```bash
curl -s -X GET "http://localhost:8080/api/v1/dashboard" \
  -H "Content-Type: application/json"
```

### Response (200 OK)

```json
{
  "siteName": "FansEdu LMS",
  "openTryouts": 2,
  "totalCourses": 5,
  "totalLevels": 3,
  "totalSubjects": 9,
  "totalSchools": 27,
  "totalStudents": 150
}
```

---

## 2. Dashboard Siswa (perlu login)

**Method:** `GET`  
**Request body:** Tidak ada (GET tidak punya body)  
**Header:** `Authorization: Bearer <token>` (wajib)

### Request

```bash
# Simpan token dulu (dari response login)
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -s -X GET "http://localhost:8080/api/v1/student/dashboard" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

Atau token langsung:

```bash
curl -s -X GET "http://localhost:8080/api/v1/student/dashboard" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Response (200 OK)

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
  "strengthAreas": ["Pilihan Ganda"],
  "improvementAreas": ["Isian Singkat"],
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
      }
    ],
    "strengthAreas": ["Pilihan Ganda"],
    "improvementAreas": ["Isian Singkat"],
    "recommendation": "Skor keseluruhan: 75.5% dari total. Fokus perbaiki: Isian Singkat. Rekomendasi: perbanyak latihan soal tipe tersebut dan ulangi materi terkait."
  }
}
```

### Response jika tidak ada token (401 Unauthorized)

```json
unauthorized
```

### Response jika token invalid/expired (401)

```
Unauthorized
```

---

## Ringkasan

| Endpoint                    | Method | Auth   | Body |
|----------------------------|--------|--------|------|
| `/api/v1/dashboard`        | GET    | Tidak  | -    |
| `/api/v1/student/dashboard`| GET    | Bearer | -    |

Kedua endpoint **tidak memakai request body** karena method-nya GET. Data dikirim hanya lewat URL dan header (termasuk `Authorization` untuk dashboard siswa).
