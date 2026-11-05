package logic

import (
	"errors"
	"testing"
)

func TestPresetProblemsAreStructured(t *testing.T) {
	for _, tc := range []struct {
		input     string
		kind      PresetErrorKind
		name, mic string
		version   int
	}{
		{"broken", PresetFileInvalid, "", "", 0},
		{`{"kind":"other"}`, PresetKindInvalid, "", "", 0},
		{`{"kind":"alignment-delay-presets","version":2}`, PresetVersionUnsupported, "", "", 2},
		{`{"kind":"alignment-delay-presets","version":1,"presets":[{"name":""}]}`, PresetNameMissing, "", "", 0},
		{`{"kind":"alignment-delay-presets","version":1,"presets":[{"name":"Room α","referenceMeters":-1}]}`, PresetReferenceInvalid, "Room α", "", 0},
		{`{"kind":"alignment-delay-presets","version":1,"presets":[{"name":"Room α","mics":[{"name":"Mic β","distanceMeters":-1}]}]}`, PresetMicInvalid, "Room α", "Mic β", 0},
	} {
		_, err := UnmarshalPresets([]byte(tc.input))
		var detail *PresetError
		if !errors.As(err, &detail) || detail.Kind != tc.kind || detail.Name != tc.name || detail.Mic != tc.mic || detail.Version != tc.version {
			t.Fatalf("wrong structured error: %#v", err)
		}
	}
	presets := []Preset{{Name: "A"}, {Name: "B"}}
	for _, tc := range []struct {
		index int
		name  string
		kind  PresetErrorKind
	}{
		{9, "A", PresetNotFound}, {0, " ", PresetNameEmpty}, {0, "B", PresetNameTaken},
	} {
		var detail *PresetError
		if err := RenamePreset(presets, tc.index, tc.name); !errors.As(err, &detail) || detail.Kind != tc.kind {
			t.Fatalf("wrong rename error: %v", err)
		}
	}
	if presets[0].Name != "A" || presets[1].Name != "B" {
		t.Fatal("validation changed names")
	}
}
