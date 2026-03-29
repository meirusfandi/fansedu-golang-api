package service

// percentileRankPercent menghitung persentil (0–100) untuk value dalam kumpulan skor,
// memakai rumus (c_less + 0.5*c_equal) / n * 100. Mengembalikan nil jika n < 2.
func percentileRankPercent(scores []float64, value float64) *float64 {
	if len(scores) < 2 {
		return nil
	}
	var less, eq int
	for _, s := range scores {
		if s < value {
			less++
		} else if s == value {
			eq++
		}
	}
	n := len(scores)
	p := (float64(less) + 0.5*float64(eq)) / float64(n) * 100
	return &p
}
