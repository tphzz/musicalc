package help

import (
	"path"
	"regexp"
	"strings"
	"testing"

	"musicalc/internal/i18n"
)

func TestGermanHelpIsCompleteAndImagesResolve(t *testing.T) {
	ref := regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)
	matches := ref.FindAllStringSubmatch("![Setup](../images/alignment-setup.png)", -1)
	if len(matches) != 1 || matches[0][1] != "../images/alignment-setup.png" {
		t.Fatalf("image matcher missed a Markdown image: %v", matches)
	}
	for _, id := range append([]string{GettingAround}, pageIDs()...) {
		src := SourceForLanguage(id, "de")
		if !strings.HasPrefix(src, "# ") || src == Source(id) {
			t.Errorf("%s lacks a German help page", id)
		}
		if id != GettingAround {
			for _, section := range []string{"## Einsatz", "## Funktionsweise", "## Bedienung", "## Ergebnisse", "## Praxis", "## Beispiele", "## Hinweise"} {
				if !strings.Contains(src, section+"\n") {
					t.Errorf("%s lacks %s", id, section)
				}
			}
		}
		for _, m := range ref.FindAllStringSubmatch(src, -1) {
			resolved := path.Clean(path.Join("de", m[1]))
			if _, err := files.ReadFile(resolved); err != nil {
				t.Errorf("%s: broken relative image reference %s: %v", id, m[1], err)
			}
			if Image(path.Base(m[1])) == nil {
				t.Errorf("%s: missing image %s", id, m[1])
			}
		}
	}
}

func TestHelpSelectionUsesAppLanguageAndEnglishFallback(t *testing.T) {
	defer i18n.Use("en")
	i18n.Use("de")
	if got := Markdown("tempo-to-delay"); !strings.Contains(got, "Verzögerungszeiten") {
		t.Fatal("German help not selected")
	}
	if got := Source("tempo-to-delay"); !strings.HasPrefix(got, "# Tempo to Delay\n") {
		t.Fatal("canonical English source changed")
	}
	if got := SourceForLanguage("tempo-to-delay", "fr"); got != Source("tempo-to-delay") {
		t.Fatal("unsupported language did not use English")
	}
	if got := SourceForLanguage("missing", "de"); got != "" {
		t.Fatal(got)
	}
	// An absent translation also falls back to English.
	if got := SourceForLanguage("tempo-to-delay", "missing-language"); got != Source("tempo-to-delay") {
		t.Fatal("missing translation did not use English")
	}
}
