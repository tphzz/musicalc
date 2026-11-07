package logic

import (
	"context"
	"math"
	"os"
	"os/exec"
	"testing"
	"time"
)

// arrival is ArrivalDelay for a valid setup
func arrival(meters, celsius float64) float64 {
	s, ok := ArrivalDelay(meters, celsius)
	if !ok {
		panic("invalid test arrival")
	}
	return s
}

// fourMeters is the design example: Mic B 4 m farther at 20 °C, equal level, normal polarity
var fourMeters = TwoMicSum{Delay: arrival(4, 20), Gain: 1}

func TestArrivalDelay(t *testing.T) {
	if got := arrival(4, 20) * 1000; !near(got, 11.6475, 0.0001) {
		t.Errorf("4 m at 20 °C: got %.4f ms, want 11.6475", got)
	}
	// 4 m at 3 °C
	if got := arrival(4, 3) * 1000; !near(got, 12.008, 0.001) {
		t.Errorf("4 m at 3 °C: got %.3f ms, want 12.008", got)
	}
	if got := arrival(4, 3) * 48000; !near(got, 576.4, 0.1) {
		t.Errorf("4 m at 3 °C: got %.1f samples, want 576.4", got)
	}
	if !(arrival(-1, 20) < 0) {
		t.Error("a negative path should give a negative delay (Mic B first)")
	}
	// Temperatures outside −40 to +50 °C and non-finite paths are rejected, not turned into 0
	for _, tc := range []struct{ m, c float64 }{{4, -600}, {4, -40.01}, {4, 50.01}, {4, math.NaN()}, {math.Inf(1), 20}, {math.NaN(), 20}} {
		if s, ok := ArrivalDelay(tc.m, tc.c); ok {
			t.Errorf("ArrivalDelay(%v m, %v °C) = %v, want rejected", tc.m, tc.c, s)
		}
	}
	for _, c := range []float64{MinTemperatureC, MaxTemperatureC} {
		if _, ok := ArrivalDelay(4, c); !ok {
			t.Errorf("%v °C is within the range", c)
		}
	}
	// Distance and time convert back exactly
	if d, ok := DistanceForDelay(arrival(1.23456789, 7), 7); !ok || !near(d, 1.23456789, 1e-12) {
		t.Errorf("distance round trip: %v %v", d, ok)
	}
}

func TestTwoMicSumLevels(t *testing.T) {
	tests := []struct {
		name string
		m    TwoMicSum
		freq float64
		want float64
	}{
		{"design case at 60 Hz", fourMeters, 60, 1.36},
		{"4 m at 3 °C, 60 Hz", TwoMicSum{Delay: arrival(4, 3), Gain: 1}, 60, 2.13},
		{"aligned equal mics add 6 dB", TwoMicSum{Gain: 1}, 60, 6.02},
		{"aligned −12 dB mic", TwoMicSum{Gain: DBToGain(-12)}, 1000, 1.94},
		{"−6 dB, inverted, at 60 Hz", TwoMicSum{Delay: fourMeters.Delay, Gain: DBToGain(-6), Inverted: true}, 60, 1.95},
		{"a peak of the design case", fourMeters, 1 / fourMeters.Delay, 6.02},
	}
	for _, tt := range tests {
		if got := tt.m.LevelDB(tt.freq); !near(got, tt.want, 0.01) {
			t.Errorf("%s: got %.3f dB, want %.2f", tt.name, got, tt.want)
		}
	}
	// A full cancellation stays finite
	if got := fourMeters.LevelDB(0.5 / fourMeters.Delay); got != MinLevelDB {
		t.Errorf("full cancellation: got %v, want %v", got, MinLevelDB)
	}
	if got := (TwoMicSum{Gain: 1, Inverted: true}).LevelDB(100); got != MinLevelDB {
		t.Errorf("aligned inverted equal mics: got %v, want %v", got, MinLevelDB)
	}
	// The sign of the delay doesn't change the level
	early := TwoMicSum{Delay: -fourMeters.Delay, Gain: 1}
	if !near(early.LevelDB(73), fourMeters.LevelDB(73), 1e-9) {
		t.Error("Mic B earlier and later should give the same level")
	}
}

