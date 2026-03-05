-- Body soal dapat berisi HTML; dukungan banyak gambar (array URL)
-- body tetap TEXT, bisa menyimpan HTML (rich text)
ALTER TABLE questions
  ADD COLUMN IF NOT EXISTS image_urls JSONB DEFAULT '[]';

COMMENT ON COLUMN questions.body IS 'Teks atau HTML isi soal (boleh berisi tag HTML dan <img> untuk gambar)';
COMMENT ON COLUMN questions.image_urls IS 'Array URL gambar tambahan untuk soal, contoh: ["https://...", "https://..."]';
