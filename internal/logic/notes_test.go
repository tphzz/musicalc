package logic

import (
	"bytes"
	"strings"
	"testing"

	sclres "musicalc/internal/logic/scl"

	scala "github.com/chinenual/go-scala"
)

func TestNoteName(t *testing.T) {
	for _, tc := range []struct {
		note, offset int
		want         string
	}{
		{69, OctaveOffsetC3, "A3"},
		{69, OctaveOffsetC4, "A4"},
		{0, OctaveOffsetC3, "C-2"},
		{0, OctaveOffsetC4, "C-1"},
		{-9, OctaveOffsetC4, "D#-2"},
		{-1, OctaveOffsetC3, "B-3"},
		{131, OctaveOffsetC3, "B8"},
		{132, OctaveOffsetC4, "C10"},
	} {
		if got := NoteName(tc.note, tc.offset); got != tc.want {
			t.Errorf("NoteName(%d, %d) = %q, want %q", tc.note, tc.offset, got, tc.want)
		}
	}
}

func TestPitchClassName(t *testing.T) {
	for pc, want := range map[int]string{0: "C", 1: "C#", 4: "E", 11: "B", 12: "C", -1: "B"} {
		if got := PitchClassName(pc); got != want {
			t.Errorf("PitchClassName(%d) = %q, want %q", pc, got, want)
		}
	}
}

func TestParseNoteName(t *testing.T) {
	for _, tc := range []struct {
		in     string
		offset int
		want   int
	}{
		{"A3", OctaveOffsetC3, 69},
		{"A4", OctaveOffsetC4, 69},
		{" c#4 ", OctaveOffsetC3, 73},
		{"C-2", OctaveOffsetC3, 0},
		{"C-1", OctaveOffsetC4, 0},
		{"G8", OctaveOffsetC3, 127},
		{"G9", OctaveOffsetC4, 127},
	} {
		got, err := ParseNoteName(tc.in, tc.offset)
		if err != nil || got != tc.want {
			t.Errorf("ParseNoteName(%q, %d) = %d, %v; want %d", tc.in, tc.offset, got, err, tc.want)
		}
	}
	for _, in := range []string{"", "A", "#3", "H3", "Db3", "E#3", "A3x", "A 3", "A+3", "G#8", "C-3", "A99999999999999999999"} {
		if got, err := ParseNoteName(in, OctaveOffsetC3); err == nil {
			t.Errorf("ParseNoteName(%q) = %d, want an error", in, got)
		}
	}
}

// TestBundledScalesParse checks every bundled file, including those with text after a pitch value
// ("100.0 cents", "2187/2048 ! name") that go-scala cannot read on its own
func TestBundledScalesParse(t *testing.T) {
	twelve := 0
	for name, info := range sclres.AvailableScales {
		scale, err := scala.ScaleFromSCLStream(bytes.NewReader(sclPitchesOnly(info.Resource.Content())))
		if err != nil {
			t.Errorf("%s: %v", info.Filename, err)
			continue
		}
		if scale.Count < 1 || scale.Tones[scale.Count-1].Cents <= 0 {
			t.Errorf("%s (%s): %d steps, no period above 1/1", info.Filename, name, scale.Count)
		}
		if scale.Count == 12 {
			twelve++
		}
	}
	if twelve != 27 {
		t.Errorf("%d bundled scales have 12 steps, want 27", twelve)
	}
	// Shi Er Lü was read as 12-TET while its pitch lines carried names
	var shiErLu string
	for name, info := range sclres.AvailableScales {
		if info.Filename == "chin_shierlu.scl" {
			shiErLu = name
		}
	}
	d := mustDiapason(t, shiErLu, 0, 60, 261.63)
	if r := d.Note(61).Frequency / d.Note(60).Frequency; !near(r, 2187.0/2048, 1e-12) {
		t.Errorf("Shi Er Lü: second step %.12f, want 2187/2048", r)
	}
}

func TestSCLPitchesOnlyKeepsDescription(t *testing.T) {
	in := "! x.scl\r\n!\r\nDescription with 100.0 cents\r\n 2\r\n!\r\n 100.0 cents\r\n   ! indented comment\r\n 2/1 ! octave\r\n"
	got := string(sclPitchesOnly([]byte(in)))
	want := "! x.scl\n!\nDescription with 100.0 cents\n2\n!\n100.0\n! indented comment\n2/1\n"
	if got != want {
		t.Errorf("sclPitchesOnly:\n%q\nwant\n%q", got, want)
	}
	if !strings.Contains(got, "Description with 100.0 cents") {
		t.Error("description line was changed")
	}
}
