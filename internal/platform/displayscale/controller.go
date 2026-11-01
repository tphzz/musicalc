package displayscale

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
)

const preferencePercent = "displayScalePercent"

// Controller owns the app's desktop zoom; Fyne combines it with the OS scale.
// An explicit launcher override takes precedence over the saved preference.
type Controller struct {
	app         fyne.App
	percent     int
	override    string
	window      fyne.Window
	logicalSize fyne.Size
}

func Percents() []int { return []int{80, 100, 115, 130, 200} }

func supportedPercent(percent int) bool {
	for _, candidate := range Percents() {
		if candidate == percent {
			return true
		}
	}
	return false
}

// NewController applies the saved multiplier before the first desktop window.
func NewController(app fyne.App) (*Controller, error) {
	c := &Controller{app: app, percent: app.Preferences().IntWithFallback(preferencePercent, 100)}
	if !supportedPercent(c.percent) {
		c.percent = 100
	}
	value := os.Getenv("FYNE_SCALE")
	if factor, err := strconv.ParseFloat(value, 64); err == nil && factor > 0 && !math.IsNaN(factor) && !math.IsInf(factor, 0) {
		c.override = value
		return c, nil
	}
	if err := os.Setenv("FYNE_SCALE", factorString(c.percent)); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Controller) Percent() int     { return c.percent }
func (c *Controller) Override() string { return c.override }

// SetPercent applies and persists the requested zoom without replacing content.
// Call it on Fyne's UI thread, like other settings changes.
func (c *Controller) SetPercent(percent int) error {
	if c.override != "" {
		return fmt.Errorf("desktop scale is set by FYNE_SCALE")
	}
	if !supportedPercent(percent) {
		return fmt.Errorf("unsupported desktop scale: %d", percent)
	}
	if err := os.Setenv("FYNE_SCALE", factorString(percent)); err != nil {
		return err
	}
	c.percent = percent
	c.app.Preferences().SetInt(preferencePercent, percent)
	c.Refresh()
	return nil
}

func factorString(percent int) string {
	return strconv.FormatFloat(float64(percent)/100, 'f', -1, 64)
}

// Refresh reapplies native scaling and the requested layout without rebuilding content.
// Keeping the requested size avoids feeding pixel-rounding drift into the next change.
func (c *Controller) Refresh() {
	settings := c.app.Settings()
	settings.SetTheme(settings.Theme())
	if c.window == nil {
		return
	}
	fixed := c.window.FixedSize()
	if fixed {
		c.window.SetFixedSize(false)
	}
	c.window.Resize(fitWindowOnScreen(c.window, c.logicalSize))
	if fixed {
		c.window.SetFixedSize(true)
	}
}
