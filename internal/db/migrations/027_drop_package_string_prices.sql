-- Harga paket disimpan dalam format int (price_early_bird, price_normal) dalam rupiah.
-- Kolom price_display dihapus karena harga tampilan digenerate dari nilai int.

ALTER TABLE packages DROP COLUMN IF EXISTS price_display;
