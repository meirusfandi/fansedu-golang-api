# cURL untuk LMS Frontend

Dokumen ini berisi contoh cURL yang **persis** bisa dipakai frontend LMS (Vite). Set `VITE_API_URL` ke base URL yang sama.

---

## Setup

```bash
# Sesuaikan dengan .env frontend (VITE_API_URL)
BASE="http://localhost:8080/api/v1"

# Setelah login/register, simpan token ke TOKEN untuk request yang butuh auth
TOKEN=""
```

**Header yang dipakai:**
- `Content-Type: application/json` (untuk POST/PUT)
- `Authorization: Bearer <token>` (untuk endpoint yang butuh login)

---

## 1. Auth

### POST /auth/register

**Request body:** `name`, `email`, `password`, `role` (optional: `"student"` | `"instructor"`)

```bash
# Daftar sebagai siswa (default)
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Budi Siswa","email":"budi@example.com","password":"rahasia123"}'

# Daftar sebagai instructor
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Pak Instruktur","email":"instruktur@example.com","password":"rahasia123","role":"instructor"}'
```

**Response 201:**
```json
{
  "user": { "id": "uuid", "name": "string", "email": "string", "role": "student" | "instructor" },
  "token": "eyJhbGc..."
}
```

---

### POST /auth/login

```bash
curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"rahasia123"}'
```

**Response 200:**
```json
{
  "user": { "id": "uuid", "name": "string", "email": "string", "role": "student" | "instructor" },
  "token": "eyJhbGc..."
}
```

Simpan `response.token` ke variabel untuk request berikutnya:
```bash
TOKEN="<paste-token-di-sini>"
```

---

### GET /auth/me

Data user saat ini (untuk header/navbar & route guard). Wajib pakai Bearer token.

```bash
curl -s "$BASE/auth/me" \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200:**
```json
{
  "id": "uuid",
  "name": "string",
  "email": "string",
  "role": "student" | "instructor"
}
```

**Response 401:** `{ "error": "unauthorized", "message": "..." }`

---

### POST /auth/logout

```bash
curl -s -X POST "$BASE/auth/logout" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 2. Program (Katalog & Detail)

### GET /programs

Daftar program dengan filter & pagination. **Tanpa auth.**

**Query params:** `category` (optional), `search` (optional), `page` (default 1), `limit` (default 12)

```bash
# Semua program, halaman 1, 12 per halaman
curl -s "$BASE/programs?page=1&limit=12"

# Dengan search
curl -s "$BASE/programs?page=1&limit=12&search=react"

# Filter kategori (slug subject)
curl -s "$BASE/programs?page=1&limit=12&category=matematika"
```

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "string",
      "title": "string",
      "shortDescription": "string",
      "thumbnail": "string",
      "price": 249000,
      "priceDisplay": "Rp249.000",
      "instructor": { "id": "uuid", "name": "string", "avatar": "string" },
      "category": "string",
      "level": "beginner",
      "duration": "string",
      "rating": 4.9,
      "reviewCount": 128
    }
  ],
  "total": 100,
  "page": 1,
  "totalPages": 9
}
```

---

### GET /programs/:slug

Detail program by slug (untuk halaman detail). **Tanpa auth.**

```bash
# Ganti <SLUG> dengan slug program (mis. dari kartu di katalog)
curl -s "$BASE/programs/react-dasar"
```

**Response 200:**
```json
{
  "id": "uuid",
  "slug": "string",
  "title": "string",
  "shortDescription": "string",
  "description": "string",
  "thumbnail": "string",
  "price": 249000,
  "priceDisplay": "Rp249.000",
  "instructor": { "id": "uuid", "name": "string", "avatar": "string" },
  "category": "string",
  "level": "beginner",
  "duration": "string",
  "rating": 4.9,
  "reviewCount": 128,
  "modules": [
    {
      "id": "uuid",
      "title": "string",
      "lessons": [
        { "id": "uuid", "title": "string", "duration": "string" }
      ]
    }
  ],
  "reviews": []
}
```

**Response 404:** `{ "error": "not_found", "message": "program not found" }`

---

## 2b. Landing — Packages (Program yang Sedang Dibuka)

### GET /packages

Daftar paket untuk section landing. Response **snake_case** (frontend bisa map ke camelCase).

```bash
curl -s "$BASE/packages"
```

**Response 200:** array of objects, contoh:

```json
[
  {
    "id": "uuid",
    "name": "string",
    "slug": "string",
    "short_description": "string | null",
    "price_display": "string | null",
    "price_early_bird": "string | null",
    "price_normal": "string | null",
    "cta_label": "string",
    "wa_message_template": "string | null",
    "cta_url": "string | null",
    "is_open": true,
    "is_bundle": false,
    "bundle_subtitle": "string | null",
    "durasi": "string | null",
    "materi": ["string"],
    "fasilitas": ["string"],
    "bonus": ["string"]
  }
]
```

Jika belum ada data, backend mengembalikan `[]`. Schema tabel: `database/landing_schema.sql`.

---

## 3. Checkout & Payment

### POST /checkout/initiate

Buat order (guest atau user sudah login). **Auth optional.**

**Request:** `programId` **atau** `programSlug`, plus `name`, `email`.

```bash
# Pakai slug (dari halaman detail program)
curl -s -X POST "$BASE/checkout/initiate" \
  -H "Content-Type: application/json" \
  -d '{
    "programSlug": "react-dasar",
    "name": "Budi Siswa",
    "email": "budi@example.com"
  }'

