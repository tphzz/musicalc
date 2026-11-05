package ui

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"musicalc/internal/logic"
	sclres "musicalc/internal/logic/scl"
	"musicalc/internal/ui/widgets"
	"strings"
	"testing"
)

func visitUI(root fyne.CanvasObject, visit func(fyne.CanvasObject)) {
	seen := map[fyne.CanvasObject]bool{}
	var walk func(fyne.CanvasObject)
	walk = func(object fyne.CanvasObject) {
		if object == nil || seen[object] {
			return
		}
		seen[object] = true
		visit(object)
		switch object := object.(type) {
		case *fyne.Container:
			for _, child := range object.Objects {
				walk(child)
			}
		case fyne.Widget:
			for _, child := range test.WidgetRenderer(object).Objects() {
				walk(child)
			}
		}
	}
	walk(root)
}
func uiText(root fyne.CanvasObject) string {
	var text []string
	visitUI(root, func(object fyne.CanvasObject) {
		switch object := object.(type) {
		case *canvas.Text:
			text = append(text, object.Text)
		case *widget.Label:
			text = append(text, object.Text)
		case *widget.RichText:
			text = append(text, object.String())
		}
	})
	return strings.Join(text, "\n")
}

func TestGermanCalculatorTextAndPhoneWidth(t *testing.T) {
	defer i18n.Use("en")
	screens := []struct {
		name, want string
		build      func() *View
	}{
		{"timecode", "Leeren", NewTimecodeTab},
		{"delay", "Ganze Note", NewTempoTab},
		{"tempo-change", "Ausgangstempo", NewTempoChangeTab},
		{"tap", "Zum Messen mindestens zweimal tippen", func() *View { return NewTapTempoTab(func() {}) }},
		{"note-frequency", "Gleichstufig", NewDiapasonTab},
		{"frequency-note", "Direktwahl", NewFrequencyToNoteTab},
		{"samples", "Länge in Samples", NewSampleLengthTab},
		{"alignment", "Referenzmikrofon", NewAlignmentDelayTab},
		{"sum", "Mikrofon B gegenüber Mikrofon A", NewTwoMicSumTab},
		{"three-to-one", "Zum Nachbarmikrofon", NewThreeToOneTab},
	}
	for _, screen := range screens {
		widths := []float32{450}
		if fyne.CurrentDevice().IsMobile() {
			widths = []float32{320, 360, 393}
		}
		for _, width := range widths {
			t.Run(fmt.Sprintf("%s-%v", screen.name, width), func(t *testing.T) {
				app := test.NewTempApp(t)
				app.Preferences().SetString(prefLanguage, "de")
				InitializeLanguage(app.Preferences())
				app.Settings().SetTheme(customTheme{isMobile: fyne.CurrentDevice().IsMobile()})
				window := app.NewWindow("MusiCalc")
				defer window.Close()
				window.Resize(fyne.NewSize(width, 650))
				content := screen.build().Content
				window.SetContent(content)
				text := uiText(content)
				if !strings.Contains(text, screen.want) || strings.Contains(text, "%!") {
					t.Fatalf("incorrect localized content: %s", text)
				}
				if actual := window.Canvas().Size().Width; actual > width+1 {
					t.Errorf("German %s expands %v-wide canvas to %v", screen.name, width, actual)
					visitUI(content, func(o fyne.CanvasObject) {
						if o.MinSize().Width < width/3 {
							return
						}
						switch o := o.(type) {
						case *widget.Select:
							t.Logf("select %q placeholder %q min %v", o.Selected, o.PlaceHolder, o.MinSize())
						case *widgets.Button:
							t.Logf("button %q min %v", o.Text, o.MinSize())
						case *widgets.TimecodeEntry:
							t.Logf("timecode %q min %v", o.PlaceHolder, o.MinSize())
						case *widgets.NumericEntry:
							t.Logf("numeric %q min %v", o.PlaceHolder, o.MinSize())
						case *widget.Label:
							t.Logf("label %q min %v", o.Text, o.MinSize())
						case *widget.RichText:
							if len(o.String()) < 80 {
								t.Logf("richtext %q min %v", o.String(), o.MinSize())
							}
						}
					})
				}
				if window.Canvas().Capture().Bounds().Dx() == 0 {
					t.Fatal("empty render")
				}
			})
		}
	}
}

