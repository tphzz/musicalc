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
	// Spill at or above the own mic: full cancellation is possible
	loud := SpillCheck(1, 3, 12, classicTarget)
	if !near(loud.SpillDB, 2.458, 0.001) || !loud.CanCancel || loud.MeetsTarget {
		t.Errorf("+12 dB gain: got %+v", loud)
	}
	equal := SpillCheck(1, 1, 0, classicTarget)
	if !equal.CanCancel || !near(equal.CombBoostDB, 6.021, 0.001) {
		t.Errorf("equal distances: got %+v", equal)
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
	for _, d := range [][2]float64{{0, 3}, {1, 0}, {-1, 3}, {1, -3}} {
		if r := SpillCheck(d[0], d[1], 0, 10); r.Valid {
			t.Errorf("%v should be invalid: %+v", d, r)
		}
	}
	if math.IsNaN(SpillCheck(1, 3, 0, 10).CombCutDB) {
		t.Error("comb cut is NaN")
	}
}
