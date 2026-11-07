package logic

import (
	"math"
	"strconv"
	"strings"
)

// ParseNumber reads typed text as a number, accepting a decimal comma as well as a point. ok is
// false for empty, malformed or non-finite text (NaN, Inf, values too large for a float64), so
// callers can tell those apart from a real 0.
func ParseNumber(s string) (v float64, ok bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(s, ",", ".")), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

// ParseFloat reads typed text as a number like ParseNumber, giving 0 for text that is not a finite
// number. Screens that must tell invalid input from 0 use ParseNumber.
func ParseFloat(s string) float64 {
	v, _ := ParseNumber(s)
	return v
}
