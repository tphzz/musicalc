package logic

import (
	"math"
	"strings"
	"testing"
)

// snareSetup is the example from the design: everything measured from the snare, reference 3 m,
// 20 °C, 48 kHz
func snareSetup() *AlignmentDelayCalculator {
	c := NewAlignmentDelayCalculator()
	c.Source = "Snare"
	c.SetTemperature(20, UnitCelsius)
	c.SetSampleRateLabel("48000")
	c.SetReferenceDistance(3, UnitMeters)
	for _, m := range []struct {
		name string
		d    float64
	}{{"Snare Top", 0.1}, {"Snare Bottom", 0.3}, {"Overhead L", 1.2}, {"Overhead R", 1.4}, {"Room L", 3}, {"Room R", 3.5}} {
		c.AddMic(m.name, m.d, UnitMeters)
	}
	c.Recalculate()
	return c
}

func TestAlignmentDelaySigned(t *testing.T) {
	c := snareSetup()
	want := []struct {
		ms        float64
		samples   int
		beyond    bool
		reference bool
	}{
		{8.444, 405, false, false},
		{7.862, 377, false, false},
		{5.241, 252, false, false},
		{4.659, 224, false, false},
		{0, 0, false, true},
		{-1.456, -70, true, false},
	}
	for i, w := range want {
		m := c.Mics[i]
		if !near(m.DelayMS, w.ms, 0.001) || m.DelaySamples != w.samples || m.IsBeyondReference != w.beyond || m.IsReference != w.reference {
			t.Errorf("%s: got %.3f ms / %d samples (beyond=%v, ref=%v), want %.3f ms / %d samples (beyond=%v, ref=%v)",
				m.Name, m.DelayMS, m.DelaySamples, m.IsBeyondReference, m.IsReference, w.ms, w.samples, w.beyond, w.reference)
		}
	}
}

func TestAlignmentDelayReferenceInFeet(t *testing.T) {
	// The same physical distance entered in feet is still the reference, despite rounding of the conversion
	c := NewAlignmentDelayCalculator()
	c.SetTemperature(20, UnitCelsius)
	c.SetReferenceDistance(ConvertDistance(3, UnitMeters, UnitFeet), UnitFeet)
	c.AddMic("Room L", 3, UnitMeters)
	c.Recalculate()
	if m := c.Mics[0]; !m.IsReference || m.IsBeyondReference || m.DelaySamples != 0 {
		t.Errorf("got ref=%v beyond=%v samples=%d, want the reference with no delay", m.IsReference, m.IsBeyondReference, m.DelaySamples)
	}
}

func TestStereoPairs(t *testing.T) {
	c := snareSetup()
	pairs := StereoPairs(c.Mics)
	if len(pairs) != 1 {
		t.Fatalf("got %d pairs, want 1 (overheads; the room pair has a mic beyond the reference): %+v", len(pairs), pairs)
	}
	p := pairs[0]
	if p.Base != "Overhead" || c.Mics[p.Left].Name != "Overhead L" || c.Mics[p.Right].Name != "Overhead R" || !near(p.DiffMS, 0.582, 0.001) {
		t.Errorf("got %+v (diff %.3f ms), want Overhead L/R differing by 0.582 ms", p, p.DiffMS)
	}
}

func TestStereoPairsMatching(t *testing.T) {
	calc := func(names []string, dists []float64) []AlignmentMic {
		c := NewAlignmentDelayCalculator()
		c.SetTemperature(20, UnitCelsius)
		c.SetReferenceDistance(3, UnitMeters)
		for i, n := range names {
			c.AddMic(n, dists[i], UnitMeters)
		}
		c.Recalculate()
		return c.Mics
	}
	tests := []struct {
		name  string
		names []string
		dists []float64
		want  int
	}{
		{"equal distances", []string{"Overhead L", "Overhead R"}, []float64{1.2, 1.2}, 0},
		{"single mic", []string{"Overhead L"}, []float64{1.2}, 0},
		{"words and case", []string{"piano left", "Piano Right"}, []float64{0.5, 0.8}, 1},
		{"right listed first", []string{"Overhead R", "Overhead L"}, []float64{1.4, 1.2}, 1},
		{"different bases", []string{"Room L", "Overhead R"}, []float64{1, 2}, 0},
		{"two lefts", []string{"Tom L", "Tom L"}, []float64{1, 2}, 0},
		{"no side", []string{"Kick In", "Kick Out"}, []float64{1, 1.5}, 0},
		{"less than a sample", []string{"OH L", "OH R"}, []float64{1.2, 1.2001}, 0},
	}
	for _, tt := range tests {
		got := StereoPairs(calc(tt.names, tt.dists))
		if len(got) != tt.want {
			t.Errorf("%s: got %d pairs, want %d", tt.name, len(got), tt.want)
		}
	}
	// The left mic is reported as Left even when the right one comes first
	if got := StereoPairs(calc([]string{"Overhead R", "Overhead L"}, []float64{1.4, 1.2})); len(got) == 1 && (got[0].Left != 1 || got[0].Right != 0) {
		t.Errorf("got left %d right %d, want left 1 right 0", got[0].Left, got[0].Right)
	}
}

func studioPresets() []Preset {
	return []Preset{
		{Name: "Studio A kit", Source: "Snare", ReferenceMeters: 3, DistanceUnit: UnitMeters, SampleRate: 48000,
			Mics: []PresetMic{{"Snare Top", 0.1}, {"Overhead L", 1.2}, {"Overhead R", 1.4}}},
		{Name: "Small booth", Source: "Snare", ReferenceMeters: 1.8, DistanceUnit: UnitFeet, SampleRate: 96000,
			Mics: []PresetMic{{"Snare Top", 0.1}}},
	}
}

