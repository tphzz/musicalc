package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"testing"
)

func TestLanguageChoiceKeyboardKeepsRequiredSelection(t *testing.T) {
	app := test.NewTempApp(t)
	app.Settings().SetTheme(NewCustomTheme())
	app.Preferences().SetString(prefLanguage, "en")
	InitializeLanguage(app.Preferences())
	defer i18n.Use("en")
	window := app.NewWindow("MusiCalc")
	defer window.Close()
	window.SetContent(widget.NewLabel("calculator"))
	dialog := NewPreferencesDialog(app.Preferences(), window, nil, func() {})
	dialog.Show()
	window.Canvas().Focus(dialog.choices.items[1])
	window.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	if dialog.choices.Selected != "de" || app.Preferences().String(prefLanguage) != "de" || window.Canvas().Focused() != dialog.choices.items[2] {
		t.Fatal("arrow selection lost its value or focus")
	}
	test.Tap(dialog.choices.items[2])
	if dialog.choices.Selected != "de" {
		t.Fatal("a selected choice was cleared")
	}
	dialog.choices.SetSelected("")
	if dialog.choices.Selected != "de" {
		t.Fatal("empty selection was accepted")
	}
	window.Canvas().Focus(dialog.choices.items[2])
	window.Canvas().Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
	if dialog.choices.Selected != "en" || app.Preferences().String(prefLanguage) != "en" || i18n.Language() != "en" {
		t.Fatal("keyboard did not change the active language")
	}
	dialog.Hide()
}