func TestGermanMenusKeepSourceAndTuningValues(t *testing.T) {
	app := test.NewTempApp(t)
	app.Settings().SetTheme(customTheme{})
	i18n.Use("de")
	defer i18n.Use("en")
	window := app.NewWindow("MusiCalc")
	defer window.Close()
	for _, tc := range []struct {
		content     fyne.CanvasObject
		value, text string
	}{
		{NewTwoMicSumTab().Content, "Drums|Kick", "Schlagzeug · Bassdrum"},
		{NewDiapasonTab().Content, sclres.DefaultScaleName, "Gleichstufig · 12-TET"},
	} {
		window.SetContent(tc.content)
		found := false
		visitUI(tc.content, func(object fyne.CanvasObject) {
			if selectMenu, ok := object.(*widgets.MenuSelect); ok && selectMenu.Selected() == tc.value {
				found = true
				if selectMenu.Text() != tc.text {
					t.Errorf("label %q, want %q", selectMenu.Text(), tc.text)
				}
			}
		})
		if !found {
			t.Fatalf("stored value %q changed", tc.value)
		}
	}
	if presetFileFilter().Name != "MusiCalc-Presets" {
		t.Fatal("file filter was translated before language selection")
	}
}

func TestGermanTwoMicSessionRetainsPolarityAndPrecision(t *testing.T) {
	app := test.NewTempApp(t)
	app.Settings().SetTheme(customTheme{})
	i18n.Use("de")
	defer i18n.Use("en")
	initial := twoMicSession{Lo: "50", Hi: "100", Freq: "60", TempC: 20, TempUnit: logic.UnitCelsius, DistUnit: logic.UnitMeters,
		SampleRate: "48000", Farther: "4.123456", FartherUnit: logic.UnitMeters, Level: "-6", Inverted: true, DelayA: "0.123456", Open: true}
	data, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	app.Preferences().SetString(prefTwoMicSession, string(data))
	NewTwoMicSumTab()
	var saved twoMicSession
	if err = json.Unmarshal([]byte(app.Preferences().String(prefTwoMicSession)), &saved); err != nil {
		t.Fatal(err)
	}
	if !saved.Inverted || saved.Farther != initial.Farther || saved.DelayA != initial.DelayA || saved.DistUnit != logic.UnitMeters {
		t.Fatalf("session changed: %#v", saved)
	}
}

func TestLocalizedContextMenuAndPopupShadow(t *testing.T) {
	app := test.NewTempApp(t)
	app.Preferences().SetString(prefLanguage, "de")
	InitializeLanguage(app.Preferences())
	defer i18n.Use("en")
	app.Settings().SetTheme(customTheme{})
	window := app.NewWindow("MusiCalc")
	defer window.Close()
	entry := widget.NewEntry()
	entry.SetText("copy me")
	window.SetContent(entry)
	window.Resize(fyne.NewSize(450, 650))
	test.TapSecondary(entry)
	text := ""
	for _, overlay := range window.Canvas().Overlays().List() {
		text += uiText(overlay)
	}
	if !strings.Contains(text, "Kopieren") || !strings.Contains(text, "Einfügen") {
		t.Fatalf("context menu is not German: %s", text)
	}
	popup := widget.NewPopUpMenu(fyne.NewMenu("", fyne.NewMenuItem("Sprache…", func() {})), window.Canvas())
	popup.ShowAtPosition(fyne.NewPos(10, 50))
	shadow := false
	for _, object := range test.WidgetRenderer(popup).Objects() {
		if background, ok := object.(*canvas.Rectangle); ok && background.Shadow.BlurRadius > 0 {
			_, _, _, alpha := background.Shadow.Color.RGBA()
			shadow = alpha > 0
		}
	}
	if !shadow {
		t.Fatal("popup has no visible separation shadow")
	}
	popup.Hide()
}
