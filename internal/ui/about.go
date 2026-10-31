package ui

import (
	_ "embed"
	"fmt"
	"strings"

	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed assets/tphzz-logo.png
var aboutLogo []byte

// DisplayVersion uses the packaged version and build on mobile, and the full
// linker version (including snapshot information) on desktop.
func DisplayVersion(linked string, metadata fyne.AppMetadata, isMobile bool) string {
	if isMobile && strings.TrimSpace(metadata.Version) != "" {
		if metadata.Build > 0 {
			return fmt.Sprintf("%s (%d)", strings.TrimSpace(metadata.Version), metadata.Build)
		}
		return strings.TrimSpace(metadata.Version)
	}
	if v := strings.TrimSpace(linked); v != "" {
		return v
	}
	return "dev"
}

// AboutDialog is a modal overlay; closing it leaves the calculator and help state intact.
type AboutDialog struct {
	widget.BaseWidget
	canvas   fyne.Canvas
	version  string
	isMobile bool
}

func NewAboutDialog(version string, parent fyne.Window) *AboutDialog {
	d := &AboutDialog{canvas: parent.Canvas(), version: version, isMobile: fyne.CurrentDevice().IsMobile()}
	d.ExtendBaseWidget(d)
	d.BaseWidget.Hide()
	return d
}

func (d *AboutDialog) Show() {
	if d.Visible() {
		return
	}
	d.BaseWidget.Show()
	d.canvas.Overlays().Add(d)
}

func (d *AboutDialog) Hide() {
	if d.Visible() {
		d.canvas.Overlays().Remove(d)
	}
	d.BaseWidget.Hide()
}

// HandleKey consumes window shortcuts while the modal is open.
func (d *AboutDialog) HandleKey(ev *fyne.KeyEvent) bool {
	if !d.Visible() {
		return false
	}
	if ev.Name == fyne.KeyEscape || ev.Name == mobile.KeyBack {
		d.Hide()
	}
	return true
}

// Tapped catches taps on the backdrop without dismissing the modal.
func (d *AboutDialog) Tapped(*fyne.PointEvent) {}

func (d *AboutDialog) CreateRenderer() fyne.WidgetRenderer {
	th := d.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	title := canvas.NewText("About MusiCalc", th.Color(theme.ColorNameForeground, variant))
	title.TextSize = th.Size(theme.SizeNameSubHeadingText)
	title.TextStyle.Bold = true
	version := widget.NewLabel(d.version)
	version.Wrapping = fyne.TextWrapBreak
	versionText := container.NewThemeOverride(version, aboutTextTheme{Theme: th})
	logo := canvas.NewImageFromResource(fyne.NewStaticResource("tphzz-logo.png", aboutLogo))
	logo.FillMode = canvas.ImageFillContain
	logoWidth := float32(112)
	if d.isMobile {
		logoWidth = 96
	}
	logo.SetMinSize(fyne.NewSize(logoWidth, logoWidth*1183/1329))
	closeButton := &aboutCloseButton{about: d}
	closeButton.Text = "Close"
	closeButton.OnTapped = d.Hide
	closeButton.ExtendBaseWidget(closeButton)
	background := canvas.NewRectangle(th.Color(theme.ColorNameOverlayBackground, variant))
	background.CornerRadius = th.Size(theme.SizeNameDialogRadius)
	contents := container.New(&aboutPanelLayout{}, title, versionText, logo, closeButton)
	panel := container.NewStack(background, contents)
	backdrop := canvas.NewRectangle(th.Color(theme.ColorNameShadow, variant))
	root := container.New(&aboutOverlayLayout{}, backdrop, panel)
	return &aboutRenderer{
		WidgetRenderer: widget.NewSimpleRenderer(root),
		dialog:         d, background: background, backdrop: backdrop, title: title,
	}
}

// Fyne delivers keys directly to the focused control, bypassing the canvas handler.
type aboutCloseButton struct {
	widgets.Button
	about *AboutDialog
}

func (b *aboutCloseButton) TypedKey(ev *fyne.KeyEvent) {
	b.about.HandleKey(ev)
	b.Button.TypedKey(ev)
}

// The About version has no label inset and uses the same size on both platforms.
type aboutTextTheme struct{ fyne.Theme }

func (t aboutTextTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNameInnerPadding:
		return 0
	}
	return t.Theme.Size(name)
}

type aboutRenderer struct {
	fyne.WidgetRenderer
	dialog               *AboutDialog
	background, backdrop *canvas.Rectangle
	title                *canvas.Text
}

func (r *aboutRenderer) Refresh() {
	th := r.dialog.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	r.background.FillColor = th.Color(theme.ColorNameOverlayBackground, variant)
	r.background.CornerRadius = th.Size(theme.SizeNameDialogRadius)
	r.backdrop.FillColor = th.Color(theme.ColorNameShadow, variant)
	r.title.Color = th.Color(theme.ColorNameForeground, variant)
	r.WidgetRenderer.Refresh()
}

const aboutPadding float32 = 16

type aboutOverlayLayout struct{}

func (*aboutOverlayLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.Size{} }

func (*aboutOverlayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	objects[0].Resize(size)
	panel := objects[1]
	width := fyne.Min(418, fyne.Max(0, size.Width-2*aboutPadding))
	panel.Resize(fyne.NewSize(width, panel.MinSize().Height))
	height := fyne.Min(size.Height, panel.MinSize().Height)
	panel.Resize(fyne.NewSize(width, height))
	panel.Move(fyne.NewPos((size.Width-width)/2, (size.Height-height)/2))
}

type aboutPanelLayout struct{}

func (*aboutPanelLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	title, version, logo, closeButton := objects[0], objects[1], objects[2], objects[3]
	headerHeight := fyne.Max(logo.MinSize().Height, title.MinSize().Height+8+version.MinSize().Height)
	return fyne.NewSize(0, aboutPadding*3+headerHeight+closeButton.MinSize().Height)
}

func (*aboutPanelLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	title, version, logo, closeButton := objects[0], objects[1], objects[2], objects[3]
	logoSize := logo.MinSize()
	textWidth := fyne.Max(0, size.Width-3*aboutPadding-logoSize.Width)
	title.Move(fyne.NewPos(aboutPadding, aboutPadding))
	title.Resize(fyne.NewSize(textWidth, title.MinSize().Height))
	version.Move(fyne.NewPos(aboutPadding, aboutPadding+title.MinSize().Height+8))
	version.Resize(fyne.NewSize(textWidth, version.MinSize().Height))
	// A width change can reflow a long snapshot version.
	version.Resize(fyne.NewSize(textWidth, version.MinSize().Height))
	logo.Resize(logoSize)
	logo.Move(fyne.NewPos(size.Width-aboutPadding-logoSize.Width, aboutPadding))
	closeSize := closeButton.MinSize()
	closeButton.Resize(closeSize)
	closeButton.Move(fyne.NewPos(size.Width-aboutPadding-closeSize.Width, size.Height-aboutPadding-closeSize.Height))
}
