package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"musicalc/internal/platform/displayscale"
	"testing"
)

func TestPreferencesChangesLanguageImmediatelyAndKeepsTheSession(t *testing.T) {
	app := test.NewTempApp(t)
	app.Preferences().SetString(prefLanguage, "en")
	InitializeLanguage(app.Preferences())
	defer i18n.Use("en")
	app.Settings().SetTheme(NewCustomTheme())
	w := app.NewWindow("MusiCalc")
	defer w.Close()
	entry := widget.NewEntry()
	entry.SetText("123.456789")
	w.SetContent(entry)
	w.Resize(fyne.NewSize(450, 650))
	w.Canvas().Focus(entry)
	changes := 0
	d := NewPreferencesDialog(app.Preferences(), w, nil, func() { changes++ })
	d.Show()
	d.Show()
	if len(w.Canvas().Overlays().List()) != 1 || w.Canvas().Focused() == entry {
		t.Fatal("invalid modal focus or overlay")
	}
	d.choices.SetSelected("de")
	if i18n.Language() != "de" || app.Preferences().String(prefLanguage) != "de" || changes != 1 {
		t.Fatal("language did not apply immediately")
	}
	if d.title.Text != "Einstellungen" {
		t.Fatal("Preferences did not translate itself")
	}
	d.choices.SetSelected("en")
	if i18n.Language() != "en" || changes != 2 {
		t.Fatal("language did not switch back")
	}
	d.choices.SetSelected("system")
	if app.Preferences().String(prefLanguage) != "system" || changes != 3 {
		t.Fatal("System was not saved as a preference")
	}
	d.Hide()
	if w.Content() != entry || entry.Text != "123.456789" || w.Canvas().Focused() != entry {
		t.Fatal("Preferences changed calculator state")
	}
}

func TestPreferencesZoomIsLiveAndRespectsLaunchOverride(t *testing.T) {
	if fyne.CurrentDevice().IsMobile() {
		t.Skip("desktop zoom")
	}
	for _, override := range []string{"", "1.5"} {
		t.Run(override, func(t *testing.T) {
			t.Setenv("FYNE_SCALE", override)
			app := test.NewTempApp(t)
			app.Preferences().SetString(prefLanguage, "en")
			InitializeLanguage(app.Preferences())
			defer i18n.Use("en")
			app.Settings().SetTheme(NewCustomTheme())
			scale, err := displayscale.NewController(app)
			if err != nil {
				t.Fatal(err)
			}
			w := app.NewWindow("MusiCalc")
			defer w.Close()
			w.SetContent(widget.NewLabel("calculator"))
			w.Resize(fyne.NewSize(450, 650))
			d := NewPreferencesDialog(app.Preferences(), w, scale, func() {})
			d.Show()
			defer d.Hide()
			if override != "" {
				if !d.scaleSelect.Disabled() || d.scaleSelect.Selected != "150%" {
					t.Fatal("launch override is not reflected")
				}
				return
			}
			for _, percent := range displayscale.Percents() {
				option := fmt.Sprintf("%d%%", percent)
				if percent == 100 {
					option = i18n.T("preferences.scale_default", 100)
				}
				d.scaleSelect.SetSelected(option)
				if scale.Percent() != percent || app.Preferences().Int("displayScalePercent") != percent {
					t.Fatal("zoom was not applied immediately")
				}
			}
			d.choices.SetSelected("de")
			if d.scaleSelect.Selected != "200%" {
				t.Fatal("translation changed the selected zoom")
			}
		})
	}
}

func TestPreferencesFitsShortViewportsAndClosesFromFocusedControls(t *testing.T) {
	app := test.NewTempApp(t)
	app.Preferences().SetString(prefLanguage, "en")
	InitializeLanguage(app.Preferences())
	defer i18n.Use("en")
	app.Settings().SetTheme(NewCustomTheme())
	w := app.NewWindow("MusiCalc")
	defer w.Close()
	entry := widget.NewEntry()
	entry.SetText("keep me")
	w.SetContent(entry)
	var scale *displayscale.Controller
	if !fyne.CurrentDevice().IsMobile() {
		t.Setenv("FYNE_SCALE", "")
		var err error
		scale, err = displayscale.NewController(app)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, width := range []float32{320, 360, 393, 450} {
		w.Resize(fyne.NewSize(width, 240))
		w.Canvas().Focus(entry)
		d := NewPreferencesDialog(app.Preferences(), w, scale, func() {})
		d.Show()
		if (d.scaleSelect == nil) != fyne.CurrentDevice().IsMobile() {
			t.Fatal("wrong platform controls")
		}
		root := test.WidgetRenderer(d).Objects()[0].(*fyne.Container)
		panel := root.Objects[1]
		if panel.Position().X < 15 || panel.Position().Y < 15 || panel.Position().X+panel.Size().Width > width-15 || panel.Position().Y+panel.Size().Height > w.Canvas().Size().Height-15 {
			t.Fatal("dialog overflows the viewport")
		}
		closeY := d.close.Position().Y + d.close.Size().Height
		if closeY > panel.Size().Height-15 {
			t.Fatal("Close is clipped")
		}
		var controls []fyne.Focusable
		for _, choice := range d.choices.items {
			controls = append(controls, choice)
		}
		controls = append(controls, d.close)
		if d.scaleSelect != nil {
			controls = append(controls, d.scaleSelect)
		}
		for _, key := range []fyne.KeyName{fyne.KeyEscape, mobile.KeyBack} {
			for _, control := range controls {
				if !d.Visible() {
					d.Show()
				}
				w.Canvas().Focus(control)
				w.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: key})
				if d.Visible() || w.Canvas().Focused() != entry || entry.Text != "keep me" {
					t.Fatal("focused dismissal lost modal/session state")
				}
			}
		}
	}
}

func TestPreferencesDismissesTheZoomPopupBeforeTheDialog(t *testing.T) {
	if fyne.CurrentDevice().IsMobile() {
		t.Skip("desktop zoom")
	}
	t.Setenv("FYNE_SCALE", "")
	app := test.NewTempApp(t)
	app.Preferences().SetString(prefLanguage, "en")
	InitializeLanguage(app.Preferences())
	defer i18n.Use("en")
	app.Settings().SetTheme(NewCustomTheme())
	scale, err := displayscale.NewController(app)
	if err != nil {
		t.Fatal(err)
	}
	w := app.NewWindow("MusiCalc")
	defer w.Close()
	w.SetContent(widget.NewLabel("calculator"))
	w.Resize(fyne.NewSize(450, 650))
	d := NewPreferencesDialog(app.Preferences(), w, scale, func() {})
	d.Show()
	d.scaleSelect.Tapped(&fyne.PointEvent{})
	if len(w.Canvas().Overlays().List()) != 2 {
		t.Fatal("zoom popup did not open")
	}
	d.scaleSelect.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if !d.Visible() || len(w.Canvas().Overlays().List()) != 1 {
		t.Fatal("Escape must dismiss the child popup first")
	}
	d.scaleSelect.Tapped(&fyne.PointEvent{})
	d.Hide()
	if len(w.Canvas().Overlays().List()) != 0 {
		t.Fatal("dialog left its child popup behind")
	}
}
