# Fansedu LMS Backend (Go + PostgreSQL)

## Environment (dev vs prod)

Aplikasi memuat env dari file:

| Mode        | File yang dimuat |
|------------|-------------------|
| Development | `.env.dev`       |
| Production  | `.env`           |

- **Development:** Buat/copy `.env.dev`. Saat `ENV` tidak diset atau `ENV=development`, file `.env.dev` akan dimuat.
- **Production:** Set `ENV=production` di environment (Docker/host), lalu isi `.env` dengan nilai production. Aplikasi akan memuat `.env`.

| ENV            | DATABASE_URL | JWT_SECRET |
|----------------|--------------|------------|
| `development`  | Opsional     | Default boleh (warning) |
| `production`   | **Wajib**    | **Wajib**, harus kuat (bukan default) |

**Jangan commit `.env` atau `.env.dev`** — keduanya ada di `.gitignore`. Pakai `.env.development.example` sebagai template: salin ke `.env.dev`, lalu isi `JWT_SECRET` sendiri (mis. `openssl rand -base64 32`). Jika `.env.dev` pernah ikut ter-commit, untrack dengan: `git rm --cached .env.dev`.

## Run (local)

Gunakan **`./cmd/api`** (dengan `./`) agar Go menjalankan paket di folder proyek, bukan di GOROOT.

**Development / dev local (pakai `.env.dev`):**
```bash
cp .env.development.example .env.dev   # sekali saja
go run ./cmd/api
go run ./cmd/api -env=dev              # atau -env=development
```

**Production (pakai `.env`, API dari server):**
```bash
go run ./cmd/api -env=prod             # atau -env=production
# Atau set ENV=production dan isi .env, lalu jalankan (mis. di Docker/host).
```

Jika di `.env` Anda memakai hostname Docker (mis. `fansedu_fansedu-db`), hostname itu hanya bisa di-resolve **di dalam jaringan Docker**. Jadi:
- **API jalan di container (Docker):** tidak masalah, DB dan API satu network.
- **API jalan di Mac/laptop (`go run ./cmd/api -env=prod`):** override `DATABASE_URL` ke host yang bisa diakses dari mesin Anda, mis. DB di localhost (port forward) atau alamat server:
  ```bash
  DATABASE_URL="postgres://user:pass@localhost:5432/fansedu?sslmode=disable" go run ./cmd/api -env=prod
  ```

## Database & migrasi

Skema PostgreSQL:
- `001_init.sql` — users, tryout_sessions, questions, attempts, courses, course_enrollments, certificates, dll.
- `002_course_content_payments.sql` — course_contents (modul/quiz/test per kelas), payments.

Jalankan migrasi (pastikan `DATABASE_URL` sudah benar di `.env` atau `.env.dev`):

```bash
go run ./cmd/migrate
```

002 idempotent (aman dijalankan ulang). 001 hanya jalankan sekali untuk DB baru.

## Endpoints (base path: `/api/v1`)

**Health**
- `GET /api/v1/health`

**Auth**
- `POST /api/v1/auth/register` — Body: `{ "name", "email", "password" }` → `{ "user", "token" }`
- `POST /api/v1/auth/login` — Body: `{ "email", "password" }` → `{ "user", "token" }`
- `POST /api/v1/auth/logout` — Bearer required
- `POST /api/v1/auth/forgot-password` — Body: `{ "email" }` (stub)
- `POST /api/v1/auth/reset-password` — Body: `{ "token", "new_password" }` (stub)

**Tryouts (public/student)**
- `GET /api/v1/tryouts/open` — Daftar tryout yang buka
- `GET /api/v1/tryouts/{tryoutId}` — Detail tryout
- `POST /api/v1/tryouts/{tryoutId}/start` — Bearer required → `{ "attempt_id", "expires_at", "time_left_seconds" }`

**Attempts (Bearer required)**
- `GET /api/v1/attempts/{attemptId}/questions` — Soal untuk attempt (tanpa kunci jawaban)
- `PUT /api/v1/attempts/{attemptId}/answers/{questionId}` — Submit jawaban
- `POST /api/v1/attempts/{attemptId}/submit` — Akhiri attempt, hitung skor, feedback

