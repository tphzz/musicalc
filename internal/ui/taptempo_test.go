package ui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"
)

// TestTapDelayFollowsValidity checks that Delay is unavailable after the first tap of a restarted
// measurement, which has no tempo to send yet. It waits for a real pause, as the screen taps with
// the current time.
func TestTapDelayFollowsValidity(t *testing.T) {
	root := openEnglishTab(t, func() *View { return NewTapTempoTab(func() {}) })
	var pad *widgets.TapPad
	visitUI(root, func(object fyne.CanvasObject) {
		if p, ok := object.(*widgets.TapPad); ok {
			pad = p
		}
	})
	if pad == nil {
		t.Fatal("no tap pad")
	}
	delay := buttonNamed(t, root, "Delay")
	if !delay.Disabled() {
		t.Fatal("Delay is available before any tap")
	}
	pad.OnTap()
	time.Sleep(300 * time.Millisecond)
	pad.OnTap()
	if delay.Disabled() {
		t.Fatal("Delay is unavailable after a valid measurement")
	}
	time.Sleep(logic.TapResetAfter + 200*time.Millisecond)
	pad.OnTap()
	if !delay.Disabled() {
		t.Error("Delay stays available after the first tap of a new measurement")
	}
}
