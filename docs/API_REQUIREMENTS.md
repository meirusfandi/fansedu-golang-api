# API — kebutuhan fungsional (ringkas)

## Geo / wilayah

- **GET** `/api/v1/geo/provinces` — response JSON array, format emsifa (`id` + `name`).
- **GET** `/api/v1/geo/regencies/:provinceId` — sama, untuk satu provinsi.
- Backend harus bisa jalan **tanpa** PostgreSQL untuk route ini (opsi deploy minimal: API + Redis).
- Cache Redis di service layer; lihat `docs/GEO_REDIS_BACKEND.md`.

## Autentikasi

Endpoint geo **public** (tanpa JWT).

Lihat juga `docs/API_SPEC.md` dan `docs/openapi.yaml`.
