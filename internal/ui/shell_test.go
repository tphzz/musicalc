package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

type insetCanvas struct {
	fyne.Canvas
	pos  fyne.Position
	area fyne.Size
}

func (c *insetCanvas) InteractiveArea() (fyne.Position, fyne.Size) { return c.pos, c.area }

type insetDriver struct {
	fyne.Driver
	canvas fyne.Canvas
}

func (d insetDriver) CanvasForObject(fyne.CanvasObject) fyne.Canvas { return d.canvas }

type insetApp struct {
	fyne.App
	driver fyne.Driver
}

func (a insetApp) Driver() fyne.Driver { return a.driver }

func TestShellPhoneSafeAreaAndKeyboard(t *testing.T) {
	app := test.NewTempApp(t)
	w := app.NewWindow("phone")
	defer w.Close()
	w.Resize(fyne.NewSize(393, 852))
	c := &insetCanvas{Canvas: w.Canvas(), pos: fyne.NewPos(0, 59), area: fyne.NewSize(393, 759)}
	fyne.SetCurrentApp(insetApp{App: app, driver: insetDriver{Driver: app.Driver(), canvas: c}})
	header := canvas.NewRectangle(color.Black)
	header.SetMinSize(fyne.NewSize(10, 48))
	content := canvas.NewRectangle(color.White)
	footer := canvas.NewRectangle(color.Black)
	footer.SetMinSize(fyne.NewSize(10, 48))
	objects := []fyne.CanvasObject{header, content, footer}
	layout := &shellLayout{isMobile: true}
	check := func(wantFooter bool) {
		t.Helper()
		layout.Layout(objects, c.area)
		if footer.Visible() != wantFooter {
			t.Fatalf("footer visible=%v, want %v", footer.Visible(), wantFooter)
		}
		if header.Position().Y != 0 || content.Position().Y != 48 {
			t.Fatal("shell applied the system safe-area inset a second time")
		}
		bottom := content.Position().Y + content.Size().Height
		if wantFooter {
			if bottom != footer.Position().Y || footer.Position().Y+footer.Size().Height != c.area.Height {
				t.Fatal("footer/content escaped the interactive area")
			}
		} else if bottom != c.area.Height {
			t.Fatal("keyboard area was not fully used")
		}
	}
	check(true) // A home-indicator inset is not the keyboard.
	c.area = fyne.NewSize(393, 493)
	check(false)
	c.area = fyne.NewSize(393, 759)
	check(true)
	w.Resize(fyne.NewSize(375, 667))
	c.pos = fyne.NewPos(0, 20)
	c.area = fyne.NewSize(375, 647)
	check(true) // Compact phone without a home indicator.
}
