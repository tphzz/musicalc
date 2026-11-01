// Package displayscale keeps desktop sizing tied to the OS display scale.
package displayscale

import (
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

// Configure selects OS scaling unless the launcher explicitly set FYNE_SCALE.
// This avoids inheriting an unrelated global Fyne Settings scale multiplier.
func Configure() error {
	if os.Getenv("FYNE_SCALE") == "" {
		return os.Setenv("FYNE_SCALE", "auto")
	}
	return nil
}

// Follow binds the requested logical size and refreshes Windows scale changes
// that Fyne 2.8 does not always observe when the physical monitor DPI is unchanged. It keeps the existing widget tree,
// logical window size and calculator state. Call the returned stop when closed.
func (c *Controller) Follow(window fyne.Window, logicalSize fyne.Size) func() {
	c.window, c.logicalSize = window, logicalSize
	if runtime.GOOS != "windows" || fyne.CurrentDevice().IsMobile() {
		return func() {}
	}
	done := make(chan struct{})
	stop := sync.OnceFunc(func() { close(done) })
	var scale observer
	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				fyne.Do(func() {
					select {
					case <-done:
						return
					default:
					}
					if scale.changed(fyne.CurrentDevice().SystemScaleForWindow(window)) {
						// The public settings notification makes the driver reload window scale
						// and font/texture caches; rebuilding content would discard user state.
						c.Refresh()
					}
				})
			}
		}
	}()
	return stop
}

type observer struct{ last float32 }

func (o *observer) changed(scale float32) bool {
	if scale <= 0 || math.IsNaN(float64(scale)) || math.IsInf(float64(scale), 0) {
		return false
	}
	previous := o.last
	o.last = scale
	return previous != 0 && previous != scale
}
