package logic

import "math"

// The 3:1 rule is about one source reaching two mics: its own mic and a neighbouring one. When
// the neighbouring mic is at least three times as far from the source, the source's spill into it
// is about 9.5 dB lower (with matched mics and gain), which keeps comb filtering shallow when the
// two mics are mixed.

// ThreeToOneRatio is the distance ratio the 3:1 guideline asks for: neighbouring mic over own mic
const ThreeToOneRatio = 3.0

// SpillTarget is how far below the source's own mic its spill into the neighbouring mic should be
type SpillTarget struct {
	Name string
	DB   float64 // dB below the own mic's contribution
}

// SpillTargets are the targets to choose from. The classic 3:1 ratio gives 9.54 dB; a full 10 dB
// needs 3.16:1.
var SpillTargets = []SpillTarget{
	{"3:1 (9.5 dB)", 20 * math.Log10(ThreeToOneRatio)},
	{"10 dB below", 10},
	{"15 dB below", 15},
}

// Geometry is how the distance ratio compares with 3:1
type Geometry int

const (
	GeometryBelow     Geometry = iota // less than 3:1
	GeometryThreshold                 // within 1 % of 3:1: meets it with no room to spare
	GeometryMeets                     // more than 3:1
)

const (
	thresholdTolerance = 0.01 // relative tolerance of "exactly 3:1"
	targetTolerance    = 0.05 // dB within which the spill counts as on target
)

// SpillResult estimates how strongly a source comes through its neighbouring mic
type SpillResult struct {
	Ratio    float64  // neighbour distance / own distance
	Geometry Geometry // how Ratio compares with 3:1

	// SpillDB is the source's direct sound in the neighbouring channel relative to its own
	// channel: 20·log10(own/neighbour) + gain. Negative is below; 0 dB or more means the
	// neighbour picks it up as loud as its own mic, or louder.
	SpillDB     float64
	TargetDB    float64 // the target, in dB below
	MeetsTarget bool    // SpillDB is at least TargetDB below (within 0.05 dB)
	// MarginDB is how far the spill is from the target: positive is room to spare, negative is
	// how much more reduction is needed
	MarginDB float64
	// MinNeighbour is the closest the neighbouring mic may be to the source while meeting the
	// target, with the other settings unchanged; in the unit of the distances
	MinNeighbour float64

	// CombBoostDB and CombCutDB bound the comb filtering when both contributions are otherwise
	// identical: mixing them boosts some frequencies and cuts others. When CanCancel is set the
	// spill is at least as loud as the own mic, so some frequencies can cancel completely and
	// CombCutDB is not used.
	CombBoostDB float64
	CombCutDB   float64
	CanCancel   bool

	Valid bool // both distances are positive
}

// SpillCheck checks one source that is own away from its mic and neighbour away from the
// neighbouring mic (both in the same unit). gainDB is how much louder the neighbouring channel is
// (gain, sensitivity, pickup); targetDB how far below the spill should be.
func SpillCheck(own, neighbour, gainDB, targetDB float64) SpillResult {
	if own <= 0 || neighbour <= 0 {
		return SpillResult{}
	}
	ratio := neighbour / own
	geometry := GeometryBelow
	switch {
	case math.Abs(ratio-ThreeToOneRatio) <= ThreeToOneRatio*thresholdTolerance:
		geometry = GeometryThreshold
	case ratio > ThreeToOneRatio:
		geometry = GeometryMeets
	}

	spill := 20*math.Log10(own/neighbour) + gainDB
	margin := -targetDB - spill
	r := SpillResult{
		Ratio:        ratio,
		Geometry:     geometry,
		SpillDB:      spill,
		TargetDB:     targetDB,
		MeetsTarget:  margin >= -targetTolerance,
		MarginDB:     margin,
		MinNeighbour: own * math.Pow(10, (targetDB+gainDB)/20),
		Valid:        true,
	}

	a := math.Pow(10, spill/20) // the spill's amplitude relative to the own mic
	r.CombBoostDB = 20 * math.Log10(1+a)
	if a >= 1 {
		r.CanCancel = true
	} else {
		r.CombCutDB = 20 * math.Log10(1-a)
	}
	return r
}
