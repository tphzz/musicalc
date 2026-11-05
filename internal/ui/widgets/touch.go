package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/widget"
)

// On mobile, Fyne only reports a tap when the finger lifts within 300 ms; longer presses become
// secondary taps, which buttons ignore. Button reacts to mobile.Touchable's TouchUp instead, so it
// fires as soon as the finger lifts, however long the press was. (Entry already handles touches
// itself: it focuses and places the cursor on TouchDown.)
//
// Touchable widgets stop a drag from reaching a scroll container underneath them, so only use
// Button outside scrolling areas.

// Button is a widget.Button that, on mobile, fires when the finger lifts, for any press length
type Button struct {
	widget.Button
	pressed bool // the current touch started on this button
}

// NewButton creates a button with a label
func NewButton(label string, tapped func()) *Button {
	b := &Button{}
	b.Text = label
	b.OnTapped = tapped
	b.ExtendBaseWidget(b)
	return b
}

// NewButtonWithIcon creates a button with a label and an icon
func NewButtonWithIcon(label string, icon fyne.Resource, tapped func()) *Button {
	b := NewButton(label, tapped)
	b.Icon = icon
	return b
}

// Tapped handles mouse clicks on desktop; on mobile the button fires from TouchUp instead
func (b *Button) Tapped(ev *fyne.PointEvent) {
	if fyne.CurrentDevice().IsMobile() {
		return
	}
	b.Button.Tapped(ev)
}

// TouchDown starts a press; the button fires on TouchUp
func (b *Button) TouchDown(*mobile.TouchEvent) { b.pressed = true }

// TouchUp fires the button if the press started on it and the finger is still on it
func (b *Button) TouchUp(ev *mobile.TouchEvent) {
	if !b.pressed {
		return
	}
	b.pressed = false
	b.Button.Tapped(&ev.PointEvent)
}

// TouchCancel ends the press without firing, e.g. when the finger slides off the button
func (b *Button) TouchCancel(*mobile.TouchEvent) { b.pressed = false }
