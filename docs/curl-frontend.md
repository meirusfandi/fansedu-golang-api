# cURL untuk Frontend — Copy-paste siap pakai

Set BASE dan TOKEN sekali, lalu jalankan per endpoint. Response sama dengan yang akan diterima frontend (fetch/axios).

```bash
BASE="http://localhost:8080/api/v1"
TOKEN="<ganti-dengan-jwt-dari-login-atau-register>"
```

---

## Public (tanpa auth)

```bash
curl -s "$BASE/roles"
curl -s "$BASE/schools"
curl -s "$BASE/health"
curl -s "$BASE/dashboard"
curl -s "$BASE/tryouts/open"
curl -s "$BASE/levels"
curl -s "$BASE/courses/"
curl -s "$BASE/courses/slug/<SLUG>"

# LMS: Katalog program (paginate, filter)
curl -s "$BASE/programs?page=1&limit=12&search=&category="
curl -s "$BASE/programs/<SLUG>"
```

---

## Checkout (tanpa login — frictionless)

Format LMS (frontend Vite): `programId` atau `programSlug`, `name`, `email`. Response: `checkoutId`, `orderId`, `total`, `program: { title, priceDisplay }`.

```bash
# 1. Initiate (LMS: programId atau programSlug + name + email)
curl -s -X POST "$BASE/checkout/initiate" \
  -H "Content-Type: application/json" \
  -d '{"programSlug":"my-course-slug","name":"User Name","email":"user@example.com"}'
# atau: "programId":"<UUID>"

# 2. Payment session (LMS: checkoutId + paymentMethod; optional promoCode)
curl -s -X POST "$BASE/checkout/payment-session" \
  -H "Content-Type: application/json" \
  -d '{"checkoutId":"<ORDER_ID>","paymentMethod":"bank_transfer"}'
# Response: paymentUrl, orderId, expiry?, virtualAccountNumber?, amount

# Legacy: course_slug + order_id masih didukung.
# Webhook (dipanggil gateway): POST $BASE/webhook/payment body: {"order_id":"<ORDER_ID>"}
```

Setelah webhook sukses, enrollment course dibuat otomatis; user bisa akses course.

---

## Auth

```bash
# Register — siswa (default)
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Budi Siswa","email":"budi@example.com","password":"rahasia123"}'

# Register — guru
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Pak Guru","email":"guru@example.com","password":"rahasia123","role":"guru"}'

# Login (simpan .token dari response untuk $TOKEN)
curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"guru@example.com","password":"rahasia123"}'

# Logout
curl -s -X POST "$BASE/auth/logout" \
  -H "Authorization: Bearer $TOKEN"

# Data user saat ini (LMS: untuk header & guard)
curl -s "$BASE/auth/me" -H "Authorization: Bearer $TOKEN"

# Register — guru (LMS)
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Pak Guru","email":"guru@example.com","password":"rahasia123","role":"guru"}'

# Ganti kata sandi
curl -s -X POST "$BASE/auth/change-password" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"current_password":"rahasia123","new_password":"rahasia456"}'
```

---

## Notifications

```bash
curl -s "$BASE/notifications" -H "Authorization: Bearer $TOKEN"
curl -s -X PATCH "$BASE/notifications/<NOTIFICATION_ID>/read" -H "Authorization: Bearer $TOKEN"
```

---

## Payments (user)

```bash
curl -s "$BASE/payments" -H "Authorization: Bearer $TOKEN"
curl -s -X POST "$BASE/payments" -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount":100000,"type":"course_purchase","reference_id":"<UUID>","proof_url":"https://..."}'
```

---

## Trainer (Guru) — semua pakai Authorization: Bearer

```bash
# GET profil (name, email, school)
curl -s "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN"

# PUT update profil — nama saja
curl -s -X PUT "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Nama Baru"}'

# PUT update profil — kaitkan sekolah (ganti UUID dengan id sekolah valid)
curl -s -X PUT "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"school_id":"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"}'

# PUT update profil — nama + sekolah sekaligus
curl -s -X PUT "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Nama Baru","school_id":"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"}'

# PUT lepas sekolah
curl -s -X PUT "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"school_id":""}'

# GET status (paid_slots, registered_students_count)
curl -s "$BASE/trainer/status" \
  -H "Authorization: Bearer $TOKEN"

# GET status + daftar siswa
curl -s "$BASE/trainer/status?students=1" \
  -H "Authorization: Bearer $TOKEN"

# POST bayar slot
curl -s -X POST "$BASE/trainer/pay" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"quantity":5}'

# POST daftarkan siswa
curl -s -X POST "$BASE/trainer/students" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Anak Siswa","email":"anak@example.com","password":"rahasia123"}'

# Daftar kelas saya (trainer)
curl -s "$BASE/trainer/courses" -H "Authorization: Bearer $TOKEN"

# Buat kelas (trainer)
curl -s -X POST "$BASE/trainer/courses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Kelas Baru","description":"Deskripsi"}'
```

