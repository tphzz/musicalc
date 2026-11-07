package logic

import (
	"math"
	"testing"
)

func TestFrequencyToNote(t *testing.T) {
	for _, tc := range []struct {
		hz       float64
		offset   int
		note100  string
		cents100 int
		note50   string
		cents50  int
	}{
		{440, OctaveOffsetC3, "A3", 0, "A3", 0},
		{440, OctaveOffsetC4, "A4", 0, "A4", 0},
		{439.9, OctaveOffsetC4, "A4", 0, "A4", 0}, // carries instead of G#4 +100
		{440 * math.Pow(2, 0.3/12), OctaveOffsetC4, "A4", 30, "A4", 30},
		{440 * math.Pow(2, 0.7/12), OctaveOffsetC4, "A4", 70, "A#4", -30},
		{5, OctaveOffsetC4, "D#-2", 49, "D#-2", 49},   // below MIDI 0
		{8.5, OctaveOffsetC3, "C-2", 67, "C#-2", -33}, // C3 naming, octave -2
		{16800, OctaveOffsetC4, "C10", 6, "C10", 6},   // above MIDI 127
		{20000, OctaveOffsetC4, "D#10", 8, "D#10", 8},
	} {
		r := FrequencyToNote(tc.hz, tc.offset)
		if !r.Valid || r.Note100 != tc.note100 || r.Cents100 != tc.cents100 || r.Note50 != tc.note50 || r.Cents50 != tc.cents50 {
			t.Errorf("%v Hz: %s %+d / %s %+d, want %s %+d / %s %+d", tc.hz, r.Note100, r.Cents100, r.Note50, r.Cents50,
				tc.note100, tc.cents100, tc.note50, tc.cents50)
		}
		if r.Cents100 < 0 || r.Cents100 > 99 || r.Cents50 < -50 || r.Cents50 > 50 {
			t.Errorf("%v Hz: cents %d / %d outside 0-99 / ±50", tc.hz, r.Cents100, r.Cents50)
		}
	}
	for _, hz := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if r := FrequencyToNote(hz, OctaveOffsetC4); r.Valid {
			t.Errorf("%v Hz gave %+v, want invalid", hz, r)
		}
	}
}

func TestGetFrequencyForNote(t *testing.T) {
	if got := GetFrequencyForNote("A", 4, OctaveOffsetC3); !near(got, 880, 1e-9) {
		t.Errorf("A4 in C3 naming = %v Hz, want 880", got)
	}
	if got := GetFrequencyForNote("C", 4, OctaveOffsetC4); !near(got, 261.6255653, 1e-6) {
		t.Errorf("C4 in C4 naming = %v Hz, want 261.6256", got)
	}
	if got := GetFrequencyForNote("H", 4, OctaveOffsetC4); got != 0 {
		t.Errorf("unknown note name gave %v, want 0", got)
	}
}
