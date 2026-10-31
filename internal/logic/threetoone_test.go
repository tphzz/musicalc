package logic

import "testing"

func TestThreeToOne(t *testing.T) {
	cases := []struct {
		name             string
		own, other       float64
		ratio, bleed     float64
		meets            bool
		minOther, maxOwn float64
	}{
		{"design example, meets", 0.3, 1.2, 4, 12.041, true, 0.9, 0.4},
		{"design example, below", 0.3, 0.6, 2, 6.021, false, 0.9, 0.2},
		{"exactly 3:1", 0.5, 1.5, 3, 9.542, true, 1.5, 0.5},
		{"just below 3:1", 0.5, 1.49, 2.98, 9.484, false, 1.5, 0.49667},
		{"other source closer than own", 2, 1, 0.5, -6.021, false, 6, 0.33333},
		{"equal distances", 1, 1, 1, 0, false, 3, 0.33333},
		{"feet work the same", 1, 10, 10, 20, true, 3, 3.33333},
	}
	for _, tc := range cases {
		r := ThreeToOne(tc.own, tc.other)
		if !r.Valid {
			t.Errorf("%s: result not valid", tc.name)
			continue
		}
		if !near(r.Ratio, tc.ratio, 1e-6) || !near(r.BleedDB, tc.bleed, 0.001) || r.Meets != tc.meets {
			t.Errorf("%s: got ratio %.4f, bleed %.3f dB, meets %v; want %.4f, %.3f dB, %v",
				tc.name, r.Ratio, r.BleedDB, r.Meets, tc.ratio, tc.bleed, tc.meets)
		}
		if !near(r.MinOther, tc.minOther, 1e-4) || !near(r.MaxOwn, tc.maxOwn, 1e-4) {
			t.Errorf("%s: limits %.5f / %.5f, want %.5f / %.5f", tc.name, r.MinOther, r.MaxOwn, tc.minOther, tc.maxOwn)
		}
		// moving to the limits lands exactly on 3:1
		if got := ThreeToOne(tc.own, r.MinOther); !got.Meets || !near(got.Ratio, 3, 1e-9) {
			t.Errorf("%s: other source at MinOther gives %+v, want exactly 3:1", tc.name, got)
		}
		if got := ThreeToOne(r.MaxOwn, tc.other); !got.Meets || !near(got.Ratio, 3, 1e-9) {
			t.Errorf("%s: mic at MaxOwn gives %+v, want exactly 3:1", tc.name, got)
		}
	}
}

func TestThreeToOneInvalid(t *testing.T) {
	for _, d := range [][2]float64{{0, 1}, {1, 0}, {-1, 2}, {1, -2}, {0, 0}} {
		if r := ThreeToOne(d[0], d[1]); r.Valid || r.Meets {
			t.Errorf("ThreeToOne(%v, %v) = %+v, want an invalid result", d[0], d[1], r)
		}
	}
}

func TestThreeToOneBleedGrowsWithRatio(t *testing.T) {
	prev := ThreeToOne(1, 1).BleedDB
	for other := 1.1; other < 20; other += 0.1 {
		got := ThreeToOne(1, other).BleedDB
		if got <= prev {
			t.Fatalf("bleed separation fell from %v to %v dB at %v", prev, got, other)
		}
		prev = got
	}
}
