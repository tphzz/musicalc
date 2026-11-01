package i18n

import "golang.org/x/text/language"

// ResolveLanguage honours an explicit choice, otherwise using the OS language.
func ResolveLanguage(saved, locale string) string {
	if saved == "en" || saved == "de" {
		return saved
	}
	tag, err := language.Parse(locale)
	if err == nil {
		base, _ := tag.Base()
		if base.String() == "de" {
			return "de"
		}
	}
	return "en"
}
