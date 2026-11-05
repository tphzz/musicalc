package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"musicalc/internal/i18n"
	"musicalc/internal/ui"
	"musicalc/internal/ui/widgets"
	"strings"
	"testing"
)

func walkWindowObjects(root fyne.CanvasObject, visit func(fyne.CanvasObject)) {
	seen := map[fyne.CanvasObject]bool{}
	var walk func(fyne.CanvasObject)
	walk = func(object fyne.CanvasObject) {
		if object == nil || seen[object] {
			return
		}
		seen[object] = true
		visit(object)
		switch value := object.(type) {
		case *fyne.Container:
			for _, child := range value.Objects {
				walk(child)
			}
		case fyne.Widget:
			for _, child := range test.WidgetRenderer(value).Objects() {
				walk(child)
			}
		}
	}
	walk(root)
}
func windowText(root fyne.CanvasObject) string {
	var values []string
	walkWindowObjects(root, func(object fyne.CanvasObject) {
		switch value := object.(type) {
		case *widget.Label:
			values = append(values, value.Text)
		case *canvas.Text:
			values = append(values, value.Text)
		}
	})
	return strings.Join(values, "\n")
}

func TestMainPreferencesRefreshKeepsActiveAndCachedCalculators(t *testing.T) {
	app := test.NewTempApp(t)
	app.Preferences().SetString("language", "en")
	ui.InitializeLanguage(app.Preferences())
	defer i18n.Use("en")
	app.Settings().SetTheme(ui.NewCustomTheme())
	w := buildMainWindow(app, nil)
	defer w.Close()
	shell := w.Content().(*fyne.Container)
	body := shell.Objects[1]
	if scroll, ok := body.(*container.Scroll); ok {
		body = scroll.Content
	}
	stack := body.(*fyne.Container)
	var bar *widgets.TabBar
	var menuButton *widgets.Button
	walkWindowObjects(shell, func(object fyne.CanvasObject) {
		if value, ok := object.(*widgets.TabBar); ok {
			bar = value
		}
		if value, ok := object.(*widgets.Button); ok && value.Icon != nil && value.Icon.Name() == theme.MenuIcon().Name() {
			menuButton = value
		}
	})
	if bar == nil || menuButton == nil {
		t.Fatal("missing shell controls")
	}
	bar.Select(2)
	active := stack.Objects[0]
	var input *widgets.NumericEntry
	walkWindowObjects(active, func(object fyne.CanvasObject) {
		if value, ok := object.(*widgets.NumericEntry); ok && input == nil {
			input = value
		}
	})
	if input == nil {
		t.Fatal("missing calculator input")
	}
	input.SetText("123.456789")
	w.Canvas().Focus(input)
	menuButton.OnTapped()
	var popup *widget.PopUpMenu
	for _, layer := range w.Canvas().Overlays().List() {
		walkWindowObjects(layer, func(object fyne.CanvasObject) {
			if value, ok := object.(*widget.PopUpMenu); ok {
				popup = value
			}
		})
	}
	if popup == nil || len(popup.Items) < 6 {
		t.Fatal("category menu did not open")
	}
	test.Tap(popup.Items[5].(fyne.Tappable))
	var preferences *ui.PreferencesDialog
	for _, layer := range w.Canvas().Overlays().List() {
		if value, ok := layer.(*ui.PreferencesDialog); ok {
			preferences = value
		}
	}
	if preferences == nil {
		t.Fatal("Preferences action is not wired")
	}
	var german fyne.Tappable
	walkWindowObjects(preferences, func(object fyne.CanvasObject) {
		if value, ok := object.(interface {
			fyne.Tappable
			AccessibilityLabel() string
		}); ok && value.AccessibilityLabel() == "Deutsch" {
			german = value
		}
	})
	if german == nil {
		t.Fatal("German choice is missing")
	}
	if touch, ok := german.(mobile.Touchable); ok && fyne.CurrentDevice().IsMobile() {
		touch.TouchDown(&mobile.TouchEvent{})
		touch.TouchUp(&mobile.TouchEvent{})
	} else {
		test.Tap(german)
	}
	if i18n.Language() != "de" || bar.Selected() != 2 || stack.Objects[0] != active || input.Text != "123.456789" || !strings.Contains(windowText(shell), i18n.T("title.tempochange")) {
		t.Fatal("language change lost shell/calculator state")
	}
	preferences.Hide()
	if w.Canvas().Focused() != input {
		t.Fatal("Preferences did not restore calculator focus")
	}
	bar.Select(0)
	if !strings.Contains(windowText(shell), "Leeren") {
		t.Fatal("cached Timecode view was not translated")
	}
	bar.Select(2)
	if stack.Objects[0] != active || input.Text != "123.456789" {
		t.Fatal("cached calculator was rebuilt")
	}
}

func TestDesktopCalculatorBodyScrollsWhenZoomLimitsTheViewport(t *testing.T) {
	if fyne.CurrentDevice().IsMobile() {
		t.Skip("desktop scrolling fallback")
	}
	app := test.NewTempApp(t)
	app.Preferences().SetString("language", "en")
	ui.InitializeLanguage(app.Preferences())
	defer i18n.Use("en")
	app.Settings().SetTheme(ui.NewCustomTheme())
	w := buildMainWindow(app, nil)
	defer w.Close()
	var bar *widgets.TabBar
	walkWindowObjects(w.Content(), func(object fyne.CanvasObject) {
		if value, ok := object.(*widgets.TabBar); ok {
			bar = value
		}
	})
	bar.Select(2)
	shell := w.Content().(*fyne.Container)
	scroll := shell.Objects[1].(*container.Scroll)
	// Exercise a calculator whose natural geometry exceeds both fitted dimensions.
	extent := canvas.NewRectangle(color.Transparent)
	extent.SetMinSize(fyne.NewSize(350, 500))
	scroll.Content.(*fyne.Container).Add(extent)
	w.Resize(fyne.NewSize(200, 300))
	if shell.MinSize().Width > 200 || w.Canvas().Size().Width > 200 || w.Canvas().Size().Height > 300 {
		t.Fatal("content minima defeated the fitted viewport")
	}
	if scroll.Content.MinSize().Height <= scroll.Size().Height || scroll.Content.MinSize().Width <= scroll.Size().Width {
		t.Fatal("constrained calculator should overflow in both axes")
	}
	scroll.ScrollToOffset(fyne.NewPos(10000, 10000))
	if scroll.Offset.X <= 0 || scroll.Offset.Y <= 0 {
		t.Fatal("overflowing calculator controls must remain reachable")
	}
	footer := shell.Objects[2].(*fyne.Container).Objects[1].(*container.Scroll)
	footer.ScrollToOffset(fyne.NewPos(10000, 0))
	if footer.Content.MinSize().Width <= footer.Size().Width || footer.Offset.X <= 0 {
		t.Fatal("all desktop tabs must remain reachable in a narrow viewport")
	}
}
