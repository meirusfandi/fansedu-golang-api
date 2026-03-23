# Contoh cURL — Geo

Ganti `BASE` dengan URL API (mis. `https://api.fansedu.web.id/api/v1`).

```bash
BASE=https://api.fansedu.web.id/api/v1

# Provinsi
curl -sS "$BASE/geo/provinces" | jq .

# Kab/kota Aceh (province id 11)
curl -sS "$BASE/geo/regencies/11" | jq .
```

Tanpa header auth.
