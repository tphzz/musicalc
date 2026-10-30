package logic

import (
	"testing"
	"unicode/utf8"
)

// maxLabelLength keeps labels readable in the tuning dropdown on a phone
const maxLabelLength = 32

func TestCatalogCoversAllScales(t *testing.T) {
	files := make(map[string]bool, len(AvailableScales))
	for _, info := range AvailableScales {
		files[info.Filename] = true
	}

	seenFile := make(map[string]bool, len(catalog))
	seenLabel := make(map[string]bool, len(catalog))
	for _, c := range catalog {
		if !files[c.filename] {
			t.Errorf("catalog entry %q has no bundled scale", c.filename)
		}
		if seenFile[c.filename] {
			t.Errorf("scale %q is listed twice", c.filename)
		}
		if seenLabel[c.label] {
			t.Errorf("label %q is used twice", c.label)
		}
		if n := utf8.RuneCountInString(c.label); n > maxLabelLength {
			t.Errorf("label %q is %d characters, more than %d", c.label, n, maxLabelLength)
		}
		seenFile[c.filename] = true
		seenLabel[c.label] = true
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
}
