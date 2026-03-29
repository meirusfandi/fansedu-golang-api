# API — kebutuhan fungsional (ringkas)

## Format error JSON

- Bentuk tunggal: `{ "error": { "code": "...", "message": "..." } }`.
- `code` stabil (UPPER_SNAKE); `message` aman untuk pengguna; detail teknis hanya di log server.
- Lihat `docs/TRYOUT_BACKEND_CONTRACT.md` untuk kontrak tryout terperinci.

## Geo / wilayah

- **GET** `/api/v1/geo/provinces` — response JSON array, format emsifa (`id` + `name`).
- **GET** `/api/v1/geo/regencies/:provinceId` — sama, untuk satu provinsi.
- Backend harus bisa jalan **tanpa** PostgreSQL untuk route ini (opsi deploy minimal: API + Redis).
- Cache Redis di service layer; lihat `docs/GEO_REDIS_BACKEND.md`.

## Autentikasi

Endpoint geo **public** (tanpa JWT).

Lihat juga `docs/API_SPEC.md` dan `docs/openapi.yaml`.
