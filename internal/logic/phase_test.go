package logic

import (
	"math"
	"testing"
)

// λ of a 60 Hz kick fundamental at 22 °C: 344.632 / 60
const kickLambda = 5.743866666666667

func TestWavelength(t *testing.T) {
	cases := []struct{ f, c, want float64 }{
		{60, 22, kickLambda},
		{1000, 20, 0.34342},
		{343.42, 20, 1},
		{0, 22, 0},
		{-50, 22, 0},
	}
	for _, tc := range cases {
		if got := Wavelength(tc.f, tc.c); !near(got, tc.want, 1e-9) {
			t.Errorf("Wavelength(%v Hz, %v °C) = %v, want %v", tc.f, tc.c, got, tc.want)
		}
	}
}

func TestPhaseOffset(t *testing.T) {
	l := kickLambda
	cases := []struct {
		name string
		diff float64
		want float64
	}{
		{"same distance", 0, 0},
		{"sixth of a wavelength", l / 6, 60},
		{"quarter wavelength", l / 4, 90},
		{"third of a wavelength", l / 3, 120},
		{"half wavelength", l / 2, 180},
		{"three quarters fold back", 0.75 * l, 90},
		{"full wavelength", l, 0},
		{"one and a half wavelengths", 1.5 * l, 180},
		{"kick example 0.8 m", 0.8, 50.141},
		{"kick example 2.6 m", 2.6, 162.956},
		{"negative difference", -0.8, 50.141},
	}
	for _, tc := range cases {
		got := PhaseOffset(tc.diff, l)
		if !near(got, tc.want, 0.001) && !(tc.want == 0 && near(got, 360, 0.001)) {
			t.Errorf("%s: PhaseOffset(%v) = %.4f°, want %.4f°", tc.name, tc.diff, got, tc.want)
		}
		if got < 0 || got > 180+1e-9 {
			t.Errorf("%s: phase %.4f° is outside 0–180°", tc.name, got)
		}
	}
	if got := PhaseOffset(1, 0); got != 0 {
		t.Errorf("PhaseOffset with zero wavelength = %v, want 0", got)
	}
}

func TestPhaseOffsetIsPeriodicAndSymmetric(t *testing.T) {
	l := 1.7
	for d := 0.0; d < 2*l; d += 0.037 {
		base := PhaseOffset(d, l)
		if got := PhaseOffset(d+l, l); !near(got, base, 1e-6) {
			t.Fatalf("PhaseOffset(%v + λ) = %v, want %v (one wavelength later)", d, got, base)
		}
		if got := PhaseOffset(-d, l); !near(got, base, 1e-9) {
			t.Fatalf("PhaseOffset(-%v) = %v, want %v (sign doesn't matter)", d, got, base)
		}
	}
}

func TestZoneForPhase(t *testing.T) {
	cases := []struct {
		deg  float64
		want PhaseZone
	}{
		{0, PhaseSweetSpot}, {59.99, PhaseSweetSpot},
		{60, PhasePartial}, {90, PhasePartial}, {120, PhasePartial},
		{120.01, PhaseDeathZone}, {180, PhaseDeathZone},
	}
	for _, tc := range cases {
		if got := ZoneForPhase(tc.deg); got != tc.want {
			t.Errorf("ZoneForPhase(%v) = %v, want %v", tc.deg, got, tc.want)
		}
	}
}

func TestSummedLevelDB(t *testing.T) {
	cases := []struct{ deg, want float64 }{
		{0, 0}, {50.141, -0.862}, {60, -1.249}, {90, -3.010}, {120, -6.021}, {162.956, -16.58},
	}
	for _, tc := range cases {
		if got := SummedLevelDB(tc.deg); !near(got, tc.want, 0.01) {
			t.Errorf("SummedLevelDB(%v°) = %.3f dB, want %.3f dB", tc.deg, got, tc.want)
		}
	}
	if got := SummedLevelDB(180); !math.IsInf(got, -1) && got > -200 {
		t.Errorf("SummedLevelDB(180°) = %v, want complete cancellation", got)
	}
	// the level only drops as the phase grows
	prev := 1.0
	for deg := 0.0; deg < 180; deg += 5 {
		got := SummedLevelDB(deg)
		if got > prev+1e-12 {
			t.Fatalf("SummedLevelDB rises from %v to %v at %v°", prev, got, deg)
		}
		prev = got
	}
}

func TestCheckPhase(t *testing.T) {
	r := CheckPhase(0.8, 60, 22)
	if !r.Valid || !near(r.Wavelength, kickLambda, 1e-9) || r.Zone != PhaseSweetSpot || !near(r.Phase, 50.141, 0.001) {
		t.Errorf("CheckPhase(0.8 m, 60 Hz, 22 °C) = %+v, want a 50° sweet spot", r)
	}
	r = CheckPhase(2.6, 60, 22)
	if r.Zone != PhaseDeathZone || !near(r.SummedDB, -16.58, 0.01) {
		t.Errorf("CheckPhase(2.6 m) = %+v, want a death zone at about -16.6 dB", r)
	}
	r = CheckPhase(1.2, 60, 22) // 75°
	if r.Zone != PhasePartial {
		t.Errorf("CheckPhase(1.2 m) zone = %v, want partial", r.Zone)
	}
	if r := CheckPhase(1, 0, 22); r.Valid {
		t.Errorf("CheckPhase with 0 Hz = %+v, want invalid", r)
	}
}

