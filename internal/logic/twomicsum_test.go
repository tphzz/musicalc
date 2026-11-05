package logic

import (
	"math"
	"testing"
)

// fourMeters is the design example: Mic B 4 m farther at 20 °C, equal level, normal polarity
var fourMeters = TwoMicSum{Delay: ArrivalDelay(4, 20), Gain: 1}

func TestArrivalDelay(t *testing.T) {
	if got := ArrivalDelay(4, 20) * 1000; !near(got, 11.6475, 0.0001) {
		t.Errorf("4 m at 20 °C: got %.4f ms, want 11.6475", got)
	}
	// The reviewer's example: 4 m at 3 °C
	if got := ArrivalDelay(4, 3) * 1000; !near(got, 12.008, 0.001) {
		t.Errorf("4 m at 3 °C: got %.3f ms, want 12.008", got)
	}
	if got := ArrivalDelay(4, 3) * 48000; !near(got, 576.4, 0.1) {
		t.Errorf("4 m at 3 °C: got %.1f samples, want 576.4", got)
	}
	if ArrivalDelay(-1, 20) >= 0 {
		t.Error("a negative path should give a negative delay (Mic B first)")
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
		{"reviewer case at 60 Hz", TwoMicSum{Delay: ArrivalDelay(4, 3), Gain: 1}, 60, 2.13},
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
	if got := fourMeters.LevelDB(0.5 / fourMeters.Delay); got != minLevelDB {
		t.Errorf("full cancellation: got %v, want %v", got, minLevelDB)
	}
	if got := (TwoMicSum{Gain: 1, Inverted: true}).LevelDB(100); got != minLevelDB {
		t.Errorf("aligned inverted equal mics: got %v, want %v", got, minLevelDB)
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
	// The reviewer's first cancellations at 3 °C
	if got := round(TwoMicSum{Delay: ArrivalDelay(4, 3), Gain: 1}.Dips(3)); !eq(got, []float64{41.6, 124.9, 208.2}) {
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
	if !near(fourMeters.PeakDB(), 6.02, 0.01) || fourMeters.DipDB() != minLevelDB {
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
	if min, max := fourMeters.RangeDB(20, 1000); min != minLevelDB || !near(max, 6.02, 0.001) {
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
