// Package i18n contains the application's embedded English and German messages.
// The active catalog can change while existing views refresh their text.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	gotext "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed messages/*.json
var messages embed.FS

type catalog struct {
	code    string
	text    map[string]string
	toolkit *gotext.Localizer
}

var (
	catalogs = loadCatalogs()
	current  atomic.Pointer[catalog]
)

func init() { Use("en") }

func loadCatalogs() map[string]*catalog {
	bundle := gotext.NewBundle(language.English)
	result := map[string]*catalog{}
	for _, code := range []string{"en", "de"} {
		data, err := messages.ReadFile("messages/" + code + ".json")
		if err != nil {
			panic(err)
		}
		texts := map[string]string{}
		if err = json.Unmarshal(data, &texts); err != nil {
			panic(err)
		}
		var builtin []*gotext.Message
		for key, text := range texts {
			if strings.HasPrefix(key, "toolkit.") {
				builtin = append(builtin, &gotext.Message{ID: strings.TrimPrefix(key, "toolkit."), Other: text})
			}
		}
		if err = bundle.AddMessages(language.Make(code), builtin...); err != nil {
			panic(err)
		}
		result[code] = &catalog{code: code, text: texts}
	}
	for code, c := range result {
		c.toolkit = gotext.NewLocalizer(bundle, code, "en")
	}
	return result
}

// Use selects a supported language. Empty or unsupported settings use English.
func Use(code string) {
	c := catalogs[code]
	if c == nil {
		c = catalogs["en"]
	}
	current.Store(c)
}

func Language() string { return current.Load().code }

// T looks up a message and formats its optional printf arguments. English is
// the fallback for untranslated keys; an unknown key remains visible for diagnosis.
func T(key string, args ...any) string {
	text, ok := current.Load().text[key]
	if !ok {
		text, ok = catalogs["en"].text[key]
	}
	if !ok {
		return key
	}
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}

// Plural uses the singular for one, and the plural for zero and all other counts.
func Plural(key string, count int, args ...any) string {
	if count == 1 {
		return T(key+".one", args...)
	}
	return T(key+".other", args...)
}

// Localize and its variants adapt Fyne's public translation callbacks so its
// context menus and dialogs use the selected app language rather than the OS locale.
func Localize(text string, data ...any) string             { return LocalizeKey(text, text, data...) }
func LocalizeKey(key, fallback string, data ...any) string { return localize(key, fallback, nil, data) }
func LocalizePlural(text string, count int, data ...any) string {
	return LocalizePluralKey(text, text, count, data...)
}
func LocalizePluralKey(key, fallback string, count int, data ...any) string {
	return localize(key, fallback, count, data)
}
func localize(key, fallback string, count any, data []any) string {
	var values any
	if len(data) > 0 {
		values = data[0]
	}
	text, err := current.Load().toolkit.Localize(&gotext.LocalizeConfig{
		DefaultMessage: &gotext.Message{ID: key, Other: fallback},
		TemplateData:   values, PluralCount: count,
	})
	if err != nil && text == "" {
		return fallback
	}
	return text
}
