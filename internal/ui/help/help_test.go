package help

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var imageRef = regexp.MustCompile(`!\[[^\]]*\]\(images/([^)\s]+)\)`)

func TestEveryPageExists(t *testing.T) {
	for _, id := range append([]string{GettingAround}, pageIDs()...) {
		src := Source(id)
		if strings.TrimSpace(src) == "" {
			t.Errorf("%s.md is missing or empty", id)
			continue
		}
		if !strings.HasPrefix(src, "# ") {
			t.Errorf("%s.md should start with its # title", id)
		}
		if strings.TrimSpace(Markdown(id)) == "" {
			t.Errorf("%s.md has no text after its title", id)
		}
	}
}

// The calculator pages share one structure, so they read the same way
func TestPagesFollowTheStructure(t *testing.T) {
	sections := []string{"## When to use it", "## How it works", "## How to use it", "## Reading the results",
		"## In practice", "## Examples", "## Good to know"}
	for _, p := range Pages {
		src := Source(p.ID)
		if !strings.HasPrefix(src, "# "+p.Title+"\n") {
			t.Errorf("%s.md should start with \"# %s\"", p.ID, p.Title)
		}
		last := -1
		for _, s := range sections {
			i := strings.Index(src, "\n"+s+"\n")
			if i < 0 {
				t.Errorf("%s.md lacks %q", p.ID, s)
				continue
			}
			if i < last {
				t.Errorf("%s.md: %q is out of order", p.ID, s)
			}
			last = i
		}
	}
}

func TestImagesExistAndAreUsed(t *testing.T) {
	used := map[string]bool{}
	for _, id := range append([]string{GettingAround}, pageIDs()...) {
		for _, m := range imageRef.FindAllStringSubmatch(Source(id), -1) {
			used[m[1]] = true
			if Image(m[1]) == nil {
				t.Errorf("%s.md refers to images/%s, which is not embedded", id, m[1])
			}
		}
	}
	for _, name := range imageNames() {
		if !used[name] {
			t.Errorf("images/%s is embedded but no page uses it", name)
		}
	}
}

func TestImagesFitThePhone(t *testing.T) {
	for _, name := range imageNames() {
		s := imageSize(Image(name))
		if s.Width > imageMaxWidth || s.Width <= 0 || s.Height <= 0 {
			t.Errorf("images/%s is shown at %v", name, s)
		}
	}
}

func TestMarkdownDropsTheTitle(t *testing.T) {
	md := Markdown(Pages[0].ID)
	if strings.HasPrefix(md, "#") && !strings.HasPrefix(md, "##") {
		t.Errorf("the title should be removed: %q", md[:40])
	}
}

// README.md is the way into the help files outside the app, so it must link every page
func TestReadmeLinksEveryPage(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range append([]string{GettingAround}, pageIDs()...) {
		if link := "](internal/ui/help/" + id + ".md)"; !strings.Contains(string(data), link) {
			t.Errorf("README.md doesn't link %s.md", id)
		}
	}
}

func pageIDs() []string {
	ids := make([]string, len(Pages))
	for i, p := range Pages {
		ids[i] = p.ID
	}
	return ids
}
