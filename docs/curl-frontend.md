# cURL untuk Frontend — Copy-paste siap pakai

Set BASE dan TOKEN sekali, lalu jalankan per endpoint. Response sama dengan yang akan diterima frontend (fetch/axios).

```bash
BASE="http://localhost:8080/api/v1"
TOKEN="<ganti-dengan-jwt-dari-login-atau-register>"
```

---

## Auth

```bash
# Register — siswa (default)
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Budi Siswa","email":"budi@example.com","password":"rahasia123"}'

# Register — guru
curl -s -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Pak Guru","email":"guru@example.com","password":"rahasia123","role":"guru"}'

# Login (simpan .token dari response untuk $TOKEN)
curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"guru@example.com","password":"rahasia123"}'

# Logout
curl -s -X POST "$BASE/auth/logout" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Trainer (Guru) — semua pakai Authorization: Bearer

```bash
# GET profil (name, email, school)
curl -s "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN"

# PUT update profil — nama saja
curl -s -X PUT "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Nama Baru"}'

# PUT update profil — kaitkan sekolah (ganti UUID dengan id sekolah valid)
curl -s -X PUT "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"school_id":"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"}'

# PUT update profil — nama + sekolah sekaligus
curl -s -X PUT "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Nama Baru","school_id":"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"}'

# PUT lepas sekolah
curl -s -X PUT "$BASE/trainer/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"school_id":""}'

# GET status (paid_slots, registered_students_count)
curl -s "$BASE/trainer/status" \
  -H "Authorization: Bearer $TOKEN"

# GET status + daftar siswa
curl -s "$BASE/trainer/status?students=1" \
  -H "Authorization: Bearer $TOKEN"

# POST bayar slot
curl -s -X POST "$BASE/trainer/pay" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"quantity":5}'

# POST daftarkan siswa
curl -s -X POST "$BASE/trainer/students" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Anak Siswa","email":"anak@example.com","password":"rahasia123"}'
```

---

## Public (tanpa auth)

```bash
curl -s "$BASE/health"
curl -s "$BASE/dashboard"
curl -s "$BASE/tryouts/open"
curl -s "$BASE/levels"
```

---

## Siswa (auth)

```bash
curl -s "$BASE/student/dashboard" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/tryouts" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/attempts" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/attempts/<ATTEMPT_ID>" -H "Authorization: Bearer $TOKEN"
curl -s "$BASE/student/certificates" -H "Authorization: Bearer $TOKEN"
```

---

## Tryouts & Attempts (auth)

```bash
# Ganti <TRYOUT_ID>, <ATTEMPT_ID>, <QUESTION_ID> dengan ID asli
curl -s "$BASE/tryouts/<TRYOUT_ID>" 
curl -s "$BASE/tryouts/<TRYOUT_ID>/leaderboard"
curl -s -X POST "$BASE/tryouts/<TRYOUT_ID>/register" -H "Authorization: Bearer $TOKEN"
curl -s -X POST "$BASE/tryouts/<TRYOUT_ID>/start" -H "Authorization: Bearer $TOKEN"

curl -s "$BASE/attempts/<ATTEMPT_ID>/questions" -H "Authorization: Bearer $TOKEN"
curl -s -X PUT "$BASE/attempts/<ATTEMPT_ID>/answers/<QUESTION_ID>" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"answer_text":"","selected_option":"A"}'
curl -s -X POST "$BASE/attempts/<ATTEMPT_ID>/submit" -H "Authorization: Bearer $TOKEN"
```

---

## Ringkasan untuk fetch/axios

| Aksi Frontend        | Method | URL                    | Body / Query                    |
|----------------------|--------|------------------------|----------------------------------|
| Register siswa       | POST   | `/api/v1/auth/register` | `{ name, email, password }`     |
| Register guru        | POST   | `/api/v1/auth/register` | `{ name, email, password, role: "guru" }` |
| Login                | POST   | `/api/v1/auth/login`   | `{ email, password }`           |
| Logout               | POST   | `/api/v1/auth/logout`  | — Header: `Authorization: Bearer <token>` |
| Get profil guru      | GET    | `/api/v1/trainer/profile` | — Header: Bearer                |
| Update profil guru   | PUT    | `/api/v1/trainer/profile` | `{ name? }` atau `{ school_id? }` |
| Status guru          | GET    | `/api/v1/trainer/status` | Optional: `?students=1`        |
| Bayar slot           | POST   | `/api/v1/trainer/pay`  | `{ quantity }`                  |
| Daftarkan siswa      | POST   | `/api/v1/trainer/students` | `{ name, email, password }`  |

Response JSON sama dengan yang diharapkan frontend (mis. `GET /trainer/profile` → `{ name, email, school? }`).
