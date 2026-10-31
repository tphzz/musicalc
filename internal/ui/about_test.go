package ui

import (
	"bytes"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"image/png"
	"math"
	"testing"

	"fyne.io/fyne/v2"
)

func TestDisplayVersion(t *testing.T) {
	tests := []struct {
		name     string
		linked   string
		metadata fyne.AppMetadata
		mobile   bool
		want     string
	}{
		{"desktop release", " 0.9.0 ", fyne.AppMetadata{}, false, "0.9.0"},
		{"desktop snapshot", "0.9.1-dev.4+abc", fyne.AppMetadata{Version: "0.9.0", Build: 39}, false, "0.9.1-dev.4+abc"},
		{"desktop default", " ", fyne.AppMetadata{}, false, "dev"},
		{"phone package", "dev", fyne.AppMetadata{Version: "0.9.0", Build: 42}, true, "0.9.0 (42)"},
		{"phone without build", "dev", fyne.AppMetadata{Version: "0.9.0"}, true, "0.9.0"},
		{"phone fallback", " 0.9.1 ", fyne.AppMetadata{}, true, "0.9.1"},
		{"phone default", "", fyne.AppMetadata{}, true, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DisplayVersion(tt.linked, tt.metadata, tt.mobile); got != tt.want {
				t.Fatalf("DisplayVersion()=%q, want %q", got, tt.want)
			}
		})
	}
}
func TestAboutModalLifecycle(t *testing.T) {
	app := test.NewTempApp(t)
	app.Settings().SetTheme(customTheme{})
	window := app.NewWindow("MusiCalc")
	defer window.Close()
	entry := widget.NewEntry()
	entry.SetText("00:00:10:00")
	window.SetContent(entry)
	window.Resize(fyne.NewSize(450, 650))
	window.Canvas().Focus(entry)
	about := NewAboutDialog("0.9.0", window)
	if about.Visible() || len(window.Canvas().Overlays().List()) != 0 {
		t.Fatal("About must start closed")
	}
	for _, key := range []fyne.KeyName{fyne.KeyEscape, mobile.KeyBack} {
		about.Show()
		about.Show()
		if len(window.Canvas().Overlays().List()) != 1 {
			t.Fatal("duplicate About overlay")
		}
		if window.Canvas().Focused() == entry {
			t.Fatal("underlying entry retained modal focus")
		}
		if !about.HandleKey(&fyne.KeyEvent{Name: fyne.KeyF1}) || !about.Visible() {
			t.Fatal("About must consume underlying shortcuts")
		}
		about.Tapped(&fyne.PointEvent{})
		if !about.Visible() {
			t.Fatal("background tap dismissed About")
		}
		if !about.HandleKey(&fyne.KeyEvent{Name: key}) || about.Visible() {
			t.Fatal("dismiss key did not close About")
		}
		if len(window.Canvas().Overlays().List()) != 0 || window.Canvas().Focused() != entry || entry.Text != "00:00:10:00" {
			t.Fatal("closing About failed to restore calculator state/focus")
		}
	}
	about.Show()
	root := test.WidgetRenderer(about).Objects()[0].(*fyne.Container)
	panel := root.Objects[1].(*fyne.Container)
	contents := panel.Objects[1].(*fyne.Container)
	button := contents.Objects[3].(*aboutCloseButton)
	button.TouchDown(&mobile.TouchEvent{})
	button.TouchUp(&mobile.TouchEvent{})
	if about.Visible() || len(window.Canvas().Overlays().List()) != 0 {
		t.Fatal("Close did not remove the modal")
	}
	about.Hide()
	if about.HandleKey(&fyne.KeyEvent{Name: fyne.KeyF1}) {
		t.Fatal("closed About consumed a shortcut")
	}
}

