package logic

import (
	"math"
	"testing"
)

const classicTarget = 9.542425094393248 // 20·log10(3)

func TestSpillTargets(t *testing.T) {
	if !near(SpillTargets[0].DB, classicTarget, 1e-12) || SpillTargets[1].DB != 10 || SpillTargets[2].DB != 15 {
		t.Errorf("unexpected targets: %+v", SpillTargets)
	}
}

func TestSpillCheckDesignCases(t *testing.T) {
	tests := []struct {
		name                 string
		own, neighbour, gain float64
		target               float64
		ratio                float64
		geometry             Geometry
		spill, margin        float64
		meets                bool
		minNeighbour         float64
		boost, cut           float64
	}{
		// A: meets the guideline with room to spare
		{"A", 1, 4, 0, classicTarget, 4, GeometryMeets, -12.041, 2.499, true, 3, 1.938, -2.499},
		// B: 3:1 geometry, but the neighbouring channel is 6 dB louder
		{"B", 1, 3, 6, 10, 3, GeometryThreshold, -3.542, -6.458, false, 6.310, 4.429, -9.501},
		// C: below 3:1
		{"C", 1, 2, 0, classicTarget, 2, GeometryBelow, -6.021, -3.522, false, 3, 3.522, -6.021},
	}
	for _, tt := range tests {
		r := SpillCheck(tt.own, tt.neighbour, tt.gain, tt.target)
		if !r.Valid || !near(r.Ratio, tt.ratio, 1e-9) || r.Geometry != tt.geometry || !near(r.SpillDB, tt.spill, 0.001) ||
			!near(r.MarginDB, tt.margin, 0.001) || r.MeetsTarget != tt.meets || !near(r.MinNeighbour, tt.minNeighbour, 0.001) ||
			!near(r.CombBoostDB, tt.boost, 0.001) || !near(r.CombCutDB, tt.cut, 0.001) || r.CanCancel {
			t.Errorf("%s: got %+v", tt.name, r)
		}
	}
}

func TestSpillCheckThreshold(t *testing.T) {
	tests := []struct {
		neighbour float64
		want      Geometry
	}{
		{3, GeometryThreshold},
		{2.98, GeometryThreshold},
		{3.02, GeometryThreshold},
		{2.9, GeometryBelow},
		{3.1, GeometryMeets},
	}
	for _, tt := range tests {
		if got := SpillCheck(1, tt.neighbour, 0, classicTarget).Geometry; got != tt.want {
			t.Errorf("1 / %v: got geometry %v, want %v", tt.neighbour, got, tt.want)
		}
	}
	// Exactly 3:1 meets the classic target, despite floating point
	if r := SpillCheck(0.3, 0.9, 0, classicTarget); !r.MeetsTarget {
		t.Errorf("exactly 3:1 should be on target: %+v", r)
	}
	// A spill a little worse than the 10 dB target is short of it
	if r := SpillCheck(1, 3, 0, 10); r.MeetsTarget || !near(r.MarginDB, -0.458, 0.001) || !near(r.MinNeighbour, 3.162, 0.001) {
		t.Errorf("3:1 against 10 dB: got %+v", r)
	}
	// Within 0.05 dB counts as on target
	if r := SpillCheck(1, 3.15, 0, 10); !r.MeetsTarget {
		t.Errorf("0.03 dB short should count as on target: %+v", r)
	}
}

func TestSpillCheckGain(t *testing.T) {
	// A quieter neighbouring channel improves the spill by the same amount
	base := SpillCheck(1, 2, 0, 10)
	quieter := SpillCheck(1, 2, -6, 10)
	if !near(quieter.SpillDB, base.SpillDB-6, 1e-9) || !near(quieter.MarginDB, base.MarginDB+6, 1e-9) {
		t.Errorf("−6 dB gain: got %+v, base %+v", quieter, base)
	}
	if quieter.MinNeighbour >= base.MinNeighbour {
		t.Errorf("a quieter neighbour may come closer: %v vs %v", quieter.MinNeighbour, base.MinNeighbour)
	}
	// Only equal amplitudes cancel completely: louder spill has a finite minimum
	loud := SpillCheck(1, 3, 12, classicTarget)
	if !near(loud.SpillDB, 2.457574906, 1e-9) || loud.CanCancel || loud.MeetsTarget ||
		!near(loud.CombCutDB, -9.708410081, 1e-9) || !near(loud.CombBoostDB, 7.336016883, 1e-9) {
		t.Errorf("+12 dB gain: got %+v, want a = 1.327, -9.708 to +7.336 dB and no cancellation", loud)
	}
	equal := SpillCheck(1, 1, 0, classicTarget)
	if !equal.CanCancel || !near(equal.CombBoostDB, 6.021, 0.001) {
		t.Errorf("equal distances: got %+v", equal)
	}
}

