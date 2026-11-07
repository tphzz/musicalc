package ui

import (
	"encoding/json"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"
)

type alignControls struct {
	root            fyne.CanvasObject
	window          fyne.Window
	temp, ref, dist *widgets.NumericEntry
	add             *widgets.Button
}

func openAlignment(t *testing.T, stored map[string]string) alignControls {
	t.Helper()
	root, w := openEnglishTabWith(t, stored, NewAlignmentDelayTab)
	c := alignControls{root: root, window: w, add: buttonNamed(t, root, "+")}
	for _, e := range numericEntries(root) {
		switch e.PlaceHolder {
		case "Temperature":
			c.temp = e
		case "Distance":
			if c.ref == nil {
				c.ref = e
			}
		}
	}
	for _, e := range numericEntries(root) {
		if e != c.ref && e != c.temp {
			c.dist = e
		}
	}
	if c.temp == nil || c.ref == nil || c.dist == nil {
		t.Fatal("missing Alignment Delay entries")
	}
	return c
}

func session(t *testing.T, s logic.SessionState) map[string]string {
	t.Helper()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]string{prefAlignmentSession: string(data)}
}

func twoMicSetup() logic.SessionState {
	return logic.SessionState{
		Preset: logic.Preset{Source: "Snare", ReferenceMeters: 3, DistanceUnit: logic.UnitMeters, SampleRate: 48000,
			Mics: []logic.PresetMic{{Name: "Top", DistanceMeters: 1}}},
		TemperatureC: 20, TemperatureUnit: logic.UnitCelsius,
	}
}

// TestAlignmentInvalidSetupShowsNoDelays checks that an impossible temperature or an unusable
// reference gives a warning and no delays, instead of a plausible result
func TestAlignmentInvalidSetupShowsNoDelays(t *testing.T) {
	c := openAlignment(t, session(t, twoMicSetup()))
	if text := visibleText(c.root); !strings.Contains(text, "5.82 ms") {
		t.Fatalf("valid setup:\n%s", text)
	}
	for _, tc := range []struct{ temp, ref, want string }{
		{"-600", "3", "temperature is not between −40 and +50 °C"},
		{"abc", "3", "temperature is not between"},
		{"", "3", "temperature is not between"},
		{"20", "-1", "reference distance is missing or not a distance of 0 or more"},
		{"20", "NaN", "reference distance is missing"},
		{"20", "", "reference distance is missing"},
	} {
		c.temp.SetText(tc.temp)
		c.ref.SetText(tc.ref)
		text := visibleText(c.root)
		if !strings.Contains(text, tc.want) || !strings.Contains(text, "–") {
			t.Errorf("temperature %q, reference %q:\n%s", tc.temp, tc.ref, text)
		}
		if strings.Contains(text, "11.65 ms") || strings.Contains(text, "5.82 ms") {
			t.Errorf("temperature %q, reference %q still shows a delay:\n%s", tc.temp, tc.ref, text)
		}
	}
	c.temp.SetText("20")
	c.ref.SetText("3")
	if text := visibleText(c.root); !strings.Contains(text, "5.82 ms") || strings.Contains(text, "No delays") {
		t.Errorf("after correcting:\n%s", text)
	}
}

// TestAlignmentRejectsNegativeDistance checks that Add refuses a distance below 0
func TestAlignmentRejectsNegativeDistance(t *testing.T) {
	c := openAlignment(t, session(t, twoMicSetup()))
	for _, bad := range []string{"-1", "NaN", "Inf"} {
		c.dist.SetText(bad)
		press(c.add)
		if text := visibleText(c.root); !strings.Contains(text, "new mic was not added") || c.dist.Text != bad {
			t.Errorf("%q:\n%s", bad, text)
		}
	}
	c.dist.SetText("2")
	press(c.add)
	if text := visibleText(c.root); strings.Contains(text, "not added") || !strings.Contains(text, "2.91 ms") {
		t.Errorf("a valid distance after a rejected one:\n%s", text)
	}
}

// TestAlignmentCorruptStorageIsReported checks that unreadable or impossible stored data gives the
// defaults (or the usable presets) and an error dialog, rather than being replaced silently
func TestAlignmentCorruptStorageIsReported(t *testing.T) {
	for _, stored := range []map[string]string{
		{prefAlignmentSession: "{not json"},
		session(t, func() logic.SessionState { s := twoMicSetup(); s.TemperatureC = -600; return s }()),
		session(t, func() logic.SessionState { s := twoMicSetup(); s.ReferenceMeters = -3; return s }()),
	} {
		c := openAlignment(t, stored)
		if !hasErrorDialog(c.window) {
			t.Errorf("%v: no error dialog", stored)
		}
		if c.temp.Text != "20" || c.ref.Text != "3" {
			t.Errorf("%v: restored %q °C, %q m; want the defaults", stored, c.temp.Text, c.ref.Text)
		}
	}
	presets := `[{"name":"Good","referenceMeters":3,"distanceUnit":"m","sampleRate":48000},` +
		`{"name":"Bad","referenceMeters":-2,"distanceUnit":"m","sampleRate":48000}]`
	c := openAlignment(t, map[string]string{prefAlignmentPresets: presets})
	if !hasErrorDialog(c.window) {
		t.Error("an invalid stored preset was dropped without telling")
	}
	got, err := restoreStoredPresets(presets)
	if len(got) != 1 || got[0].Name != "Good" || err == nil {
		t.Errorf("restored presets %+v, %v", got, err)
	}
	if c := openAlignment(t, map[string]string{prefAlignmentPresets: "[broken"}); !hasErrorDialog(c.window) {
		t.Error("an unreadable preset list was dropped without telling")
	}
	// A valid session and presets open without a dialog
	if c := openAlignment(t, session(t, twoMicSetup())); hasErrorDialog(c.window) {
		t.Error("a valid session showed an error")
	}
}
