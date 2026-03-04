# Curl — Admin Levels (Jenjang: SD, SMP, SMA)

Base URL: **http://localhost:8080/api/v1**  
Semua endpoint butuh **Bearer token** (login sebagai admin dulu).

```bash
export TOKEN="<token_dari_login_admin>"
```

Ganti `LEVEL_ID` dengan UUID level dari response list/detail.

---

## POST /api/v1/admin/levels — Tambah jenjang

```bash
curl -s -X POST http://localhost:8080/api/v1/admin/levels \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SMP",
    "slug": "smp",
    "description": "Sekolah Menengah Pertama",
    "sort_order": 2,
    "icon_url": "https://example.com/icons/smp.svg"
  }'
```

Body (semua kecuali name bisa opsional):
- `name` (wajib)
- `slug` (opsional, auto dari name bila kosong)
- `description` (opsional)
- `sort_order` (opsional, default 0)
- `icon_url` (opsional)

---

## GET /api/v1/admin/levels/{id} — Detail satu jenjang

```bash
curl -s http://localhost:8080/api/v1/admin/levels/LEVEL_ID \
  -H "Authorization: Bearer $TOKEN"
```

Response: object satu level (id, name, slug, description, sort_order, icon_url, created_at, updated_at).

---

## GET /api/v1/admin/levels/{id}/subjects — Jenjang + daftar bidang/mata pelajaran

```bash
curl -s http://localhost:8080/api/v1/admin/levels/LEVEL_ID/subjects \
  -H "Authorization: Bearer $TOKEN"
```

Response: level object + array `subjects` (id, name, slug, description, icon_url, sort_order, ...).

---

## PUT /api/v1/admin/levels/{id} — Update jenjang

```bash
curl -s -X PUT http://localhost:8080/api/v1/admin/levels/LEVEL_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SMP",
    "slug": "smp",
    "description": "Sekolah Menengah Pertama (diperbarui)",
    "sort_order": 2,
    "icon_url": "https://example.com/icons/smp-new.svg"
  }'
```

Hanya field yang dikirim yang di-update (partial update).

---

## Ringkasan

| Method | Path | Keterangan |
|--------|------|-------------|
| POST   | `/api/v1/admin/levels`           | Tambah jenjang |
| GET    | `/api/v1/admin/levels/{id}`      | Detail jenjang |
| GET    | `/api/v1/admin/levels/{id}/subjects` | Jenjang + daftar bidang |
| PUT    | `/api/v1/admin/levels/{id}`      | Update jenjang |

Semua butuh header: `Authorization: Bearer <token>`.
