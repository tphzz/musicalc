package i18n

import "testing"

func TestResolveLanguageUsesSystemUntilExplicitlyOverridden(t *testing.T) {
	for _, tc := range []struct{ saved, locale, want string }{
		{"", "de-DE", "de"}, {"system", "de-AT", "de"}, {"system", "de-CH", "de"},
		{"", "en-US", "en"}, {"system", "fr-FR", "en"}, {"system", "", "en"},
		{"system", "not_a_locale", "en"}, {"invalid", "de-DE", "de"},
		{"en", "de-DE", "en"}, {"de", "en-US", "de"},
	} {
		if got := ResolveLanguage(tc.saved, tc.locale); got != tc.want {
			t.Errorf("%q / %q: got %q want %q", tc.saved, tc.locale, got, tc.want)
		}
	}
}
