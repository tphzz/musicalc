package ui

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// zoneColors and zoneNames describe the phase zones in the UI
var (
	zoneColors = map[logic.PhaseZone]fyne.ThemeColorName{
		logic.PhaseSweetSpot: theme.ColorNameSuccess,
		logic.PhasePartial:   theme.ColorNameWarning,
		logic.PhaseDeathZone: theme.ColorNameError,
	}
	zoneNames = map[logic.PhaseZone]string{
		logic.PhaseSweetSpot: "Sweet spot · in phase",
		logic.PhasePartial:   "Partial cancellation",
		logic.PhaseDeathZone: "Death zone · cancellation",
	}
)

// minusSign formats negative numbers with a real minus sign instead of a hyphen
func minusSign(s string) string {
	return strings.Replace(s, "-", "−", 1)
}

// labeledRow is a two-column row: a label and its control, like the other calculators
func labeledRow(label string, control fyne.CanvasObject) fyne.CanvasObject {
	return container.NewGridWithColumns(2, widget.NewLabel(label), control)
}

// withSuffix puts a muted unit after a control
func withSuffix(control fyne.CanvasObject, suffix *widget.RichText) fyne.CanvasObject {
	return container.NewBorder(nil, nil, nil, suffix, control)
}

// newResultsScroll puts result cards in a scroll view and returns a function that lays them out
// again. A wrapped text can gain or lose a line without changing its width, which Fyne doesn't
// notice by itself, so call it after updating the cards' texts.
func newResultsScroll(objects ...fyne.CanvasObject) (fyne.CanvasObject, func()) {
	list := container.NewVBox(objects...)
	// A little inset keeps the card shadows from being clipped by the scroller
	scroll := container.NewVScroll(container.New(layout.NewCustomPaddedLayout(2, 4, 2, 2), list))
	return scroll, func() {
		list.Refresh()
		scroll.Refresh()
	}
}

// zoneDot is the small coloured circle in front of a zone name
func zoneDot(colorName fyne.ThemeColorName) fyne.CanvasObject {
	dot := canvas.NewCircle(theme.Color(colorName))
	return container.NewCenter(container.NewGridWrap(fyne.NewSquareSize(10), dot))
}

