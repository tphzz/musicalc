package ui

import (
	"encoding/json"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"
)

type twoMicControls struct {
	root                                  fyne.CanvasObject
	lo, hi, freq, temp, farther, level, a *widgets.NumericEntry
	fartherUnit, polarity                 *widget.Select
	align                                 *widget.Button
}

// openTwoMic shows Two-Mic Sum restored from session s
func openTwoMic(t *testing.T, s twoMicSession) twoMicControls {
	t.Helper()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := openEnglishTabWith(t, map[string]string{prefTwoMicSession: string(data)}, NewTwoMicSumTab)
	e := numericEntries(root)
	if len(e) != 7 {
		t.Fatalf("got %d numeric entries, want range, frequency, temperature, farther, level and delay", len(e))
	}
	// The results scroll (with Mic B's inputs) comes before the setup card in the layout tree
	c := twoMicControls{root: root, farther: e[0], level: e[1], a: e[2], lo: e[3], hi: e[4], freq: e[5], temp: e[6]}
	visitUI(root, func(o fyne.CanvasObject) {
		switch x := o.(type) {
		case *widget.Select:
			if len(x.Options) == 2 && x.Options[1] == unitMS {
				c.fartherUnit = x
			}
			if len(x.Options) == 2 && x.Options[1] == "Inverted" {
				c.polarity = x
			}
		case *widget.Button:
			if x.Text == "Align" || x.Text == "Reset" {
				c.align = x
			}
		}
	})
	if c.fartherUnit == nil || c.polarity == nil || c.align == nil {
		t.Fatal("missing Two-Mic controls")
	}
	return c
}

// baseSession is the design example: Mic B 4 m farther at 20 °C, kick range
func baseSession() twoMicSession {
	return twoMicSession{Lo: "50", Hi: "100", Freq: "60", TempC: 20, TempUnit: logic.UnitCelsius, DistUnit: logic.UnitMeters,
		SampleRate: "48000", Farther: "4", FartherUnit: logic.UnitMeters, Level: "0", DelayA: "0", Open: true}
}

func (c twoMicControls) text() string { return visibleText(c.root) }

// TestTwoMicAlignKeepsResidual: 4 m at 20 °C arrives after 11.6475 ms and Align writes 11.65 ms,
// and the 2.45 µs left over still matters to equal inverted copies at cymbal frequencies
func TestTwoMicAlignKeepsResidual(t *testing.T) {
	s := baseSession()
	s.Lo, s.Hi, s.Freq, s.Inverted = "3000", "12000", "6000", true
	c := openTwoMic(t, s)
	test.Tap(c.align)
	if c.a.Text != "11.65" {
		t.Fatalf("Align wrote %q ms, want 11.65", c.a.Text)
	}
	text := c.text()
	if strings.Contains(text, "complete cancellation") {
		t.Errorf("a 2.45 µs residual is reported as complete cancellation:\n%s", text)
	}
	for _, want := range []string{"Aligned", "At 6000 Hz: −20.7 dB", "2.45 µs earlier"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q:\n%s", want, text)
		}
	}
	// Normal polarity: no dips or peaks below 20 kHz, and the audible range is still about +6 dB
	s.Inverted = false
	c = openTwoMic(t, s)
	test.Tap(c.align)
	if text := c.text(); !strings.Contains(text, "None below 20 kHz: across the audible range the level stays between +5.9 and +6.0 dB") {
		t.Errorf("aligned, normal polarity:\n%s", text)
	}
}

// TestTwoMicInvalidInputShowsMessage checks every input that cannot be used: no result, no hang,
// and a message saying which value to correct
func TestTwoMicInvalidInputShowsMessage(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(c twoMicControls)
		want string
	}{
		{"NaN distance", func(c twoMicControls) { c.farther.SetText("NaN") }, "how much farther Mic B is as a number"},
		{"infinite delay", func(c twoMicControls) { c.a.SetText("Inf") }, "delay on Mic A as 0 ms or more"},
		{"negative delay", func(c twoMicControls) { c.a.SetText("-1") }, "delay on Mic A as 0 ms or more"},
		{"NaN from", func(c twoMicControls) { c.lo.SetText("NaN") }, "range within 10 Hz to 20 kHz"},
		{"from above to", func(c twoMicControls) { c.lo.SetText("200") }, "range within 10 Hz to 20 kHz"},
		{"above 20 kHz", func(c twoMicControls) { c.hi.SetText("25000") }, "range within 10 Hz to 20 kHz"},
		{"impossible temperature", func(c twoMicControls) { c.temp.SetText("-600") }, "temperature from −40 to +50 °C"},
		{"temperature not a number", func(c twoMicControls) { c.temp.SetText("abc") }, "temperature from −40 to +50 °C"},
		{"temperature being typed", func(c twoMicControls) { c.temp.SetText("-") }, "temperature from −40 to +50 °C"},
		{"distance too far to count in samples", func(c twoMicControls) { c.farther.SetText("1e308") }, "how much farther Mic B is as a number"},
		{"delay too long to count in samples", func(c twoMicControls) { c.a.SetText("1e300") }, "delay on Mic A as 0 ms or more"},
		{"cursor outside the plot", func(c twoMicControls) { c.freq.SetText("1500") }, "frequency the graph shows, 20 to 1000 Hz"},
		{"level beyond any use", func(c twoMicControls) { c.level.SetText("1e6") }, "Mic B's level as a number of dB"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := openTwoMic(t, baseSession())
			tc.edit(c)
			text := c.text()
			if !strings.Contains(text, tc.want) {
				t.Errorf("missing message %q:\n%s", tc.want, text)
			}
			if strings.Contains(text, "Combined level") || strings.Contains(text, "At 60 Hz") {
				t.Errorf("a result is still shown:\n%s", text)
			}
		})
	}
	// A cursor outside the band but inside the plot is fine
	c := openTwoMic(t, baseSession())
	c.freq.SetText("500")
	if !strings.Contains(c.text(), "At 500 Hz") {
		t.Errorf("cursor 500 Hz inside the 20–1000 Hz plot was rejected:\n%s", c.text())
	}
}

