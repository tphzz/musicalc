package logic

import "math"

// semitoneSplit is a pitch offset in whole semitones plus cents, in both notations the app shows
type semitoneSplit struct {
	semitones int // whole semitones down, then 0 to 99 cents
	cents     int
	nearest   int // nearest semitone, then -50 to +50 cents
	cents50   int
}

// splitSemitones rounds a finite pitch offset to whole cents once and splits that both ways, so a
// value just below a semitone carries into it instead of showing 100 cents
func splitSemitones(semitones float64) semitoneSplit {
	total := int(math.Round(semitones * 100))
	down := floorDiv(total, 100)
	nearest := int(math.Round(float64(total) / 100))
	return semitoneSplit{down, total - down*100, nearest, total - nearest*100}
}

// floorDiv divides rounding towards minus infinity, so -9 semitones fall in octave -1, not 0
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// floorMod is the remainder that goes with floorDiv; it has the sign of b
func floorMod(a, b int) int {
	return a - floorDiv(a, b)*b
}
