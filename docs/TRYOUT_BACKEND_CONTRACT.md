# Kontrak API Tryout (backend)

## Konvensi JSON (request & response)

Semua kunci properti JSON untuk **tryout, soal, attempt, jawaban, leaderboard terkait tryout** memakai **camelCase** (contoh: `tryoutSessionId`, `sortOrder`, `maxScore`, `opensAt`, `answerText`, `selectedOption`).

- **Respons API** selalu camelCase (struct `encoding/json` di server).
- **Permintaan (body)** disarankan camelCase. Untuk **admin** (tryout & soal), server tetap bisa menerima **snake_case** sebagai alias input (mis. `sort_order`, `max_score`) agar kompatibel dengan klien lama — ini tidak mengubah bentuk **respons**.

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

## CORS (browser / Chrome)

`Access-Control-Allow-Headers` memakai daftar eksplisit yang mencakup **`Authorization`** (wildcard `*` tidak lagi dianggap mencakup header ini pada preflight di browser modern). Origin diatur lewat env **`CORS_ORIGINS`** (koma); jika perlu kredensial cookie, jangan memakai `*` untuk origin — daftar domain frontend secara eksplisit.

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

- `review[]` — per soal (urut `sortOrder`): `questionId`, `sortOrder`, `questionType`, `questionTypeLabel` (mis. "Pilihan ganda"), `questionBody`, jawaban siswa `answerText` / `selectedOption`, kunci `correctOption` / `correctText` (jika ada), `isCorrect`, `scoreGot`, `maxScore`, `analysisSummary` (satu kalimat), `analysisDetail` (narasi per soal sesuai tipe: pilihan vs isian, benar/salah, skor), modul (`moduleKey`, `moduleLabel`, …).
- `overallAnalysis` — ringkasan tryout: `totalQuestions`, `answeredCount`, `unansweredCount`, `correctCount`, `wrongCount`, `unscoredCount`, `scorePercent`, `scoreGot`, `maxScore`, `byQuestionType[]` (`type`, `label`, `total`, `correct`, `wrong`, `unscored`, `scoreGot`, `maxScore` per jenis soal), `summary` (paragraf Bahasa Indonesia).
- `moduleAnalysis[]` — agregat: `moduleKey`, `moduleLabel`, `questionCount`, `correctCount`, `wrongCount`, `unscoredCount` (selaras dengan `isCorrect` / skor per soal).
- `percentile` pada submit / attempt: **dihilangkan (omit)** jika belum bisa dihitung (mis. peserta submit &lt; 2). Bukan placeholder `0` saat skor final sudah ada. Jika ada ≥2 skor pada tryout yang sama, diisi 0–100 (rank persentil dalam kelompok peserta).

Penilaian otomatis (tryout: biner, tanpa setengah poin):

- **PG / benar–salah:** kunci diambil dari kolom `correct_option` **atau** dari entri opsi di JSON `options` yang memiliki `correct` / `isCorrect` / `is_correct` / `isTrue` (dan field kunci `key`, `value`, `id`, `option`, atau `label`). Dengan kunci: benar → penuh; salah / kosong → 0 + `isCorrect: false`. Tanpa kunci sama sekali: `scoreGot = 0`, `isCorrect: null`.
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
