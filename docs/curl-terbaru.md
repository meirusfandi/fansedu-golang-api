# cURL Terbaru — Semua Endpoint LMS & Landing

Base URL = `VITE_API_URL` di frontend (mis. `http://localhost:8080/api/v1`). Copy-paste perintah di bawah.

---

## Setup

```bash
BASE="http://localhost:8080/api/v1"
TOKEN=""   # isi setelah login untuk endpoint yang butuh Bearer
```

---

## Landing — Packages

```bash
# Daftar paket (section "Program yang Sedang Dibuka"). Response snake_case.
curl -s "$BASE/packages"
curl -s "$BASE/packages" | jq .
```

---

## Katalog & Detail Program

```bash
# Katalog (pagination, search, category)
curl -s "$BASE/programs?page=1&limit=12"
curl -s "$BASE/programs?page=1&limit=12&search=osn"
curl -s "$BASE/programs?page=1&limit=12&category=matematika"

# Detail program by slug — dari courses ATAU packages (fallback)
# Slug dari packages (landing):
curl -s "$BASE/programs/algorithm-programming-foundation"
curl -s "$BASE/programs/pelatihan-intensif-osn-k-2026"
curl -s "$BASE/programs/paket-hemat-foundation-osn"
# Slug dari courses (jika ada):
curl -s "$BASE/programs/<SLUG_COURSE>"
```

---

## Auth

```bash
# Register
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Budi Siswa","email":"budi@example.com","password":"rahasia123"}'

curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Pak Instruktur","email":"instruktur@example.com","password":"rahasia123","role":"instructor"}'

# Login (simpan token ke $TOKEN)
curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"rahasia123"}'

# Profil saat ini
curl -s "$BASE/auth/me" -H "Authorization: Bearer $TOKEN"

# Logout
curl -s -X POST "$BASE/auth/logout" -H "Authorization: Bearer $TOKEN"
```

---

## Checkout (tanpa login / guest)

### 1. POST /checkout/initiate

Body: `programSlug` **atau** `programId`, `name`, `email`. Response 201: `checkoutId`, `orderId`, `total`, `program: { title, priceDisplay }`.

```bash
BASE="http://localhost:8080/api/v1"

# Paket: Pelatihan Intensif OSN-K 2026
curl -s -X POST "$BASE/checkout/initiate" \
  -H "Content-Type: application/json" \
  -d '{"programSlug":"pelatihan-intensif-osn-k-2026","name":"Budi Siswa","email":"budi@example.com"}'

# Paket: Algorithm & Programming Foundation
curl -s -X POST "$BASE/checkout/initiate" \
  -H "Content-Type: application/json" \
  -d '{"programSlug":"algorithm-programming-foundation","name":"Budi Siswa","email":"budi@example.com"}'

# Paket: Paket Hemat (bundle)
curl -s -X POST "$BASE/checkout/initiate" \
  -H "Content-Type: application/json" \
  -d '{"programSlug":"paket-hemat-foundation-osn","name":"Budi Siswa","email":"budi@example.com"}'

# Pakai programId (UUID) jika punya
# curl -s -X POST "$BASE/checkout/initiate" \
#   -H "Content-Type: application/json" \
#   -d '{"programId":"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx","name":"Budi","email":"budi@example.com"}'
```

**Contoh response 201:**
```json
{
  "checkoutId": "uuid-order-id",
  "orderId": "uuid-order-id",
  "total": 349000,
  "program": { "title": "Pelatihan Intensif OSN-K 2026 Informatika", "priceDisplay": "Rp349.000" }
}
```

Simpan `checkoutId` (sama dengan `orderId`) untuk langkah payment-session.

---

### 2. POST /checkout/payment-session

Body: `checkoutId`, `paymentMethod`, `promoCode` (opsional). Response 200: `paymentUrl`, `orderId`, `expiry`, `virtualAccountNumber`, `amount`.

```bash
# Ganti <CHECKOUT_ID> dengan checkoutId dari response initiate
curl -s -X POST "$BASE/checkout/payment-session" \
  -H "Content-Type: application/json" \
  -d '{"checkoutId":"<CHECKOUT_ID>","paymentMethod":"bank_transfer","promoCode":""}'

# paymentMethod contoh: bank_transfer, virtual_account, ewallet
```

**Satu blok (initiate lalu payment-session):**
```bash
BASE="http://localhost:8080/api/v1"

# Initiate
RES=$(curl -s -X POST "$BASE/checkout/initiate" \
  -H "Content-Type: application/json" \
  -d '{"programSlug":"pelatihan-intensif-osn-k-2026","name":"Budi Siswa","email":"budi@example.com"}')
echo "$RES" | jq .
CHECKOUT_ID=$(echo "$RES" | jq -r '.checkoutId')

# Payment session (pakai checkoutId dari atas)
curl -s -X POST "$BASE/checkout/payment-session" \
  -H "Content-Type: application/json" \
  -d "{\"checkoutId\":\"$CHECKOUT_ID\",\"paymentMethod\":\"bank_transfer\",\"promoCode\":\"\"}" | jq .
```

---

## Student (Bearer required)

