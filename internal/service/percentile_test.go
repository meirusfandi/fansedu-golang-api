package service

import (
	"math"
	"testing"
)

func TestPercentileRankPercent(t *testing.T) {
	if p := percentileRankPercent([]float64{10}, 10); p != nil {
		t.Fatalf("expected nil for n<2, got %v", *p)
	}
	// Empat skor 10,20,20,30 — nilai 20 → (1 + 0.5*2) / 4 * 100 = 50
	p := percentileRankPercent([]float64{10, 20, 20, 30}, 20)
	if p == nil || math.Abs(*p-50) > 0.001 {
		t.Fatalf("got %v want ~50", p)
	}
	// Satu-satunya yang lebih rendah dari 30
	p = percentileRankPercent([]float64{10, 20, 20, 30}, 30)
	if p == nil || math.Abs(*p-87.5) > 0.001 {
		t.Fatalf("got %v want ~87.5", p)
	}
}
