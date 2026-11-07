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
		if !near(r.TimeStretchPercent, tc.stretch, 0.005) {
			t.Errorf("%v → %v: time stretch %.2f %%, want %.2f %%", tc.original, tc.new, r.TimeStretchPercent, tc.stretch)
		}
	}
}

// TestTempoChangeCentsCarry checks that a shift just below a whole semitone carries instead of
// showing 100 cents
func TestTempoChangeCentsCarry(t *testing.T) {
	for _, tc := range []struct {
		original, new    float64
		semis, cents     int
		semis50, cents50 int
	}{
		{120, 119.98, 0, 0, 0, 0},
		{120, 120 * math.Pow(2, 2.996/12), 3, 0, 3, 0},
		{120, 120 * math.Pow(2, -0.996/12), -1, 0, -1, 0},
		{120, 100, -4, 84, -3, -16},
		{120, 120 * math.Pow(2, -0.5/12), -1, 50, -1, 50},
		{120, 120 * math.Pow(2, 0.5/12), 0, 50, 1, -50},
	} {
		r := CalculateFromNewTempo(tc.original, tc.new)
		if !r.Valid || r.Semitones != tc.semis || r.Cents != tc.cents || r.Semitones50Cent != tc.semis50 || r.Cents50Cent != tc.cents50 {
			t.Errorf("%v → %v: %d st %d c / %d st %d c, want %d %d / %d %d", tc.original, tc.new,
				r.Semitones, r.Cents, r.Semitones50Cent, r.Cents50Cent, tc.semis, tc.cents, tc.semis50, tc.cents50)
		}
	}
}

func TestTempoChangeDirectionsAndRange(t *testing.T) {
	if r := CalculateFromTimeStretch(120, 120); !r.Valid || !near(r.NewTempo, 100, 1e-9) {
		t.Errorf("120 BPM stretched to 120 %% = %+v, want 100 BPM", r)
	}
	if r := CalculateFromTranspose(100, 12, 0); !r.Valid || !near(r.NewTempo, 200, 1e-9) || !near(r.TimeStretchPercent, 50, 1e-9) {
		t.Errorf("100 BPM + 12 semitones = %+v, want 200 BPM at 50 %%", r)
	}
	for _, tc := range []struct{ original, new float64 }{
		{120, 4.99}, {120, 1000}, {4.99, 120}, {120, 0}, {120, -100}, {math.NaN(), 120}, {120, math.Inf(1)},
	} {
		if r := CalculateFromNewTempo(tc.original, tc.new); r.Valid {
			t.Errorf("%v → %v BPM is outside the range but gave %+v", tc.original, tc.new, r)
		}
	}
	if r := CalculateFromNewTempo(MinTempoBPM, MaxTempoBPM); !r.Valid {
		t.Error("the range limits themselves must be valid")
	}
	if r := CalculateFromTimeStretch(120, 0); r.Valid {
		t.Error("a 0 % stretch must be invalid")
	}
}

func TestNormalizeTranspose(t *testing.T) {
	for _, tc := range []struct {
		original             float64
		semis, cents         int
		wantSemis, wantCents int
	}{
		{120, 0, 51, 1, -49},
		{120, 0, 150, 1, 50},
		{120, 0, 151, 2, -49},
		{120, 2, -51, 1, 49},
		{120, 0, 50, 0, 50},
		{120, 0, -50, 0, -50},
		{120, 0, 0, 0, 0},
		{0, 0, 151, 2, -49}, // invalid original: wrapped, not limited
	} {
		s, c := NormalizeTranspose(tc.original, tc.semis, tc.cents)
		if s != tc.wantSemis || c != tc.wantCents {
			t.Errorf("NormalizeTranspose(%v, %d, %d) = %d, %d; want %d, %d", tc.original, tc.semis, tc.cents, s, c, tc.wantSemis, tc.wantCents)
		}
	}
	// Limited shifts stay inside the range and are written as 0 to 99 cents
	for _, semis := range []int{60, -60, 1000000, -1000000} {
		s, c := NormalizeTranspose(120, semis, 0)
		r := CalculateFromTranspose(120, s, c)
		if !r.Valid || c < 0 || c > 99 || r.NewTempo > MaxTempoBPM || r.NewTempo < MinTempoBPM {
			t.Errorf("120 BPM %+d st: limited to %d st %d c = %+v, want a valid tempo and 0-99 cents", semis, s, c, r)
		}
	}
}