// TestTwoMicUnitSwitchKeepsValue checks that switching between ms and metres only changes how the
// value is shown: 1 ms stays 1 ms, with its first dip at 500 Hz
func TestTwoMicUnitSwitchKeepsValue(t *testing.T) {
	s := baseSession()
	s.Farther, s.FartherUnit = "1", unitMS
	c := openTwoMic(t, s)
	for i := 0; i < 3; i++ {
		c.fartherUnit.SetSelected(logic.UnitMeters)
		c.fartherUnit.SetSelected(unitMS)
	}
	if c.farther.Text != "1" {
		t.Errorf("1 ms became %q ms after switching units", c.farther.Text)
	}
	if !strings.Contains(c.text(), "500 · 1500") {
		t.Errorf("first dips moved from 500 and 1500 Hz:\n%s", c.text())
	}
	// A measured time stays fixed when the temperature changes; its distance display follows
	c.fartherUnit.SetSelected(logic.UnitMeters)
	before := c.farther.Text
	c.temp.SetText("0")
	if c.farther.Text == before {
		t.Errorf("the distance for 1 ms did not follow the temperature (%s m)", before)
	}
	if !strings.Contains(c.text(), "500 · 1500") {
		t.Errorf("a temperature change moved the dips of a measured time:\n%s", c.text())
	}
}

// TestTwoMicAlignedInvertedBoost checks the signed wording: a louder inverted Mic B at no time
// difference raises every frequency by +9.5 dB
func TestTwoMicAlignedInvertedBoost(t *testing.T) {
	s := baseSession()
	s.Farther, s.Level, s.Inverted = "0", "12", true
	c := openTwoMic(t, s)
	text := c.text()
	if !strings.Contains(text, "changes by the same +9.5 dB") || strings.Contains(text, "drops by") {
		t.Errorf("aligned inverted +12 dB:\n%s", text)
	}
	s.Level = "0"
	c = openTwoMic(t, s)
	if !strings.Contains(c.text(), "Complete cancellation") {
		t.Errorf("equal inverted copies at no time difference should cancel completely:\n%s", c.text())
	}
}

// TestTwoMicSessionPrecision checks the stored full-precision value and its precedence rules
func TestTwoMicSessionPrecision(t *testing.T) {
	meters := 1.23456789
	s := baseSession()
	s.Farther, s.Basis, s.Meters = "1.23", string(farDistance), &meters
	c := openTwoMic(t, s)
	want := logic.FormatTrim(mustArrival(t, meters)*1000, 2) + " ms later"
	if !strings.Contains(c.text(), "Mic B arrives "+want) {
		t.Errorf("canonical 1.23456789 m not used:\n%s", c.text())
	}
	// Saved again with the full value
	var saved twoMicSession
	if err := json.Unmarshal([]byte(fyne.CurrentApp().Preferences().String(prefTwoMicSession)), &saved); err != nil ||
		saved.Meters == nil || *saved.Meters != meters || saved.Basis != string(farDistance) {
		t.Errorf("saved session %+v, %v", saved, err)
	}
	// An older session without the full value reads the text
	legacy := baseSession()
	legacy.Farther = "2"
	if c := openTwoMic(t, legacy); !strings.Contains(c.text(), "Mic B arrives 5.82 ms later") {
		t.Errorf("legacy session:\n%s", c.text())
	}
	// Partial or contradictory metadata is not guessed at
	for _, bad := range []twoMicSession{
		func() twoMicSession { b := baseSession(); b.Basis = string(farDistance); return b }(),
		func() twoMicSession { b := baseSession(); b.Basis = "volume"; b.Meters = &meters; return b }(),
		func() twoMicSession { b := baseSession(); b.Basis = string(farTime); b.Meters = &meters; return b }(),
	} {
		if c := openTwoMic(t, bad); !strings.Contains(c.text(), "how much farther Mic B is as a number") {
			t.Errorf("session %+v was used:\n%s", bad, c.text())
		}
	}
	// The stored value wins over contradictory or missing typed text, and the entry shows it
	for _, text := range []string{"4", ""} {
		one := 1.0
		b := baseSession()
		b.Farther, b.Basis, b.Meters = text, string(farDistance), &one
		c := openTwoMic(t, b)
		if c.farther.Text != "1" || !strings.Contains(c.text(), "Mic B arrives 2.91 ms later") {
			t.Errorf("text %q with stored 1 m: entry %q\n%s", text, c.farther.Text, c.text())
		}
	}
	// Zero is a real stored value, not "missing"
	zero := 0.0
	z := baseSession()
	z.Farther, z.Basis, z.Meters = "0", string(farDistance), &zero
	if c := openTwoMic(t, z); !strings.Contains(c.text(), "Both arrive together") {
		t.Errorf("stored 0 m:\n%s", c.text())
	}
}

func mustArrival(t *testing.T, meters float64) float64 {
	t.Helper()
	s, ok := logic.ArrivalDelay(meters, 20)
	if !ok {
		t.Fatal("arrival")
	}
	return s
}
