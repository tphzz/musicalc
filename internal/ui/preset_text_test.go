package ui

import (
	"musicalc/internal/i18n"
	"musicalc/internal/logic"
	"strings"
	"testing"
)

func TestGermanPresetFeedbackPreservesNamesAndUnits(t *testing.T) {
	i18n.Use("de")
	defer i18n.Use("en")
	err := &logic.PresetError{Kind: logic.PresetMicInvalid, Name: "Room α", Mic: "Mic β"}
	text := sentence(err).Error()
	if !strings.Contains(text, "ungültigen Abstand") || !strings.Contains(text, "Room α") || !strings.Contains(text, "Mic β") {
		t.Fatal(text)
	}
	preset := logic.Preset{Source: "Snare α", ReferenceMeters: 3, DistanceUnit: logic.UnitMeters, Mics: []logic.PresetMic{{Name: "Mic β"}}}
	if got := presetSummary(preset); got != "Snare α · Ref. 3 m · 1 Mikrofon" {
		t.Fatal(got)
	}
	i18n.Use("en")
	if got := presetSummary(preset); got != "Snare α · ref 3 m · 1 mic" {
		t.Fatal(got)
	}
}
