package displayscale

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"os"
	"testing"
)

func TestDesktopZoomPersistsRequestedOSRelativeFactor(t *testing.T) {
	t.Setenv("FYNE_SCALE", "")
	app := test.NewTempApp(t)
	controller, err := NewController(app)
	if err != nil {
		t.Fatal(err)
	}
	if controller.Percent() != 100 || controller.Override() != "" {
		t.Fatal("incorrect default")
	}
	for _, tc := range []struct {
		percent int
		factor  string
	}{{80, "0.8"}, {100, "1"}, {115, "1.15"}, {130, "1.3"}, {200, "2"}} {
		if err := controller.SetPercent(tc.percent); err != nil {
			t.Fatal(err)
		}
		if controller.Percent() != tc.percent || app.Preferences().Int("displayScalePercent") != tc.percent || os.Getenv("FYNE_SCALE") != tc.factor {
			t.Fatalf("zoom %d was not applied and persisted", tc.percent)
		}
	}
	before := os.Getenv("FYNE_SCALE")
	if err := controller.SetPercent(0); err == nil {
		t.Fatal("accepted invalid zoom")
	}
	if os.Getenv("FYNE_SCALE") != before || controller.Percent() != 200 {
		t.Fatal("invalid zoom changed state")
	}
}

func TestSavedZoomAndLaunchOverride(t *testing.T) {
	for _, tc := range []struct {
		saved        int
		launch, want string
		percent      int
	}{{130, "", "1.3", 130}, {130, "auto", "1.3", 130}, {19, "", "1", 100}, {130, "1.5", "1.5", 130}} {
		t.Run(tc.launch+tc.want, func(t *testing.T) {
			t.Setenv("FYNE_SCALE", tc.launch)
			app := test.NewTempApp(t)
			app.Preferences().SetInt("displayScalePercent", tc.saved)
			c, err := NewController(app)
			if err != nil {
				t.Fatal(err)
			}
			if c.Percent() != tc.percent || os.Getenv("FYNE_SCALE") != tc.want {
				t.Fatal("wrong startup zoom")
			}
			if tc.launch == "1.5" {
				if c.Override() != tc.launch {
					t.Fatal("launch override lost")
				}
				if err := c.SetPercent(80); err == nil {
					t.Fatal("overrode launch setting")
				}
				if app.Preferences().Int("displayScalePercent") != tc.saved || os.Getenv("FYNE_SCALE") != tc.launch {
					t.Fatal("override changed preferences")
				}
			}
		})
	}
}

type sizedWindow struct {
	fyne.Window
	sizes []fyne.Size
	fixed []bool
}

func (w *sizedWindow) Resize(size fyne.Size) { w.sizes = append(w.sizes, size); w.Window.Resize(size) }
func (w *sizedWindow) SetFixedSize(fixed bool) {
	w.fixed = append(w.fixed, fixed)
	w.Window.SetFixedSize(fixed)
}

type windowsDriver struct {
	fyne.Driver
	windows []fyne.Window
}

func (d windowsDriver) AllWindows() []fyne.Window { return d.windows }

type windowsApp struct {
	fyne.App
	driver fyne.Driver
}

func (a windowsApp) Driver() fyne.Driver { return a.driver }

func TestZoomReappliesLogicalSizeAndKeepsFixedWindowState(t *testing.T) {
	t.Setenv("FYNE_SCALE", "")
	base := test.NewTempApp(t)
	window := base.NewWindow("scale test")
	defer window.Close()
	input := widget.NewEntry()
	input.SetText("123.456789")
	window.SetContent(input)
	window.Resize(fyne.NewSize(450, 650))
	window.SetFixedSize(true)
	window.Canvas().Focus(input)
	spy := &sizedWindow{Window: window}
	app := windowsApp{App: base, driver: windowsDriver{Driver: base.Driver(), windows: []fyne.Window{spy}}}
	controller, err := NewController(app)
	if err != nil {
		t.Fatal(err)
	}
	stop := controller.Follow(spy, fyne.NewSize(450, 650))
	defer stop()
	if err := controller.SetPercent(130); err != nil {
		t.Fatal(err)
	}
	if len(spy.sizes) != 1 || spy.sizes[0] != fyne.NewSize(450, 650) || len(spy.fixed) != 2 || spy.fixed[0] || !spy.fixed[1] || !window.FixedSize() {
		t.Fatal("zoom did not reapply the fixed window's logical size")
	}

	window.Resize(fyne.NewSize(450.5, 650.8333)) // Simulate the driver's rounded canvas report.
	if err := controller.SetPercent(100); err != nil {
		t.Fatal(err)
	}
	if spy.sizes[len(spy.sizes)-1] != fyne.NewSize(450, 650) {
		t.Fatal("zoom accumulated pixel-rounding drift")
	}
	if window.Content() != input || input.Text != "123.456789" || window.Canvas().Focused() != input {
		t.Fatal("zoom changed the active input")
	}
}