```bash
curl -s "$BASE/student/dashboard"       -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/profile"         -H "Authorization: Bearer $TOKEN"
curl -s -X PUT "$BASE/student/profile"   -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Nama Baru","email":"email@baru.com"}'
curl -s "$BASE/student/courses"         -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/transactions"    -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/certificates"    -H "Authorization: Bearer $TOKEN"
```

---

## Instructor (Bearer required)

```bash
curl -s "$BASE/instructor/courses"   -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/instructor/students"  -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/instructor/earnings"  -H "Authorization: Bearer $TOKEN"
```

---

## Admin — Tryout Analisis & Grafik (Bearer, role admin)

Analisis per tryout (grafik per soal) dan analisis AI per siswa. Ganti `<TRYOUT_ID>` dan `<ATTEMPT_ID>` dengan UUID dari tryout/attempt.

```bash
# Login admin dulu, simpan token
# RES=$(curl -s -X POST "$BASE/auth/login" -H "Content-Type: application/json" \
#   -d '{"email":"admin@example.com","password":"..."}')
# TOKEN=$(echo "$RES" | jq -r '.token')

# 1. Analisis & data grafik per tryout (per nomor soal: jawaban, benar/salah, distribusi pilihan A/B/C/D)
curl -s "$BASE/admin/tryouts/<TRYOUT_ID>/analysis" -H "Authorization: Bearer $TOKEN" | jq .

# 2. Daftar siswa yang submit tryout (untuk link ke analisis AI per siswa)
curl -s "$BASE/admin/tryouts/<TRYOUT_ID>/students" -H "Authorization: Bearer $TOKEN" | jq .

# 3. Analisis AI per siswa (per attempt) — summary, recap, strength/improvement, rekomendasi
curl -s "$BASE/admin/tryouts/<TRYOUT_ID>/attempts/<ATTEMPT_ID>/ai-analysis" -H "Authorization: Bearer $TOKEN" | jq .
```

**Contoh urutan (ambil tryout ID → students → attempt_id → ai-analysis):**
```bash
BASE="http://localhost:8080/api/v1"
TOKEN="<TOKEN_ADMIN>"

# Daftar tryout, ambil satu tryout_id
curl -s "$BASE/admin/tryouts" -H "Authorization: Bearer $TOKEN" | jq '.[0].id'

# Analisis grafik tryout
TRYOUT_ID="<uuid-tryout>"
curl -s "$BASE/admin/tryouts/$TRYOUT_ID/analysis" -H "Authorization: Bearer $TOKEN" | jq .

# Daftar siswa tryout, ambil attempt_id salah satu siswa
curl -s "$BASE/admin/tryouts/$TRYOUT_ID/students" -H "Authorization: Bearer $TOKEN" | jq '.[0].attempt_id'

# Analisis AI untuk attempt tersebut
ATTEMPT_ID="<uuid-attempt>"
curl -s "$BASE/admin/tryouts/$TRYOUT_ID/attempts/$ATTEMPT_ID/ai-analysis" -H "Authorization: Bearer $TOKEN" | jq .
```

---

## Satu blok lengkap (copy-paste)

```bash
BASE="http://localhost:8080/api/v1"

# Packages (landing)
curl -s "$BASE/packages" | jq .

# Katalog program
curl -s "$BASE/programs?page=1&limit=12" | jq .

# Detail program — slug dari packages
curl -s "$BASE/programs/pelatihan-intensif-osn-k-2026" | jq .

# Login & simpan token
RES=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"rahasia123"}')
TOKEN=$(echo "$RES" | jq -r '.token')
echo "TOKEN=$TOKEN"

# Profil & kursus
curl -s "$BASE/auth/me" -H "Authorization: Bearer $TOKEN" | jq .
curl -s "$BASE/student/courses" -H "Authorization: Bearer $TOKEN" | jq .
```

Tanpa `jq`: hapus `| jq .` dan `| jq -r '.token'`.

---

## Ringkasan URL

| Endpoint | Method | Auth |
|----------|--------|------|
| `/packages` | GET | - |
| `/programs?page=1&limit=12` | GET | - |
| `/programs/:slug` | GET | - |
| `/auth/register` | POST | - |
| `/auth/login` | POST | - |
| `/auth/me` | GET | Bearer |
| `/auth/logout` | POST | Bearer |
| `/checkout/initiate` | POST | - |
| `/checkout/payment-session` | POST | - |
| `/student/profile` | GET / PUT | Bearer |
| `/student/courses` | GET | Bearer |
| `/student/transactions` | GET | Bearer |
| `/student/certificates` | GET | Bearer |
| `/instructor/courses` | GET | Bearer |
| `/instructor/students` | GET | Bearer |
| `/instructor/earnings` | GET | Bearer |
| `/admin/tryouts/:tryoutId/analysis` | GET | Bearer (admin) |
| `/admin/tryouts/:tryoutId/students` | GET | Bearer (admin) |
| `/admin/tryouts/:tryoutId/attempts/:attemptId/ai-analysis` | GET | Bearer (admin) |

**Error response:** `{ "error": "kode", "message": "pesan" }` (400, 401, 404, 422).
