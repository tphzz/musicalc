package logic

import (
	"math"
	"time"
)

const (
	// TapResetAfter is the pause after which the next tap starts a new measurement
	TapResetAfter = 2 * time.Second
	// tapMinInterval ignores taps closer than this (above 400 BPM) as accidental double taps
	tapMinInterval = 150 * time.Millisecond
	// tapMaxIntervals is how many recent intervals the tempo is averaged over
	tapMaxIntervals = 8
)

// TapResult is the tempo measured from the taps so far
type TapResult struct {
	Taps   int     // taps in the current measurement
	BPM    float64 // 60000 / mean interval in ms; valid from the second tap
	Spread float64 // standard deviation of the per-interval tempos, in BPM
	Valid  bool    // at least two taps, so BPM and Spread are meaningful
}

// TapTempo measures a tempo from taps along with the beat
type TapTempo struct {
	taps  []time.Time // recent taps, at most tapMaxIntervals+1
	count int         // taps in the current measurement, including ones dropped from taps
}

// Tap records a tap at the given time and returns the updated measurement
func (t *TapTempo) Tap(now time.Time) TapResult {
	if n := len(t.taps); n > 0 {
		gap := now.Sub(t.taps[n-1])
		switch {
		case gap > TapResetAfter:
			t.Reset()
		case gap < tapMinInterval:
			return t.Result()
		}
	}
	t.taps = append(t.taps, now)
	if len(t.taps) > tapMaxIntervals+1 {
		t.taps = t.taps[len(t.taps)-tapMaxIntervals-1:]
	}
	t.count++
	return t.Result()
}

// Reset forgets all taps
func (t *TapTempo) Reset() {
	t.taps = t.taps[:0]
	t.count = 0
}

// Active reports whether the measurement is still running, i.e. the last tap was recent enough
// that the next one continues it
func (t *TapTempo) Active(now time.Time) bool {
	n := len(t.taps)
	return n > 0 && now.Sub(t.taps[n-1]) <= TapResetAfter
}

// Result returns the current measurement
func (t *TapTempo) Result() TapResult {
	r := TapResult{Taps: t.count}
	if len(t.taps) < 2 {
		return r
	}

	intervals := len(t.taps) - 1
	total := t.taps[intervals].Sub(t.taps[0])
	r.BPM = 60000 / (float64(total.Milliseconds()) / float64(intervals))
	r.Valid = true

	var sumSq float64
	for i := 1; i < len(t.taps); i++ {
		bpm := 60000 / float64(t.taps[i].Sub(t.taps[i-1]).Milliseconds())
		sumSq += (bpm - r.BPM) * (bpm - r.BPM)
	}
	r.Spread = math.Sqrt(sumSq / float64(intervals))
	return r
}
