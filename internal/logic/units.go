package logic

import (
	"fmt"
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

// SpeedOfSound returns the speed of sound in air in m/s at the given temperature in °C
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