func TestPresetFileRoundTrip(t *testing.T) {
	data, err := MarshalPresets(studioPresets())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"kind": "alignment-delay-presets"`) {
		t.Errorf("export misses the kind: %s", data)
	}
	got, err := UnmarshalPresets(data)
	if err != nil {
		t.Fatal(err)
	}
	want := studioPresets()
	if len(got) != len(want) {
		t.Fatalf("got %d presets, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Name != want[i].Name || !SameSetup(got[i], want[i]) {
			t.Errorf("preset %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
	// An empty list exports as an empty array, not null
	empty, _ := MarshalPresets(nil)
	if p, err := UnmarshalPresets(empty); err != nil || len(p) != 0 {
		t.Errorf("empty export: got %v, %v", p, err)
	}
}

func TestPresetFileRejectsBadFiles(t *testing.T) {
	tests := map[string]string{
		"broken JSON":    `{"kind": `,
		"not an object":  `[1, 2]`,
		"wrong kind":     `{"app":"MusiCalc","kind":"tap-tempo","version":1,"presets":[]}`,
		"newer version":  `{"app":"MusiCalc","kind":"alignment-delay-presets","version":2,"presets":[]}`,
		"no version":     `{"app":"MusiCalc","kind":"alignment-delay-presets","presets":[]}`,
		"no name":        `{"kind":"alignment-delay-presets","version":1,"presets":[{"name":"  ","referenceMeters":3}]}`,
		"negative ref":   `{"kind":"alignment-delay-presets","version":1,"presets":[{"name":"A","referenceMeters":-1}]}`,
		"negative mic":   `{"kind":"alignment-delay-presets","version":1,"presets":[{"name":"A","referenceMeters":3,"mics":[{"name":"x","distanceMeters":-0.5}]}]}`,
		"string as dist": `{"kind":"alignment-delay-presets","version":1,"presets":[{"name":"A","referenceMeters":"3"}]}`,
	}
	for name, data := range tests {
		if _, err := UnmarshalPresets([]byte(data)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestPresetFileFillsDefaults(t *testing.T) {
	got, err := UnmarshalPresets([]byte(`{"kind":"alignment-delay-presets","version":1,"presets":[{"name":" A ","referenceMeters":3}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if p := got[0]; p.Name != "A" || p.DistanceUnit != UnitMeters || p.SampleRate != 48000 {
		t.Errorf("got %+v, want name trimmed, unit m and 48000 Hz", p)
	}
}

func TestMergePresetsRenamesDuplicates(t *testing.T) {
	existing := studioPresets()
	imported := []Preset{{Name: "small booth"}, {Name: "Live room"}, {Name: "Small booth"}}
	merged, renamed := MergePresets(existing, imported)
	var names []string
	for _, p := range merged {
		names = append(names, p.Name)
	}
	want := "Studio A kit|Small booth|small booth (2)|Live room|Small booth (3)"
	if strings.Join(names, "|") != want {
		t.Errorf("got %v, want %s", names, want)
	}
	if strings.Join(renamed, "|") != "small booth (2)|Small booth (3)" {
		t.Errorf("renamed: got %v", renamed)
	}
	if len(existing) != 2 {
		t.Errorf("merge changed the existing slice")
	}
}

func TestRenameAndDeletePreset(t *testing.T) {
	p := studioPresets()
	if err := RenamePreset(p, 0, "small BOOTH"); err == nil {
		t.Error("renaming to another preset's name must fail")
	}
	if err := RenamePreset(p, 0, "   "); err == nil {
		t.Error("an empty name must fail")
	}
	if err := RenamePreset(p, 5, "X"); err == nil {
		t.Error("a missing preset must fail")
	}
	if err := RenamePreset(p, 1, "SMALL booth"); err != nil || p[1].Name != "SMALL booth" {
		t.Errorf("changing only the case of its own name: got %v, %q", err, p[1].Name)
	}
	if err := RenamePreset(p, 0, "  Studio B  "); err != nil || p[0].Name != "Studio B" {
		t.Errorf("rename: got %v, %q", err, p[0].Name)
	}
	p = DeletePreset(p, 0)
	if len(p) != 1 || p[0].Name != "SMALL booth" {
		t.Errorf("delete: got %+v", p)
	}
	if p = DeletePreset(p, 3); len(p) != 1 {
		t.Errorf("deleting a missing preset changed the list")
	}
}

func TestPresetSummary(t *testing.T) {
	p := studioPresets()
	if got := PresetSummary(p[0]); got != "Snare · ref 3 m · 3 mics" {
		t.Errorf("got %q", got)
	}
	if got := PresetSummary(p[1]); got != "Snare · ref 5.91 ft · 1 mic" {
		t.Errorf("got %q", got)
	}
	if got := PresetSummary(Preset{ReferenceMeters: 2}); got != "ref 2 m · 0 mics" {
		t.Errorf("got %q", got)
	}
}

func TestSameSetup(t *testing.T) {
	a := studioPresets()[0]
	b := a
	b.Mics = append([]PresetMic(nil), a.Mics...)
	b.Name = "Other name"
	if !SameSetup(a, b) {
		t.Error("presets differing only in name must be the same setup")
	}
	b.Mics[1].DistanceMeters = 1.25
	if SameSetup(a, b) {
		t.Error("a changed mic distance must differ")
	}
	c := a
	c.ReferenceMeters = math.Nextafter(3, 4) + 0.001
	if SameSetup(a, c) {
		t.Error("a changed reference must differ")
	}
}
