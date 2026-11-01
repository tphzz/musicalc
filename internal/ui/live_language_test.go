package ui

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"
	"strings"
	"testing"
)

func TestExistingCalculatorViewsRelabelInPlace(t *testing.T) {
	for _, screen := range []struct {
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
	} {
		t.Run(screen.name, func(t *testing.T) {
			app := test.NewTempApp(t)
			app.Preferences().SetString(prefLanguage, "en")
			InitializeLanguage(app.Preferences())
			app.Settings().SetTheme(NewCustomTheme())
			i18n.Use("en")
			defer i18n.Use("en")
			window := app.NewWindow("MusiCalc")
			defer window.Close()
			view := screen.build()
			window.SetContent(view.Content)
			window.Resize(fyne.NewSize(450, 650))
			type inputState struct {
				entry    *widget.Entry
				text     string
				row, col int
			}
			var inputs []inputState
			var focus fyne.Focusable
			visitUI(view.Content, func(object fyne.CanvasObject) {
				var entry *widget.Entry
				switch value := object.(type) {
				case *widgets.NumericEntry:
					entry = &value.Entry
				case *widgets.TimecodeEntry:
					entry = &value.Entry
				case *widget.Entry:
					entry = value
				}
				if entry != nil {
					if focus == nil {
						focus = object.(fyne.Focusable)
					}
					inputs = append(inputs, inputState{entry, entry.Text, entry.CursorRow, entry.CursorColumn})
				}
			})
			if focus != nil {
				window.Canvas().Focus(focus)
			}
			for _, code := range []string{"de", "en", "de"} {
				i18n.Use(code)
				view.RefreshText()
				if code == "de" && !strings.Contains(uiText(view.Content), screen.want) {
					t.Fatalf("existing %s view did not switch to German", screen.name)
				}
				if window.Content() != view.Content || window.Canvas().Focused() != focus {
					t.Fatal("language refresh replaced the content or focus")
				}
				for _, input := range inputs {
					if input.entry.Text != input.text || input.entry.CursorRow != input.row || input.entry.CursorColumn != input.col {
						t.Fatal("language refresh rewrote an input")
					}
				}
			}
		})
	}
}

func TestMultiMicResultsRelabelWithoutChangingStoredSessions(t *testing.T) {
	cases := []struct {
		name, key string
		state     any
		build     func() *View
		messages  []string
	}{
		{"alignment", prefAlignmentSession, logic.SessionState{Preset: logic.Preset{Source: "Keep my source", ReferenceMeters: 3, DistanceUnit: logic.UnitMeters, SampleRate: 48000, Mics: []logic.PresetMic{{Name: "OH L", DistanceMeters: 1}, {Name: "OH R", DistanceMeters: 1.2}}}, TemperatureC: 20, TemperatureUnit: logic.UnitCelsius}, NewAlignmentDelayTab, []string{"alignment.reference"}},
		{"sum", prefTwoMicSession, twoMicSession{Lo: "50", Hi: "100", Freq: "60", TempC: 20, TempUnit: logic.UnitCelsius, DistUnit: logic.UnitMeters, SampleRate: "48000", Farther: "1.23456789", FartherUnit: logic.UnitMeters, Level: "-2.50", Inverted: true, DelayA: "0.234567", Open: false}, NewTwoMicSumTab, []string{"sum.comb_title", "sum.compared"}},
		{"three", prefThreeToOneSession, threeToOneSession{Source: "Keep my source", OwnMeters: 1, NeighbourMeters: 4, HasOwn: true, HasNeighbour: true, Target: 0, Unit: logic.UnitMeters, Open: false}, NewThreeToOneTab, []string{"three.meets", "three.spill"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := test.NewTempApp(t)
			app.Preferences().SetString(prefLanguage, "en")
			InitializeLanguage(app.Preferences())
			defer i18n.Use("en")
			app.Settings().SetTheme(NewCustomTheme())
			data, err := json.Marshal(tc.state)
			if err != nil {
				t.Fatal(err)
			}
			app.Preferences().SetString(tc.key, string(data))
			window := app.NewWindow("MusiCalc")
			defer window.Close()
			view := tc.build()
			window.SetContent(view.Content)
			window.Resize(fyne.NewSize(450, 650))
			stored := app.Preferences().String(tc.key)
			selections := map[*widget.Select]int{}
			menuValues := map[*widgets.MenuSelect]string{}
			visitUI(view.Content, func(object fyne.CanvasObject) {
				switch value := object.(type) {
				case *widget.Select:
					selections[value] = value.SelectedIndex()
				case *widgets.MenuSelect:
					menuValues[value] = value.Selected()
				}
			})
			for _, code := range []string{"de", "en", "de"} {
				i18n.Use(code)
				view.RefreshText()
				text := uiText(view.Content)
				for _, key := range tc.messages {
					if !strings.Contains(text, i18n.T(key)) {
						t.Fatalf("%s result did not refresh %s", code, key)
					}
				}
				if app.Preferences().String(tc.key) != stored {
					t.Fatal("language change saved a different session")
				}
				for control, index := range selections {
					if control.SelectedIndex() != index {
						t.Fatal("language change altered a semantic choice")
					}
				}
				for control, value := range menuValues {
					if control.Selected() != value {
						t.Fatal("language change altered a source/tuning ID")
					}
				}
			}
		})
	}
}