func TestTwoMicSumDipsAndPeaks(t *testing.T) {
	round := func(fs []float64) []float64 {
		out := make([]float64, len(fs))
		for i, f := range fs {
			out[i] = math.Round(f*10) / 10
		}
		return out
	}
	eq := func(a, b []float64) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	if got := round(fourMeters.Dips(3)); !eq(got, []float64{42.9, 128.8, 214.6}) {
		t.Errorf("dips: got %v", got)
	}
	if got := round(fourMeters.Peaks(3)); !eq(got, []float64{85.9, 171.7, 257.6}) {
		t.Errorf("peaks: got %v", got)
	}
	inv := TwoMicSum{Delay: fourMeters.Delay, Gain: 1, Inverted: true}
	if got := round(inv.Dips(3)); !eq(got, []float64{85.9, 171.7, 257.6}) {
		t.Errorf("inverted dips: got %v", got)
	}
	if got := round(inv.Peaks(3)); !eq(got, []float64{42.9, 128.8, 214.6}) {
		t.Errorf("inverted peaks: got %v", got)
	}
	// The first cancellations of 4 m at 3 °C
	if got := round(TwoMicSum{Delay: arrival(4, 3), Gain: 1}.Dips(3)); !eq(got, []float64{41.6, 124.9, 208.2}) {
		t.Errorf("dips at 3 °C: got %v", got)
	}
	if (TwoMicSum{Gain: 1}).Dips(3) != nil {
		t.Error("aligned mics have no comb")
	}
	// From a frequency on: the cymbal range of the design example
	if got := round(fourMeters.DipsFrom(1200, 2)); !eq(got, []float64{1244.9, 1330.8}) {
		t.Errorf("dips from 1200 Hz: got %v", got)
	}
	if got := round(inv.PeaksFrom(42.9, 1)); !eq(got, []float64{42.9}) {
		t.Errorf("inverted peaks from 42.9 Hz: got %v", got)
	}
	if got := fourMeters.CombSpacing(); !near(got, 85.855, 0.001) || (TwoMicSum{Gain: 1}).CombSpacing() != 0 {
		t.Errorf("comb spacing: got %v", got)
	}
	// Every listed dip really is a minimum
	for _, f := range fourMeters.Dips(5) {
		if fourMeters.LevelDB(f) > -60 {
			t.Errorf("dip at %.1f Hz is only %.1f dB", f, fourMeters.LevelDB(f))
		}
	}
	// Peak and dip depth follow the level
	m := TwoMicSum{Delay: fourMeters.Delay, Gain: DBToGain(-6)}
	if !near(m.PeakDB(), 3.53, 0.01) || !near(m.DipDB(), -6.04, 0.01) {
		t.Errorf("−6 dB: peak %.2f, dip %.2f", m.PeakDB(), m.DipDB())
	}
	if !near(fourMeters.PeakDB(), 6.02, 0.01) || fourMeters.DipDB() != MinLevelDB {
		t.Errorf("equal: peak %.2f, dip %.2f", fourMeters.PeakDB(), fourMeters.DipDB())
	}
}

func TestTwoMicSumRange(t *testing.T) {
	min, max := fourMeters.RangeDB(50, 100)
	if !near(min, -5.82, 0.02) || !near(max, 6.02, 0.01) {
		t.Errorf("design case 50–100 Hz: got %.2f to %.2f", min, max)
	}
	c := TwoMicSum{Delay: fourMeters.Delay, Gain: DBToGain(-6), Inverted: true}
	if min, max := c.RangeDB(50, 100); !near(min, -6.04, 0.02) || !near(max, 3.27, 0.02) {
		t.Errorf("−6 dB inverted 50–100 Hz: got %.2f to %.2f", min, max)
	}
	// Over a wide range the extremes are the full comb, found exactly
	if min, max := fourMeters.RangeDB(20, 1000); min != MinLevelDB || !near(max, 6.02, 0.001) {
		t.Errorf("20–1000 Hz: got %.2f to %.2f", min, max)
	}
	if min, max := (TwoMicSum{Gain: 1}).RangeDB(50, 100); !near(min, 6.02, 0.01) || !near(max, 6.02, 0.01) {
		t.Errorf("aligned: got %.2f to %.2f", min, max)
	}
}

