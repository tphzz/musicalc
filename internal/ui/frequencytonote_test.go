package ui

import (
	"strings"
	"testing"
)

// TestFrequencyToNoteClearsInvalidInput checks that invalid input does not leave an earlier note
// looking current, and that a correct frequency shows it again
func TestFrequencyToNoteClearsInvalidInput(t *testing.T) {
	root := openEnglishTab(t, NewFrequencyToNoteTab)
	entries := numericEntries(root)
	if len(entries) != 1 {
		t.Fatalf("got %d numeric entries, want the frequency", len(entries))
	}
	freq := entries[0]
	// 450 Hz is A3 +39 cents in C3 naming, in both notations
	freq.SetText("450")
	if text := visibleText(root); !strings.Contains(text, "+39") {
		t.Fatalf("450 Hz in C3 naming:\n%s", text)
	}
	freq.SetText("439.9")
	if text := visibleText(root); strings.Contains(text, "+100") {
		t.Fatalf("439.9 Hz shows 100 cents:\n%s", text)
	}
	for _, bad := range []string{"0", "-5", "abc", "NaN", "Inf", ""} {
		freq.SetText("450")
		freq.SetText(bad)
		if text := visibleText(root); strings.Contains(text, "+39") {
			t.Errorf("%q left the earlier result visible:\n%s", bad, text)
		}
	}
	freq.SetText("450")
	if text := visibleText(root); !strings.Contains(text, "+39") {
		t.Errorf("450 Hz after invalid input:\n%s", text)
	}
}

// TestTempoToDelayRejectsInvalidTempo checks that a missing or unusable tempo shows no values
// instead of silently using 120 BPM. The cards follow a data binding, so the test waits for them.
func TestTempoToDelayRejectsInvalidTempo(t *testing.T) {
	root := openEnglishTab(t, NewTempoTab)
	shows := func(want string) func() bool {
		return func() bool { return strings.Contains(visibleText(root), want) }
	}
	if !eventually(shows("500.00 ms")) {
		t.Fatalf("120 BPM:\n%s", visibleText(root))
	}
	entries := numericEntries(root)
	if len(entries) == 0 {
		t.Fatal("no tempo entry")
	}
	for _, bad := range []string{"0", "-120", "abc", "NaN", ""} {
		entries[0].SetText(bad)
		gone := func() bool { return !strings.Contains(visibleText(root), " ms") }
		if !eventually(gone) {
			t.Errorf("tempo %q still shows delays:\n%s", bad, visibleText(root))
		}
	}
	entries[0].SetText("60")
	if !eventually(shows("1000.00 ms")) {
		t.Errorf("60 BPM:\n%s", visibleText(root))
	}
}