// TestSpillCheckEnvelope checks the comb bounds 20·log10(1+a) and 20·log10(|1-a|) just below, at
// and above unity, at a = 2 and above, where even the minimum is a boost
func TestSpillCheckEnvelope(t *testing.T) {
	for _, tc := range []struct {
		gain      float64 // with own = neighbour, the spill amplitude is 10^(gain/20)
		cancel    bool
		cut, peak float64
	}{
		{-20, false, -0.915149811, 0.827853703},
		{-0.1, false, -38.826238, 5.970744},
		{0, true, 0, 6.020599913},
		{0.1, false, -38.726238, 6.070744},
		{20 * math.Log10(2), false, 0, 9.542425094},
		{20, false, 19.084850189, 20.827853703},
	} {
		r := SpillCheck(1, 1, tc.gain, 10)
		if !r.Valid || r.CanCancel != tc.cancel || !near(r.CombBoostDB, tc.peak, 1e-3) || (!tc.cancel && !near(r.CombCutDB, tc.cut, 1e-3)) {
			t.Errorf("gain %+.4f dB: cancel %v, cut %.6f, peak %.6f; want %v, %.6f, %.6f", tc.gain, r.CanCancel, r.CombCutDB, r.CombBoostDB, tc.cancel, tc.cut, tc.peak)
		}
	}
}

func TestSpillCheckMinNeighbourMeetsTarget(t *testing.T) {
	for _, gain := range []float64{-6, 0, 3, 6} {
		for _, target := range SpillTargets {
			r := SpillCheck(0.5, 1, gain, target.DB)
			at := SpillCheck(0.5, r.MinNeighbour, gain, target.DB)
			if !near(at.SpillDB, -target.DB, 1e-9) || !at.MeetsTarget {
				t.Errorf("gain %v, %s: moving to %.3f gives %.3f dB, want %.3f", gain, target.Name, r.MinNeighbour, at.SpillDB, -target.DB)
			}
		}
	}
}

func TestSpillCheckUnitsAndInvalid(t *testing.T) {
	m := SpillCheck(1, 3, 6, 10)
	ft := SpillCheck(ConvertDistance(1, UnitMeters, UnitFeet), ConvertDistance(3, UnitMeters, UnitFeet), 6, 10)
	if !near(m.SpillDB, ft.SpillDB, 1e-9) || !near(ft.MinNeighbour, ConvertDistance(m.MinNeighbour, UnitMeters, UnitFeet), 1e-9) {
		t.Errorf("feet: got %+v, metres %+v", ft, m)
	}
	for _, d := range [][2]float64{{0, 3}, {1, 0}, {-1, 3}, {1, -3}, {math.NaN(), 3}, {math.Inf(1), 3}, {1, math.Inf(1)}} {
		if r := SpillCheck(d[0], d[1], 0, 10); r.Valid {
			t.Errorf("%v should be invalid: %+v", d, r)
		}
	}
	// Non-finite or overflowing gain is not a result
	for _, g := range []float64{math.NaN(), math.Inf(1), 1e6, -1e6} {
		if r := SpillCheck(1, 3, g, 10); r.Valid {
			t.Errorf("gain %v should be invalid: %+v", g, r)
		}
	}
	if r := SpillCheck(1, 3, 0, math.NaN()); r.Valid {
		t.Error("NaN target should be invalid")
	}
	if math.IsNaN(SpillCheck(1, 3, 0, 10).CombCutDB) {
		t.Error("comb cut is NaN")
	}
}

// TestSpillCheckEqualLevelsFromGain checks that a distance ratio exactly made up by the gain counts
// as equal levels, although the floating-point amplitude is off by a few units in the last place
func TestSpillCheckEqualLevelsFromGain(t *testing.T) {
	for _, tc := range []struct{ own, neighbour, gain float64 }{
		{1, 10, 20}, {0.5, 5, 20}, {3, 30, 20}, {1, 3, classicTarget}, {1, 2, 20 * math.Log10(2)},
	} {
		if r := SpillCheck(tc.own, tc.neighbour, tc.gain, 10); !r.Valid || !r.CanCancel || !near(r.CombBoostDB, 6.0206, 1e-3) {
			t.Errorf("%v m / %v m, %+.4f dB: %+v, want equal levels that can cancel", tc.own, tc.neighbour, tc.gain, r)
		}
	}
}

// TestMinimumAdviceMeetsTarget rounds every advised minimum up at the shown precision, in metres
// and feet, and checks that the advice meets the target
func TestMinimumAdviceMeetsTarget(t *testing.T) {
	for _, unit := range []string{UnitMeters, UnitFeet} {
		for _, target := range SpillTargets {
			for _, gain := range []float64{-12, -6, -0.3, 0, 0.3, 6, 12} {
				for _, own := range []float64{0.02, 0.1, 0.37, 1, 2.5} {
					r := SpillCheck(own, own*2, gain, target.DB)
					shown := CeilTo(FromMeters(r.MinNeighbour, unit), 2)
					at := SpillCheck(own, ToMeters(shown, unit), gain, target.DB)
					// The advice never undercuts the exact minimum beyond floating-point noise
					if exact := FromMeters(r.MinNeighbour, unit); !r.Valid || !at.MeetsTarget || shown < exact*(1-1e-9) {
						t.Errorf("%s, %s, gain %v, own %v: advice %v %s gives %.4f dB", unit, target.Name, gain, own, shown, unit, at.SpillDB)
					}
				}
			}
		}
	}
}
