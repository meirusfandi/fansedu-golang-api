# Admin API: Tryout / Quiz & Soal (Questions)

Base path: **`/api/v1/admin`**. Semua endpoint memerlukan **Authorization: Bearer &lt;admin-jwt&gt;** dan role admin.

---

## Tryout (Event / Quiz)

| Method | Path | Deskripsi |
|--------|------|------------|
| GET | `/tryouts` | Daftar semua tryout/quiz |
| POST | `/tryouts` | Buat tryout baru |
| PUT | `/tryouts/{tryoutId}` | Update tryout |
| DELETE | `/tryouts/{tryoutId}` | Hapus tryout |

**Body POST/PUT tryout (contoh):**
```json
{
  "title": "Simulasi OSN Informatika 2025",
  "short_title": "OSN 2025-1",
  "description": "Latihan simulasi OSN.",
  "duration_minutes": 90,
  "questions_count": 25,
  "level": "medium",
  "opens_at": "2025-01-01T00:00:00Z",
  "closes_at": "2025-01-02T00:00:00Z",
  "max_participants": 200,
  "status": "open"
}
```

---

## Soal (Questions) — nested under tryout

Semua CRUD soal di bawah **`/tryouts/{tryoutId}/questions`**.

| Method | Path | Deskripsi |
|--------|------|------------|
| GET | `/tryouts/{tryoutId}/questions` | Daftar semua soal tryout/quiz |
| GET | `/tryouts/{tryoutId}/questions/{questionId}` | Detail satu soal |
| POST | `/tryouts/{tryoutId}/questions` | Tambah soal |
| PUT | `/tryouts/{tryoutId}/questions/{questionId}` | Update soal |
| DELETE | `/tryouts/{tryoutId}/questions/{questionId}` | Hapus soal |

**Tipe soal (`type`):**
- `short` — jawaban singkat (text)
- `multiple_choice` — pilihan ganda (options: array string)
- `true_false` — benar/salah (options: ["Benar", "Salah"] atau serupa)

**Body soal:**
- **`body`** — Isi soal. Boleh **plain text** atau **HTML** (rich text). Bisa berisi tag HTML dan `<img src="...">` untuk menyisipkan gambar di dalam teks.
- **`image_url`** — (opsional) Satu URL gambar utama untuk soal.
- **`image_urls`** — (opsional) Array URL gambar tambahan, contoh: `["https://...", "https://..."]`. Berguna untuk form yang mengizinkan banyak gambar.

**Body POST soal (contoh — teks + HTML + gambar):**
```json
{
  "sort_order": 1,
  "type": "multiple_choice",
  "body": "<p>Perhatikan gambar berikut.</p><p>Nilai <strong>x</strong> adalah ...</p>",
  "image_url": "https://cdn.example.com/soal1.png",
  "image_urls": ["https://cdn.example.com/soal1.png", "https://cdn.example.com/diagram.png"],
  "options": ["10", "12", "15", "20"],
  "max_score": 5
}
```

**Body PUT soal (partial, semua field opsional):**
```json
{
  "sort_order": 2,
  "type": "short",
  "body": "<p>Updated question with <em>HTML</em>.</p>",
  "image_url": "https://...",
  "image_urls": ["https://..."],
  "options": null,
  "max_score": 2
}
```

**Response satu soal (GET/POST/PUT):**
```json
{
  "id": "uuid",
  "tryout_session_id": "uuid",
  "sort_order": 1,
  "type": "multiple_choice",
  "body": "<p>Perhatikan gambar berikut.</p><p>Nilai <strong>x</strong> adalah ...</p>",
  "image_url": "https://cdn.example.com/soal1.png",
  "image_urls": ["https://cdn.example.com/soal1.png", "https://cdn.example.com/diagram.png"],
  "options": ["10", "12", "15", "20"],
  "max_score": 5
}
```

---

## Contoh URL lengkap

- List tryout: `GET /api/v1/admin/tryouts`
- List soal: `GET /api/v1/admin/tryouts/{tryoutId}/questions`
- Detail soal: `GET /api/v1/admin/tryouts/{tryoutId}/questions/{questionId}`
- Tambah soal: `POST /api/v1/admin/tryouts/{tryoutId}/questions`
- Update soal: `PUT /api/v1/admin/tryouts/{tryoutId}/questions/{questionId}`
- Hapus soal: `DELETE /api/v1/admin/tryouts/{tryoutId}/questions/{questionId}`