func TestGraphRange(t *testing.T) {
	tests := []struct{ lo, hi, wantLo, wantHi float64 }{
		{50, 100, 20, 1000},        // kick: the design's graph
		{3000, 12000, 1200, 20000}, // cymbals: capped at 20 kHz
		{20, 40, 10, 400},          // low bass: floor at 10 Hz
		{0, 100, 20, 1000},         // invalid: the default
		{100, 50, 20, 1000},
		{9000, 19000, 3600, 20000},
	}
	for _, tt := range tests {
		lo, hi := GraphRange(tt.lo, tt.hi)
		if !near(lo, tt.wantLo, 1e-9) || !near(hi, tt.wantHi, 1e-9) {
			t.Errorf("%v–%v: got %v–%v, want %v–%v", tt.lo, tt.hi, lo, hi, tt.wantLo, tt.wantHi)
		}
	}
}

func TestSampleFrequenciesIncludeCombPoints(t *testing.T) {
	fs := SampleFrequencies(20, 1000, 100, fourMeters)
	for _, d := range fourMeters.Dips(3) {
		found := false
		for _, f := range fs {
			if near(f, d, 1e-9) {
				found = true
			}
		}
		if !found {
			t.Errorf("dip %.2f Hz missing from the samples", d)
		}
	}
	for i := 1; i < len(fs); i++ {
		if fs[i] < fs[i-1] {
			t.Fatal("samples are not sorted")
		}
	}
	if SampleFrequencies(0, 100, 10, fourMeters) != nil || SampleFrequencies(100, 50, 10, fourMeters) != nil {
		t.Error("invalid ranges should give no samples")
	}
}

