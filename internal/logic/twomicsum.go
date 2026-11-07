package logic

import (
	"math"
	"math/cmplx"
	"sort"
)

// The Two-Mic Sum model: the same signal reaches Mic A and, later, Mic B. Mixing them adds a
// delayed copy, which reinforces some frequencies and cancels others (comb filtering). Levels are
// relative to Mic A alone. Room reflections and the mics' own responses are not modelled.

// TwoMicSum describes how Mic B joins Mic A in the mix
type TwoMicSum struct {
	Delay    float64 // seconds Mic B arrives after Mic A in the mix (negative: before)
	Gain     float64 // Mic B's level relative to Mic A, as an amplitude ratio (1 = equal)
	Inverted bool    // Mic B's polarity is inverted
}

// The band of interest the response graph accepts, in Hz
const (
	MinBandHz = 10.0
	MaxBandHz = 20000.0
)

// BandInRange reports whether lo to hi is an accepted band of interest: MinBandHz <= lo < hi <=
// MaxBandHz
func BandInRange(lo, hi float64) bool {
	return lo >= MinBandHz && lo < hi && hi <= MaxBandHz
}

// ArrivalDelay is the time sound needs for an extra path of meters at celsius, in seconds. ok is
// false for a non-finite path or a temperature outside the accepted range, so an impossible setup
// does not turn into a plausible arrival.
func ArrivalDelay(meters, celsius float64) (seconds float64, ok bool) {
	if math.IsNaN(meters) || math.IsInf(meters, 0) || !TemperatureInRange(celsius) {
		return 0, false
	}
	return meters / SpeedOfSound(celsius), true
}

// DistanceForDelay is the extra path in metres that takes seconds at celsius; ok as for ArrivalDelay
func DistanceForDelay(seconds, celsius float64) (meters float64, ok bool) {
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) || !TemperatureInRange(celsius) {
		return 0, false
	}
	return seconds * SpeedOfSound(celsius), true
}

// DBToGain converts a level in dB to an amplitude ratio
func DBToGain(db float64) float64 {
	return math.Pow(10, db/20)
}

// Valid reports whether the model can be evaluated: a finite delay and a finite positive gain
func (m TwoMicSum) Valid() bool {
	return !math.IsNaN(m.Delay) && !math.IsInf(m.Delay, 0) && m.Gain > 0 && !math.IsInf(m.Gain, 1)
}

// sign is +1 for normal and −1 for inverted polarity
func (m TwoMicSum) sign() float64 {
	if m.Inverted {
		return -1
	}
	return 1
}

// MinLevelDB is the floor that stands for complete cancellation, so levels stay finite. It is a
// numerical convention, not a measurable level: anything at the floor cancels at least 120 dB.
const MinLevelDB = -120.0

// LevelDB is the combined level at freq, in dB relative to Mic A alone
func (m TwoMicSum) LevelDB(freq float64) float64 {
	sum := 1 + complex(m.Gain*m.sign(), 0)*cmplx.Exp(complex(0, -2*math.Pi*freq*m.Delay))
	a := cmplx.Abs(sum)
	if a <= 0 {
		return MinLevelDB
	}
	return math.Max(20*math.Log10(a), MinLevelDB)
}

// PeakDB and DipDB are the highest and lowest combined levels the comb reaches
func (m TwoMicSum) PeakDB() float64 { return 20 * math.Log10(1+m.Gain) }

func (m TwoMicSum) DipDB() float64 {
	d := math.Abs(1 - m.Gain)
	if d <= 0 {
		return MinLevelDB
	}
	return math.Max(20*math.Log10(d), MinLevelDB)
}

// Aligned reports whether the two arrive exactly together, so there is no comb. Even a residual
// of a few microseconds leaves a comb that matters at high frequencies, so no tolerance is used.
func (m TwoMicSum) Aligned() bool {
	return m.Delay == 0
}

