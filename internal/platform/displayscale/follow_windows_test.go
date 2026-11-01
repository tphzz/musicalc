package displayscale

import (
	"sync/atomic"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

type scaleDevice struct {
	fyne.Device
	value atomic.Int32
	reads atomic.Int32
}

func (d *scaleDevice) IsMobile() bool { return false }
func (d *scaleDevice) SystemScaleForWindow(fyne.Window) float32 {
	d.reads.Add(1)
	return float32(d.value.Load()) / 100
}

type scaleDriver struct {
	fyne.Driver
	device fyne.Device
}

func (d scaleDriver) Device() fyne.Device { return d.device }

type scaleApp struct {
	fyne.App
	driver fyne.Driver
}

func (a scaleApp) Driver() fyne.Driver { return a.driver }

func TestFollowRefreshesWithoutReplacingInputAndStops(t *testing.T) {
	base := test.NewApp()
	defer base.Quit()
	device := &scaleDevice{Device: base.Driver().Device()}
	device.value.Store(100)
	app := scaleApp{App: base, driver: scaleDriver{Driver: base.Driver(), device: device}}
	fyne.SetCurrentApp(app)
	defer fyne.SetCurrentApp(base)
	window := app.NewWindow("scale test")
	defer window.Close()
	input := widget.NewEntry()
	input.SetText("123.456789")
	window.SetContent(input)
	window.Canvas().Focus(input)
	var changes atomic.Int32
	app.Settings().AddListener(func(fyne.Settings) { changes.Add(1) })
	controller, err := NewController(app)
	if err != nil {
		t.Fatal(err)
	}
	stop := controller.Follow(window, fyne.NewSize(450, 650))
	defer stop()
	eventually(t, func() bool { return device.reads.Load() > 0 })
	initial := changes.Load()
	device.value.Store(150)
	eventually(t, func() bool { return changes.Load() > initial })
	if window.Content() != input || input.Text != "123.456789" || window.Canvas().Focused() != input {
		t.Fatal("scale refresh replaced the content, value or keyboard focus")
	}
	stop()
	stop() // window-close and app-exit cleanup may both call it.
	// Give the worker an opportunity to observe stop, then ensure a whole poll
	// interval passes without another device query.
	time.Sleep(30 * time.Millisecond)
	reads := device.reads.Load()
	time.Sleep(1100 * time.Millisecond)
	if device.reads.Load() != reads {
		t.Fatal("scale observer continued after stop")
	}
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for display-scale refresh")
}
