package logic

type TempoResult struct {
	DelayMS float64
	ModHz   float64
	Valid   bool // the tempo and note length are finite and positive
}

// GetTempoData returns the delay time of a note length (multiplier quarter notes) at bpm, and the
// modulation rate with one cycle per note. A tempo that is not a finite positive number gives an
// invalid result rather than a default tempo.
func GetTempoData(bpm float64, multiplier float64) TempoResult {
	if !positiveFinite(bpm) || !positiveFinite(multiplier) {
		return TempoResult{}
	}
	d := (60000.0 / bpm) * multiplier
	if !positiveFinite(d) {
		return TempoResult{}
	}
	return TempoResult{DelayMS: d, ModHz: 1000.0 / d, Valid: true}
}
