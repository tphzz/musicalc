// Command capture renders the Multi-Mic screens shown in the help pages, with the help pages'
// examples entered, as PNG files for annotate.py:
//
//	go run ./utils/helpimages/capture <capture folder>
//	python utils/helpimages/annotate.py <capture folder>
//
// The screens are drawn by Fyne's software renderer in the app's dark theme, 400 points wide at
// scale 2 (800 px), without the window's header and tab bar. Text rendering can differ slightly
// from a screenshot of the running app; the content is the app's own.
package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"musicalc/internal/i18n"
	"musicalc/internal/logic"
	"musicalc/internal/ui"
)

// darkTheme is the app's theme with the dark colours, as on the help images
type darkTheme struct{ fyne.Theme }

func (t darkTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(name, theme.VariantDark)
}

type screen struct {
	file   string
	prefs  map[string]any
	build  func() *ui.View
	height float32
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: capture <capture folder>")
		os.Exit(2)
	}
	out := os.Args[1]
	if err := os.MkdirAll(out, 0o755); err != nil {
		fail(err)
	}
	alignment := logic.SessionState{
		Preset: logic.Preset{Source: "Snare", ReferenceMeters: 3, DistanceUnit: logic.UnitMeters, SampleRate: 48000, Mics: []logic.PresetMic{
			{Name: "Snare Top", DistanceMeters: 0.1}, {Name: "Overhead L", DistanceMeters: 1.2}, {Name: "Overhead R", DistanceMeters: 1.4},
			{Name: "Room L", DistanceMeters: 3}, {Name: "Room R", DistanceMeters: 3.5},
		}},
		TemperatureC: 20, TemperatureUnit: logic.UnitCelsius,
	}
	kick := map[string]any{"Lo": "50", "Hi": "100", "Freq": "60", "TempC": 20, "TempUnit": logic.UnitCelsius, "DistUnit": logic.UnitMeters,
		"SampleRate": "48000", "Farther": "4", "FartherUnit": logic.UnitMeters, "Basis": "distance", "Meters": 4.0, "Level": "0", "DelayA": "0", "Open": false}
	aligned := map[string]any{}
	for k, v := range kick {
		aligned[k] = v
	}
	aligned["DelayA"] = "11.65"
	three := map[string]any{"source": "Singer A", "ownMeters": 1, "neighbourMeters": 3, "hasOwn": true, "hasNeighbour": true,
		"gainDB": 6, "target": 1, "unit": logic.UnitMeters, "open": false}

	for _, s := range []screen{
		{"cap-alignment.png", map[string]any{"alignmentDelay.session": alignment}, ui.NewAlignmentDelayTab, 680},
		{"cap-sum.png", map[string]any{"twoMicSum.session": kick}, ui.NewTwoMicSumTab, 1100},
		{"cap-sum-aligned.png", map[string]any{"twoMicSum.session": aligned}, ui.NewTwoMicSumTab, 1100},
		{"cap-31.png", map[string]any{"threeToOne.session": three}, ui.NewThreeToOneTab, 900},
	} {
		if err := capture(filepath.Join(out, s.file), s); err != nil {
			fail(err)
		}
		fmt.Println("captured", s.file)
	}
}

func capture(path string, s screen) error {
	app := test.NewApp()
	defer app.Quit()
	for key, v := range s.prefs {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		app.Preferences().SetString(key, string(data))
	}
	app.Preferences().SetString("language", "en")
	ui.InitializeLanguage(app.Preferences())
	i18n.Use("en")
	app.Settings().SetTheme(darkTheme{ui.NewCustomTheme()})
	w := app.NewWindow("MusiCalc")
	view := s.build()
	w.SetContent(view.Content)
	if c, ok := w.Canvas().(software.WindowlessCanvas); ok {
		c.SetScale(2)
	}
	// Wrapped text only knows its height once it has a width: lay out twice, as a shown window does
	w.Resize(fyne.NewSize(400, s.height))
	view.Content.Refresh()
	w.Resize(fyne.NewSize(400, s.height+1))
	w.Resize(fyne.NewSize(400, s.height))
	var img image.Image = w.Canvas().Capture()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