// TestCombEnumerationTerminates checks that invalid input gives no frequencies instead of an
// endless loop. It runs in a child process with a deadline, so a regression cannot hang the suite.
func TestCombEnumerationTerminates(t *testing.T) {
	if os.Getenv("MUSICALC_COMB_CHILD") == "1" {
		for _, m := range []TwoMicSum{
			{Delay: math.NaN(), Gain: 1}, {Delay: math.Inf(1), Gain: 1}, {Delay: math.Inf(-1), Gain: 1},
			{Delay: 0.001, Gain: math.NaN()}, {Delay: 0.001, Gain: math.Inf(1)}, {Delay: 0.001, Gain: 0},
		} {
			if got := m.DipsFrom(20, 3); got != nil {
				t.Errorf("%+v: dips %v, want none", m, got)
			}
			_ = m.PeaksFrom(20, 3)
			_ = SampleFrequencies(20, 1000, 100, m)
		}
		good := TwoMicSum{Delay: 0.001, Gain: 1}
		for _, from := range []float64{math.NaN(), math.Inf(1), -1} {
			if got := good.DipsFrom(from, 3); got != nil {
				t.Errorf("from %v: dips %v, want none", from, got)
			}
		}
		// Huge but finite values stay bounded and finite
		for _, m := range []TwoMicSum{{Delay: 1e300, Gain: 1}, {Delay: 1e-300, Gain: 1}} {
			for _, f := range m.DipsFrom(1e300, 3) {
				if math.IsNaN(f) || math.IsInf(f, 0) {
					t.Errorf("%+v: non-finite dip %v", m, f)
				}
			}
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCombEnumerationTerminates$", "-test.v")
	cmd.Env = append(os.Environ(), "MUSICALC_COMB_CHILD=1")
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("comb enumeration did not finish within the deadline:\n%s", out)
	}
	if err != nil {
		t.Fatalf("child failed: %v\n%s", err, out)
	}
}

// TestResidualDelayKeepsComb checks that a residual of microseconds left by 0.01 ms delay steps is
// not treated as alignment: equal inverted copies still cancel only partly at high frequencies
func TestResidualDelayKeepsComb(t *testing.T) {
	a := arrival(4, 20)
	residual := a - 0.01165 // Align writes 11.65 ms
	if !near(residual*1e6, -2.454720168, 1e-6) {
		t.Fatalf("residual %.9f us", residual*1e6)
	}
	m := TwoMicSum{Delay: residual, Gain: 1, Inverted: true}
	if m.Aligned() {
		t.Fatal("a residual of 2.45 us must not count as aligned")
	}
	for _, tc := range []struct{ f, want float64 }{{6000, -20.676437}, {12000, -14.665139}, {20000, -10.250227}} {
		if got := m.LevelDB(tc.f); !near(got, tc.want, 1e-5) {
			t.Errorf("%v Hz: %.6f dB, want %.6f", tc.f, got, tc.want)
		}
	}
	if got := (TwoMicSum{Delay: 0.004e-3, Gain: 1, Inverted: true}).LevelDB(20000); !near(got, -6.066238, 1e-5) {
		t.Errorf("0.004 ms residual at 20 kHz: %.6f dB, want -6.066238", got)
	}
	if got := (TwoMicSum{Delay: 50e-9, Gain: 1, Inverted: true}).LevelDB(20000); !near(got, -44.04, 0.01) {
		t.Errorf("50 ns residual at 20 kHz: %.3f dB, want about -44.04", got)
	}
	// Exact zero: +6.02 dB normal, the floor inverted
	if got := (TwoMicSum{Gain: 1}).LevelDB(1000); !near(got, 6.020600, 1e-6) {
		t.Errorf("exact zero, normal: %.6f", got)
	}
	if got := (TwoMicSum{Gain: 1, Inverted: true}).LevelDB(1000); got != MinLevelDB {
		t.Errorf("exact zero, inverted: %v, want the floor", got)
	}
	// Delay sign and both polarities with unequal gain: the magnitude is symmetric in the sign
	for _, inv := range []bool{false, true} {
		p, q := TwoMicSum{Delay: 3e-4, Gain: 0.7, Inverted: inv}, TwoMicSum{Delay: -3e-4, Gain: 0.7, Inverted: inv}
		for _, f := range []float64{100, 1234, 5000} {
			if !near(p.LevelDB(f), q.LevelDB(f), 1e-9) {
				t.Errorf("inverted %v at %v Hz: %v vs %v", inv, f, p.LevelDB(f), q.LevelDB(f))
			}
		}
	}
	// A louder inverted copy at zero delay is a boost: |1 - 10^(12/20)| = +9.487 dB
	if got := (TwoMicSum{Gain: DBToGain(12), Inverted: true}).DipDB(); !near(got, 9.487448450, 1e-6) {
		t.Errorf("aligned inverted +12 dB: %.9f, want +9.487448450", got)
	}
}

func TestBandAndPointRange(t *testing.T) {
	for _, tc := range []struct {
		lo, hi float64
		ok     bool
	}{{10, 20000, true}, {50, 100, true}, {9.99, 100, false}, {100, 100, false}, {100, 50, false}, {50, 20000.1, false}, {math.NaN(), 100, false}} {
		if got := BandInRange(tc.lo, tc.hi); got != tc.ok {
			t.Errorf("BandInRange(%v, %v) = %v", tc.lo, tc.hi, got)
		}
	}
	if lo, hi := fourMeters.RangeDB(60, 60); !near(lo, fourMeters.LevelDB(60), 1e-12) || lo != hi {
		t.Errorf("a single frequency: %v to %v", lo, hi)
	}
}
