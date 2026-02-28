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

**Development (pakai `.env.dev`):**
```bash
cp .env.development.example .env.dev   # sekali saja
go run ./cmd/api
```

**Production (pakai `.env`):** Set `ENV=production` dan isi `.env`, lalu jalankan (mis. di Docker/host).

## Database & migrasi

Skema PostgreSQL ada di `internal/db/migrations/001_init.sql` (users, tryout_sessions, questions, attempts, attempt_answers, attempt_feedback, courses, course_enrollments, certificates, password_reset_tokens).

Jalankan migrasi sekali (pastikan `DATABASE_URL` sudah benar di `.env` atau `.env.dev`):

```bash
go run ./cmd/migrate
```

Jangan jalankan ulang setelah skema sudah ada (DDL tidak idempotent).

## Endpoints

**Health**
- `GET /v1/health`

**Auth**
- `POST /v1/auth/register` — Body: `{ "name", "email", "password" }` → `{ "user", "token" }`
- `POST /v1/auth/login` — Body: `{ "email", "password" }` → `{ "user", "token" }`
- `POST /v1/auth/logout` — Bearer required
- `POST /v1/auth/forgot-password` — Body: `{ "email" }` (stub)
- `POST /v1/auth/reset-password` — Body: `{ "token", "new_password" }` (stub)

**Tryouts (public/student)**
- `GET /v1/tryouts/open` — Daftar tryout yang buka
- `GET /v1/tryouts/{tryoutId}` — Detail tryout
- `POST /v1/tryouts/{tryoutId}/start` — Bearer required → `{ "attempt_id", "expires_at", "time_left_seconds" }`

**Attempts (Bearer required)**
- `GET /v1/attempts/{attemptId}/questions` — Soal untuk attempt (tanpa kunci jawaban)
- `PUT /v1/attempts/{attemptId}/answers/{questionId}` — Submit jawaban
- `POST /v1/attempts/{attemptId}/submit` — Akhiri attempt, hitung skor, feedback

**Student (Bearer required)**
- `GET /v1/student/dashboard` — Ringkasan, open tryouts, recent attempts, strength/improvement
- `GET /v1/student/attempts` — Riwayat attempt
- `GET /v1/student/attempts/{attemptId}` — Detail attempt
- `GET /v1/student/certificates` — Daftar sertifikat

**Courses**
- `GET /v1/courses` — Daftar kursus
- `POST /v1/courses/{courseId}/enroll` — Bearer required — Daftar kelas

**Admin (Bearer + role admin)**
- `GET /v1/admin/overview` — Statistik
- `POST /v1/admin/tryouts` — Buat tryout
- `PUT /v1/admin/tryouts/{tryoutId}` — Update tryout
- `DELETE /v1/admin/tryouts/{tryoutId}` — Hapus tryout
- `POST /v1/admin/tryouts/{tryoutId}/questions` — Tambah soal
- `PUT /v1/admin/questions/{questionId}` — Update soal
- `DELETE /v1/admin/questions/{questionId}` — Hapus soal
- `POST /v1/admin/courses` — Buat kursus
- `GET /v1/admin/courses/{courseId}/enrollments` — Daftar enrollment
- `POST /v1/admin/certificates` — Terbitkan sertifikat

# fansedu-golang-api