# Atau pakai programId (UUID)
curl -s -X POST "$BASE/checkout/initiate" \
  -H "Content-Type: application/json" \
  -d '{
    "programId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "name": "Budi Siswa",
    "email": "budi@example.com"
  }'
```

**Response 201:**
```json
{
  "checkoutId": "uuid-order-id",
  "orderId": "uuid-order-id",
  "total": 249000,
  "program": {
    "title": "React Dasar",
    "priceDisplay": "Rp249.000"
  }
}
```

Simpan `checkoutId` (sama dengan `orderId`) untuk langkah payment-session.

---

### POST /checkout/payment-session

Buat sesi pembayaran (redirect ke gateway atau dapat VA). Pakai `checkoutId` dari response initiate.

**Request:** `checkoutId`, `paymentMethod`, `promoCode` (optional)

```bash
curl -s -X POST "$BASE/checkout/payment-session" \
  -H "Content-Type: application/json" \
  -d '{
    "checkoutId": "<ORDER_ID_DARI_INITIATE>",
    "paymentMethod": "bank_transfer",
    "promoCode": ""
  }'
```

**paymentMethod** contoh: `bank_transfer`, `virtual_account`, `ewallet` (sesuai gateway).

**Response 200:**
```json
{
  "paymentUrl": "https://...",
  "orderId": "uuid",
  "expiry": "",
  "virtualAccountNumber": "",
  "amount": 249000
}
```

Frontend: redirect user ke `paymentUrl`, atau tampilkan `virtualAccountNumber` + `amount` untuk transfer manual.

---

## 4. Student Dashboard

Semua endpoint di bawah **butuh auth** (Header: `Authorization: Bearer $TOKEN`). Role: **student**.

### GET /student/courses

Daftar program yang sudah di-enroll.

```bash
curl -s "$BASE/student/courses" \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200:**
```json
{
  "data": [
    {
      "id": "enrollment-uuid",
      "program": {
        "id": "uuid",
        "slug": "string",
        "title": "string",
        "thumbnail": "string"
      },
      "progressPercent": 0,
      "enrolledAt": "2026-03-08T10:00:00Z",
      "lastAccessedAt": "2026-03-08T10:00:00Z"
    }
  ]
}
```

---

### GET /student/transactions

Riwayat transaksi (order) user.

```bash
curl -s "$BASE/student/transactions" \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "orderId": "uuid",
      "status": "paid",
      "total": 249000,
      "programs": [{ "title": "string" }],
      "paidAt": "2026-03-08T10:00:00Z"
    }
  ]
}
```

---

### GET /student/profile

Profil siswa (response sama dengan GET /auth/me: id, name, email, role).

```bash
curl -s "$BASE/student/profile" -H "Authorization: Bearer $TOKEN"
```

**Response 200:** `{ "id": "uuid", "name": "string", "email": "string", "role": "student" }`

---

### PUT /student/profile

