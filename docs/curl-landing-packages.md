# cURL — Landing & Packages

Base URL sama dengan `VITE_API_URL` di frontend. Copy-paste perintah berikut (set `BASE` sekali di awal).

---

## Setup

```bash
BASE="http://localhost:8080/api/v1"
TOKEN=""   # isi setelah login untuk endpoint yang butuh auth
```

---

## Landing — Packages (Program yang Sedang Dibuka)

**GET /packages** — Daftar paket untuk section landing. Tanpa auth. Response snake_case.

```bash
curl -s "$BASE/packages"
```

Dengan format JSON rapi (perlu `jq`):

```bash
curl -s "$BASE/packages" | jq .
```

Contoh response (array; kosong `[]` jika belum ada data):

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
    "cta_label": "Daftar",
    "wa_message_template": "string | null",
    "cta_url": "string | null",
    "is_open": true,
    "is_bundle": false,
    "bundle_subtitle": "string | null",
    "durasi": "string | null",
    "materi": ["Item 1", "Item 2"],
    "fasilitas": ["Fasilitas 1"],
    "bonus": ["Bonus 1"]
  }
]
```

---

## Endpoint LMS lain (ringkas)

```bash
# Katalog program (pagination)
curl -s "$BASE/programs?page=1&limit=12"
curl -s "$BASE/programs?page=1&limit=12&search=react"

# Detail program by slug
curl -s "$BASE/programs/<SLUG>"

# Checkout (guest)
curl -s -X POST "$BASE/checkout/initiate" \
  -H "Content-Type: application/json" \
  -d '{"programSlug":"<SLUG>","name":"Nama","email":"email@example.com"}'

curl -s -X POST "$BASE/checkout/payment-session" \
  -H "Content-Type: application/json" \
  -d '{"checkoutId":"<ORDER_ID>","paymentMethod":"bank_transfer"}'

# Auth
curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"rahasia123"}'

curl -s "$BASE/auth/me" -H "Authorization: Bearer $TOKEN"

# Student (butuh Bearer)
curl -s "$BASE/student/courses"      -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/transactions" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/profile"      -H "Authorization: Bearer $TOKEN"

# Instructor (butuh Bearer)
curl -s "$BASE/instructor/courses"   -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/instructor/students"  -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/instructor/earnings"  -H "Authorization: Bearer $TOKEN"
```

---

## Satu blok lengkap (Landing + Packages + Login + Profil)

```bash
BASE="http://localhost:8080/api/v1"

# 1. Landing — packages
curl -s "$BASE/packages" | jq .

# 2. Katalog program
curl -s "$BASE/programs?page=1&limit=12" | jq .

# 3. Login & simpan token
RES=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"budi@example.com","password":"rahasia123"}')
echo "$RES" | jq .
TOKEN=$(echo "$RES" | jq -r '.token')

# 4. Profil & kursus saya
curl -s "$BASE/auth/me" -H "Authorization: Bearer $TOKEN" | jq .
curl -s "$BASE/student/courses" -H "Authorization: Bearer $TOKEN" | jq .
```

`jq` opsional; tanpa `jq` hapus `| jq .` atau `| jq -r '.token'`.
