# Curl — Admin Roles (untuk Administrator)

Base URL: **http://localhost:8080/api/v1**  
Semua endpoint memakai pattern **/api/v1/**.

Semua endpoint butuh **Bearer token** (login sebagai admin dulu).

---

## 1. Login admin → dapat token

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"me.rusfandi@gmail.com","password":"mr@Condong1105"}'
```

Ambil nilai `token` dari response, lalu set variabel:

```bash
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## 2. List semua role

```bash
curl -s http://localhost:8080/api/v1/admin/roles \
  -H "Authorization: Bearer $TOKEN"
```

Response contoh:

```json
[
  {"id":"...","name":"Admin","slug":"admin","description":"Administrator sistem",...},
  {"id":"...","name":"Pembimbing","slug":"pembimbing",...},
  {"id":"...","name":"Siswa","slug":"siswa",...},
  {"id":"...","name":"Pengajar","slug":"pengajar",...}
]
```

---

## 3. Detail satu role (by ID)

Ganti `ROLE_UUID` dengan id role dari list.

```bash
curl -s http://localhost:8080/api/v1/admin/roles/ROLE_UUID \
  -H "Authorization: Bearer $TOKEN"
```

---

## 4. Tambah role baru

```bash
curl -s -X POST http://localhost:8080/api/v1/admin/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Moderator",
    "slug": "moderator",
    "description": "Moderator konten dan diskusi",
    "icon_url": "https://cdn.example.com/icons/moderator.svg"
  }'
```

- **name** (wajib): nama role  
- **slug** (opsional): untuk URL; kalau kosong diisi otomatis dari name  
- **description** (opsional)  
- **icon_url** (opsional): URL ikon

---

## 5. Update role

Ganti `ROLE_UUID` dengan id role yang mau diubah.

```bash
curl -s -X PUT http://localhost:8080/api/v1/admin/roles/ROLE_UUID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Moderator",
    "slug": "moderator",
    "description": "Deskripsi diperbarui",
    "icon_url": "https://cdn.example.com/icons/moderator-new.svg"
  }'
```

Hanya field yang dikirim yang di-update (partial update).

---

## 6. Hapus role

```bash
curl -s -X DELETE http://localhost:8080/api/v1/admin/roles/ROLE_UUID \
  -H "Authorization: Bearer $TOKEN"
```

Response: **204 No Content** jika sukses.

---

## Ringkasan endpoint

| Method | Path | Keterangan |
|--------|------|------------|
| GET | `/api/v1/admin/roles` | Daftar semua role |
| GET | `/api/v1/admin/roles/{id}` | Detail role |
| POST | `/api/v1/admin/roles` | Tambah role |
| PUT | `/api/v1/admin/roles/{id}` | Update role |
| DELETE | `/api/v1/admin/roles/{id}` | Hapus role |

Semua butuh header: `Authorization: Bearer <token>`.
