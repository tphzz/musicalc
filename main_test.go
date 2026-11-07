package main

import (
	"bytes"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"image/png"
	"musicalc/internal/i18n"
	"strings"
	"testing"
)

func TestCategoryMenuActions(t *testing.T) {
	categories := []CategoryInfo{{Name: "First"}, {Name: "Second"}}
	selected, languageCalls, aboutCalls := -1, 0, 0
	menu := categoryMenu(categories, func(i int) { selected = i }, func() { languageCalls++ }, func() { aboutCalls++ })
	if len(menu.Items) != 5 || !menu.Items[2].IsSeparator || menu.Items[3].Label != "Preferences…" || menu.Items[4].Label != "About MusiCalc" {
		t.Fatal("categories must be separated from Preferences and About")
	}
	menu.Items[0].Action()
	if selected != 0 {
		t.Fatal("wrong first category")
	}
	menu.Items[1].Action()
	menu.Items[3].Action()
	menu.Items[4].Action()
	if selected != 1 || languageCalls != 1 || aboutCalls != 1 {
		t.Fatal("non-category actions changed selection")
	}
	native := categoryMenu(categories, func(int) {}, func() {}, nil)
	if len(native.Items) != 4 || !native.Items[2].IsSeparator || native.Items[3].Label != "Preferences…" {
		t.Fatal("native About must not be duplicated")
	}
}

func TestGermanCategoryMenu(t *testing.T) {
	i18n.Use("de")
	defer i18n.Use("en")
	menu := categoryMenu(nil, func(int) {}, func() {}, func() {})
	if len(menu.Items) != 2 || menu.Items[0].Label != "Einstellungen…" || menu.Items[1].Label != "Über MusiCalc" {
		t.Fatal("menu is not German")
	}
}

// TestEmbeddedAppIcon checks that the window icon is compiled in as a complete PNG, so it does not
// depend on the working directory the app starts from
func TestEmbeddedAppIcon(t *testing.T) {
	img, err := png.Decode(bytes.NewReader(appIcon))
	if err != nil {
		t.Fatalf("embedded app icon is not a PNG: %v", err)
	}
	if b := img.Bounds(); b.Dx() < 64 || b.Dy() < 64 {
		t.Errorf("embedded app icon is %v, want at least 64×64", b.Size())
	}
	// Fyne's test app ignores SetIcon, so record the call
	a := &iconRecorder{App: test.NewTempApp(t)}
	setAppIcon(a)
	if a.icon == nil || !bytes.Equal(a.icon.Content(), appIcon) {
		t.Error("the app does not use the embedded icon")
	}
}

type iconRecorder struct {
	fyne.App
	icon fyne.Resource
}

func (r *iconRecorder) SetIcon(icon fyne.Resource) { r.icon = icon }

func TestWriteVersion(t *testing.T) {
	previous := version
	t.Cleanup(func() { version = previous })
	for _, value := range []string{"0.9.0-snapshot.gdbc3c2c", "0.9.0-snapshot", "0.9.0"} {
		version = value
		var out bytes.Buffer
		if !writeVersion([]string{"--version"}, &out) || out.String() != value+"\n" {
			t.Fatalf("version output=%q, want %q", out.String(), value+"\n")
		}
	}
	for _, args := range [][]string{nil, {"other"}, {"--version", "other"}} {
		var out bytes.Buffer
		if writeVersion(args, &out) || strings.TrimSpace(out.String()) != "" {
			t.Fatalf("unexpected version handling for %v", args)
		}
	}
}

func TestCategoryPopupKeepsApprovedWidth(t *testing.T) {
	app := test.NewTempApp(t)
	window := app.NewWindow("MusiCalc")
	defer window.Close()
	window.Resize(fyne.NewSize(320, 650))
	popup := showCategoryMenu(window, []CategoryInfo{{Name: "Time & Tempo"}}, func(int) {}, func() {}, func() {})
	defer popup.Hide()
	if popup.Size().Width != 240 || popup.Position().X < 0 || popup.Position().X+popup.Size().Width > 320 {
		t.Fatalf("popup geometry: %v / %v", popup.Position(), popup.Size())
	}
}
