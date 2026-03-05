# Curl — Leaderboard Tryout

## Request

```bash
curl -s -w "\nHTTP_STATUS:%{http_code}\n" \
  "http://localhost:8080/api/v1/tryouts/6b5517c4-7708-409c-9dd7-eaa74878a007/leaderboard" \
  -H "Content-Type: application/json"
```

Ganti UUID di atas dengan `tryoutId` yang dipakai di halaman detail tryout.

---

## Response (200 OK)

Body selalu **array JSON** (bisa kosong `[]`):

```json
[
  {
    "rank": 1,
    "user_id": "uuid-user-1",
    "user_name": "Nama Siswa Satu",
    "school_name": "SMAN 1 Solok",
    "best_score": 85.5,
    "best_time_seconds": 1200,
    "has_attempt": true
  },
  {
    "rank": 2,
    "user_id": "uuid-user-2",
    "user_name": "Nama Siswa Dua",
    "school_name": null,
    "best_score": null,
    "has_attempt": false
  }
]
```

### Field tiap item

| Field              | Tipe    | Keterangan                                      |
|--------------------|--------|--------------------------------------------------|
| `rank`             | number | Peringkat (1, 2, 3, ...)                        |
| `user_id`          | string | UUID user                                        |
| `user_name`        | string | Nama siswa                                       |
| `school_name`      | string \| null | Nama sekolah (null jika tidak ada)     |
| `best_score`       | number \| null | Nilai tertinggi (null jika belum submit) |
| `best_time_seconds`| number \| null | Waktu pengerjaan detik (opsional)       |
| `has_attempt`      | boolean| Apakah sudah pernah submit                       |

---

## Cek di frontend

1. **Response selalu array**  
   Jika tidak ada data, backend mengembalikan `[]`, bukan `null`. Gunakan `Array.isArray(data)` lalu `data.map(...)`.

2. **Parse JSON**  
   `const data = await response.json()` → `data` harus array.

3. **Tampilkan nilai**  
   - `best_score` bisa `null` → tampilkan `"-"` atau `"0"` jika null.  
   - `school_name` bisa `null` → tampilkan string kosong atau "-".

Contoh (JavaScript):

```js
const res = await fetch(`${API_BASE}/tryouts/${tryoutId}/leaderboard`, {
  headers: { 'Content-Type': 'application/json' },
});
const list = await res.json();
// list selalu array ([] atau [...])
const items = Array.isArray(list) ? list : [];
```

---

## Error

- **404** — Tryout tidak ditemukan (UUID salah atau tryout dihapus).
- **500** — Error server; cek log backend.
