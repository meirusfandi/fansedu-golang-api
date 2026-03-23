# API spec — Geo

Base: `{origin}/api/v1`

## GET /geo/provinces

**Response 200** — `application/json`

Body: array of:

```json
{ "id": "11", "name": "ACEH" }
```

Sumber data: cache-aside dari upstream emsifa (lihat `GEO_REDIS_BACKEND.md`).

## GET /geo/regencies/{provinceId}

**Path**

- `provinceId` — string ID provinsi (contoh: `11`).

**Response 200** — array format sama seperti provinsi.

**404** — provinsi tidak ditemukan di upstream.

**502** — gagal mengambil upstream (bukan 404).
