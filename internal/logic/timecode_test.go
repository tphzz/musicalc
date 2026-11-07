package logic

import (
	"math"
	"testing"
)

func TestTimecodeNonDropFrame(t *testing.T) {
	f30 := GetFPSFormat("30 fps")
	f25 := GetFPSFormat("25 fps")
	for _, tc := range []struct {
		format     FPSFormat
		h, m, s, f int
		frames     int
	}{
		{f30, 0, 0, 1, 15, 45},
		{f30, 0, 1, 0, 0, 1800},
		{f25, 0, 1, 0, 0, 1500},
		{f30, 1, 0, 0, 0, 108000},
		{GetFPSFormat("23.976 fps"), 0, 0, 1, 0, 24}, // timecode counts the nominal 24 frames
		{GetFPSFormat("59.94 fps"), 0, 0, 1, 0, 60},
	} {
		if got := TimecodeToFrames(tc.h, tc.m, tc.s, tc.f, tc.format); got != tc.frames {
			t.Errorf("%s %02d:%02d:%02d:%02d = %d frames, want %d", tc.format.Name, tc.h, tc.m, tc.s, tc.f, got, tc.frames)
		}
	}
	if got := AddTimecodes(0, 0, 1, 0, 0, 0, 0, 15, f30); got.Timecode != "00:00:01:15" || got.TotalFrames != 45 {
		t.Errorf("00:00:01:00 + 00:00:00:15 = %+v", got)
	}
	// A larger subtrahend gives zero
	if got := SubtractTimecodes(0, 0, 1, 0, 0, 0, 2, 0, f30); got.Timecode != "00:00:00:00" || got.TotalFrames != 0 {
		t.Errorf("1 s − 2 s = %+v", got)
	}
	// Components at or above their range carry, by decision: 30 frames at 25 fps are 1 s 5 f
	if got := FramesToTimecode(TimecodeToFrames(0, 0, 0, 30, f25), f25).Timecode; got != "00:00:01:05" {
		t.Errorf("00:00:00:30 at 25 fps = %s, want 00:00:01:05", got)
	}
	// 29.94 is an alias of 29.97 non-drop
	if a, b := GetFPSFormat("29.94 fps"), GetFPSFormat("29.97 fps"); a.FPS != b.FPS || a.DropFrame || b.DropFrame ||
		math.Abs(a.FPS-30000.0/1001) > 1e-12 {
		t.Errorf("29.94 %+v, 29.97 %+v", a, b)
	}
	if GetFPSFormat("no such rate").Name != "30 fps" {
		t.Error("an unknown rate should fall back to 30 fps")
	}
}

// TestTimecodeDropFrame checks 29.97 drop frame: two frame numbers are skipped every minute except
// every tenth, every frame of 25 hours converts back, and no skipped label is produced
func TestTimecodeDropFrame(t *testing.T) {
	df := GetFPSFormat("29.97 fps (df)")
	if !df.DropFrame {
		t.Fatal("29.97 df is not drop frame")
	}
	for _, tc := range []struct {
		h, m, s, f int
		frames     int
	}{
		{0, 0, 59, 29, 1799},
		{0, 1, 0, 2, 1800}, // 00:01:00;00 and ;01 do not exist
		{0, 9, 59, 29, 17981},
		{0, 10, 0, 0, 17982}, // every tenth minute keeps them
		{1, 0, 0, 0, 107892},
		{23, 59, 59, 29, 2589407},
	} {
		if got := TimecodeToFrames(tc.h, tc.m, tc.s, tc.f, df); got != tc.frames {
			t.Errorf("%02d:%02d:%02d;%02d = %d frames, want %d", tc.h, tc.m, tc.s, tc.f, got, tc.frames)
		}
		if got := FramesToTimecode(tc.frames, df); got.Hours != tc.h || got.Minutes != tc.m || got.Seconds != tc.s || got.Frames != tc.f {
			t.Errorf("frame %d = %s", tc.frames, got.Timecode)
		}
	}
	for n := 0; n < 30*3600*25; n++ {
		tc := FramesToTimecode(n, df)
		if tc.Minutes%10 != 0 && tc.Seconds == 0 && tc.Frames < 2 {
			t.Fatalf("frame %d is the skipped label %s", n, tc.Timecode)
		}
		if back := TimecodeToFrames(tc.Hours, tc.Minutes, tc.Seconds, tc.Frames, df); back != n {
			t.Fatalf("frame %d → %s → %d", n, tc.Timecode, back)
		}
	}
	// A skipped label is accepted and counts as the frame two numbers earlier, by decision
	if got := FramesToTimecode(TimecodeToFrames(0, 1, 0, 0, df), df).Timecode; got != "00:00:59:28" {
		t.Errorf("00:01:00;00 = %s, want 00:00:59:28", got)
	}
}

func TestTempoToDelay(t *testing.T) {
	for _, tc := range []struct {
		bpm, mult, ms, hz float64
	}{
		{120, 1, 500, 2},
		{120, 1.5, 750, 1.333333333},
		{120, 2.0 / 3.0, 333.333333333, 3},
		{120, 4, 2000, 0.5},
		{90, 0.25, 166.666666667, 6},
	} {
		r := GetTempoData(tc.bpm, tc.mult)
		if !r.Valid || !near(r.DelayMS, tc.ms, 1e-6) || !near(r.ModHz, tc.hz, 1e-6) {
			t.Errorf("%v BPM × %v: %+v, want %v ms / %v Hz", tc.bpm, tc.mult, r, tc.ms, tc.hz)
		}
	}
	for _, bpm := range []float64{0, -120, math.NaN(), math.Inf(1)} {
		if r := GetTempoData(bpm, 1); r.Valid {
			t.Errorf("%v BPM gave %+v, want invalid rather than a default tempo", bpm, r)
		}
	}
	if r := GetTempoData(1e-320, 1); r.Valid {
		t.Errorf("a tempo so small the delay overflows gave %+v", r)
	}
}