// combFrequencies returns the first n frequencies above from (and above zero) where the comb has
// its dips (dips=true) or peaks. With normal polarity dips sit at odd multiples of 1/(2·delay) and
// peaks at multiples of 1/delay; inverting the polarity swaps them.
// The loop is bounded by the result count: every step from the first index on gives a frequency at
// or above from, so at most a few extra steps are needed for floating-point rounding. Invalid
// input gives no frequencies instead of looping.
func (m TwoMicSum) combFrequencies(from float64, n int, dips bool) []float64 {
	if !m.Valid() || m.Aligned() || n <= 0 || !(from >= 0) || math.IsInf(from, 1) {
		return nil
	}
	t := math.Abs(m.Delay)
	offset := 0.0
	if dips != m.Inverted { // normal dips and inverted peaks sit at the half steps
		offset = 0.5
	}
	k := math.Max(0, math.Ceil(from*t-offset))
	out := make([]float64, 0, n)
	last := math.Inf(-1)
	for step := 0; len(out) < n && step < n+4; step, k = step+1, k+1 {
		if k+1 == k {
			break // the index no longer advances: the next values would repeat
		}
		if f := (k + offset) / t; f > last && f > 0 && f >= from && !math.IsInf(f, 1) {
			out = append(out, f)
			last = f
		}
	}
	return out
}

// Dips returns the first n frequencies where Mic B cancels Mic A most
func (m TwoMicSum) Dips(n int) []float64 { return m.combFrequencies(0, n, true) }

// Peaks returns the first n frequencies where Mic B reinforces Mic A most
func (m TwoMicSum) Peaks(n int) []float64 { return m.combFrequencies(0, n, false) }

// DipsFrom and PeaksFrom return the first n dips or peaks at or above from Hz
func (m TwoMicSum) DipsFrom(from float64, n int) []float64  { return m.combFrequencies(from, n, true) }
func (m TwoMicSum) PeaksFrom(from float64, n int) []float64 { return m.combFrequencies(from, n, false) }

// CombSpacing is the distance in Hz between neighbouring dips (or peaks); 0 when aligned
func (m TwoMicSum) CombSpacing() float64 {
	if m.Aligned() || !m.Valid() {
		return 0
	}
	return 1 / math.Abs(m.Delay)
}

// RangeDB returns the lowest and highest combined level between lo and hi Hz (either order; a
// single frequency when they are equal). lo and hi must be positive and finite.
func (m TwoMicSum) RangeDB(lo, hi float64) (min, max float64) {
	if hi < lo {
		lo, hi = hi, lo
	}
	if lo == hi {
		v := m.LevelDB(lo)
		return v, v
	}
	freqs := SampleFrequencies(lo, hi, 400, m)
	min, max = math.Inf(1), math.Inf(-1)
	for _, f := range freqs {
		v := m.LevelDB(f)
		min, max = math.Min(min, v), math.Max(max, v)
	}
	return min, max
}

// SampleFrequencies returns n+1 frequencies spread evenly on a log scale from lo to hi, plus the
// exact dips and peaks of the comb in between, so plots and ranges catch the full depth
func SampleFrequencies(lo, hi float64, n int, m TwoMicSum) []float64 {
	if !(lo > 0) || !(hi > lo) || math.IsInf(hi, 1) || n < 1 {
		return nil
	}
	freqs := make([]float64, 0, n+64)
	for i := 0; i <= n; i++ {
		freqs = append(freqs, lo*math.Pow(hi/lo, float64(i)/float64(n)))
	}
	if m.Valid() && !m.Aligned() {
		step := 0.5 / math.Abs(m.Delay) // dips and peaks alternate every half step
		first := math.Ceil(lo/step) * step
		for f, added := first, 0; f <= hi && added < 2000 && f+step > f; f, added = f+step, added+1 {
			freqs = append(freqs, f)
		}
	}
	sort.Float64s(freqs)
	return freqs
}

// GraphRange is the frequency range a response graph shows for a range of interest: wide enough to
// show the comb around it, limited to the audible range
func GraphRange(lo, hi float64) (float64, float64) {
	if lo <= 0 || hi <= lo {
		return 20, 1000
	}
	gLo := math.Max(10, lo/2.5)
	gHi := math.Min(20000, hi*10)
	if gHi <= gLo*2 { // keep at least an octave
		gHi = math.Min(20000, gLo*2)
	}
	return gLo, gHi
}
