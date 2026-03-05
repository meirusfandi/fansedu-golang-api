-- Soal (pertanyaan) dapat menambahkan gambar di dalamnya (URL gambar)
ALTER TABLE questions
  ADD COLUMN IF NOT EXISTS image_url TEXT;

COMMENT ON COLUMN questions.image_url IS 'URL gambar yang ditampilkan di dalam soal (opsional)';
