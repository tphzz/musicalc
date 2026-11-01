package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"testing"
)

func TestLanguageChoicesHandleDirectMobileKeyDelivery(t *testing.T) {
	app := test.NewTempApp(t)
	app.Settings().SetTheme(NewCustomTheme())
	app.Preferences().SetString(prefLanguage, "en")
	InitializeLanguage(app.Preferences())
	defer i18n.Use("en")
	window := app.NewWindow("MusiCalc")
	defer window.Close()
	entry := widget.NewEntry()
	entry.SetText("123.456789")
	window.SetContent(entry)
	window.Canvas().Focus(entry)
	dialog := NewPreferencesDialog(app.Preferences(), window, nil, func() {})
	for _, key := range []fyne.KeyName{fyne.KeyEscape, mobile.KeyBack} {
		var choices []fyne.Focusable
		visitUI(dialog.choices, func(object fyne.CanvasObject) {
			if choice, ok := object.(fyne.Focusable); ok {
				choices = append(choices, choice)
			}
		})
		if len(choices) != 3 {
			t.Fatalf("expected three focusable choices, got %d", len(choices))
		}
		for _, choice := range choices {
			dialog.Show()
			window.Canvas().Focus(choice)
			// This is the mobile driver's route: no canvas shortcut dispatch.
			window.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: key})
			if dialog.Visible() {
				t.Fatalf("%s did not dismiss the focused choice", key)
			}
			if window.Canvas().Focused() != entry || entry.Text != "123.456789" {
				t.Fatal("dismissal lost the prior state")
			}
		}
	}
}
