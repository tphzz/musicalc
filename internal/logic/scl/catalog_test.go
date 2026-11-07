package logic

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// maxLabelLength keeps the selected tuning ("Group · Name") readable in the dropdown on a phone
const maxLabelLength = 32

func TestCatalogCoversAllScales(t *testing.T) {
	files := make(map[string]bool, len(AvailableScales))
	for _, info := range AvailableScales {
		files[info.Filename] = true
	}

	seenFile := make(map[string]bool)
	seenLabel := make(map[string]bool)
	for _, g := range catalog {
		if len(g.scales) == 0 {
			t.Errorf("group %q is empty", g.group)
		}
		for _, sc := range g.scales {
			label := g.group + " · " + sc.name
			if !files[sc.filename] {
				t.Errorf("catalog entry %q has no bundled scale", sc.filename)
			}
			if seenFile[sc.filename] {
				t.Errorf("scale %q is listed twice", sc.filename)
			}
			if seenLabel[label] {
				t.Errorf("label %q is used twice", label)
			}
			if n := utf8.RuneCountInString(label); n > maxLabelLength {
				t.Errorf("label %q is %d characters, more than %d", label, n, maxLabelLength)
			}
			seenFile[sc.filename] = true
			seenLabel[label] = true
		}
	}
	for file := range files {
		if !seenFile[file] {
			t.Errorf("bundled scale %q has no catalog entry", file)
		}
	}
}

func TestTuningsStartWithDefault(t *testing.T) {
	tunings := Tunings()
	if len(tunings) != len(AvailableScales) {
		t.Fatalf("got %d tunings, want %d", len(tunings), len(AvailableScales))
	}
	if tunings[0].Name != DefaultScaleName {
		t.Errorf("first tuning is %q, want the default %q", tunings[0].Name, DefaultScaleName)
	}
	if want := tunings[0].Group + " · " + tunings[0].Short; tunings[0].Label != want {
		t.Errorf("label is %q, want %q", tunings[0].Label, want)
	}
}

// TestCatalogTonics checks the tonic each 1/1 is meant for: C unless the file name names another
// one ("..._on_d.scl")
func TestCatalogTonics(t *testing.T) {
	tonics := map[string]int{"c": tonicC, "d": tonicD, "e": tonicE}
	for _, g := range catalog {
		for _, sc := range g.scales {
			want := tonicC
			if i := strings.LastIndex(sc.filename, "_on_"); i >= 0 {
				letter := strings.TrimSuffix(sc.filename[i+len("_on_"):], ".scl")
				var ok bool
				if want, ok = tonics[letter]; !ok {
					t.Fatalf("%s: tonic %q is not in the test table", sc.filename, letter)
				}
			}
			if sc.tonic != want {
				t.Errorf("%s: tonic %d, want %d", sc.filename, sc.tonic, want)
			}
		}
	}
	for _, tuning := range Tunings() {
		if tuning.Tonic < 0 || tuning.Tonic > 11 {
			t.Errorf("%s: tonic %d is not a pitch class", tuning.Label, tuning.Tonic)
		}
	}
}
