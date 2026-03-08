# API Admin Tryout Detail – Contoh cURL untuk Frontend

Base URL: `https://api.fansedu.web.id/api/v1` (atau dari env, mis. `http://localhost:8080/api/v1`)

Semua endpoint admin memakai header: `Authorization: Bearer <token_admin>`

---

## 1. Detail tryout (event info)

```bash
curl -s -X GET "https://api.fansedu.web.id/api/v1/admin/tryouts/{tryoutId}" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Ganti `{tryoutId}` dengan UUID tryout. Response: title, opens_at, closes_at, duration_minutes, questions_count, level, status, description, dll.

---

## 2. Daftar soal tryout

```bash
curl -s -X GET "https://api.fansedu.web.id/api/v1/admin/tryouts/{tryoutId}/questions" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Response: array of `{ id, tryout_session_id, sort_order, type, body, options, max_score, image_url?, image_urls? }`.

---

## 3. Leaderboard tryout (boleh tanpa auth)

```bash
curl -s -X GET "https://api.fansedu.web.id/api/v1/tryouts/{tryoutId}/leaderboard"
```

Response: array of `{ rank?, user_id?, user_name?, name?, school_name?, score?, best_score? }`. Field bisa sedikit beda; frontend bisa pakai `score`/`best_score` dan `user_name`/`name`, `school_name`.

---

## 4. Statistik per satu soal (baru)

```bash
curl -s -X GET "https://api.fansedu.web.id/api/v1/admin/tryouts/{tryoutId}/questions/{questionId}/stats" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Ganti `{tryoutId}` dan `{questionId}` dengan UUID.  
Response 200 contoh:

```json
{
  "participants_count": 50,
  "answered_count": 48,
  "correct_count": 32,
  "wrong_count": 16,
  "correct_percent": 66.67,
  "wrong_percent": 33.33
}
```

404 = tryout atau soal tidak ditemukan (tampilkan "–" di UI).

---

## 5. Statistik semua soal sekaligus (bulk, disarankan)

```bash
curl -s -X GET "https://api.fansedu.web.id/api/v1/admin/tryouts/{tryoutId}/questions/stats" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

Response 200 contoh:

```json
{
  "participants_count": 50,
  "questions": [
    {
      "question_id": "uuid-soal-1",
      "answered_count": 48,
      "correct_count": 32,
      "wrong_count": 16,
      "correct_percent": 66.67,
      "wrong_percent": 33.33
    },
    {
      "question_id": "uuid-soal-2",
      "answered_count": 48,
      "correct_count": 40,
      "wrong_count": 8,
      "correct_percent": 83.33,
      "wrong_percent": 16.67
    }
  ]
}
```

404 = tryout tidak ditemukan.

---

## Ringkasan endpoint (halaman admin tryout detail)

| Method | Path | Auth | Keterangan |
|--------|------|------|------------|
| GET | `/api/v1/admin/tryouts/:tryoutId` | Bearer token | Detail event |
| GET | `/api/v1/admin/tryouts/:tryoutId/questions` | Bearer token | Daftar soal |
| GET | `/api/v1/tryouts/:tryoutId/leaderboard` | Optional | Leaderboard |
| GET | `/api/v1/admin/tryouts/:tryoutId/questions/stats` | Bearer token | Statistik semua soal (bulk) |
| GET | `/api/v1/admin/tryouts/:tryoutId/questions/:questionId/stats` | Bearer token | Statistik satu soal |

Untuk development lokal, ganti host dengan `http://localhost:PORT/api/v1` dan gunakan token admin dari login.
