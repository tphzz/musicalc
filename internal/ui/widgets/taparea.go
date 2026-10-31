package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

// TapArea makes any content tappable, like a card header that opens and closes the card.
// Like Button, it fires on mobile when the finger lifts, for any press length; only use it
// outside scrolling areas.
type TapArea struct {
	widget.BaseWidget
	Content  fyne.CanvasObject
	OnTapped func()

	pressed bool // the current touch started on this area
}

// NewTapArea creates a tappable area around content
func NewTapArea(content fyne.CanvasObject, tapped func()) *TapArea {
	a := &TapArea{Content: content, OnTapped: tapped}
	a.ExtendBaseWidget(a)
	return a
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (a *TapArea) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(a.Content)
}

// Cursor shows the pointer on desktop, so the area reads as clickable
func (a *TapArea) Cursor() desktop.Cursor { return desktop.PointerCursor }

// Tapped handles mouse clicks on desktop; on mobile the area fires from TouchUp instead
func (a *TapArea) Tapped(*fyne.PointEvent) {
	if fyne.CurrentDevice().IsMobile() {
		return
	}
	a.fire()
}

// TouchDown starts a press; the area fires on TouchUp
func (a *TapArea) TouchDown(*mobile.TouchEvent) { a.pressed = true }

// TouchUp fires the area if the press started on it
func (a *TapArea) TouchUp(*mobile.TouchEvent) {
	if !a.pressed {
		return
	}
	a.pressed = false
	a.fire()
}

// TouchCancel ends the press without firing
func (a *TapArea) TouchCancel(*mobile.TouchEvent) { a.pressed = false }

func (a *TapArea) fire() {
	if a.OnTapped != nil {
		a.OnTapped()
	}
}
