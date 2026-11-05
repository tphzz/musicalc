package ui

import (
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/test"
	"musicalc/internal/i18n"
	"testing"
)

func TestLanguagePreferenceAndToolkit(t *testing.T) {
	app := test.NewTempApp(t)
	defer i18n.Use("en")
	for _, tc := range []struct{ saved, want, copy string }{
		{"", "en", "Copy"}, {"de", "de", "Kopieren"}, {"invalid", "en", "Copy"}, {"en", "en", "Copy"},
	} {
		app.Preferences().SetString(prefLanguage, tc.saved)
		initializeLanguage(app.Preferences(), "en-US")
		if i18n.Language() != tc.want || lang.L("Copy") != tc.copy {
			t.Fatalf("saved %q gave %s / %s", tc.saved, i18n.Language(), lang.L("Copy"))
		}
	}
}
