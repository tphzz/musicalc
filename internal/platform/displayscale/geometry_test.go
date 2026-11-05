package displayscale

import (
	"fyne.io/fyne/v2"
	"math"
	"testing"
)

func TestZoomFitsTheMonitorWithoutChangingTheRequestedBaseline(t *testing.T) {
	requested := fyne.NewSize(450, 650)
	available := fyne.NewSize(1000, 720)
	if got := fitToAvailable(requested, 1, available); got != requested {
		t.Fatal("normal size should be unchanged")
	}
	for _, scale := range []float32{0.8, 1.2, 1.3, 2, 4} {
		got := fitToAvailable(requested, scale, available)
		if got.Width > requested.Width || got.Height > requested.Height || math.Ceil(float64(got.Width)*float64(scale)) > 1000 || math.Ceil(float64(got.Height)*float64(scale)) > 720 {
			t.Fatalf("scale %v overflows the monitor: %v", scale, got)
		}
	}
	if got := fitToAvailable(requested, 1, available); got != requested {
		t.Fatal("returning to normal size must restore the requested baseline")
	}
}
