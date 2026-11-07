package logic

import (
	"math"
	"testing"
)

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestIsFahrenheit(t *testing.T) {
	for _, u := range []string{"°F", "F", "f", " °F "} {
		if !IsFahrenheit(u) {
			t.Errorf("IsFahrenheit(%q) = false, want true", u)
		}
	}
	for _, u := range []string{"°C", "C", "", "ft", "m"} {
		if IsFahrenheit(u) {
			t.Errorf("IsFahrenheit(%q) = true, want false", u)
		}
	}
}

func TestTemperatureConversion(t *testing.T) {
	cases := []struct{ c, f float64 }{{0, 32}, {100, 212}, {22, 71.6}, {-40, -40}, {37, 98.6}}
	for _, tc := range cases {
		if got := ConvertTemperature(tc.c, UnitCelsius, UnitFahrenheit); !near(got, tc.f, 1e-9) {
			t.Errorf("%v °C → %v °F, want %v", tc.c, got, tc.f)
		}
		if got := ConvertTemperature(tc.f, UnitFahrenheit, UnitCelsius); !near(got, tc.c, 1e-9) {
			t.Errorf("%v °F → %v °C, want %v", tc.f, got, tc.c)
		}
		if got := ConvertTemperature(tc.c, UnitCelsius, UnitCelsius); got != tc.c {
			t.Errorf("same-unit conversion changed %v to %v", tc.c, got)
		}
	}
	// the old "C"/"F" labels still work
	if got := ToCelsius(71.6, "F"); !near(got, 22, 1e-9) {
		t.Errorf("ToCelsius(71.6, F) = %v, want 22", got)
	}
}

func TestDistanceConversion(t *testing.T) {
	cases := []struct{ m, ft float64 }{{0.3048, 1}, {1, 3.280839895}, {4, 13.12335958}, {0, 0}}
	for _, tc := range cases {
		if got := ConvertDistance(tc.m, UnitMeters, UnitFeet); !near(got, tc.ft, 1e-8) {
			t.Errorf("%v m → %v ft, want %v", tc.m, got, tc.ft)
		}
		if got := ConvertDistance(tc.ft, UnitFeet, UnitMeters); !near(got, tc.m, 1e-8) {
			t.Errorf("%v ft → %v m, want %v", tc.ft, got, tc.m)
		}
	}
	// a round trip keeps the value
	for _, v := range []float64{0.01, 1.5, 4, 123.456} {
		if got := ConvertDistance(ConvertDistance(v, UnitMeters, UnitFeet), UnitFeet, UnitMeters); !near(got, v, 1e-12) {
			t.Errorf("round trip of %v gave %v", v, got)
		}
	}
}

func TestSpeedOfSound(t *testing.T) {
	cases := []struct{ c, want float64 }{{0, 331.3}, {20, 343.42}, {22, 344.632}, {-10, 325.24}}
	for _, tc := range cases {
		if got := SpeedOfSound(tc.c); !near(got, tc.want, 1e-9) {
			t.Errorf("SpeedOfSound(%v) = %v, want %v", tc.c, got, tc.want)
		}
	}
}

func TestFormatTrim(t *testing.T) {
	cases := []struct {
		v    float64
		d    int
		want string
	}{
		{13.123359, 2, "13.12"}, {4, 2, "4"}, {4.5, 2, "4.5"}, {0.001, 2, "0"}, {-0.001, 2, "0"},
		{100, 0, "100"}, {1.005, 3, "1.005"}, {71.6, 2, "71.6"},
	}
	for _, tc := range cases {
		if got := FormatTrim(tc.v, tc.d); got != tc.want {
			t.Errorf("FormatTrim(%v, %d) = %q, want %q", tc.v, tc.d, got, tc.want)
		}
	}
}

// Switching the display unit must not change the physical setup, so all delays stay the same
func TestAlignmentDelayUnitSwitchKeepsDelays(t *testing.T) {
	metric := NewAlignmentDelayCalculator()
	metric.SetTemperature(22, UnitCelsius)
	metric.SetSampleRateLabel("48000")
	metric.SetReferenceDistance(4, UnitMeters)
	metric.AddMic("Kick In", 1, UnitMeters)
	metric.AddMic("Kick Out", 1.5, UnitMeters)
	metric.Recalculate()

	imperial := NewAlignmentDelayCalculator()
	imperial.SetTemperature(ConvertTemperature(22, UnitCelsius, UnitFahrenheit), UnitFahrenheit)
	imperial.SetSampleRateLabel("48000")
	imperial.SetReferenceDistance(ConvertDistance(4, UnitMeters, UnitFeet), UnitFeet)
	imperial.AddMic("Kick In", ConvertDistance(1, UnitMeters, UnitFeet), UnitFeet)
	imperial.AddMic("Kick Out", ConvertDistance(1.5, UnitMeters, UnitFeet), UnitFeet)
	imperial.Recalculate()

	want := []struct {
		ms      float64
		samples int
	}{{8.705, 418}, {7.254, 348}}
	for i, w := range want {
		for _, calc := range []*AlignmentDelayCalculator{metric, imperial} {
			m := calc.Mics[i]
			if !near(m.DelayMS, w.ms, 0.001) || m.DelaySamples != w.samples || m.IsBeyondReference {
				t.Errorf("mic %d: got %.3f ms / %d samples (beyond=%v), want %.3f ms / %d samples",
					i, m.DelayMS, m.DelaySamples, m.IsBeyondReference, w.ms, w.samples)
			}
		}
	}
}

func TestFormatTwoDecimals(t *testing.T) {
	for _, tc := range []struct {
		v    float64
		want string
	}{
		{3600, "3600"}, {3599.9999999999995, "3600"}, {82.354, "82.35"}, {1.5, "1.50"}, {-16.67, "-16.67"}, {-0.001, "0"}, {0, "0"},
	} {
		if got := FormatTwoDecimals(tc.v); got != tc.want {
			t.Errorf("FormatTwoDecimals(%v) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestCeilTo(t *testing.T) {
	for _, tc := range []struct {
		v    float64
		want float64
	}{
		{0.112468265, 0.12},
		{0.12, 0.12},
		{0.1 + 0.02, 0.12}, // 0.12000000000000001
		{3.0000000000000004, 3},
		{2.9999999999999996, 3},
		{0.120001, 0.13},
		{0.369, 0.37},
		{0, 0},
		{-0.112, -0.11},
	} {
		if got := CeilTo(tc.v, 2); !near(got, tc.want, 1e-12) {
			t.Errorf("CeilTo(%v, 2) = %v, want %v", tc.v, got, tc.want)
		}
	}
	if got := CeilTo(math.Inf(1), 2); !math.IsInf(got, 1) {
		t.Errorf("CeilTo(+Inf) = %v; callers must reject non-finite minimums before showing them", got)
	}
}
