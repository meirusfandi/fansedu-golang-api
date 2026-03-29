# Kontrak API Tryout (backend)

## Format error (semua endpoint API v1)

Respons error memakai bentuk tunggal:

```json
{
  "error": {
    "code": "STABLE_CODE",
    "message": "Teks aman untuk pengguna (tanpa SQL, path, stack)."
  }
}
```

Kode di-normalisasi ke **UPPER_SNAKE**. Detail teknis hanya di log server. Set `EXPOSE_INTERNAL_ERRORS=true` hanya di lingkungan dev jika perlu melihat pesan asli pada 5xx.

## TryoutResponse

- `opensAt` dan `closesAt` selalu diserialisasi sebagai ISO-8601 (RFC3339) dari `time.Time`.
- Field wajib untuk jadwal: `opensAt`, `closesAt`, `status`, `durationMinutes`, dll.

## Siswa (Bearer + password guard)

| Method | Path | Catatan |
|--------|------|---------|
| GET | `/api/v1/student/tryouts/:tryoutId/status` | `canRegister`, `canStartExam`, `isRegistered`, `hasAttempted`, `opensAt`, `closesAt`, `tryoutStatus`, `startDisabledReason` (kode seperti `NOT_REGISTERED`, `BEFORE_OPENS_AT`, `ALREADY_SUBMITTED`). |
| POST | `/api/v1/student/tryouts/:tryoutId/register` | Idempoten (`200` / `201`). Setelah sukses, status & leaderboard konsisten untuk user terdaftar. |
| POST | `/api/v1/student/tryouts/:tryoutId/start` | Ditolak di server jika belum daftar, sebelum `opensAt`, setelah `closesAt`, atau sudah submit (satu attempt per user per tryout). |

## Leaderboard (publik / Bearer)

| Method | Path | Catatan |
|--------|------|---------|
| GET | `/api/v1/tryouts/:tryoutId/leaderboard` | Hanya peserta **terdaftar** (`tryout_registrations`). |
| GET | `/api/v1/tryouts/:tryoutId/leaderboard/top` | Redis ZSET difilter ke user terdaftar. |
| GET | `/api/v1/tryouts/:tryoutId/leaderboard/rank` | `200` + `{ "inLeaderboard": false }` jika user belum terdaftar; tidak mengembalikan rank/skor palsu. |

## Guru / Trainer (Bearer, role trainer/guru)

| Method | Path | Catatan |
|--------|------|---------|
| GET | `/api/v1/guru/tryouts/:tryoutId/paper` | Metadata tryout + array soal (`durationMinutes`, `questions`, …). |
| PUT | `/api/v1/guru/tryouts/:tryoutId/paper` | `501` + `FEATURE_NOT_IMPLEMENTED` (bulk update belum tersedia). |
| GET | `/api/v1/trainer/tryouts/:tryoutId/paper` | Sama seperti guru. |
| PUT | `/api/v1/trainer/tryouts/:tryoutId/paper` | Sama seperti guru. |
| GET | `/api/v1/guru/tryouts/:tryoutId/analysis` | Ringkasan analisis per tryout (admin service). |

Path trainer setara untuk `analysis`, `students`, `ai-analysis`.
