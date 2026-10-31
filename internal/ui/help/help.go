// Package help holds the help pages of the calculators: one Markdown file per calculator, with
// its images, all embedded in the app. README.md links to the same files, which GitHub renders
// with their images, so the in-app help is also the user guide.
package help

import (
	"embed"
	"strings"

	"fyne.io/fyne/v2"
)

//go:embed *.md images/*.png
var files embed.FS

// Page is one help page
type Page struct {
	ID       string // file name without .md
	Title    string
	Category string
}

// GettingAround is the page about the app as a whole; it opens the user guide
const GettingAround = "getting-around"

// Pages lists the calculators' help pages in the order of the app's categories and tabs
var Pages = []Page{
	{"timecode", "Timecode", "Time & Tempo"},
	{"tempo-to-delay", "Tempo to Delay", "Time & Tempo"},
	{"tempo-change", "Tempo Change", "Time & Tempo"},
	{"tap-tempo", "Tap Tempo", "Time & Tempo"},
	{"note-to-frequency", "Note to Frequency", "Frequency & Pitch"},
	{"frequency-to-note", "Frequency to Note", "Frequency & Pitch"},
	{"sample-length", "Sample Length", "Analysis"},
	{"alignment-delay", "Alignment Delay", "Multi-Mic"},
	{"two-mic-sum", "Two-Mic Sum", "Multi-Mic"},
	{"three-to-one", "3:1 Rule", "Multi-Mic"},
}

// Source returns a page's Markdown as written, or "" when there is no such page
func Source(id string) string {
	data, err := files.ReadFile(id + ".md")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

// Markdown returns a page for showing in the app: without its "# Title" line, which the help
// view shows in its title bar instead
func Markdown(id string) string {
	src := Source(id)
	if strings.HasPrefix(src, "# ") {
		if i := strings.Index(src, "\n"); i >= 0 {
			return strings.TrimLeft(src[i+1:], "\n")
		}
		return ""
	}
	return src
}

// Image returns an embedded image by its file name, or nil
func Image(name string) fyne.Resource {
	data, err := files.ReadFile("images/" + name)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource(name, data)
}

// imageNames lists the embedded images
func imageNames() []string {
	entries, _ := files.ReadDir("images")
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}
