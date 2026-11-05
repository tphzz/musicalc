package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
	"musicalc/internal/i18n"
)

const prefLanguage = "language"

// InitializeLanguage resolves the saved choice or OS language before building widgets.
func InitializeLanguage(preferences fyne.Preferences) {
	initializeLanguage(preferences, string(lang.SystemLocale()))
}

func initializeLanguage(preferences fyne.Preferences, locale string) {
	i18n.Use(i18n.ResolveLanguage(preferences.String(prefLanguage), locale))
	lang.L = i18n.Localize
	lang.N = i18n.LocalizePlural
	lang.X = i18n.LocalizeKey
	lang.XN = i18n.LocalizePluralKey
}
