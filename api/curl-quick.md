# Curl Cepat — Base URL: http://localhost:8080/api/v1

Semua endpoint memakai pattern **/api/v1/** (contoh: `/api/v1/health`, `/api/v1/admin/roles`).  
Setelah login, ganti `$TOKEN` dengan token dari response login (atau di bash: `export TOKEN="eyJ..."`).

---

## Tanpa auth

```bash
# Health
curl -s http://localhost:8080/api/v1/health

# Register
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"StrongP@ssw0rd"}'

# Login (student)
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"jane@example.com","password":"StrongP@ssw0rd"}'

# Login (admin)
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"me.rusfandi@gmail.com","password":"mr@Condong1105"}'

# Tryouts open
curl -s http://localhost:8080/api/v1/tryouts/open

# Tryout detail (ganti TRYYOUT_UUID)
curl -s http://localhost:8080/api/v1/tryouts/TRYYOUT_UUID

# Leaderboard tryout (ganti TRYYOUT_UUID dengan id tryout)
curl -s http://localhost:8080/api/v1/tryouts/TRYYOUT_UUID/leaderboard

# Daftar kursus
curl -s http://localhost:8080/api/v1/courses/
```

---

## Dengan Bearer token (ganti $TOKEN)

```bash
# Logout
curl -s -X POST http://localhost:8080/api/v1/auth/logout -H "Authorization: Bearer $TOKEN"

# Start tryout (ganti TRYYOUT_UUID)
curl -s -X POST http://localhost:8080/api/v1/tryouts/TRYYOUT_UUID/start -H "Authorization: Bearer $TOKEN"

# Soal attempt (ganti ATTEMPT_UUID)
curl -s http://localhost:8080/api/v1/attempts/ATTEMPT_UUID/questions -H "Authorization: Bearer $TOKEN"

# Submit jawaban (ganti ATTEMPT_UUID, QUESTION_UUID)
curl -s -X PUT "http://localhost:8080/api/v1/attempts/ATTEMPT_UUID/answers/QUESTION_UUID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"selected_option":"B"}'

# Submit attempt (ganti ATTEMPT_UUID)
curl -s -X POST http://localhost:8080/api/v1/attempts/ATTEMPT_UUID/submit -H "Authorization: Bearer $TOKEN"

# Dashboard siswa
curl -s http://localhost:8080/api/v1/student/dashboard -H "Authorization: Bearer $TOKEN"

# Riwayat attempt
curl -s http://localhost:8080/api/v1/student/attempts -H "Authorization: Bearer $TOKEN"

# Sertifikat
curl -s http://localhost:8080/api/v1/student/certificates -H "Authorization: Bearer $TOKEN"

# Enroll kursus (ganti COURSE_UUID)
curl -s -X POST "http://localhost:8080/api/v1/courses/COURSE_UUID/enroll" -H "Authorization: Bearer $TOKEN"
```

---

## Admin (token harus role admin)

```bash
# Overview
curl -s http://localhost:8080/api/v1/admin/overview -H "Authorization: Bearer $TOKEN"

# Buat tryout
curl -s -X POST http://localhost:8080/api/v1/admin/tryouts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Simulasi OSN","duration_minutes":90,"questions_count":25,"level":"medium","opens_at":"2025-06-01T00:00:00Z","closes_at":"2025-06-02T23:59:59Z","status":"open"}'

# Buat soal (ganti TRYYOUT_UUID)
curl -s -X POST "http://localhost:8080/api/v1/admin/tryouts/TRYYOUT_UUID/questions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"sort_order":1,"type":"multiple_choice","body":"Soal contoh?","options":["A","B","C","D"],"max_score":1}'

# Buat kursus
curl -s -X POST http://localhost:8080/api/v1/admin/courses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Pembinaan OSN","description":"Kelas OSN"}'

# Daftar enrollment per kursus (ganti COURSE_UUID)
curl -s "http://localhost:8080/api/v1/admin/courses/COURSE_UUID/enrollments" -H "Authorization: Bearer $TOKEN"

# Terbitkan sertifikat
curl -s -X POST http://localhost:8080/api/v1/admin/certificates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"USER_UUID","tryout_session_id":"TRYYOUT_UUID"}'

# --- ROLES (Administrator) ---
# Daftar semua role
curl -s http://localhost:8080/api/v1/admin/roles -H "Authorization: Bearer $TOKEN"

# Detail role by ID (ganti ROLE_UUID)
curl -s http://localhost:8080/api/v1/admin/roles/ROLE_UUID -H "Authorization: Bearer $TOKEN"

# Tambah role baru
curl -s -X POST http://localhost:8080/api/v1/admin/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Moderator","slug":"moderator","description":"Moderator konten","icon_url":"https://example.com/icon.png"}'

# Update role (ganti ROLE_UUID)
curl -s -X PUT http://localhost:8080/api/v1/admin/roles/ROLE_UUID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Moderator","slug":"moderator","description":"Deskripsi baru","icon_url":null}'

# Hapus role (ganti ROLE_UUID)
curl -s -X DELETE http://localhost:8080/api/v1/admin/roles/ROLE_UUID -H "Authorization: Bearer $TOKEN"

# --- LEVELS (Administrator - jenjang SD/SMP/SMA) ---
# Tambah jenjang
curl -s -X POST http://localhost:8080/api/v1/admin/levels \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"SMP","slug":"smp","description":"Sekolah Menengah Pertama","sort_order":2}'

# Detail jenjang (ganti LEVEL_ID)
curl -s http://localhost:8080/api/v1/admin/levels/LEVEL_ID -H "Authorization: Bearer $TOKEN"

# Jenjang + daftar bidang/mata pelajaran (ganti LEVEL_ID)
curl -s http://localhost:8080/api/v1/admin/levels/LEVEL_ID/subjects -H "Authorization: Bearer $TOKEN"

# Update jenjang (ganti LEVEL_ID)
curl -s -X PUT http://localhost:8080/api/v1/admin/levels/LEVEL_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"SMP","slug":"smp","description":"Sekolah Menengah Pertama (updated)","sort_order":2}'
```
