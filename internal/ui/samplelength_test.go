package ui

import "testing"

// TestSampleLengthExactAndInvalid checks an exact sample count with a whole-millisecond display,
// and that invalid input clears the results instead of leaving them looking current
func TestSampleLengthExactAndInvalid(t *testing.T) {
	entries := numericEntries(openEnglishTab(t, NewSampleLengthTab))
	if len(entries) != 4 {
		t.Fatalf("got %d numeric entries, want beats, samples, ms and tempo", len(entries))
	}
	beats, samples, ms, tempo := entries[0], entries[1], entries[2], entries[3]
	if samples.Text != "88200" || ms.Text != "2000" {
		t.Fatalf("defaults show %s samples / %s ms, want 88200 / 2000", samples.Text, ms.Text)
	}

	beats.SetText("3")
	tempo.SetText("50")
	if samples.Text != "158760" || ms.Text != "3600" {
		t.Errorf("3 beats at 50 BPM, 44.1 kHz: %s samples / %s ms, want 158760 / 3600", samples.Text, ms.Text)
	}

	beats.SetText("0")
	if samples.Text != "" || ms.Text != "" {
		t.Errorf("0 beats left %q samples / %q ms, want both cleared", samples.Text, ms.Text)
	}

	beats.SetText("4")
	ms.SetText("1000")
	if samples.Text != "44100" || tempo.Text != "240" {
		t.Errorf("1000 ms, 4 beats: %s samples / %s BPM, want 44100 / 240", samples.Text, tempo.Text)
	}
	samples.SetText("22050")
	if ms.Text != "500" || tempo.Text != "480" {
		t.Errorf("22050 samples, 4 beats: %s ms / %s BPM, want 500 / 480", ms.Text, tempo.Text)
	}
	samples.SetText("0")
	if ms.Text != "" || tempo.Text != "" {
		t.Errorf("0 samples left %q ms / %q BPM, want both cleared", ms.Text, tempo.Text)
	}
}
