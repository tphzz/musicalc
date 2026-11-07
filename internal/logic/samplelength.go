package logic

import "math"

type SamplerResult struct {
	Samples int
	MS      float64
	Valid   bool // rate, tempo and beats are finite and positive and the length fits
}

// maxSamples bounds sample counts (about six days at 192 kHz) so the relative snapping tolerance
// of SamplesIn stays far below one sample
const maxSamples = 1e11

// sampleSnap is the relative distance within which a count is taken as an exact whole number
const sampleSnap = 1e-12

// GetSampleData returns the length of beats at bpm, in milliseconds and in whole samples at rate
func GetSampleData(rate, bpm, beats float64) SamplerResult {
	if !positiveFinite(rate) || !positiveFinite(bpm) || !positiveFinite(beats) {
		return SamplerResult{}
	}
	ms := 60000.0 * beats / bpm
	samples, ok := SamplesIn(ms, rate)
	if !ok {
		return SamplerResult{}
	}
	return SamplerResult{Samples: samples, MS: ms, Valid: true}
}

// SamplesIn returns how many whole samples at rate fit in ms milliseconds. A count within a
// relative 1e-12 of a whole number is that number, so an exact count such as 3.6 s at 44.1 kHz =
// 158760 does not lose a sample to floating-point error; any other count is rounded down.
func SamplesIn(ms, rate float64) (int, bool) {
	if !positiveFinite(ms) || !positiveFinite(rate) {
		return 0, false
	}
	exact := ms * rate / 1000.0
	if !(exact < maxSamples) {
		return 0, false
	}
	if whole := math.Round(exact); math.Abs(exact-whole) <= exact*sampleSnap {
		return int(whole), true
	}
	return int(math.Floor(exact)), true
}

func positiveFinite(v float64) bool {
	return v > 0 && !math.IsInf(v, 1)
}
