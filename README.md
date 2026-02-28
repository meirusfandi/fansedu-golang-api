# Fansedu LMS Backend (Go + PostgreSQL)

## Environment (dev vs prod)

Aplikasi memuat env dari file:

| Mode        | File yang dimuat |
|------------|-------------------|
| Development | `.env.dev`       |
| Production  | `.env`           |

- **Development:** Buat/copy `.env.dev`. Saat `ENV` tidak diset atau `ENV=development`, file `.env.dev` akan dimuat.
- **Production:** Set `ENV=production` di environment (Docker/host), lalu isi `.env` dengan nilai production. Aplikasi akan memuat `.env`.

| ENV            | DATABASE_URL | JWT_SECRET |
|----------------|--------------|------------|
| `development`  | Opsional     | Default boleh (warning) |
| `production`   | **Wajib**    | **Wajib**, harus kuat (bukan default) |

`.env` ada di `.gitignore` (jangan commit secret production).

## Run (local)

**Development (pakai `.env.dev`):**
```bash
cp .env.development.example .env.dev   # sekali saja
go run ./cmd/api
```

**Production (pakai `.env`):** Set `ENV=production` dan isi `.env`, lalu jalankan (mis. di Docker/host).

## Endpoints (MVP)

- `GET /v1/health`
- `POST /v1/auth/register` (stub)
- `POST /v1/auth/login` (stub)

# fansedu-golang-api
