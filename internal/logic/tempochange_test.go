package logic

import (
	"math"
	"testing"
)

func TestTempoChangeTranspose(t *testing.T) {
	for _, tc := range []struct {
		original, new float64
		semis, cents  int // standard notation: whole semitones down, then 0 to 99 cents
		semis50       int // 50-cent notation: nearest semitone, then ±50 cents
		cents50       int
		stretch       float64
	}{
		{140, 170, 3, 36, 3, 36, 82.35},
		{120, 120 * math.Pow(2, 3.7/12), 3, 70, 4, -30, 80.76},
		{120, 120 * math.Pow(2, -2.4/12), -3, 60, -2, -40, 114.87},
		{120, 60, -12, 0, -12, 0, 200},
	} {
		r := CalculateFromNewTempo(tc.original, tc.new)
		if r.Semitones != tc.semis || r.Cents != tc.cents {
			t.Errorf("%v → %v: %d semitones %d cents, want %d %d", tc.original, tc.new, r.Semitones, r.Cents, tc.semis, tc.cents)
		}
		if r.Semitones50Cent != tc.semis50 || r.Cents50Cent != tc.cents50 {
			t.Errorf("%v → %v: 50-cent %d semitones %d cents, want %d %d", tc.original, tc.new, r.Semitones50Cent, r.Cents50Cent, tc.semis50, tc.cents50)
		}
		if math.Abs(r.TimeStretchPercent-tc.stretch) > 0.005 {
			t.Errorf("%v → %v: time stretch %.2f %%, want %.2f %%", tc.original, tc.new, r.TimeStretchPercent, tc.stretch)
		}
	}
}