---

## Guru (LMS — role `guru`; `instructor` tetap diterima sebagai alias saat fallback)

```bash
curl -s "$BASE/guru/courses" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/guru/students" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/guru/earnings" -H "Authorization: Bearer $TOKEN"
```

---

## Siswa (auth)

```bash
curl -s "$BASE/student/dashboard" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/courses" -H "Authorization: Bearer $TOKEN"
# Riwayat transaksi (order) — LMS
curl -s "$BASE/student/transactions" -H "Authorization: Bearer $TOKEN"
# Kelas berdasarkan subject siswa
curl -s "$BASE/student/courses/by-subject" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/payments" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/tryouts" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/attempts" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/attempts/<ATTEMPT_ID>" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/certificates" -H "Authorization: Bearer $TOKEN"
```

---

## Course chat & forum (user ter-enroll)

```bash
curl -s "$BASE/courses/<COURSE_ID>/messages" -H "Authorization: Bearer $TOKEN"
curl -s -X POST "$BASE/courses/<COURSE_ID>/messages" -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"message":"Pesan"}'
curl -s "$BASE/courses/<COURSE_ID>/discussions" -H "Authorization: Bearer $TOKEN"
curl -s -X POST "$BASE/courses/<COURSE_ID>/discussions" -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"title":"Judul","body":"Isi"}'
curl -s "$BASE/discussions/<DISCUSSION_ID>" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/discussions/<DISCUSSION_ID>/replies" -H "Authorization: Bearer $TOKEN"
curl -s -X POST "$BASE/discussions/<DISCUSSION_ID>/replies" -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" -d '{"body":"Balasan"}'
```

---

## Tryouts & Attempts (auth)

```bash
# Ganti <TRYOUT_ID>, <ATTEMPT_ID>, <QUESTION_ID> dengan ID asli
curl -s "$BASE/tryouts/<TRYOUT_ID>" 
curl -s "$BASE/tryouts/<TRYOUT_ID>/leaderboard"
curl -s -X POST "$BASE/tryouts/<TRYOUT_ID>/register" -H "Authorization: Bearer $TOKEN"
curl -s -X POST "$BASE/tryouts/<TRYOUT_ID>/start" -H "Authorization: Bearer $TOKEN"

curl -s "$BASE/attempts/<ATTEMPT_ID>/questions" -H "Authorization: Bearer $TOKEN"
curl -s -X PUT "$BASE/attempts/<ATTEMPT_ID>/answers/<QUESTION_ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"answer_text":"","selected_option":"A"}'
curl -s -X POST "$BASE/attempts/<ATTEMPT_ID>/submit" -H "Authorization: Bearer $TOKEN"
```

---

## Admin — Laporan per kelas (rekap skor tryout, kehadiran, progress)

**Butuh token admin.**

```bash
# Laporan satu kelas: rekap skor tryout, kehadiran, progress tiap siswa
curl -s "$BASE/admin/reports/courses/<COURSE_ID>" -H "Authorization: Bearer $TOKEN"
```

Response berisi:
- `course`: id, title, description kelas
- `generated_at`: waktu generate laporan
- `students`: array siswa di kelas, tiap item berisi:
  - `student_id`, `student_name`, `student_email`
  - `enrolled_at`, `enrollment_status`
  - `progress`: status (enrolled/in_progress/completed), `completed_at`
  - `tryout_scores`: array skor tryout (tryout_id, tryout_title, attempt_id, score, max_score, percentile, submitted_at)
  - `attendance`: `tryouts_participated` (jumlah tryout yang diselesaikan), `last_activity_at`

---

## Ringkasan untuk fetch/axios

| Aksi Frontend        | Method | URL                    | Body / Query                    |
|----------------------|--------|------------------------|----------------------------------|
| Register siswa       | POST   | `/api/v1/auth/register` | `{ name, email, password }`     |
| Register guru        | POST   | `/api/v1/auth/register` | `{ name, email, password, role: "guru" }` |
| Login                | POST   | `/api/v1/auth/login`   | `{ email, password }`           |
| Logout               | POST   | `/api/v1/auth/logout`  | — Header: `Authorization: Bearer <token>` |
| Get profil guru      | GET    | `/api/v1/trainer/profile` | — Header: Bearer                |
| Update profil guru   | PUT    | `/api/v1/trainer/profile` | `{ name? }` atau `{ school_id? }` |
| Status guru          | GET    | `/api/v1/trainer/status` | Optional: `?students=1`        |
| Bayar slot           | POST   | `/api/v1/trainer/pay`  | `{ quantity }`                  |
| Daftarkan siswa      | POST   | `/api/v1/trainer/students` | `{ name, email, password }`  |

Response JSON sama dengan yang diharapkan frontend (mis. `GET /trainer/profile` → `{ name, email, school? }`).