**Student (Bearer required)**
- `GET /api/v1/student/dashboard` — Ringkasan, open tryouts, recent attempts, strength/improvement
- `GET /api/v1/student/attempts` — Riwayat attempt
- `GET /api/v1/student/attempts/{attemptId}` — Detail attempt
- `GET /api/v1/student/certificates` — Daftar sertifikat

**Courses**
- `GET /api/v1/courses` — Daftar kursus
- `POST /api/v1/courses/{courseId}/enroll` — Bearer required — Daftar kelas

**Levels (jenjang pendidikan: SD, SMP, SMA)**
- `GET /api/v1/levels` — Daftar jenjang
- `GET /api/v1/levels/{id}` — Detail jenjang beserta daftar bidang/mata pelajaran

**Admin (Bearer + role admin)**

- **Dashboard overview:** `GET /api/v1/admin/overview` — total_students, total_users, active_tryouts, total_courses, total_enrollments, avg_score, total_certificates
- **Manajemen user:**  
  - `GET /api/v1/admin/users` — Daftar user (query: `?role=student|admin`)  
  - `GET /api/v1/admin/users/{userId}` — Detail user  
  - `POST /api/v1/admin/users` — Tambah user (body: email, password, name, role, avatar_url)  
  - `PUT /api/v1/admin/users/{userId}` — Edit user (body: name, email, role, avatar_url, password opsional)
- **Manajemen kelas (courses):**  
  - `GET /api/v1/admin/courses` — Daftar kelas  
  - `GET /api/v1/admin/courses/{courseId}` — Detail kelas  
  - `POST /api/v1/admin/courses` — Buat kelas (body: title, description)  
  - `PUT /api/v1/admin/courses/{courseId}` — Edit kelas  
  - `GET /api/v1/admin/courses/{courseId}/enrollments` — Daftar enrollment  
  - `GET /api/v1/admin/courses/{courseId}/contents` — Daftar konten (modul/quiz/test)  
  - `POST /api/v1/admin/courses/{courseId}/contents` — Tambah konten (body: title, description, sort_order, type: module|quiz|test, content)  
  - `PUT /api/v1/admin/courses/{courseId}/contents/{contentId}` — Edit konten  
  - `DELETE /api/v1/admin/courses/{courseId}/contents/{contentId}` — Hapus konten
- **Payment (placeholder):**  
  - `GET /api/v1/admin/payments` — Daftar pembayaran (query: `?limit=50`)  
  - `POST /api/v1/admin/payments` — Catat pembayaran (body: user_id, amount_cents, currency, type, reference_id, description, status)
- **Report bulanan:**  
  - `GET /api/v1/admin/reports/monthly?year=2025&month=2` — new_enrollments, payments_count, total_revenue_cents
- **Tryout & soal (event/quiz):**  
  - `GET /api/v1/admin/tryouts` — Daftar semua tryout/quiz  
  - `POST /api/v1/admin/tryouts` — Buat tryout (body: title, short_title, description, duration_minutes, questions_count, level, opens_at, closes_at, max_participants, status)  
  - `PUT /api/v1/admin/tryouts/{tryoutId}` — Update tryout  
  - `DELETE /api/v1/admin/tryouts/{tryoutId}` — Hapus tryout  
  - `GET /api/v1/admin/tryouts/{tryoutId}/questions` — Daftar soal tryout/quiz  
  - `GET /api/v1/admin/tryouts/{tryoutId}/questions/{questionId}` — Detail satu soal  
  - `POST /api/v1/admin/tryouts/{tryoutId}/questions` — Tambah soal (body: sort_order, type, body, options, max_score; type: short | multiple_choice | true_false)  
  - `PUT /api/v1/admin/tryouts/{tryoutId}/questions/{questionId}` — Update soal  
  - `DELETE /api/v1/admin/tryouts/{tryoutId}/questions/{questionId}` — Hapus soal  
- `POST /api/v1/admin/certificates` — Terbitkan sertifikat

# fansedu-golang-api
