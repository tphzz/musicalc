package ui

import (
	_ "embed"
	"fmt"
	"musicalc/internal/i18n"
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
	title    *canvas.Text
	close    *aboutCloseButton
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
	title := canvas.NewText(i18n.T("menu.about"), th.Color(theme.ColorNameForeground, variant))
	title.TextSize = th.Size(theme.SizeNameSubHeadingText)
	title.TextStyle.Bold = true
	version := widget.NewLabel(d.version)
	version.Wrapping = fyne.TextWrapBreak
	versionText := container.NewThemeOverride(version, modalTextTheme{Theme: th, size: 16})
	logo := canvas.NewImageFromResource(fyne.NewStaticResource("tphzz-logo.png", aboutLogo))
	logo.FillMode = canvas.ImageFillContain
	logoWidth := float32(112)
	if d.isMobile {
		logoWidth = 96
	}
	logo.SetMinSize(fyne.NewSize(logoWidth, logoWidth*1183/1329))
	closeButton := &aboutCloseButton{about: d}
	closeButton.Text = i18n.T("action.close")
	closeButton.OnTapped = d.Hide
	closeButton.ExtendBaseWidget(closeButton)
	d.title, d.close = title, closeButton
	contents := container.New(&aboutPanelLayout{}, title, versionText, logo, closeButton)
	return newModalRenderer(&d.BaseWidget, title, contents)
}

// RefreshText relabels an already-created dialog without replacing its renderer.
func (d *AboutDialog) RefreshText() {
	if d.title == nil {
		return
	}
	d.title.Text = i18n.T("menu.about")
	d.title.Refresh()
	d.close.SetText(i18n.T("action.close"))
	d.Refresh()
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

type aboutPanelLayout struct{}

func (*aboutPanelLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	title, version, logo, closeButton := objects[0], objects[1], objects[2], objects[3]
	headerHeight := fyne.Max(logo.MinSize().Height, title.MinSize().Height+8+version.MinSize().Height)
	return fyne.NewSize(0, modalPadding*3+headerHeight+closeButton.MinSize().Height)
}

func (*aboutPanelLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	title, version, logo, closeButton := objects[0], objects[1], objects[2], objects[3]
	logoSize := logo.MinSize()
	textWidth := fyne.Max(0, size.Width-3*modalPadding-logoSize.Width)
	title.Move(fyne.NewPos(modalPadding, modalPadding))
	title.Resize(fyne.NewSize(textWidth, title.MinSize().Height))
	version.Move(fyne.NewPos(modalPadding, modalPadding+title.MinSize().Height+8))
	version.Resize(fyne.NewSize(textWidth, version.MinSize().Height))
	// A width change can reflow a long snapshot version.
	version.Resize(fyne.NewSize(textWidth, version.MinSize().Height))
	logo.Resize(logoSize)
	logo.Move(fyne.NewPos(size.Width-modalPadding-logoSize.Width, modalPadding))
	closeSize := closeButton.MinSize()
	closeButton.Resize(closeSize)
	closeButton.Move(fyne.NewPos(size.Width-modalPadding-closeSize.Width, size.Height-modalPadding-closeSize.Height))
}
