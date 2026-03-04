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
  "site_name": "FansEdu LMS",
  "open_tryouts": 2,
  "total_courses": 5,
  "total_levels": 3,
  "total_subjects": 9,
  "total_schools": 27,
  "total_students": 150
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
  "strength_areas": ["Aljabar", "Geometri"],
  "improvement_areas": ["Statistika"],
  "recommendation": "Fokus latihan statistika untuk tryout berikutnya."
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
