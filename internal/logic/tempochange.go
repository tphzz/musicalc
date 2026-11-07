package logic

import "math"

// The supported tempo range of Tempo Change, for the original and the new tempo
const (
	MinTempoBPM = 5.0
	MaxTempoBPM = 999.99
)

// TempoChangeResult holds all calculated values for tempo change
type TempoChangeResult struct {
	NewTempo           float64
	TimeStretchPercent float64
	TempoVariation     float64
	Semitones          int // whole semitones down, then Cents from 0 to 99
	Cents              int
	Semitones50Cent    int // nearest semitone, then Cents50Cent from -50 to +50
	Cents50Cent        int
	Valid              bool // both tempos are within the supported range
}

// TempoInRange reports whether bpm is within the supported tempo range
func TempoInRange(bpm float64) bool {
	return bpm >= MinTempoBPM && bpm <= MaxTempoBPM
}

// CalculateFromNewTempo calculates all values when the new tempo is known. The semitones are the
// pitch change of playing the original faster or slower (varispeed); a pitch-preserving time
// stretch changes the length only.
func CalculateFromNewTempo(originalTempo, newTempo float64) TempoChangeResult {
	if !TempoInRange(originalTempo) || !TempoInRange(newTempo) {
		return TempoChangeResult{}
	}

	split := splitSemitones(12.0 * math.Log2(newTempo/originalTempo))
	return TempoChangeResult{
		NewTempo:           newTempo,
		TimeStretchPercent: originalTempo / newTempo * 100.0,
		TempoVariation:     (newTempo - originalTempo) / originalTempo * 100.0,
		Semitones:          split.semitones,
		Cents:              split.cents,
		Semitones50Cent:    split.nearest,
		Cents50Cent:        split.cents50,
		Valid:              true,
	}
}

// CalculateFromTimeStretch calculates all values when time stretch % is known
func CalculateFromTimeStretch(originalTempo, timeStretchPercent float64) TempoChangeResult {
	if !(timeStretchPercent > 0) {
		return TempoChangeResult{}
	}
	return CalculateFromNewTempo(originalTempo, originalTempo/(timeStretchPercent/100.0))
}

// CalculateFromTranspose calculates all values when transpose semitones/cents are known
func CalculateFromTranspose(originalTempo float64, semitones, cents int) TempoChangeResult {
	totalSemitones := float64(semitones) + float64(cents)/100.0
	return CalculateFromNewTempo(originalTempo, originalTempo*math.Pow(2.0, totalSemitones/12.0))
}

// NormalizeTranspose wraps cents into -50 to +50 by moving whole semitones, then limits the shift
// so the new tempo stays within the supported range. A limited shift is written as whole semitones
// down plus 0 to 99 cents, rounded towards the inside of the range. An original tempo outside the
// range is left for CalculateFromTranspose to reject.
func NormalizeTranspose(originalTempo float64, semitones, cents int) (int, int) {
	if cents > 50 {
		k := (cents - 50 + 99) / 100
		semitones, cents = semitones+k, cents-100*k
	} else if cents < -50 {
		k := (-50 - cents + 99) / 100
		semitones, cents = semitones-k, cents+100*k
	}
	if !TempoInRange(originalTempo) {
		return semitones, cents
	}

	total := semitones*100 + cents
	maxCents := int(math.Floor(1200 * math.Log2(MaxTempoBPM/originalTempo)))
	minCents := int(math.Ceil(1200 * math.Log2(MinTempoBPM/originalTempo)))
	switch {
	case total > maxCents:
		total = maxCents
	case total < minCents:
		total = minCents
	default:
		return semitones, cents
	}
	down := floorDiv(total, 100)
	return down, total - down*100
}
