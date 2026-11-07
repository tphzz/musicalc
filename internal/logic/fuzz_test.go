package logic

import (
	"math"
	"sort"
	"testing"
)

// FuzzCombFrequencies checks that comb enumeration always ends with a bounded, sorted list of
// finite frequencies at or above its start, for any delay, gain and start
func FuzzCombFrequencies(f *testing.F) {
	for _, seed := range [][3]float64{
		{0.01165, 1, 20}, {-0.01165, 0.5, 1200}, {2.45e-6, 1, 3000}, {math.NaN(), 1, 20}, {math.Inf(1), 1, 20},
		{0.001, math.NaN(), 20}, {0.001, 1, math.NaN()}, {1e300, 1, 1e300}, {1e-300, 1, 0}, {0, 1, 20},
		{1e300, 1, 20}, {1e18, 1, 20}, // huge delays: the index stops advancing
	} {
		f.Add(seed[0], seed[1], seed[2], false)
	}
	f.Fuzz(func(t *testing.T, delay, gain, from float64, inverted bool) {
		m := TwoMicSum{Delay: delay, Gain: gain, Inverted: inverted}
		for _, fs := range [][]float64{m.DipsFrom(from, 3), m.PeaksFrom(from, 3)} {
			if len(fs) > 3 || !sort.Float64sAreSorted(fs) {
				t.Fatalf("%+v from %v: %v", m, from, fs)
			}
			for i := 1; i < len(fs); i++ {
				if !(fs[i] > fs[i-1]) {
					t.Fatalf("%+v from %v: repeated frequency in %v", m, from, fs)
				}
			}
			for _, x := range fs {
				if math.IsNaN(x) || math.IsInf(x, 0) || x < from || x <= 0 {
					t.Fatalf("%+v from %v: frequency %v", m, from, x)
				}
			}
		}
		if m.Valid() && from > 0 && from < 1e6 {
			for _, x := range SampleFrequencies(from, from*10, 50, m) {
				if math.IsNaN(x) || math.IsInf(x, 0) {
					t.Fatalf("%+v: sample %v", m, x)
				}
			}
		}
	})
}

// FuzzParseNumber checks that a number is either rejected or finite, and that ParseFloat agrees
func FuzzParseNumber(f *testing.F) {
	for _, s := range []string{"", "0", "1,5", " 2.5 ", "-0", "NaN", "Inf", "-Infinity", "1e400", "1e-400", "abc", "1,000.5", "0x10"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, ok := ParseNumber(s)
		if ok && (math.IsNaN(v) || math.IsInf(v, 0)) {
			t.Fatalf("%q accepted as %v", s, v)
		}
		if !ok && v != 0 {
			t.Fatalf("%q rejected but gave %v", s, v)
		}
		if got := ParseFloat(s); got != v {
			t.Fatalf("ParseFloat(%q) = %v, ParseNumber %v", s, got, v)
		}
	})
}
