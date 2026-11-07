package logic

import (
	"fmt"
	"math"
	"strings"
)

// Unit labels shown in the unit dropdowns
const (
	UnitMeters     = "m"
	UnitFeet       = "ft"
	UnitCelsius    = "°C"
	UnitFahrenheit = "°F"
)

// IsFahrenheit reports whether a temperature unit label means Fahrenheit ("°F", "F" or "f")
func IsFahrenheit(unit string) bool {
	return strings.EqualFold(strings.TrimPrefix(strings.TrimSpace(unit), "°"), "F")
}

// ToCelsius converts a temperature in the given unit to °C
func ToCelsius(value float64, unit string) float64 {
	if IsFahrenheit(unit) {
		return (value - 32) * 5 / 9
	}
	return value
}

// FromCelsius converts a temperature in °C to the given unit
func FromCelsius(celsius float64, unit string) float64 {
	if IsFahrenheit(unit) {
		return celsius*9/5 + 32
	}
	return celsius
}

// ConvertTemperature converts a temperature between units, e.g. 22 °C → 71.6 °F
func ConvertTemperature(value float64, from, to string) float64 {
	return FromCelsius(ToCelsius(value, from), to)
}

// ConvertDistance converts a distance between units, e.g. 4 m → 13.12 ft
func ConvertDistance(value float64, from, to string) float64 {
	return FromMeters(ToMeters(value, from), to)
}

// The ambient temperatures the calculators accept. Within them the linear speed-of-sound
// approximation below is accurate to well under 1 %; humidity is not taken into account.
const (
	MinTemperatureC = -40.0
	MaxTemperatureC = 50.0
)

// TemperatureInRange reports whether celsius is within the accepted ambient temperatures
func TemperatureInRange(celsius float64) bool {
	return celsius >= MinTemperatureC && celsius <= MaxTemperatureC
}

// SpeedOfSound returns the speed of sound in air in m/s at the given temperature in °C, a linear
// approximation meant for TemperatureInRange
func SpeedOfSound(celsius float64) float64 {
	return 331.3 + 0.606*celsius
}

// FormatTrim formats v with at most the given decimals and drops trailing zeros: 4.50 → "4.5"
func FormatTrim(v float64, decimals int) string {
	s := fmt.Sprintf("%.*f", decimals, v)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}

// FormatTwoDecimals formats v with two decimals, dropping ".00" from whole values: 3599.9999999999995
// → "3600", 82.354 → "82.35", 1.5 → "1.50"
func FormatTwoDecimals(v float64) string {
	s := strings.TrimSuffix(fmt.Sprintf("%.2f", v), ".00")
	if s == "-0" {
		return "0"
	}
	return s
}

// CeilTo rounds v up to the given number of decimals, for minimums shown to the user: 0.112468
// → 0.12. A value that is already on a step, within floating-point noise, stays: 0.12 → 0.12.
func CeilTo(v float64, decimals int) float64 {
	scale := math.Pow(10, float64(decimals))
	scaled := v * scale
	if r := math.Round(scaled); math.Abs(scaled-r) <= 1e-9*math.Max(1, math.Abs(scaled)) {
		return r / scale
	}
	return math.Ceil(scaled) / scale
}