Update profil (name, email). Hanya field yang dikirim yang di-update.

```bash
curl -s -X PUT "$BASE/student/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Nama Baru","email":"email@baru.com"}'
```

**Response 200:** sama seperti GET /student/profile (data terbaru).

---

### GET /student/certificates

Daftar sertifikat user.

```bash
curl -s "$BASE/student/certificates" -H "Authorization: Bearer $TOKEN"
```

---

## 5. Instructor Dashboard

Semua endpoint di bawah **butuh auth**. Role: **instructor** atau **guru**.

### GET /instructor/courses

Daftar program yang diajar.

```bash
curl -s "$BASE/instructor/courses" \
  -H "Authorization: Bearer $TOKEN"
```

---

### GET /instructor/students

Daftar siswa (enrollment) per program yang diajar.

```bash
curl -s "$BASE/instructor/students" \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200:**
```json
{
  "data": [
    {
      "userId": "uuid",
      "name": "string",
      "email": "string",
      "programTitle": "string",
      "progressPercent": 0
    }
  ]
}
```

---

### GET /instructor/earnings

Ringkasan pendapatan per periode (saat ini stub, bisa kembangkan nanti).

```bash
curl -s "$BASE/instructor/earnings" \
  -H "Authorization: Bearer $TOKEN"
```

**Response 200:**
```json
{
  "data": [
    { "period": "2026-03", "revenue": 0, "newStudents": 0 }
  ]
}
```

---

## 6. Error Response

Semua error API mengembalikan body konsisten:

```json
{
  "error": "kode_error",
  "message": "Pesan untuk user"
}
```

**Status code umum:**
- `400` — Bad Request (validasi)
- `401` — Unauthorized (belum login / token invalid)
- `403` — Forbidden (role tidak boleh akses)
- `404` — Not Found
- `422` — Validation Error

---

## Ringkasan Flow UI → cURL

| Langkah di UI              | cURL / Endpoint                    |
|----------------------------|------------------------------------|
| Landing — section Program  | `GET /packages`                    |
| Landing / Katalog          | `GET /programs?page=1&limit=12`    |
| Klik kartu program        | `GET /programs/:slug`             |
| Klik "Daftar" / Checkout   | Navigate ke checkout, isi nama & email |
| Submit form checkout       | `POST /checkout/initiate`          |
| Pilih metode bayar, Bayar  | `POST /checkout/payment-session`   |
| Redirect ke gateway        | Buka `paymentUrl` dari response    |
| Setelah bayar, success     | Redirect ke success page          |
| Dashboard siswa            | `GET /student/courses` (perlu login) |
| Login                      | `POST /auth/login` → simpan token |
| Header / guard             | `GET /auth/me`                    |

---

## Contoh lengkap (bash)

```bash
BASE="http://localhost:8080/api/v1"

# 1. Lihat katalog
curl -s "$BASE/programs?page=1&limit=12" | jq .

# 2. Detail satu program (ganti slug)
curl -s "$BASE/programs/react-dasar" | jq .

# 3. Initiate checkout
RES=$(curl -s -X POST "$BASE/checkout/initiate" \
  -H "Content-Type: application/json" \
  -d '{"programSlug":"react-dasar","name":"Budi","email":"budi@example.com"}')
echo "$RES" | jq .
CHECKOUT_ID=$(echo "$RES" | jq -r '.checkoutId')

# 4. Payment session
curl -s -X POST "$BASE/checkout/payment-session" \
  -H "Content-Type: application/json" \
  -d "{\"checkoutId\":\"$CHECKOUT_ID\",\"paymentMethod\":\"bank_transfer\"}" | jq .

# 5. Login
LOGIN_RES=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"rahasia123"}')
TOKEN=$(echo "$LOGIN_RES" | jq -r '.token')

# 6. Profil & kursus saya
curl -s "$BASE/auth/me" -H "Authorization: Bearer $TOKEN" | jq .
curl -s "$BASE/student/courses" -H "Authorization: Bearer $TOKEN" | jq .
curl -s "$BASE/student/transactions" -H "Authorization: Bearer $TOKEN" | jq .
```

`jq` optional (untuk format JSON). Tanpa `jq`, hapus `| jq .` atau ganti dengan `| jq` saja.
