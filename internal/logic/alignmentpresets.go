package logic

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// Presets store a studio's standard Alignment Delay setups. A preset holds everything about the
// physical setup but not the temperature, which changes from day to day.

// PresetMic is one microphone of a preset
type PresetMic struct {
	Name           string  `json:"name"`
	DistanceMeters float64 `json:"distanceMeters"`
}

// Preset is a stored Alignment Delay setup. Distances are kept in metres; DistanceUnit is the
// unit they are shown in.
type Preset struct {
	Name            string      `json:"name"`
	Source          string      `json:"source"`
	ReferenceMeters float64     `json:"referenceMeters"`
	DistanceUnit    string      `json:"distanceUnit"`
	SampleRate      int         `json:"sampleRate"`
	Mics            []PresetMic `json:"mics"`
}

// PresetFile is the JSON file that presets are exported to and imported from
type PresetFile struct {
	App     string   `json:"app"`
	Kind    string   `json:"kind"`
	Version int      `json:"version"`
	Presets []Preset `json:"presets"`
}

const (
	presetFileApp     = "MusiCalc"
	presetFileKind    = "alignment-delay-presets"
	presetFileVersion = 1
)

// SessionState is the Alignment Delay setup on screen, restored when the app starts again.
// Unlike a preset it includes the temperature, and it remembers which preset was loaded.
type SessionState struct {
	Preset
	TemperatureC    float64 `json:"temperatureC"`
	TemperatureUnit string  `json:"temperatureUnit"`
	LoadedPreset    string  `json:"loadedPreset"`
}

// MarshalPresets writes presets as an export file
func MarshalPresets(presets []Preset) ([]byte, error) {
	if presets == nil {
		presets = []Preset{}
	}
	return json.MarshalIndent(PresetFile{
		App: presetFileApp, Kind: presetFileKind, Version: presetFileVersion, Presets: presets,
	}, "", "  ")
}

// UnmarshalPresets reads an export file and checks that every preset is usable
func UnmarshalPresets(data []byte) ([]Preset, error) {
	var f PresetFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, &PresetError{Kind: PresetFileInvalid}
	}
	if f.Kind != presetFileKind {
		return nil, &PresetError{Kind: PresetKindInvalid}
	}
	if f.Version < 1 || f.Version > presetFileVersion {
		return nil, &PresetError{Kind: PresetVersionUnsupported, Version: f.Version}
	}
	for i := range f.Presets {
		if err := validatePreset(&f.Presets[i]); err != nil {
			return nil, err
		}
	}
	return f.Presets, nil
}

// validDistance reports whether d can be a distance: a finite number, not negative
func validDistance(d float64) bool {
	return !math.IsNaN(d) && !math.IsInf(d, 0) && d >= 0
}

// validatePreset checks a preset read from a file and fills in missing optional values
func validatePreset(p *Preset) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return &PresetError{Kind: PresetNameMissing}
	}
	if !validDistance(p.ReferenceMeters) {
		return &PresetError{Kind: PresetReferenceInvalid, Name: p.Name}
	}
	for _, m := range p.Mics {
		if !validDistance(m.DistanceMeters) {
			return &PresetError{Kind: PresetMicInvalid, Name: p.Name, Mic: m.Name}
		}
	}
	if !strings.EqualFold(p.DistanceUnit, UnitFeet) {
		p.DistanceUnit = UnitMeters
	}
	if p.SampleRate <= 0 {
		p.SampleRate = 48000
	}
	return nil
}

// FindPreset returns the index of the preset with this name (ignoring case), or -1
func FindPreset(presets []Preset, name string) int {
	name = strings.TrimSpace(name)
	for i, p := range presets {
		if strings.EqualFold(p.Name, name) {
			return i
		}
	}
	return -1
}

// MergePresets adds imported presets to the existing ones. An imported preset whose name is
// already taken is renamed "Name (2)", "Name (3)" and so on; renamed lists the new names.
func MergePresets(existing, imported []Preset) (merged []Preset, renamed []string) {
	merged = append([]Preset(nil), existing...)
	for _, p := range imported {
		if FindPreset(merged, p.Name) >= 0 {
			base := p.Name
			for n := 2; ; n++ {
				candidate := fmt.Sprintf("%s (%d)", base, n)
				if FindPreset(merged, candidate) < 0 {
					p.Name = candidate
					break
				}
			}
			renamed = append(renamed, p.Name)
		}
		merged = append(merged, p)
	}
	return merged, renamed
}

// RenamePreset renames the preset at index. The new name must not be empty or belong to
// another preset.
func RenamePreset(presets []Preset, index int, newName string) error {
	newName = strings.TrimSpace(newName)
	if index < 0 || index >= len(presets) {
		return &PresetError{Kind: PresetNotFound}
	}
	if newName == "" {
		return &PresetError{Kind: PresetNameEmpty}
	}
	if i := FindPreset(presets, newName); i >= 0 && i != index {
		return &PresetError{Kind: PresetNameTaken, Name: presets[i].Name}
	}
	presets[index].Name = newName
	return nil
}

// DeletePreset removes the preset at index
func DeletePreset(presets []Preset, index int) []Preset {
	if index < 0 || index >= len(presets) {
		return presets
	}
	return append(presets[:index:index], presets[index+1:]...)
}

// PresetSummary describes a preset in one line, e.g. "Snare · ref 3 m · 6 mics"
func PresetSummary(p Preset) string {
	unit := p.DistanceUnit
	if unit == "" {
		unit = UnitMeters
	}
	parts := []string{}
	if s := strings.TrimSpace(p.Source); s != "" {
		parts = append(parts, s)
	}
	parts = append(parts, "ref "+FormatTrim(FromMeters(p.ReferenceMeters, unit), 2)+" "+unit)
	mics := "mics"
	if len(p.Mics) == 1 {
		mics = "mic"
	}
	parts = append(parts, fmt.Sprintf("%d %s", len(p.Mics), mics))
	return strings.Join(parts, " · ")
}

// SameSetup reports whether two presets describe the same setup (their names are ignored)
func SameSetup(a, b Preset) bool {
	if !strings.EqualFold(strings.TrimSpace(a.Source), strings.TrimSpace(b.Source)) ||
		math.Abs(a.ReferenceMeters-b.ReferenceMeters) > 1e-9 ||
		!strings.EqualFold(a.DistanceUnit, b.DistanceUnit) ||
		a.SampleRate != b.SampleRate || len(a.Mics) != len(b.Mics) {
		return false
	}
	for i := range a.Mics {
		if a.Mics[i].Name != b.Mics[i].Name || math.Abs(a.Mics[i].DistanceMeters-b.Mics[i].DistanceMeters) > 1e-9 {
			return false
		}
	}
	return true
}