func TestAboutLayout(t *testing.T) {
	for _, tt := range []struct {
		name          string
		width, height float32
		mobile        bool
		version       string
	}{
		{"desktop", 450, 650, false, "0.9.0"},
		{"iphone", 393, 759, true, "0.9.0 (42)"},
		{"compact phone", 375, 647, true, "0.9.0 (42)"},
		{"android", 360, 728, true, "0.9.0 (42)"},
		{"snapshot", 375, 647, true, "0.9.1-dev.12345678901234567890+abcdef0123456789"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := test.NewTempApp(t)
			app.Settings().SetTheme(customTheme{isMobile: tt.mobile})
			window := app.NewWindow("MusiCalc")
			defer window.Close()
			window.SetPadded(false)
			window.SetContent(canvas.NewRectangle(color.Transparent))
			window.Resize(fyne.NewSize(tt.width, tt.height))
			// The software driver has desktop device flags; select the mobile
			// renderer explicitly to test phone geometry without a native device.
			about := &AboutDialog{canvas: window.Canvas(), version: tt.version, isMobile: tt.mobile}
			about.ExtendBaseWidget(about)
			about.BaseWidget.Hide()
			about.Show()
			renderer := test.WidgetRenderer(about)
			root := renderer.Objects()[0].(*fyne.Container)
			panel := root.Objects[1].(*fyne.Container)
			contents := panel.Objects[1].(*fyne.Container)
			title, version, logo, closeButton := contents.Objects[0], contents.Objects[1], contents.Objects[2], contents.Objects[3]
			near := func(got, want float32) {
				t.Helper()
				if math.IsNaN(float64(got)) || math.IsInf(float64(got), 0) || math.Abs(float64(got-want)) > 0.01 {
					t.Fatalf("geometry %g, want %g", got, want)
				}
			}
			near(panel.Size().Width, tt.width-32)
			near(panel.Position().X, 16)
			near(panel.Position().Y, (tt.height-panel.Size().Height)/2)
			near(title.Position().X, 16)
			near(title.Position().Y, 16)
			near(logo.Position().X+logo.Size().Width, panel.Size().Width-16)
			near(logo.Position().Y, 16)
			logoWidth := float32(112)
			if tt.mobile {
				logoWidth = 96
			}
			near(logo.Size().Width, logoWidth)
			near(logo.Size().Height, logoWidth*1183/1329)
			near(closeButton.Position().X+closeButton.Size().Width, panel.Size().Width-16)
			near(closeButton.Position().Y+closeButton.Size().Height, panel.Size().Height-16)
			textBottom := version.Position().Y + version.MinSize().Height
			if textBottom+16 > closeButton.Position().Y+0.01 {
				t.Fatal("version overlaps Close")
			}
			imageObject := logo.(*canvas.Image)
			if imageObject.FillMode != canvas.ImageFillContain || len(imageObject.Resource.Content()) == 0 {
				t.Fatal("missing or distorted logo")
			}
			for _, themeVariant := range []fyne.ThemeVariant{theme.VariantDark, theme.VariantLight} {
				app.Settings().SetTheme(themeWithVariant{Theme: customTheme{isMobile: tt.mobile}, variant: themeVariant})
				about.Refresh()
				if got, want := panel.Objects[0].(*canvas.Rectangle).FillColor, app.Settings().Theme().Color(theme.ColorNameOverlayBackground, themeVariant); got != want {
					t.Fatalf("panel theme color=%v, want %v", got, want)
				}
				if window.Canvas().Capture().Bounds().Dx() == 0 {
					t.Fatal("empty render")
				}
			}
		})
	}
}

type themeWithVariant struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

func (t themeWithVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(name, t.variant)
}

func TestAboutLogoHasTransparency(t *testing.T) {
	image, err := png.Decode(bytes.NewReader(aboutLogo))
	if err != nil {
		t.Fatal(err)
	}
	bounds := image.Bounds()
	if bounds.Dx() != 1329 || bounds.Dy() != 1183 {
		t.Fatalf("unexpected logo size %v", bounds)
	}
	_, _, _, alpha := image.At(0, 0).RGBA()
	if alpha != 0 {
		t.Fatal("logo retains an opaque background")
	}
	foundOpaque := false
	for y := bounds.Min.Y; y < bounds.Max.Y && !foundOpaque; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha = image.At(x, y).RGBA()
			if alpha == 65535 {
				foundOpaque = true
				break
			}
		}
	}
	if !foundOpaque {
		t.Fatal("logo artwork is empty")
	}
}

func TestAboutDismissWithCloseFocused(t *testing.T) {
	app := test.NewTempApp(t)
	w := app.NewWindow("MusiCalc")
	defer w.Close()
	entry := widget.NewEntry()
	entry.SetText("retained input")
	w.SetContent(entry)
	w.Resize(fyne.NewSize(450, 650))
	w.Canvas().Focus(entry)
	about := NewAboutDialog("0.9.0", w)
	for _, key := range []fyne.KeyName{fyne.KeyEscape, mobile.KeyBack, fyne.KeySpace} {
		about.Show()
		w.Canvas().FocusNext()
		focused := w.Canvas().Focused()
		if focused == nil || focused == entry {
			t.Fatal("Close did not receive focus")
		}
		focused.TypedKey(&fyne.KeyEvent{Name: fyne.KeyF1})
		if !about.Visible() {
			t.Fatal("F1 dismissed About")
		}
		focused.TypedKey(&fyne.KeyEvent{Name: key})
		if about.Visible() || len(w.Canvas().Overlays().List()) != 0 {
			t.Fatalf("%s did not dismiss About with Close focused", key)
		}
		if w.Canvas().Focused() != entry || entry.Text != "retained input" {
			t.Fatal("dismissal failed to restore underlying input/focus")
		}
	}
}
