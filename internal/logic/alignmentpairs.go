package logic

import "strings"

// PairDiff is a stereo pair (e.g. "Overhead L" and "Overhead R") whose mics get different delays.
// Delaying them separately changes the timing between the two channels and shifts the stereo image.
type PairDiff struct {
	Base        string  // the name without its side, e.g. "Overhead"
	Left, Right int     // indexes of the two mics in the calculator's Mics
	DiffMS      float64 // how much the two delays differ, always positive
}

// stereoSides are the name endings that mark the left and right mic of a pair
var stereoSides = []struct{ left, right string }{
	{" l", " r"},
	{" left", " right"},
}

// splitStereoName splits "Overhead L" into ("overhead", true for left) and reports whether the
// name ends with a side at all. The base is lower-case, so pairs match regardless of case.
func splitStereoName(name string) (base string, left bool, ok bool) {
	lower := strings.ToLower(strings.TrimSpace(name))
	for _, s := range stereoSides {
		if strings.HasSuffix(lower, s.left) {
			return strings.TrimSpace(strings.TrimSuffix(lower, s.left)), true, true
		}
		if strings.HasSuffix(lower, s.right) {
			return strings.TrimSpace(strings.TrimSuffix(lower, s.right)), false, true
		}
	}
	return "", false, false
}

// StereoPairs finds left/right pairs whose delays differ by at least one sample. Mics farther than
// the reference are skipped: they can't be delayed anyway. Call it after Recalculate.
func StereoPairs(mics []AlignmentMic) []PairDiff {
	var pairs []PairDiff
	used := make(map[int]bool)
	for i, a := range mics {
		baseA, leftA, ok := splitStereoName(a.Name)
		if !ok || baseA == "" || used[i] {
			continue
		}
		for j := i + 1; j < len(mics); j++ {
			b := mics[j]
			baseB, leftB, ok := splitStereoName(b.Name)
			if !ok || used[j] || baseB != baseA || leftB == leftA {
				continue
			}
			used[i], used[j] = true, true
			if a.IsBeyondReference || b.IsBeyondReference {
				break
			}
			diffSamples := a.DelaySamples - b.DelaySamples
			if diffSamples < 0 {
				diffSamples = -diffSamples
			}
			if diffSamples < 1 {
				break
			}
			left, right := i, j
			if !leftA {
				left, right = j, i
			}
			diff := a.DelayMS - b.DelayMS
			if diff < 0 {
				diff = -diff
			}
			// The base name as typed, without its side
			name := strings.TrimSpace(mics[left].Name)
			base := strings.TrimSpace(name[:len(name)-len(sideSuffix(name))])
			pairs = append(pairs, PairDiff{Base: base, Left: left, Right: right, DiffMS: diff})
			break
		}
	}
	return pairs
}

// sideSuffix returns the side ending of a stereo mic name as typed, e.g. " L" or " Left"
func sideSuffix(name string) string {
	lower := strings.ToLower(name)
	for _, s := range stereoSides {
		for _, side := range []string{s.left, s.right} {
			if strings.HasSuffix(lower, side) {
				return name[len(name)-len(side):]
			}
		}
	}
	return ""
}
