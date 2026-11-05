package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// iconOnlyTabWidth is the width of a tab without a label (mobile), so a single tab doesn't stretch
const iconOnlyTabWidth = 72

// TabBarItem is one tab in a TabBar. Tabs without Text show a larger icon only.
type TabBarItem struct {
	Text string
	Icon fyne.Resource
}

// TabBar is a bottom tab bar with left-aligned tabs. Unlike container.AppTabs it doesn't stretch
// tabs across the width on mobile, and its tabs switch when the finger lifts, however long the
// press was (Fyne ignores mobile taps longer than 300 ms).
type TabBar struct {
	widget.BaseWidget
	OnSelected func(index int)

	selected int
	tabs     *fyne.Container
}

// NewTabBar creates an empty tab bar; add tabs with SetItems
func NewTabBar() *TabBar {
	t := &TabBar{tabs: container.New(layout.NewCustomPaddedHBoxLayout(0))}
	t.ExtendBaseWidget(t)
	return t
}

// SetItems replaces the tabs and selects the tab at index selected, without calling OnSelected
func (t *TabBar) SetItems(items []TabBarItem, selected int) {
	t.selected = selected
	objects := make([]fyne.CanvasObject, len(items))
	for i, item := range items {
		objects[i] = newTabButton(t, i, item)
	}
	t.tabs.Objects = objects
	t.tabs.Refresh()
}

// SetLabels updates the existing tabs without changing selection or button identity.
// Labels correspond to the items most recently passed to SetItems.
func (t *TabBar) SetLabels(labels []string) {
	for i, label := range labels {
		if i >= len(t.tabs.Objects) {
			break
		}
		button := t.tabs.Objects[i].(*tabButton)
		button.item.Text = label
		button.Refresh()
	}
	t.tabs.Refresh()
}

// Selected returns the index of the selected tab
func (t *TabBar) Selected() int {
	return t.selected
}

// Select selects the tab at index and calls OnSelected; selecting the current tab does nothing
func (t *TabBar) Select(index int) {
	if index == t.selected || index < 0 || index >= len(t.tabs.Objects) {
		return
	}
	t.selected = index
	for _, o := range t.tabs.Objects {
		o.Refresh()
	}
	if t.OnSelected != nil {
		t.OnSelected(index)
	}
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (t *TabBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.tabs)
}

type tabButton struct {
	widget.BaseWidget
	bar     *TabBar
	index   int
	item    TabBarItem
	hovered bool
	pressed bool
}

func newTabButton(bar *TabBar, index int, item TabBarItem) *tabButton {
	b := &tabButton{bar: bar, index: index, item: item}
	b.ExtendBaseWidget(b)
	return b
}

// Tapped selects the tab (mouse click, or a mobile tap shorter than 300 ms)
func (b *tabButton) Tapped(*fyne.PointEvent) { b.bar.Select(b.index) }

// TouchDown shows the pressed state
func (b *tabButton) TouchDown(*mobile.TouchEvent) {
	b.pressed = true
	b.Refresh()
}

// TouchUp selects the tab if the press started on it
func (b *tabButton) TouchUp(*mobile.TouchEvent) {
	if !b.pressed {
		return
	}
	b.pressed = false
	b.Refresh()
	b.bar.Select(b.index)
}

// TouchCancel ends the press without selecting, e.g. when the finger slides off
func (b *tabButton) TouchCancel(*mobile.TouchEvent) {
	b.pressed = false
	b.Refresh()
}

// MouseIn shows the hover state on desktop
func (b *tabButton) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

// MouseMoved is required by desktop.Hoverable
func (b *tabButton) MouseMoved(*desktop.MouseEvent) {}

// MouseOut clears the hover state
func (b *tabButton) MouseOut() {
	b.hovered = false
	b.Refresh()
}

func (b *tabButton) CreateRenderer() fyne.WidgetRenderer {
	r := &tabButtonRenderer{
		button:     b,
		background: canvas.NewRectangle(color.Transparent),
		indicator:  canvas.NewRectangle(color.Transparent),
		icon:       canvas.NewImageFromResource(b.item.Icon),
		label:      canvas.NewText(b.item.Text, color.Transparent),
	}
	r.icon.FillMode = canvas.ImageFillContain
	r.label.TextStyle = fyne.TextStyle{Bold: true}
	r.applyTheme()
	return r
}

