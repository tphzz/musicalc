package logic

import "math"

// PhaseZone classifies how two mics at different distances from a source combine at a frequency
type PhaseZone int

const (
	// PhaseSweetSpot: less than PhaseSafeDegrees apart, the signals reinforce each other
	PhaseSweetSpot PhaseZone = iota
	// PhasePartial: between PhaseSafeDegrees and PhaseDeathDegrees, partial cancellation
	PhasePartial
	// PhaseDeathZone: more than PhaseDeathDegrees apart, the frequency largely cancels
	PhaseDeathZone
)

const (
	// PhaseSafeDegrees is the largest phase offset that still counts as a sweet spot
	PhaseSafeDegrees = 60.0
	// PhaseDeathDegrees is the phase offset above which the result is a death zone
	PhaseDeathDegrees = 120.0
)

// Wavelength returns the wavelength in metres of a frequency at the given air temperature in °C,
// or 0 for a frequency that isn't positive
func Wavelength(freqHz, celsius float64) float64 {
	if freqHz <= 0 {
		return 0
	}
	return SpeedOfSound(celsius) / freqHz
}

// PhaseOffset returns the phase offset in degrees, folded to 0–180, between two mics whose
// distances to the source differ by pathDiff metres. 0° means in phase, 180° fully opposed.
func PhaseOffset(pathDiff, wavelength float64) float64 {
	if wavelength <= 0 {
		return 0
	}
	deg := math.Mod(math.Abs(pathDiff)/wavelength*360, 360)
	if deg > 180 {
		deg = 360 - deg
	}
	return deg
}

// ZoneForPhase classifies a phase offset in degrees (0–180)
func ZoneForPhase(deg float64) PhaseZone {
	switch {
	case deg < PhaseSafeDegrees:
		return PhaseSweetSpot
	case deg <= PhaseDeathDegrees:
		return PhasePartial
	default:
		return PhaseDeathZone
	}
}

// SummedLevelDB returns the level, relative to perfectly in-phase signals, of two equally loud
// signals of one frequency that are deg degrees apart: 0 dB in phase, -3 dB at 90°, -Inf at 180°
func SummedLevelDB(deg float64) float64 {
	return 20 * math.Log10(math.Abs(math.Cos(deg*math.Pi/360)))
}

// PhaseResult is the outcome of a phase-safe distance check
type PhaseResult struct {
	Wavelength float64 // metres
	Phase      float64 // degrees, 0–180
	Zone       PhaseZone
	SummedDB   float64 // level of both mics summed at the frequency, relative to in phase
	Valid      bool    // the frequency was positive
}

// CheckPhase checks a path difference in metres at a frequency and air temperature in °C
func CheckPhase(pathDiff, freqHz, celsius float64) PhaseResult {
	wl := Wavelength(freqHz, celsius)
	if wl <= 0 {
		return PhaseResult{}
	}
	deg := PhaseOffset(pathDiff, wl)
	return PhaseResult{Wavelength: wl, Phase: deg, Zone: ZoneForPhase(deg), SummedDB: SummedLevelDB(deg), Valid: true}
}

// SafeTolerance is how far a path difference may be from a sweet spot and still be in it:
// one sixth of a wavelength, i.e. PhaseSafeDegrees
func SafeTolerance(wavelength float64) float64 {
	return wavelength * PhaseSafeDegrees / 360
}

// SweetSpots returns the path differences up to maxDist where the mics are exactly in phase:
// 0, λ, 2λ, ...
func SweetSpots(wavelength, maxDist float64) []float64 {
	return multiplesOf(wavelength, 0, maxDist)
}

// DeathZones returns the path differences up to maxDist where the mics are exactly opposed:
// λ/2, 3λ/2, ...
func DeathZones(wavelength, maxDist float64) []float64 {
	return multiplesOf(wavelength, 0.5, maxDist)
}

func multiplesOf(wavelength, offset, maxDist float64) []float64 {
	if wavelength <= 0 {
		return nil
	}
	var out []float64
	for k := 0; ; k++ {
		d := (float64(k) + offset) * wavelength
		if d > maxDist+1e-9 {
			return out
		}
		out = append(out, d)
	}
}

// ZoneSegment is a range of path differences, in metres, that falls into one zone
type ZoneSegment struct {
	Start, End float64
	Zone       PhaseZone
}

// ZoneSegments splits the path differences from 0 to maxDist into consecutive zones. Within each
// wavelength the zones change at 1/6, 1/3, 2/3 and 5/6 of it (60°, 120°, 240° and 300°).
func ZoneSegments(wavelength, maxDist float64) []ZoneSegment {
	if wavelength <= 0 || maxDist <= 0 {
		return nil
	}
	pattern := []struct {
		end  float64 // fraction of a wavelength
		zone PhaseZone
	}{
		{1.0 / 6, PhaseSweetSpot},
		{2.0 / 6, PhasePartial},
		{4.0 / 6, PhaseDeathZone},
		{5.0 / 6, PhasePartial},
		{1, PhaseSweetSpot},
	}
	var segs []ZoneSegment
	start := 0.0
	for k := 0; start < maxDist; k++ {
		for _, p := range pattern {
			end := math.Min((float64(k)+p.end)*wavelength, maxDist)
			if end <= start {
				continue
			}
			if n := len(segs); n > 0 && segs[n-1].Zone == p.zone {
				segs[n-1].End = end // the sweet spot spans the end of one wavelength and the start of the next
			} else {
				segs = append(segs, ZoneSegment{Start: start, End: end, Zone: p.zone})
			}
			start = end
			if start >= maxDist {
				break
			}
		}
	}
	return segs
}
