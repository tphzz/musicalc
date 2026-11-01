package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"strings"
	"testing"
)

func TestInitializeLanguageUsesOSAndSavedOverride(t *testing.T) {
	app := test.NewTempApp(t)
	defer i18n.Use("en")
	for _, tc := range []struct{ saved, locale, want string }{{"", "de-DE", "de"}, {"system", "de-CH", "de"}, {"en", "de-DE", "en"}, {"de", "en-US", "de"}, {"invalid", "de-DE", "de"}, {"system", "fr-FR", "en"}} {
		app.Preferences().SetString(prefLanguage, tc.saved)
		initializeLanguage(app.Preferences(), tc.locale)
		if i18n.Language() != tc.want || lang.L("Copy") != map[string]string{"en": "Copy", "de": "Kopieren"}[tc.want] {
			t.Fatalf("wrong language for %q / %q", tc.saved, tc.locale)
		}
	}
}

func TestAboutRefreshesExistingTextWithoutChangingVersion(t *testing.T) {
	app := test.NewTempApp(t)
	app.Settings().SetTheme(NewCustomTheme())
	i18n.Use("en")
	defer i18n.Use("en")
	w := app.NewWindow("MusiCalc")
	defer w.Close()
	w.SetContent(widget.NewLabel("calculator"))
	w.Resize(fyne.NewSize(450, 650))
	about := NewAboutDialog("0.9.0-test", w)
	about.Show()
	renderer := test.WidgetRenderer(about)
	about.Hide()
	i18n.Use("de")
	about.RefreshText()
	about.Show()
	defer about.Hide()
	if test.WidgetRenderer(about) != renderer || !strings.Contains(uiText(about), "Über MusiCalc") || !strings.Contains(uiText(about), "0.9.0-test") {
		t.Fatal("About lost its content or translation")
	}
}

func TestHelpRefreshesItsPageAndPreservesScroll(t *testing.T) {
	app := test.NewTempApp(t)
	app.Settings().SetTheme(NewCustomTheme())
	i18n.Use("en")
	defer i18n.Use("en")
	w := app.NewWindow("MusiCalc")
	defer w.Close()
	view := NewHelpView("getting-around", "title.timecode", func() {})
	w.SetContent(view.Content)
	w.Resize(fyne.NewSize(450, 400))
	var scroll *container.Scroll
	visitUI(view.Content, func(object fyne.CanvasObject) {
		if value, ok := object.(*container.Scroll); ok {
			scroll = value
		}
	})
	if scroll == nil {
		t.Fatal("help has no scroll view")
	}
	scroll.Offset = fyne.NewPos(0, 50)
	scroll.Refresh()
	i18n.Use("de")
	view.RefreshText()
	if w.Content() != view.Content || scroll.Offset.Y != 50 || !strings.Contains(uiText(view.Content), "Hilfe") {
		t.Fatal("help refresh replaced the view or lost its scroll position")
	}
}
