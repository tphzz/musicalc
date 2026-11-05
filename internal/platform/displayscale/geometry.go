package displayscale

import (
	"fyne.io/fyne/v2"
	"math"
)

func fitToAvailable(requested fyne.Size, scale float32, available fyne.Size) fyne.Size {
	if scale <= 0 || math.IsNaN(float64(scale)) || math.IsInf(float64(scale), 0) || available.Width <= 0 || available.Height <= 0 {
		return requested
	}
	fit := func(value, pixels float32) float32 {
		if math.Ceil(float64(value)*float64(scale)) <= float64(pixels) {
			return value
		}
		return fyne.Max(0, (pixels-1)/scale) // Leave a pixel for the driver's ceiling.
	}
	return fyne.NewSize(fit(requested.Width, available.Width), fit(requested.Height, available.Height))
}