// NewPhaseDistanceTab creates the Phase-Safe Distance calculator: where the path difference between
// two mics on one source makes a frequency reinforce (sweet spots) or cancel (death zones)
func NewPhaseDistanceTab() fyne.CanvasObject {
	var update func()
	restoring := false // the last session is being put on screen
	refresh := func() {
		if update != nil && !restoring {
			update()
		}
	}

	// Typical fundamentals; picking one fills in the frequency
	preset := func(label string, hz float64) widgets.MenuOption {
		return widgets.MenuOption{Label: label, Value: logic.FormatTrim(hz, 2)}
	}
	presets := []widgets.MenuGroup{
		{Name: "Drums", Options: []widgets.MenuOption{
			preset("Kick", 60), preset("Floor tom", 80), preset("Rack tom", 120), preset("Snare", 200)}},
		{Name: "Bass", Options: []widgets.MenuOption{
			preset("Low B (5-string)", 30.87), preset("Low E", 41.2), preset("A string", 55)}},
		{Name: "Guitars", Options: []widgets.MenuOption{
			preset("Low E", 82.41), preset("A string", 110)}},
		{Name: "Vocals", Options: []widgets.MenuOption{
			preset("Male voice", 110), preset("Female voice", 220)}},
	}

	freqEntry := widgets.NewNumericEntry()
	freqEntry.SetPlaceHolder("Frequency")
	presetSelect := widgets.NewMenuSelect(presets, func(hz string) { freqEntry.SetText(hz) })
	presetSelect.Placeholder = "Custom"
	freqEntry.OnChanged = func(s string) {
		// show the matching preset, or "Custom" for any other frequency
		presetSelect.SetSelected(logic.FormatTrim(logic.ParseFloat(s), 2))
		refresh()
	}

	temp := newTemperatureInput("Temperature", logic.UnitCelsius, refresh)
	tempUnit := newUnitSelect([]string{logic.UnitCelsius, logic.UnitFahrenheit}, []*unitInput{temp}, func(string) { refresh() })

	path := newDistanceInput("Path difference", logic.UnitMeters, refresh)
	pathSuffix := widgets.NewMutedText(logic.UnitMeters)
	distUnit := newUnitSelect([]string{logic.UnitMeters, logic.UnitFeet}, []*unitInput{path}, func(unit string) {
		widgets.SetStyledText(pathSuffix, unit, "")
		refresh()
	})

	// Labels draw their own inner padding; overlapping it keeps texts a padding / line apart
	textGap := theme.Size(theme.SizeNamePadding) - 2*theme.Size(theme.SizeNameInnerPadding)
	rowGap := theme.Size(theme.SizeNameLineSpacing) - 2*theme.Size(theme.SizeNameInnerPadding)
	vbox := func(objects ...fyne.CanvasObject) *fyne.Container {
		return container.New(layout.NewCustomPaddedVBoxLayout(rowGap), objects...)
	}
	wrapped := func(rt *widget.RichText) *widget.RichText {
		rt.Wrapping = fyne.TextWrapWord
		return rt
	}

	// Wavelength
	wlValue := widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true})
	wlCaption := widgets.NewMutedText("")
	wlDetail := wrapped(widgets.NewMutedText(""))
	wlCard := widgets.NewCard(vbox(widgets.NewBaselineRow(textGap, wlValue, wlCaption), wlDetail))

	// Path difference check
	phaseValue := widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true})
	zoneText := widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{})
	sumText := wrapped(widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{}))
	resultRow := widgets.NewBaselineRow(textGap, phaseValue, zoneText)
	// The entry and the zone map have no inner text padding to overlap, so they get normal spacing
	checkCard := widgets.NewCard(container.NewVBox(
		widgets.NewMutedText("Path difference between the two mics"),
		withSuffix(path.Entry, pathSuffix),
		vbox(resultRow, sumText),
	))

	// Zone map
	mapTitle := widgets.NewMutedText("")
	zoneMap := widgets.NewZoneMap()
	mapCard := widgets.NewCard(container.NewVBox(mapTitle, zoneMap))

	// Zone list
	trailing := func(rt *widget.RichText) *widget.RichText {
		rt.Segments[0].(*widget.TextSegment).Style.Alignment = fyne.TextAlignTrailing
		return rt
	}
	sweetValues := trailing(widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{}))
	deathValues := trailing(widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{}))
	zoneRow := func(dotColor fyne.ThemeColorName, name string, values *widget.RichText) fyne.CanvasObject {
		title := widgets.NewStyledText(name, theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{Bold: true})
		return container.NewBorder(nil, nil, container.NewHBox(zoneDot(dotColor), title), nil, values)
	}
	tolerance := wrapped(widgets.NewMutedText(""))
	zonesCard := widgets.NewCard(vbox(
		zoneRow(theme.ColorNameSuccess, "Sweet spots", sweetValues),
		zoneRow(theme.ColorNameError, "Death zones", deathValues),
		tolerance,
	))

	// Setup card: source, frequency, temperature and units; the path difference stays in the
	// check card, as it is the value people vary
	var content *fyne.Container
	var saveSession func()
	setup := newSetupCard(container.NewVBox(
		labeledRow("Source", presetSelect),
		labeledRow("Frequency", withSuffix(freqEntry, widgets.NewMutedText("Hz"))),
		labeledRow("Temperature", container.NewBorder(nil, nil, nil, container.New(fixedWidthLayout{width: 70}, tempUnit), temp.Entry)),
		labeledRow("Units", distUnit),
		wrapped(widgets.NewMutedText("Pick a source for its typical fundamental, or type any frequency. "+
			"Enter the path difference between the two mics below.")),
	), func(bool) {
		if content != nil {
			content.Refresh() // the card's height changed
		}
		if saveSession != nil {
			saveSession()
		}
	})

	results, relayout := newResultsScroll(wlCard, checkCard, mapCard, zonesCard)
	update = func() {
		defer relayout()
		defer saveSession()
		freq := logic.ParseFloat(freqEntry.Text)
		tempC := temp.Base
		unit := distUnit.Selected
		dist := func(meters float64) string {
			return logic.FormatTrim(logic.FromMeters(meters, unit), 2) + " " + unit
		}
		distList := func(values []float64) string {
			parts := make([]string, len(values))
			for i, m := range values {
				parts[i] = logic.FormatTrim(logic.FromMeters(m, unit), 2)
			}
			return strings.Join(parts, " · ") + " " + unit
		}

		title := presetSelect.Text()
		if title == "" {
			title = "Custom"
		}
		setup.SetTitle(title)
		freqShown := "–"
		if freq > 0 {
			freqShown = logic.FormatTrim(freq, 2)
		}
		tempSummary := "–"
		if !temp.Empty() {
			tempSummary = minusSign(logic.FormatTrim(logic.FromCelsius(tempC, tempUnit.Selected), 1))
		}
		setup.SetSummary(freqShown + " Hz · " + tempSummary + " " + tempUnit.Selected + " · " + unit)

		wl := logic.Wavelength(freq, tempC)
		if wl <= 0 {
			widgets.SetStyledText(wlValue, "–", "")
			widgets.SetStyledText(wlCaption, "enter a frequency", "")
			widgets.SetStyledText(wlDetail, "", "")
			checkCard.Hide()
			mapCard.Hide()
			zonesCard.Hide()
			return
		}
		checkCard.Show()
		mapCard.Show()
		zonesCard.Show()

		hz := logic.FormatTrim(freq, 2)
		widgets.SetStyledText(wlValue, dist(wl), "")
		widgets.SetStyledText(wlCaption, "wavelength at "+hz+" Hz", "")
		tempShown := logic.FormatTrim(logic.FromCelsius(tempC, tempUnit.Selected), 1)
		widgets.SetStyledText(wlDetail, fmt.Sprintf("½ λ = %s · %.1f m/s at %s %s",
			dist(wl/2), logic.SpeedOfSound(tempC), minusSign(tempShown), tempUnit.Selected), "")

		maxDist := 2 * wl
		marker := -1.0
		if path.Empty() {
			checkCard.SetOutline("")
			resultRow.Hide()
			// the hint goes in the wrapping line, so it doesn't widen the layout on phones
			widgets.SetStyledText(sumText, "Enter how much farther one mic is from the source", theme.ColorNamePlaceHolder)
		} else {
			resultRow.Show()
			r := logic.CheckPhase(path.Base, freq, tempC)
			color := zoneColors[r.Zone]
			checkCard.SetOutline(color)
			widgets.SetStyledText(phaseValue, fmt.Sprintf("%.0f°", r.Phase), color)
			widgets.SetStyledText(zoneText, zoneNames[r.Zone], color)
			level := "Complete cancellation"
			if !math.IsInf(r.SummedDB, -1) && r.SummedDB > -60 {
				level = minusSign(fmt.Sprintf("%.1f dB", r.SummedDB))
			} else if !math.IsInf(r.SummedDB, -1) {
				level = "Below −60 dB"
			}
			widgets.SetStyledText(sumText, fmt.Sprintf("%s at %s Hz when both mics are summed", level, hz), color)
			if math.Abs(path.Base) <= maxDist {
				marker = math.Abs(path.Base)
			}
		}

		segments := logic.ZoneSegments(wl, maxDist)
		bands := make([]widgets.ZoneBand, len(segments))
		for i, s := range segments {
			bands[i] = widgets.ZoneBand{Start: s.Start, End: s.End, Color: zoneColors[s.Zone]}
		}
		var labels []widgets.ZoneLabel
		for _, at := range []float64{0, wl / 2, wl, 1.5 * wl, 2 * wl} {
			labels = append(labels, widgets.ZoneLabel{At: at, Text: logic.FormatTrim(logic.FromMeters(at, unit), 2)})
		}
		zoneMap.SetZones(maxDist, bands, labels, marker)
		widgets.SetStyledText(mapTitle, "Zones up to "+dist(maxDist), "")

		widgets.SetStyledText(sweetValues, distList(logic.SweetSpots(wl, maxDist)), "")
		widgets.SetStyledText(deathValues, distList(logic.DeathZones(wl, maxDist)), "")
		widgets.SetStyledText(tolerance, fmt.Sprintf("Safe within ±%s of a sweet spot (±%.0f°)",
			dist(logic.SafeTolerance(wl)), logic.PhaseSafeDegrees), "")
	}

	saveSession = func() {
		if restoring {
			return
		}
		sess := phaseSession{Frequency: logic.ParseFloat(freqEntry.Text), TemperatureC: temp.Base,
			TemperatureUnit: tempUnit.Selected, DistanceUnit: distUnit.Selected,
			PathMeters: path.Base, HasPath: !path.Empty(), Open: setup.IsOpen()}
		if data, err := json.Marshal(sess); err == nil {
			fyne.CurrentApp().Preferences().SetString(prefPhaseSession, string(data))
		}
	}

	// Restore the last setup; the first start opens the card with a kick drum at 22 °C
	sess := phaseSession{Frequency: 60, TemperatureC: 22, TemperatureUnit: logic.UnitCelsius,
		DistanceUnit: logic.UnitMeters, Open: true}
	if data := fyne.CurrentApp().Preferences().String(prefPhaseSession); data != "" {
		_ = json.Unmarshal([]byte(data), &sess)
	}
	restoring = true
	if sess.TemperatureUnit != "" {
		tempUnit.SetSelected(sess.TemperatureUnit)
	}
	if sess.DistanceUnit != "" {
		distUnit.SetSelected(sess.DistanceUnit)
	}
	temp.setBase(sess.TemperatureC)
	if sess.HasPath {
		path.setBase(sess.PathMeters)
	}
	freqEntry.SetText(logic.FormatTrim(sess.Frequency, 2))
	setup.SetOpen(sess.Open)
	restoring = false
	refresh()

	content = container.NewBorder(
		container.New(layout.NewCustomPaddedLayout(2, 2, 2, 2), setup.Card),
		nil, nil, nil,
		results,
	)
	return content
}

// prefPhaseSession stores the Phase-Safe Distance setup, so it is back when the app starts again
const prefPhaseSession = "phase.session"

// phaseSession is the Phase-Safe Distance setup on screen
type phaseSession struct {
	Frequency       float64 `json:"frequency"`
	TemperatureC    float64 `json:"temperatureC"`
	TemperatureUnit string  `json:"temperatureUnit"`
	DistanceUnit    string  `json:"distanceUnit"`
	PathMeters      float64 `json:"pathMeters"`
	HasPath         bool    `json:"hasPath"`
	Open            bool    `json:"open"`
}
