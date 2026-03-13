-- Keuntungan khusus: semua peserta yang sudah terdaftar dianggap sudah terverifikasi
-- (one-time benefit for early registrants; new signups still require email verification)

UPDATE users
SET
  email_verified = TRUE,
  email_verified_at = COALESCE(email_verified_at, NOW())
WHERE email_verified = FALSE;
