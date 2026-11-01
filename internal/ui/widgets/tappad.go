package widgets

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TapPadStyle selects how a TapPad is drawn
type TapPadStyle int

const (
	// TapPadLarge fills its space: beat dot, title and hint (the Tap Tempo tab)
	TapPadLarge TapPadStyle = iota
	// TapPadCompact is a full-width pad with a large "TAP" and a status line (mobile)
	TapPadCompact
	// TapPadInline is a single line: bold "Tap" and a status (desktop)
	TapPadInline
)

// beatDotSize is the diameter of the large pad's beat dot
const beatDotSize = 56

// TapPad is a surface to tap along with the beat. It reacts when the finger or mouse button goes
// down, not on release like buttons, so taps land on the beat; and it has no Tapped handler, so a
// tap can't be counted twice.
type TapPad struct {
	widget.BaseWidget
	OnTap func()

	style    TapPadStyle
	title    string
	status   string
	active   bool // a measurement is running
	flashing bool // briefly highlighting the beat dot after a tap
}

// NewTapPad creates a tap pad; onTap runs on every tap
func NewTapPad(style TapPadStyle, title, status string, onTap func()) *TapPad {
	p := &TapPad{OnTap: onTap, style: style, title: title, status: status}
	p.ExtendBaseWidget(p)
	return p
}

// Active reports whether a measurement is running.
func (p *TapPad) Active() bool { return p.active }

// SetActive switches between the idle and the tapping look
func (p *TapPad) SetActive(active bool) {
	if p.active == active {
		return
	}
	p.active = active
	p.Refresh()
}

// SetTitle changes the title (the large pad's main text)
func (p *TapPad) SetTitle(title string) {
	p.title = title
	p.Refresh()
}

// SetStatus changes the status or hint line
func (p *TapPad) SetStatus(status string) {
	p.status = status
	p.Refresh()
}

// Flash highlights the beat dot for a moment to confirm a tap
func (p *TapPad) Flash() {
	p.flashing = true
	p.Refresh()
	time.AfterFunc(120*time.Millisecond, func() {
		fyne.Do(func() {
			p.flashing = false
			p.Refresh()
		})
	})
}

func (p *TapPad) tap() {
	if p.OnTap != nil {
		p.OnTap()
	}
}

// TouchDown counts a tap on mobile as soon as the finger touches
func (p *TapPad) TouchDown(*mobile.TouchEvent) { p.tap() }

// TouchUp is required by mobile.Touchable
func (p *TapPad) TouchUp(*mobile.TouchEvent) {}

// TouchCancel is required by mobile.Touchable
func (p *TapPad) TouchCancel(*mobile.TouchEvent) {}

// MouseDown counts a tap on desktop as soon as the primary button goes down
func (p *TapPad) MouseDown(ev *desktop.MouseEvent) {
	if ev.Button == desktop.MouseButtonPrimary {
		p.tap()
	}
}

// MouseUp is required by desktop.Mouseable
func (p *TapPad) MouseUp(*desktop.MouseEvent) {}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (p *TapPad) CreateRenderer() fyne.WidgetRenderer {
	r := &tapPadRenderer{
		pad:        p,
		background: canvas.NewRectangle(color.Transparent),
		dot:        canvas.NewCircle(color.Transparent),
	}
	titleSize, statusSize := theme.SizeNameText, theme.SizeNameText
	switch p.style {
	case TapPadLarge:
		titleSize = theme.SizeNameSubHeadingText
	case TapPadCompact:
		titleSize = theme.SizeNameHeadingText
	}
	r.title = NewStyledText(p.title, theme.ColorNameForeground, titleSize, fyne.TextStyle{Bold: true})
	r.status = NewStyledText(p.status, theme.ColorNamePlaceHolder, statusSize, fyne.TextStyle{})

	switch p.style {
	case TapPadLarge:
		dot := container.NewGridWrap(fyne.NewSquareSize(beatDotSize), r.dot)
		r.title.Wrapping = fyne.TextWrapWord
		r.status.Wrapping = fyne.TextWrapWord
		r.title.Segments[0].(*widget.TextSegment).Style.Alignment = fyne.TextAlignCenter
		r.status.Segments[0].(*widget.TextSegment).Style.Alignment = fyne.TextAlignCenter
		r.content = container.NewVBox(layout.NewSpacer(), container.NewCenter(dot), r.title, r.status, layout.NewSpacer())
	case TapPadCompact:
		r.content = container.NewVBox(container.NewCenter(r.title), container.NewCenter(r.status))
	default:
		// Overlap the texts' inner padding so they sit a padding apart
		gap := theme.Padding() - 2*theme.InnerPadding()
		r.content = container.New(layout.NewCustomPaddedHBoxLayout(gap), r.title, r.status)
	}
	r.applyState()
	return r
}

type tapPadRenderer struct {
	pad           *TapPad
	background    *canvas.Rectangle
	dot           *canvas.Circle
	title, status *widget.RichText
	content       fyne.CanvasObject
}

// inset keeps the visible text inset at the card padding (labels add their own inner padding)
func (r *tapPadRenderer) inset() float32 {
	th := r.pad.Theme()
	p := th.Size(SizeNameCardPadding)
	if p == 0 {
		p = 2 * th.Size(theme.SizeNameInnerPadding)
	}
	return fyne.Max(p-th.Size(theme.SizeNameInnerPadding), 0)
}

func (r *tapPadRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	in := r.inset()
	r.content.Move(fyne.NewPos(in, in))
	r.content.Resize(size.Subtract(fyne.NewSquareSize(2 * in)))
}

func (r *tapPadRenderer) MinSize() fyne.Size {
	return r.content.MinSize().Add(fyne.NewSquareSize(2 * r.inset()))
}

func (r *tapPadRenderer) Refresh() {
	r.applyState()
	r.Layout(r.pad.Size())
	canvas.Refresh(r.pad)
}

func (r *tapPadRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.content}
}

func (r *tapPadRenderer) Destroy() {}

func (r *tapPadRenderer) applyState() {
	p := r.pad
	th := p.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	r.background.CornerRadius = th.Size(theme.SizeNameCardRadius)
	r.background.Shadow = canvas.Shadow{
		Color:      th.Color(theme.ColorNameShadow, v),
		BlurRadius: 8,
		Offset:     fyne.NewPos(0, 2),
		Variant:    canvas.DropShadow,
	}
	titleColor, statusColor := theme.ColorNameForeground, theme.ColorNamePlaceHolder
	if p.active {
		r.background.FillColor = th.Color(theme.ColorNameSelection, v)
		r.background.StrokeColor = th.Color(theme.ColorNamePrimary, v)
		r.background.StrokeWidth = 2
		statusColor = theme.ColorNameForeground
		if p.style != TapPadLarge {
			titleColor = theme.ColorNamePrimary
		}
	} else {
		r.background.FillColor = th.Color(theme.ColorNameButton, v)
		r.background.StrokeColor = th.Color(theme.ColorNameInputBorder, v)
		r.background.StrokeWidth = 1
		if p.style == TapPadLarge {
			titleColor = theme.ColorNamePlaceHolder
		}
	}
	r.background.Refresh()

	switch {
	case p.flashing:
		r.dot.FillColor = th.Color(theme.ColorNameForeground, v)
	case p.active:
		r.dot.FillColor = th.Color(theme.ColorNamePrimary, v)
	default:
		r.dot.FillColor = th.Color(theme.ColorNameInputBorder, v)
	}
	r.dot.Refresh()

	SetStyledText(r.title, p.title, titleColor)
	SetStyledText(r.status, p.status, statusColor)
}
