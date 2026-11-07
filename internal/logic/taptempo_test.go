package logic

import (
	"testing"
	"time"
)

// tapAt taps at the given millisecond offsets from a fixed start and returns the last result
func tapAt(t *TapTempo, offsetsMs ...int) TapResult {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var r TapResult
	for _, ms := range offsetsMs {
		r = t.Tap(start.Add(time.Duration(ms) * time.Millisecond))
	}
	return r
}

func TestTapTempoNeedsTwoTaps(t *testing.T) {
	var tt TapTempo
	r := tapAt(&tt, 0)
	if r.Valid || r.Taps != 1 {
		t.Fatalf("after one tap got %+v, want 1 tap and not valid", r)
	}
	r = tapAt(&tt, 500)
	if !r.Valid || r.Taps != 2 || !near(r.BPM, 120, 0.001) {
		t.Fatalf("after two taps 500 ms apart got %+v, want 120 BPM", r)
	}
}

func TestTapTempoSteadyTaps(t *testing.T) {
	var tt TapTempo
	r := tapAt(&tt, 0, 500, 1000, 1500, 2000, 2500)
	if !near(r.BPM, 120, 0.001) || !(r.Spread <= 0.001) || r.Taps != 6 {
		t.Fatalf("steady 500 ms taps got %+v, want 120 BPM, no spread, 6 taps", r)
	}
}

func TestTapTempoSpread(t *testing.T) {
	var tt TapTempo
	r := tapAt(&tt, 0, 480, 1000, 1490, 2000)
	if !near(r.BPM, 120, 0.001) {
		t.Fatalf("got %.3f BPM, want 120 (mean interval 500 ms)", r.BPM)
	}
	if !(r.Spread >= 1 && r.Spread <= 6) {
		t.Fatalf("got spread %.3f BPM, want a few BPM for ±20 ms jitter", r.Spread)
	}
}

func TestTapTempoPauseRestarts(t *testing.T) {
	var tt TapTempo
	tapAt(&tt, 0, 600, 1200)
	r := tapAt(&tt, 3700) // 2.5 s after the last tap
	if r.Valid || r.Taps != 1 {
		t.Fatalf("after a 2.5 s pause got %+v, want a new measurement with 1 tap", r)
	}
	r = tapAt(&tt, 4200)
	if !near(r.BPM, 120, 0.001) {
		t.Fatalf("restarted measurement got %.3f BPM, want 120", r.BPM)
	}
}

func TestTapTempoAveragesLastEightIntervals(t *testing.T) {
	var tt TapTempo
	// Four slow taps (1000 ms apart), then nine fast ones (500 ms): only the last 8 intervals count
	r := tapAt(&tt, 0, 1000, 2000, 3000, 3500, 4000, 4500, 5000, 5500, 6000, 6500, 7000, 7500)
	if !near(r.BPM, 120, 0.001) {
		t.Fatalf("got %.3f BPM, want 120 from the last 8 intervals", r.BPM)
	}
	if r.Taps != 13 {
		t.Fatalf("got %d taps, want all 13 counted", r.Taps)
	}
}

func TestTapTempoIgnoresDoubleTaps(t *testing.T) {
	var tt TapTempo
	r := tapAt(&tt, 0, 100, 500, 1000)
	if !near(r.BPM, 120, 0.001) || r.Taps != 3 {
		t.Fatalf("got %+v, want the 100 ms double tap ignored: 3 taps at 120 BPM", r)
	}
}

func TestTapTempoActiveAndReset(t *testing.T) {
	var tt TapTempo
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	tt.Tap(start)
	if !tt.Active(start.Add(time.Second)) || tt.Active(start.Add(3*time.Second)) {
		t.Fatal("measurement should stay active for 2 s after the last tap only")
	}
	tt.Reset()
	if r := tt.Result(); r.Taps != 0 || r.Valid {
		t.Fatalf("after Reset got %+v, want no taps", r)
	}
}

// TestTapTempoSubMillisecond checks that intervals are not cut to whole milliseconds, which
// biased the tempo upwards
func TestTapTempoSubMillisecond(t *testing.T) {
	var tt TapTempo
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var r TapResult
	for i := 0; i < 9; i++ {
		r = tt.Tap(start.Add(time.Duration(i) * 500600 * time.Microsecond))
	}
	if want := 60 / 0.5006; !r.Valid || !near(r.BPM, want, 1e-9) || !(r.Spread <= 1e-9) {
		t.Fatalf("taps 500.6 ms apart got %+v, want %.6f BPM and no spread", r, want)
	}
}
