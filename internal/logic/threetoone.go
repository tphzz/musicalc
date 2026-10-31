package logic

import "math"

// ThreeToOneRatio is the minimum ratio of the distance to another source over the distance to a
// mic's own source that the 3:1 rule asks for
const ThreeToOneRatio = 3.0

// ThreeToOneResult checks one mic against the 3:1 rule
type ThreeToOneResult struct {
	Ratio   float64 // distance to the other source / distance to the own source
	BleedDB float64 // how much quieter the other source arrives, assuming equally loud sources
	Meets   bool    // Ratio is at least ThreeToOneRatio

	MinOther float64 // the closest the other source may be for the rule to hold
	MaxOwn   float64 // the farthest the mic may be from its own source for the rule to hold
	Valid    bool    // both distances were positive
}

// ThreeToOne checks a mic that is own metres (or feet) from its source and other from another
// source. Both distances must use the same unit; the limits are returned in it too.
func ThreeToOne(own, other float64) ThreeToOneResult {
	if own <= 0 || other <= 0 {
		return ThreeToOneResult{}
	}
	ratio := other / own
	return ThreeToOneResult{
		Ratio:    ratio,
		BleedDB:  20 * math.Log10(ratio),
		Meets:    ratio >= ThreeToOneRatio-1e-9,
		MinOther: own * ThreeToOneRatio,
		MaxOwn:   other / ThreeToOneRatio,
		Valid:    true,
	}
}
