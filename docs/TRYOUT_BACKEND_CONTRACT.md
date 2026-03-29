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

## Submit tryout (siswa)

`POST /api/v1/attempts/{attemptId}/submit` — setelah sukses, body mencakup:

- `review[]` — per soal (urut `sortOrder`): `questionId`, `sortOrder`, `questionType`, `questionBody`, jawaban siswa `answerText` / `selectedOption` (opsional), kunci pembahasan `correctOption` / `correctText` (jika diset di bank soal), `isCorrect`, `scoreGot`, `maxScore`, `analysisSummary` (teks singkat untuk siswa), modul (`moduleKey`, `moduleLabel`, `moduleId`, `moduleTitle`, `bidang`, `tags`).
- `moduleAnalysis[]` — agregat: `moduleKey`, `moduleLabel`, `questionCount`, `correctCount`, `wrongCount`, `unscoredCount` (selaras dengan `isCorrect` / skor per soal).
- `percentile` pada submit / attempt: **dihilangkan (omit)** jika belum bisa dihitung (mis. peserta submit &lt; 2). Bukan placeholder `0` saat skor final sudah ada. Jika ada ≥2 skor pada tryout yang sama, diisi 0–100 (rank persentil dalam kelompok peserta).

Penilaian otomatis (tryout: biner, tanpa setengah poin):

- **PG / benar–salah:** dengan `correct_option`: benar → `scoreGot = maxScore`, `isCorrect: true`; salah / kosong → `scoreGot = 0`, `isCorrect: false`. Tanpa `correct_option`: `scoreGot = 0`, `isCorrect: null` (belum dinilai otomatis) meski siswa memilih opsi.
- **Isian singkat:** dengan `correct_text`: sama biner (0 atau penuh, `isCorrect` true/false). Tanpa `correct_text`: `scoreGot = 0`, `isCorrect: null` (tidak ada skor parsial 50%).

Soal (GET lembar ujian, tanpa kunci):

- Field opsional untuk pengelompokan modul di FE: `moduleId`, `moduleTitle`, `bidang`, `tags` — tidak menyertakan `correctOption` / `correctText`.

Admin (penyusunan soal):

- `POST|PUT .../admin/tryouts/{tryoutId}/questions` — boleh mengirim `moduleId`, `moduleTitle`, `bidang`, `tags`, `correctOption`, `correctText`.

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
