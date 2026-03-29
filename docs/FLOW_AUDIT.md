# Audit alur (referensi frontend ↔ backend)

- **Error handling:** Frontend disarankan memetakan `error.code` (UPPER_SNAKE) dan HTTP status; jangan menampilkan `message` mentah sebagai sumber kebenaran jika sudah ada copy tetap per status.
- **Tryout siswa:** Ikuti `docs/TRYOUT_BACKEND_CONTRACT.md` — status endpoint sebagai sumber kebenaran tombol daftar / mulai.
- **Guru:** `GET .../guru/tryouts/:id/paper` untuk lembar soal; `PUT` mengembalikan `501` sampai implementasi bulk ada.
- **Checkout:** `POST .../checkout/orders/:orderId/payment-proof` memvalidasi tipe/ukuran file; error memakai kode `FILE_TOO_LARGE`, `INVALID_FILE_TYPE`, dll.

Lihat juga `docs/API_REQUIREMENTS.md` dan `docs/API_SPEC.md`.
