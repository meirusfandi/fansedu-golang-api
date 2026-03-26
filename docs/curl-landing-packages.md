# cURL — Packages (landing publik & admin)

Contoh siap pakai; sesuaikan `BASE_URL` dan token JWT admin.

---

## Publik (landing) — tanpa auth

Daftar paket program landing (JSON: `id`, `name`, `slug`, harga, `linked_courses`, dll.).

```bash
curl -sS -X GET "${BASE_URL:-http://localhost:8080}/api/v1/packages" \
  -H "Accept: application/json"
```

---

## Admin — butuh JWT + permission `landing.manage`

Ganti `YOUR_ADMIN_JWT` dengan token dari `POST /api/v1/auth/login` (user admin/super_admin yang punya izin landing).

```bash
AUTH="Authorization: Bearer YOUR_ADMIN_JWT"
BASE="${BASE_URL:-http://localhost:8080}/api/v1/admin/landing"
```

### List paket

```bash
curl -sS -X GET "$BASE/packages" \
  -H "Accept: application/json" \
  -H "$AUTH"
```

### Buat paket (minimal + `linked_course_ids` opsional)

```bash
curl -sS -X POST "$BASE/packages" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "$AUTH" \
  -d '{
    "name": "Paket Bundle A+B",
    "slug": "paket-bundle-a-b",
    "short_description": "Akses 2 kelas sekaligus",
    "price_early_bird": 499000,
    "price_normal": 799000,
    "is_open": true,
    "is_bundle": true,
    "durasi": "8 Minggu",
    "materi": ["Modul algoritma", "Latihan OSN"],
    "fasilitas": ["Live class", "Rekaman"],
    "bonus": [],
    "linked_course_ids": ["UUID-KELAS-1", "UUID-KELAS-2"]
  }'
```

### Update paket (ganti `PACKAGE_UUID`)

```bash
curl -sS -X PUT "$BASE/packages/PACKAGE_UUID" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "$AUTH" \
  -d '{
    "name": "Paket Bundle A+B (updated)",
    "slug": "paket-bundle-a-b",
    "price_early_bird": 449000,
    "price_normal": 749000,
    "linked_course_ids": ["UUID-KELAS-1", "UUID-KELAS-2"]
  }'
```

**Catatan:** `linked_course_ids` hanya diproses jika dikirim di body (update); untuk hanya mengubah field lain, omit key itu.

### Hapus paket

```bash
curl -sS -X DELETE "$BASE/packages/PACKAGE_UUID" \
  -H "Accept: application/json" \
  -H "$AUTH"
```

---

## Ringkasan path

| Tujuan        | Method | Path                                      |
|---------------|--------|-------------------------------------------|
| Publik        | GET    | `/api/v1/packages`                        |
| Admin list    | GET    | `/api/v1/admin/landing/packages`          |
| Admin create  | POST   | `/api/v1/admin/landing/packages`          |
| Admin update  | PUT    | `/api/v1/admin/landing/packages/{id}`     |
| Admin delete  | DELETE | `/api/v1/admin/landing/packages/{id}`     |

**Header auth:** `Authorization: Bearer <jwt>`.

**Alternatif tautan kelas dari sisi course:** `PUT /api/v1/admin/courses/{courseId}/linked-packages` dengan body `{"package_ids":["..."]}` (permission `courses.manage`).

---

## Setup cepat + `jq` (opsional)

```bash
BASE_URL="http://localhost:8080"
curl -sS "${BASE_URL}/api/v1/packages" | jq .
```

```bash
RES=$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"..."}')
TOKEN=$(echo "$RES" | jq -r '.token')
AUTH="Authorization: Bearer $TOKEN"
BASE="${BASE_URL}/api/v1/admin/landing"
curl -sS -X GET "$BASE/packages" -H "Accept: application/json" -H "$AUTH" | jq .
```
