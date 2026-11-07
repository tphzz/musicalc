package ui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"musicalc/internal/ui/widgets"
)

// openEnglishTab shows a calculator in an English test app with its own preferences
func openEnglishTab(t *testing.T, build func() *View) fyne.CanvasObject {
	t.Helper()
	root, _ := openEnglishTabWith(t, nil, build)
	return root
}

// openEnglishTabWith is openEnglishTab with stored preferences (such as a saved session) set
// before the calculator is built; it also returns the window, e.g. to look for dialogs
func openEnglishTabWith(t *testing.T, stored map[string]string, build func() *View) (fyne.CanvasObject, fyne.Window) {
	t.Helper()
	app := test.NewTempApp(t)
	for key, value := range stored {
		app.Preferences().SetString(key, value)
	}
	app.Preferences().SetString(prefLanguage, "en")
	InitializeLanguage(app.Preferences())
	i18n.Use("en")
	t.Cleanup(func() { i18n.Use("en") })
	window := app.NewWindow("MusiCalc")
	t.Cleanup(window.Close)
	view := build()
	window.SetContent(view.Content)
	window.Resize(fyne.NewSize(450, 650))
	return view.Content, window
}

// visibleText is the text a user can see: like uiText, but hidden objects and everything inside
// them are skipped, so a hidden stale result cannot satisfy an assertion
func visibleText(root fyne.CanvasObject) string {
	var text []string
	seen := map[fyne.CanvasObject]bool{}
	var walk func(fyne.CanvasObject)
	walk = func(object fyne.CanvasObject) {
		if object == nil || seen[object] || !object.Visible() {
			return
		}
		seen[object] = true
		switch x := object.(type) {
		case *canvas.Text:
			text = append(text, x.Text)
		case *widget.Label:
			text = append(text, x.Text)
		case *widget.RichText:
			text = append(text, x.String())
		}
		switch x := object.(type) {
		case *fyne.Container:
			for _, child := range x.Objects {
				walk(child)
			}
		case fyne.Widget:
			for _, child := range test.WidgetRenderer(x).Objects() {
				walk(child)
			}
		}
	}
	walk(root)
	return strings.Join(text, "\n")
}

// eventually polls cond for up to two seconds, for screens that update through data bindings
func eventually(cond func() bool) bool {
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		if cond() {
			return true
		}
	}
	return cond()
}

// hasErrorDialog reports whether an error dialog is open in the window
func hasErrorDialog(w fyne.Window) bool {
	return w.Canvas().Overlays().Top() != nil
}

// numericEntries returns a screen's numeric entries in layout order
func numericEntries(root fyne.CanvasObject) []*widgets.NumericEntry {
	var entries []*widgets.NumericEntry
	visitUI(root, func(object fyne.CanvasObject) {
		if e, ok := object.(*widgets.NumericEntry); ok {
			entries = append(entries, e)
		}
	})
	return entries
}

// press activates a button the way a finger does, which also works for mouse clicks on desktop
func press(b *widgets.Button) {
	b.TouchDown(&mobile.TouchEvent{})
	b.TouchUp(&mobile.TouchEvent{})
}

// buttonNamed returns the screen's button with the given text
func buttonNamed(t *testing.T, root fyne.CanvasObject, text string) *widgets.Button {
	t.Helper()
	var found *widgets.Button
	visitUI(root, func(object fyne.CanvasObject) {
		if b, ok := object.(*widgets.Button); ok && b.Text == text {
			found = b
		}
	})
	if found == nil {
		t.Fatalf("no %q button", text)
	}
	return found
}
