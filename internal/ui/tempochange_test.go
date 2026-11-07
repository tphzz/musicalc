package ui

import (
	"strings"
	"testing"

	"musicalc/internal/logic"
)

// TestTempoChangeCarryAndRange checks the cent carry and that tempos outside the supported range
// clear the derived values
func TestTempoChangeCarryAndRange(t *testing.T) {
	root := openEnglishTab(t, NewTempoChangeTab)
	entries := numericEntries(root)
	if len(entries) != 5 {
		t.Fatalf("got %d numeric entries, want original, new, stretch, semitones and cents", len(entries))
	}
	newTempo, stretch, semis, cents := entries[1], entries[2], entries[3], entries[4]
	if stretch.Text != "120" || semis.Text != "-4" || cents.Text != "+84" {
		t.Fatalf("120 → 100 BPM shows %s %%, %s st %s c; want 120, -4, +84", stretch.Text, semis.Text, cents.Text)
	}
	if !strings.Contains(uiText(root), "-16.67 %") {
		t.Error("120 → 100 BPM does not show the -16.67 % tempo change")
	}

	newTempo.SetText("119.98")
	if semis.Text != "0" || cents.Text != "0" {
		t.Errorf("120 → 119.98 BPM: %s st %s c, want 0 and 0 rather than -1 and +100", semis.Text, cents.Text)
	}

	newTempo.SetText("2000")
	if stretch.Text != "" || semis.Text != "" || cents.Text != "" {
		t.Errorf("2000 BPM left %q %%, %q st, %q c; want them cleared", stretch.Text, semis.Text, cents.Text)
	}

	stretch.SetText("50")
	if newTempo.Text != "240" || semis.Text != "+12" || cents.Text != "0" {
		t.Errorf("50 %% stretch: %s BPM, %s st %s c; want 240, +12, 0", newTempo.Text, semis.Text, cents.Text)
	}

	// A shift beyond the range is limited so the new tempo stays inside it
	semis.SetText("+60")
	got := logic.ParseFloat(newTempo.Text)
	if semis.Text != "+36" || cents.Text != "+70" || !(got > 999 && got <= logic.MaxTempoBPM) {
		t.Errorf("+60 semitones from 120 BPM: %s st %s c, %q BPM; want +36 +70 just below %v BPM",
			semis.Text, cents.Text, newTempo.Text, logic.MaxTempoBPM)
	}
}
