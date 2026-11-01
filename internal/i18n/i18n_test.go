package i18n

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestLanguageDefaultSelectionAndFallback(t *testing.T) {
	defer Use("en")
	for _, tc := range []struct{ code, language, close string }{
		{"", "en", "Close"}, {"en", "en", "Close"}, {"de", "de", "Schließen"}, {"fr", "en", "Close"},
	} {
		Use(tc.code)
		if Language() != tc.language || T("action.close") != tc.close {
			t.Fatalf("%q selected %q: %q", tc.code, Language(), T("action.close"))
		}
	}
	if got := T("missing.key"); got != "missing.key" {
		t.Fatalf("unknown key: %q", got)
	}
}

func TestCatalogsHaveSameKeysAndFormatArguments(t *testing.T) {
	var en, de map[string]string
	for name, dest := range map[string]*map[string]string{"en": &en, "de": &de} {
		data, err := messages.ReadFile("messages/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, dest); err != nil {
			t.Fatal(err)
		}
	}
	keys := func(m map[string]string) []string {
		var keys []string
		for k, v := range m {
			if strings.TrimSpace(v) == "" {
				t.Errorf("%s is empty", k)
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return keys
	}
	if !reflect.DeepEqual(keys(en), keys(de)) {
		t.Fatal("catalog keys differ")
	}
	// printf verbs must preserve their order and type in both catalogs.
	verb := regexp.MustCompile(`%([-+ #0]*[0-9]*(?:\.[0-9]+)?[a-zA-Z%])`)
	expected := []string{"%d", "%.1f", "%.0f", "%.2f"}
	if got := verb.FindAllString("%d taps · %.1f BPM · %.0f samples · %.2f Hz", -1); !reflect.DeepEqual(got, expected) {
		t.Fatalf("format matcher missed arguments: %v", got)
	}
	if reflect.DeepEqual(verb.FindAllString("%d taps · %.1f BPM", -1), verb.FindAllString("%d taps · %.2f BPM", -1)) {
		t.Fatal("incompatible floating-point formats were accepted")
	}
	for key, s := range en {
		if !reflect.DeepEqual(verb.FindAllString(s, -1), verb.FindAllString(de[key], -1)) {
			t.Errorf("%s has incompatible format arguments", key)
		}
	}
}

func TestFormattedMessagesAndPlurals(t *testing.T) {
	defer Use("en")
	for _, code := range []string{"en", "de"} {
		Use(code)
		if got := T("tap.delay_status", 3, 120.5); strings.Contains(got, "%!") || !strings.Contains(got, "120.5") {
			t.Fatalf("formatted status: %q", got)
		}
		for _, n := range []int{0, 1, 2} {
			if got := Plural("preset.imported", n, n); strings.Contains(got, "%!") || !strings.Contains(got, fmt.Sprint(n)) {
				t.Fatalf("plural %d: %q", n, got)
			}
		}
	}
	Use("de")
	if got := Plural("preset.imported", 1, 1); got != "1 Preset importiert." {
		t.Fatal(got)
	}
}

func TestToolkitFollowsAppLanguage(t *testing.T) {
	defer Use("en")
	Use("de")
	if got := Localize("Copy"); got != "Kopieren" {
		t.Fatal(got)
	}
	if got := LocalizeKey("file.parent", "Parent"); got != "Übergeordneter Ordner" {
		t.Fatal(got)
	}
	Use("en")
	if got := Localize("Copy"); got != "Copy" {
		t.Fatal(got)
	}
	if got := Localize("unknown {{.Name}}", map[string]any{"Name": "name"}); got != "unknown name" {
		t.Fatal(got)
	}
}
