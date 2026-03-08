# cURL — API v1 (Backend terbaru)

Base URL: `http://localhost:8080/api/v1`  
Ganti `localhost:8080` jika server jalan di host/port lain.  
Untuk endpoint yang butuh auth, ganti `<TOKEN>` dengan JWT dari login/register.

---

## Health & Public

```bash
# Health check
curl -s http://localhost:8080/api/v1/health

# Dashboard umum (tanpa auth)
curl -s http://localhost:8080/api/v1/dashboard
```

---

## Auth

### Register (siswa atau guru)

```bash
# Daftar sebagai siswa (default / role kosong)
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Budi Siswa","email":"budi@example.com","password":"rahasia123"}'

# Daftar sebagai siswa (role eksplisit)
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Siti Siswa","email":"siti@example.com","password":"rahasia123","role":"student"}'

# Daftar sebagai guru
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Pak Guru","email":"guru@example.com","password":"rahasia123","role":"guru"}'
```

### Login

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"guru@example.com","password":"rahasia123"}'
```

Simpan nilai `token` dari response untuk dipakai di header `Authorization: Bearer <TOKEN>`.

### Logout (auth)

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <TOKEN>"
```

---

## Trainer (Guru) — butuh Auth + role guru

```bash
# GET profil guru (name, email, school)
# Data sekolah: objek school ada jika guru terhubung ke sekolah (school_id). Jika tidak, school tidak diisi → frontend bisa tampilkan "Info sekolah akan tampil setelah backend mengembalikan data dari GET /trainer/profile (termasuk objek school)."
curl -s -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/trainer/profile

# PUT update profil (name, optional school_id)
# school_id: UUID sekolah untuk mengaitkan guru; kirim "" atau null untuk melepas.
curl -s -X PUT http://localhost:8080/api/v1/trainer/profile \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Nama Baru Guru"}'

# Contoh kaitkan guru ke sekolah (ganti SCHOOL_UUID dengan id dari master data sekolah)
# curl -s -X PUT ... -d '{"school_id":"SCHOOL_UUID"}'

# GET status (paid_slots, registered_students_count)
curl -s -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/trainer/status

# GET status + daftar siswa (?students=1)
curl -s -H "Authorization: Bearer <TOKEN>" "http://localhost:8080/api/v1/trainer/status?students=1"

# POST bayar slot (naikkan paid_slots)
curl -s -X POST http://localhost:8080/api/v1/trainer/pay \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"quantity":5}'

# POST daftarkan siswa (hanya jika registered_students_count < paid_slots)
curl -s -X POST http://localhost:8080/api/v1/trainer/students \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Anak Siswa","email":"anak@example.com","password":"rahasia123"}'
```

---

## Tryouts

```bash
# List tryout open (tanpa auth)
curl -s http://localhost:8080/api/v1/tryouts/open

# Detail tryout
curl -s http://localhost:8080/api/v1/tryouts/<TRYOUT_ID>

# Leaderboard tryout
curl -s http://localhost:8080/api/v1/tryouts/<TRYOUT_ID>/leaderboard

# Daftar tryout (auth: siswa)
curl -s -X POST http://localhost:8080/api/v1/tryouts/<TRYOUT_ID>/register \
  -H "Authorization: Bearer <TOKEN>"

# Mulai attempt (auth: siswa)
curl -s -X POST http://localhost:8080/api/v1/tryouts/<TRYOUT_ID>/start \
  -H "Authorization: Bearer <TOKEN>"
```

---

## Attempts (auth: siswa)

```bash
# Soal attempt
curl -s -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/attempts/<ATTEMPT_ID>/questions

# Submit jawaban per soal
curl -s -X PUT "http://localhost:8080/api/v1/attempts/<ATTEMPT_ID>/answers/<QUESTION_ID>" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"answer_text":"...","selected_option":"A"}'

# Submit attempt (selesai)
curl -s -X POST http://localhost:8080/api/v1/attempts/<ATTEMPT_ID>/submit \
  -H "Authorization: Bearer <TOKEN>"
```

---

## Student (auth: siswa)

```bash
# Dashboard siswa
curl -s -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/student/dashboard

# List tryouts siswa
curl -s -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/student/tryouts

# List attempts siswa
curl -s -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/student/attempts

# Detail attempt
curl -s -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/student/attempts/<ATTEMPT_ID>

# Sertifikat
curl -s -H "Authorization: Bearer <TOKEN>" http://localhost:8080/api/v1/student/certificates
```

---

## Courses

```bash
# List kursus (tanpa auth)
curl -s http://localhost:8080/api/v1/courses/

# Enroll kursus (auth)
curl -s -X POST http://localhost:8080/api/v1/courses/<COURSE_ID>/enroll \
  -H "Authorization: Bearer <TOKEN>"
```

---

## Levels (public)

```bash
curl -s http://localhost:8080/api/v1/levels
curl -s http://localhost:8080/api/v1/levels/<ID>
```

---

## Contoh alur cepat

```bash
# 1. Register guru
RESP=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Pak Guru","email":"guru@example.com","password":"rahasia123","role":"guru"}')
echo "$RESP"

# 2. Ambil token (jika pakai jq)
TOKEN=$(echo "$RESP" | jq -r '.token')

# 3. Lihat profil guru
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/trainer/profile

# 4. Update nama
curl -s -X PUT http://localhost:8080/api/v1/trainer/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Guru Updated"}'

# 5. Bayar slot
curl -s -X POST http://localhost:8080/api/v1/trainer/pay \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"quantity":3}'

# 6. Daftarkan siswa
curl -s -X POST http://localhost:8080/api/v1/trainer/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Siswa Satu","email":"siswa1@example.com","password":"rahasia123"}'
```