type tabButtonRenderer struct {
	button     *tabButton
	background *canvas.Rectangle
	indicator  *canvas.Rectangle
	icon       *canvas.Image
	label      *canvas.Text
}

func (r *tabButtonRenderer) iconOnly() bool {
	return r.button.item.Text == ""
}

// iconSize follows AppTabs: 1.5x the inline icon size when the tab has no label
func (r *tabButtonRenderer) iconSize() float32 {
	size := r.button.Theme().Size(theme.SizeNameInlineIcon)
	if r.iconOnly() {
		return size * 1.5
	}
	return size
}

func (r *tabButtonRenderer) indicatorHeight() float32 {
	return 2 * r.button.Theme().Size(theme.SizeNameSeparatorThickness)
}

func (r *tabButtonRenderer) Layout(size fyne.Size) {
	th := r.button.Theme()
	r.indicator.Move(fyne.NewPos(0, 0))
	r.indicator.Resize(fyne.NewSize(size.Width, r.indicatorHeight()))

	iconSize := r.iconSize()
	inner := th.Size(theme.SizeNameInnerPadding)
	if r.iconOnly() {
		iconPos := fyne.NewPos((size.Width-iconSize)/2, (size.Height-iconSize)/2)
		r.icon.Move(iconPos)
		r.icon.Resize(fyne.NewSquareSize(iconSize))
		r.label.Hide()

		// Press feedback is a pill behind the icon, like Android's navigation bar, rather than a
		// box over the whole tab
		pill := fyne.NewSize(fyne.Min(size.Width, iconSize+4*inner), iconSize+inner)
		r.background.CornerRadius = pill.Height / 2
		r.background.Move(fyne.NewPos((size.Width-pill.Width)/2, iconPos.Y-inner/2))
		r.background.Resize(pill)
		return
	}

	r.background.CornerRadius = th.Size(theme.SizeNameInputRadius)
	r.background.Move(fyne.NewPos(0, r.indicatorHeight()))
	r.background.Resize(fyne.NewSize(size.Width, size.Height-r.indicatorHeight()))
	r.icon.Move(fyne.NewPos(inner, (size.Height-iconSize)/2))
	r.icon.Resize(fyne.NewSquareSize(iconSize))
	labelSize := r.label.MinSize()
	r.label.Move(fyne.NewPos(inner+iconSize+th.Size(theme.SizeNamePadding), (size.Height-labelSize.Height)/2))
	r.label.Resize(labelSize)
	r.label.Show()
}

func (r *tabButtonRenderer) MinSize() fyne.Size {
	th := r.button.Theme()
	inner := th.Size(theme.SizeNameInnerPadding)
	iconSize := r.iconSize()
	if r.iconOnly() {
		return fyne.NewSize(iconOnlyTabWidth, iconSize+2*inner+r.indicatorHeight())
	}
	labelSize := r.label.MinSize()
	return fyne.NewSize(
		inner+iconSize+th.Size(theme.SizeNamePadding)+labelSize.Width+inner,
		fyne.Max(iconSize, labelSize.Height)+2*inner+r.indicatorHeight(),
	)
}

func (r *tabButtonRenderer) Refresh() {
	r.label.Text = r.button.item.Text
	r.applyTheme()
	r.Layout(r.button.Size())
	canvas.Refresh(r.button)
}

func (r *tabButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.indicator, r.icon, r.label}
}

func (r *tabButtonRenderer) Destroy() {}

func (r *tabButtonRenderer) applyTheme() {
	th := r.button.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	selected := r.button.index == r.button.bar.selected

	switch {
	case r.button.pressed:
		// A light primary tint; the theme's pressed color is a strong grey that washes out the icon
		r.background.FillColor = th.Color(theme.ColorNameFocus, v)
	case r.button.hovered:
		r.background.FillColor = th.Color(theme.ColorNameHover, v)
	default:
		r.background.FillColor = color.Transparent
	}
	if selected {
		r.indicator.FillColor = th.Color(theme.ColorNamePrimary, v)
		r.label.Color = th.Color(theme.ColorNamePrimary, v)
	} else {
		r.indicator.FillColor = color.Transparent
		r.label.Color = th.Color(theme.ColorNameForeground, v)
	}
	r.label.TextSize = th.Size(theme.SizeNameText)
	r.background.Refresh()
	r.indicator.Refresh()
	r.label.Refresh()
}
