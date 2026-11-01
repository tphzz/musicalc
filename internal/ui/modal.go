package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const modalPadding float32 = 16

func newModalRenderer(base *widget.BaseWidget, title *canvas.Text, content fyne.CanvasObject) fyne.WidgetRenderer {
	th := base.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	background := canvas.NewRectangle(th.Color(theme.ColorNameOverlayBackground, variant))
	background.CornerRadius = th.Size(theme.SizeNameDialogRadius)
	backdrop := canvas.NewRectangle(th.Color(theme.ColorNameShadow, variant))
	root := container.New(&modalOverlayLayout{}, backdrop, container.NewStack(background, content))
	return &modalRenderer{WidgetRenderer: widget.NewSimpleRenderer(root), base: base, background: background, backdrop: backdrop, title: title}
}

type modalRenderer struct {
	fyne.WidgetRenderer
	base                 *widget.BaseWidget
	background, backdrop *canvas.Rectangle
	title                *canvas.Text
}

func (r *modalRenderer) Refresh() {
	th := r.base.Theme()
	variant := fyne.CurrentApp().Settings().ThemeVariant()
	r.background.FillColor = th.Color(theme.ColorNameOverlayBackground, variant)
	r.background.CornerRadius = th.Size(theme.SizeNameDialogRadius)
	r.backdrop.FillColor = th.Color(theme.ColorNameShadow, variant)
	r.title.Color = th.Color(theme.ColorNameForeground, variant)
	r.WidgetRenderer.Refresh()
}

type modalOverlayLayout struct{}

func (*modalOverlayLayout) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.Size{} }
func (*modalOverlayLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	objects[0].Resize(size)
	panel := objects[1]
	width := fyne.Min(418, fyne.Max(0, size.Width-2*modalPadding))
	panel.Resize(fyne.NewSize(width, panel.MinSize().Height))
	height := fyne.Min(fyne.Max(0, size.Height-2*modalPadding), panel.MinSize().Height)
	panel.Resize(fyne.NewSize(width, height))
	panel.Move(fyne.NewPos((size.Width-width)/2, (size.Height-height)/2))
}

// modalTextTheme keeps explanatory text inset-free at its specified design size.
type modalTextTheme struct {
	fyne.Theme
	size float32
}

func (t modalTextTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return t.size
	case theme.SizeNameInnerPadding:
		return 0
	}
	return t.Theme.Size(name)
}