func TestSweetSpotsAndDeathZones(t *testing.T) {
	l := kickLambda
	sweet := SweetSpots(l, 12)
	wantSweet := []float64{0, l, 2 * l}
	if len(sweet) != len(wantSweet) {
		t.Fatalf("SweetSpots = %v, want %v", sweet, wantSweet)
	}
	for i := range sweet {
		if !near(sweet[i], wantSweet[i], 1e-9) {
			t.Errorf("sweet spot %d = %v, want %v", i, sweet[i], wantSweet[i])
		}
	}
	death := DeathZones(l, 12)
	wantDeath := []float64{l / 2, 1.5 * l}
	if len(death) != len(wantDeath) {
		t.Fatalf("DeathZones = %v, want %v", death, wantDeath)
	}
	for i := range death {
		if !near(death[i], wantDeath[i], 1e-9) {
			t.Errorf("death zone %d = %v, want %v", i, death[i], wantDeath[i])
		}
	}
	// every listed distance really is in phase / opposed
	for _, d := range sweet {
		if p := PhaseOffset(d, l); !near(p, 0, 1e-6) && !near(p, 360, 1e-6) {
			t.Errorf("sweet spot %v has phase %v°", d, p)
		}
	}
	for _, d := range death {
		if p := PhaseOffset(d, l); !near(p, 180, 1e-6) {
			t.Errorf("death zone %v has phase %v°", d, p)
		}
	}
	if got := SweetSpots(0, 10); got != nil {
		t.Errorf("SweetSpots with zero wavelength = %v, want none", got)
	}
	if got := SafeTolerance(l); !near(got, l/6, 1e-12) || !near(got, 0.957, 0.001) {
		t.Errorf("SafeTolerance = %v, want λ/6 ≈ 0.957 m", got)
	}
}

func TestZoneSegments(t *testing.T) {
	for _, tc := range []struct{ lambda, max float64 }{{kickLambda, 12}, {1.7, 5}, {2, 1}, {0.3, 2.95}} {
		segs := ZoneSegments(tc.lambda, tc.max)
		if len(segs) == 0 {
			t.Fatalf("no segments for λ=%v", tc.lambda)
		}
		if segs[0].Start != 0 || !near(segs[len(segs)-1].End, tc.max, 1e-9) {
			t.Errorf("λ=%v: segments cover %v–%v, want 0–%v", tc.lambda, segs[0].Start, segs[len(segs)-1].End, tc.max)
		}
		for i, s := range segs {
			if s.End <= s.Start {
				t.Errorf("λ=%v: empty segment %+v", tc.lambda, s)
			}
			if i > 0 {
				if !near(s.Start, segs[i-1].End, 1e-9) {
					t.Errorf("λ=%v: gap between %+v and %+v", tc.lambda, segs[i-1], s)
				}
				if s.Zone == segs[i-1].Zone {
					t.Errorf("λ=%v: neighbouring segments share zone %v", tc.lambda, s.Zone)
				}
			}
			// the zone of a segment matches the phase anywhere inside it
			for _, f := range []float64{0.1, 0.5, 0.9} {
				d := s.Start + f*(s.End-s.Start)
				if got := ZoneForPhase(PhaseOffset(d, tc.lambda)); got != s.Zone {
					t.Errorf("λ=%v: %.3f m is in segment zone %v but has zone %v", tc.lambda, d, s.Zone, got)
				}
			}
		}
	}
	// the 60 Hz kick map from the design: green, amber, red, amber, green, amber, red, amber, green
	segs := ZoneSegments(kickLambda, 2*kickLambda)
	want := []PhaseZone{PhaseSweetSpot, PhasePartial, PhaseDeathZone, PhasePartial, PhaseSweetSpot, PhasePartial, PhaseDeathZone, PhasePartial, PhaseSweetSpot}
	if len(segs) != len(want) {
		t.Fatalf("got %d segments, want %d: %+v", len(segs), len(want), segs)
	}
	for i := range want {
		if segs[i].Zone != want[i] {
			t.Errorf("segment %d zone = %v, want %v", i, segs[i].Zone, want[i])
		}
	}
	if !near(segs[0].End, kickLambda/6, 1e-9) || !near(segs[2].Start, kickLambda/3, 1e-9) || !near(segs[2].End, 2*kickLambda/3, 1e-9) {
		t.Errorf("zone boundaries are off: %+v", segs[:3])
	}
	if ZoneSegments(0, 5) != nil || ZoneSegments(1, 0) != nil {
		t.Error("ZoneSegments should return nothing for a zero wavelength or range")
	}
}
