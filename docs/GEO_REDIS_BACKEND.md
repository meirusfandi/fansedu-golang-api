# Geo / Wilayah — cache Redis di backend (FansEdu API)

Backend menyajikan data provinsi & kabupaten/kota dengan **format JSON yang sama** seperti [emsifa/api-wilayah-indonesia](https://www.emsifa.com/api-wilayah-indonesia/api/provinces.json): array objek `{ "id": "...", "name": "..." }`.

## Endpoint

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/api/v1/geo/provinces` | Daftar provinsi |
| GET | `/api/v1/geo/regencies/{provinceId}` | Kab/kota untuk provinsi (contoh: `11` = Aceh) |

Tidak memerlukan autentikasi.

## Pola cache-aside (Redis)

1. `GET` Redis key.
2. Jika **ada** → kembalikan body JSON apa adanya (byte-for-byte dari upstream / cache).
3. Jika **kosong / miss** → `HTTP GET` ke upstream emsifa, `SET` Redis dengan TTL, lalu response.

Upstream default:

- Base: `https://www.emsifa.com/api-wilayah-indonesia/api` (override via `GEO_UPSTREAM_BASE_URL`)
- Provinsi: `{base}/provinces.json`
- Kab/kota: `{base}/regencies/{provinceId}.json`

## Key Redis & TTL

| Data | Key | TTL default |
|------|-----|-------------|
| Provinsi | `fansedu:geo:provinces:v1` | 30 hari (`GEO_CACHE_TTL_SECONDS`) |
| Kab/kota per provinsi | `fansedu:geo:regencies:v1:{provinceId}` | sama |

Versi `:v1` memungkinkan invalidasi massal dengan mengganti versi di kode jika format berubah.

## Environment

| Variable | Contoh | Keterangan |
|----------|--------|------------|
| `REDIS_URL` | `redis://localhost:6379/0` | Kosong = **tanpa cache** (selalu hit upstream). |
| `GEO_UPSTREAM_BASE_URL` | `https://www.emsifa.com/api-wilayah-indonesia/api` | Optional. |
| `GEO_CACHE_TTL_SECONDS` | `2592000` | 30 hari = 30×24×3600. |

## Pseudo-code

```text
function provinces():
  key = "fansedu:geo:provinces:v1"
  cached = redis.GET(key)
  if cached: return 200, cached
  body = http.GET(upstream + "/provinces.json")
  redis.SET(key, body, TTL)
  return 200, body

function regencies(provinceId):
  key = "fansedu:geo:regencies:v1:" + provinceId
  cached = redis.GET(key)
  if cached: return 200, cached
  body = http.GET(upstream + "/regencies/" + provinceId + ".json")
  if 404: return 404
  redis.SET(key, body, TTL)
  return 200, body
```

## Frontend

Setelah backend siap, FE bisa memakai `VITE_GEO_SOURCE=internal` dan memanggil `{VITE_API_URL}/api/v1/geo/...` (bukan emsifa langsung). Cache browser (localStorage) di FE tetap opsional untuk mengurangi request ke API.
