package logic

import (
	"math"
	"math/big"
	"testing"
)

func TestSampleLength(t *testing.T) {
	for _, tc := range []struct {
		rate, bpm, beats float64
		samples          int
		ms               float64
	}{
		{44100, 120, 4, 88200, 2000},
		{48000, 120, 4, 96000, 2000},
		{44100, 50, 3, 158760, 3600}, // exactly 3.6 s, once one sample short
		{44100, 130, 1, 20353, 461.538461538},
		{192000, 5, 40000, 92160000000, 480000000}, // large exact count: snapped, not over-counted
	} {
		r := GetSampleData(tc.rate, tc.bpm, tc.beats)
		if !r.Valid || r.Samples != tc.samples || !near(r.MS, tc.ms, 1e-6) {
			t.Errorf("%v Hz, %v BPM, %v beats = %+v, want %d samples, %v ms", tc.rate, tc.bpm, tc.beats, r, tc.samples, tc.ms)
		}
	}
	for _, tc := range []struct{ rate, bpm, beats float64 }{
		{0, 120, 4}, {44100, 0, 4}, {44100, 120, 0}, {44100, -120, 4}, {math.NaN(), 120, 4}, {44100, 120, math.Inf(1)}, {192000, 1e-300, 4}, {192000, 5, 1e6},
	} {
		if r := GetSampleData(tc.rate, tc.bpm, tc.beats); r.Valid {
			t.Errorf("%v Hz, %v BPM, %v beats gave %+v, want invalid", tc.rate, tc.bpm, tc.beats, r)
		}
	}
}

// TestSampleLengthExactCounts compares every whole-sample count with exact integer arithmetic,
// for tempos in tenths of a BPM
func TestSampleLengthExactCounts(t *testing.T) {
	for _, rate := range []int64{22050, 44100, 48000, 88200, 96000} {
		for bpm10 := int64(400); bpm10 <= 2400; bpm10++ {
			for _, beats := range []int64{1, 2, 3, 4, 8, 16} {
				got := GetSampleData(float64(rate), float64(bpm10)/10, float64(beats)).Samples
				want := new(big.Int).Div(big.NewInt(600*beats*rate), big.NewInt(bpm10)).Int64()
				if int64(got) != want {
					t.Fatalf("%d Hz, %.1f BPM, %d beats: %d samples, want %d", rate, float64(bpm10)/10, beats, got, want)
				}
			}
		}
	}
}

func TestSamplesIn(t *testing.T) {
	if n, ok := SamplesIn(1000, 48000); !ok || n != 48000 {
		t.Errorf("1 s at 48 kHz = %d, %v; want 48000", n, ok)
	}
	// A genuinely fractional count still rounds down
	if n, ok := SamplesIn(0.0104, 48000); !ok || n != 0 {
		t.Errorf("0.0104 ms at 48 kHz = %d, %v; want 0", n, ok)
	}
	if n, ok := SamplesIn(1000.0/3, 3); !ok || n != 1 {
		t.Errorf("1/3 s at 3 Hz = %d, %v; want 1", n, ok)
	}
	for _, tc := range []struct{ ms, rate float64 }{{0, 48000}, {-1, 48000}, {1, 0}, {1e300, 48000}, {math.NaN(), 48000}} {
		if n, ok := SamplesIn(tc.ms, tc.rate); ok {
			t.Errorf("SamplesIn(%v, %v) = %d, want invalid", tc.ms, tc.rate, n)
		}
	}
}
