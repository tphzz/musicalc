package ui

import (
	"encoding/json"
	"fmt"
	"math"
	"musicalc/internal/i18n"
	"strings"

	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// prefTwoMicSession stores the Two-Mic Sum setup, so it is back when the app starts again
const prefTwoMicSession = "twoMicSum.session"

// twoMicSession is the Two-Mic Sum setup on screen; entries are kept as typed
type twoMicSession struct {
	Lo, Hi, Freq string
	TempC        float64
	TempUnit     string
	DistUnit     string
	SampleRate   string
	Farther      string
	FartherUnit  string
	// The physical "Farther by" value: Basis "distance" with Meters or "time" with Seconds. Absent
	// in sessions from older versions, which are read from Farther and FartherUnit instead.
	Basis    string   `json:",omitempty"`
	Meters   *float64 `json:",omitempty"`
	Seconds  *float64 `json:",omitempty"`
	Level    string
	Inverted bool
	DelayA   string
	Open     bool
}

// sourceRange is a typical range of interest and a frequency for the cursor
type sourceRange struct{ lo, hi, cursor float64 }

// unitMS is the "Farther by" unit for entering a measured time difference
const unitMS = "ms"

// NewTwoMicSumTab creates the Two-Mic Sum calculator: the same signal reaches two mics at
// different times; what does mixing them do to the combined sound?
func NewTwoMicSumTab() *View {
	texts := newTextRefresh()
	var update func()
	restoring := false
	refresh := func() {
		if update != nil && !restoring {
			update()
		}
	}

	// Sources pick a range of interest and a frequency for the cursor
	ranges := map[string]sourceRange{}
	group := func(name string, opts ...any) widgets.MenuGroup {
		g := widgets.MenuGroup{Name: i18n.T("source." + name)}
		for i := 0; i < len(opts); i += 2 {
			label, r := opts[i].(string), opts[i+1].(sourceRange)
			key := name + "|" + label
			ranges[key] = r
			g.Options = append(g.Options, widgets.MenuOption{Label: i18n.T("source." + label), Value: key})
		}
		return g
	}
	groups := []widgets.MenuGroup{
		group("Drums", "Kick", sourceRange{50, 100, 60}, "Floor tom", sourceRange{60, 120, 80},
			"Rack tom", sourceRange{100, 250, 120}, "Snare", sourceRange{150, 300, 200}, "Cymbals", sourceRange{3000, 12000, 6000}),
		group("Bass", "Low B (5-string)", sourceRange{30, 100, 30.87}, "Low E", sourceRange{40, 120, 41.2},
			"A string", sourceRange{55, 165, 55}),
		group("Guitars", "Low E", sourceRange{80, 330, 82.41}, "A string", sourceRange{110, 440, 110}),
		group("Vocals", "Male voice", sourceRange{100, 400, 110}, "Female voice", sourceRange{200, 800, 220}),
	}

	loEntry, hiEntry, freqEntry := widgets.NewNumericEntry(), widgets.NewNumericEntry(), widgets.NewNumericEntry()
	texts.placeholder(loEntry, "field.from")
	texts.placeholder(hiEntry, "field.to")
	texts.placeholder(freqEntry, "field.frequency")
	presetSelect := widgets.NewMenuSelect(groups, nil)
	presetSelect.Placeholder = i18n.T("field.custom")
	presetSelect.OnChanged = func(key string) {
		r := ranges[key]
		restoring = true
		loEntry.SetText(logic.FormatTrim(r.lo, 2))
		hiEntry.SetText(logic.FormatTrim(r.hi, 2))
		freqEntry.SetText(logic.FormatTrim(r.cursor, 2))
		restoring = false
		refresh()
	}
	// matchPreset shows the source whose range and cursor are entered, or "Custom"
	matchPreset := func() {
		lo, hi, f := logic.ParseFloat(loEntry.Text), logic.ParseFloat(hiEntry.Text), logic.ParseFloat(freqEntry.Text)
		for _, g := range groups {
			for _, o := range g.Options {
				r := ranges[o.Value]
				if r.lo == lo && r.hi == hi && math.Abs(r.cursor-f) < 1e-9 {
					presetSelect.SetSelected(o.Value)
					return
				}
			}
		}
		presetSelect.SetSelected("")
	}
	for _, e := range []*widgets.NumericEntry{loEntry, hiEntry, freqEntry} {
		e.OnChanged = func(string) {
			if !restoring {
				matchPreset()
			}
			refresh()
		}
	}

	temp := newTemperatureInput(i18n.T("field.temperature"), logic.UnitCelsius, refresh)
	texts.placeholder(temp.Entry, "field.temperature")
	tempUnit := newUnitSelect([]string{logic.UnitCelsius, logic.UnitFahrenheit}, []*unitInput{temp}, func(string) { refresh() })
	sampleRate := widget.NewSelect([]string{"44100", "48000", "88200", "96000", "192000"}, func(string) { refresh() })
	sampleRate.Selected = "48000"

	// Mic B: how much farther (a distance, or a measured time), its level and polarity; and a delay on Mic A
	farther := widgets.NewNumericEntry()
	texts.placeholder(farther, "field.distance")
	fartherUnit := widget.NewSelect([]string{logic.UnitMeters, unitMS}, nil)
	fartherUnit.Selected = logic.UnitMeters
	// The "Farther by" value keeps what was entered, a distance or a measured time, at full
	// precision. Switching its unit only changes how it is shown; editing it chooses the basis.
	var basis farBasis
	var basisMeters, basisSeconds float64
	converting := false
	farther.OnChanged = func(text string) {
		if converting {
			return
		}
		v, ok := logic.ParseNumber(text)
		switch {
		case strings.TrimSpace(text) == "":
			basis = farEmpty
		case !ok:
			basis = farInvalid
		case fartherUnit.Selected == unitMS:
			basis, basisSeconds = farTime, v/1000
		default:
			basis, basisMeters = farDistance, logic.ToMeters(v, fartherUnit.Selected)
		}
		refresh()
	}
	// showFarther writes the stored value in the selected unit; a distance and a time convert through
	// the speed of sound, which needs a valid temperature
	showFarther := func() {
		var v float64
		ok := true
		unit := fartherUnit.Selected
		switch {
		case basis == farDistance && unit != unitMS:
			v = logic.FromMeters(basisMeters, unit)
		case basis == farDistance:
			v, ok = logic.ArrivalDelay(basisMeters, temp.Base)
			v *= 1000
		case basis == farTime && unit == unitMS:
			v = basisSeconds * 1000
		case basis == farTime:
			v, ok = logic.DistanceForDelay(basisSeconds, temp.Base)
			v = logic.FromMeters(v, unit)
		default:
			return // empty or invalid text stays as typed
		}
		if !ok {
			return
		}
		converting = true
		farther.SetText(logic.FormatTrim(v, 2))
		converting = false
	}
	fartherUnit.OnChanged = func(string) {
		showFarther()
		refresh()
	}
	// The distance unit for the whole tab; "Farther by" offers it and ms
	distUnit := widget.NewSelect([]string{logic.UnitMeters, logic.UnitFeet}, nil)
	distUnit.Selected = logic.UnitMeters
	distUnit.OnChanged = func(unit string) {
		fartherUnit.Options = []string{unit, unitMS}
		if fartherUnit.Selected != unitMS {
			fartherUnit.SetSelected(unit) // shows the value in the new unit
		}
		fartherUnit.Refresh()
		refresh()
	}
	level := widgets.NewNumericEntry()
	level.SetPlaceHolder("0")
	level.OnChanged = func(string) { refresh() }
	polarity := widget.NewSelect([]string{i18n.T("polarity.normal"), i18n.T("polarity.inverted")}, func(string) { refresh() })
	polarity.Selected = i18n.T("polarity.normal")
	texts.selectOptions(polarity, []string{"polarity.normal", "polarity.inverted"})
	delayA := widgets.NewNumericEntry()
	delayA.SetPlaceHolder("0")
	delayA.OnChanged = func(string) { refresh() }

	// fixed gives a select a fixed width next to an entry
	fixed := func(o fyne.CanvasObject, w float32) fyne.CanvasObject {
		return container.New(fixedWidthLayout{width: w}, o)
	}

	var content *fyne.Container
	var saveSession func()
	setup := newSetupCard(container.NewVBox(
		texts.row("field.source", presetSelect),
		texts.row("field.range", container.NewGridWithColumns(2, loEntry, hiEntry)),
		texts.row("field.frequency", withSuffix(freqEntry, widgets.NewMutedText("Hz"))),
		texts.row("field.temperature", container.NewBorder(nil, nil, nil, fixed(tempUnit, 70), temp.Entry)),
		texts.row("field.units", distUnit),
		texts.row("field.sample_rate", sampleRate),
		texts.wrappedMuted("sum.source_hint"),
	), func(bool) {
		if content != nil {
			content.Refresh() // the card's height changed
		}
		if saveSession != nil {
			saveSession()
		}
	})

	// Texts
	textGap := theme.Size(theme.SizeNameLineSpacing) - 2*theme.Size(theme.SizeNameInnerPadding)
	vbox := func(objects ...fyne.CanvasObject) *fyne.Container {
		return container.New(layout.NewCustomPaddedVBoxLayout(textGap), objects...)
	}
	plain := func() *widget.RichText {
		rt := widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{})
		rt.Wrapping = fyne.TextWrapWord
		return rt
	}
	bold := func(text string) *widget.RichText {
		rt := widgets.NewStyledText(text, theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{Bold: true})
		rt.Wrapping = fyne.TextWrapWord
		return rt
	}

	compared := bold("")
	texts.add(func() { widgets.SetStyledText(compared, i18n.T("sum.compared"), "") })
	inputsCard := widgets.NewCard(container.NewVBox(
		compared,
		texts.row("sum.farther", container.NewBorder(nil, nil, nil, fixed(fartherUnit, 70), farther)),
		texts.row("field.level", withSuffix(level, widgets.NewMutedText("dB"))),
		texts.row("field.polarity", polarity),
		texts.row("sum.delay_a", withSuffix(delayA, widgets.NewMutedText("ms"))),
	))

	// Arrival: the main result, with Align to try the matching delay on Mic A
	arrivalTitle := widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true})
	arrivalTitle.Wrapping = fyne.TextWrapWord
	arrivalText := wrappedMuted("")
	alignButton := widget.NewButton(i18n.T("action.align"), nil)
	arrivalCard := widgets.NewCard(vbox(
		container.NewBorder(nil, nil, nil, container.NewCenter(alignButton), arrivalTitle),
		arrivalText,
	))

	graph := widgets.NewResponseGraph()
	readout := bold("")
	graphCard := widgets.NewCard(container.NewVBox(
		texts.wrappedMuted("sum.combined_level"),
		graph,
		readout,
	))
	combLines := []*widget.RichText{plain(), plain(), plain()}
	combTitle := bold("")
	texts.add(func() { widgets.SetStyledText(combTitle, i18n.T("sum.comb_title"), "") })
	combCard := widgets.NewCard(vbox(combTitle, combLines[0], combLines[1], combLines[2]))
	suggestTitle, suggestText := bold(""), plain()
	suggestCard := widgets.NewCard(vbox(suggestTitle, suggestText))
	modelCard := widgets.NewCard(texts.wrappedMuted("sum.model"))

	results, relayout := newResultsScroll(inputsCard, arrivalCard, graphCard, combCard, suggestCard, modelCard)

	ms := func(seconds float64) string { return fmt.Sprintf("%.2f", math.Abs(seconds)*1000) }
	signedDB := func(v float64) string { return minusSign(fmt.Sprintf("%+.1f", v)) }
	freqList := func(fs []float64) string {
		if len(fs) == 0 {
			return i18n.T("sum.none_audible")
		}
		parts := make([]string, len(fs))
		for i, f := range fs {
			parts[i] = logic.FormatTrim(math.Round(f*10)/10, 1)
		}
		return strings.Join(parts, " · ") + " Hz …"
	}
	// audible keeps the frequencies up to 20 kHz
	audible := func(fs []float64) []float64 {
		var out []float64
		for _, f := range fs {
			if f <= 20000 {
				out = append(out, f)
			}
		}
		return out
	}
	setLines := func(lines []*widget.RichText, texts ...string) {
		for i, l := range lines {
			if i < len(texts) {
				widgets.SetStyledText(l, texts[i], "")
				l.Show()
			} else {
				l.Hide()
			}
		}
	}

	var renderResult func()
	update = func() {
		lo, loOK := logic.ParseNumber(loEntry.Text)
		hi, hiOK := logic.ParseNumber(hiEntry.Text)
		cursor, cursorOK := logic.ParseNumber(freqEntry.Text)
		cursorEmpty := strings.TrimSpace(freqEntry.Text) == ""
		rate := logic.ParseSampleRate(sampleRate.Selected)
		hasArrival := basis != farEmpty
		bandOK := loOK && hiOK && logic.BandInRange(lo, hi)
		// A distance, or showing the time as one, needs the speed of sound
		needsTemp := basis == farDistance || fartherUnit.Selected != unitMS
		tempOK := temp.Valid() && logic.TemperatureInRange(temp.Base) // "abc" or "-" is not 0 °C
		// A value shown in the other kind of unit follows the temperature; typed text stays as typed
		if (basis == farDistance) == (fartherUnit.Selected == unitMS) && (basis == farDistance || basis == farTime) {
			showFarther()
		}
		var fv, arrival, delay, net, levelDB, gLo, gHi, cursorDB, lowDB, highDB float64
		var m logic.TwoMicSum
		var dipFreqs, peakFreqs []float64
		invalid := "" // the message key of the first unusable input
		if bandOK {
			gLo, gHi = logic.GraphRange(lo, hi)
		}
		levelDB, levelOK := logic.ParseNumber(level.Text)
		if strings.TrimSpace(level.Text) == "" {
			levelDB, levelOK = 0, true
		}
		delayMS, delayOK := logic.ParseNumber(delayA.Text)
		if strings.TrimSpace(delayA.Text) == "" {
			delayMS, delayOK = 0, true
		}
		delay = delayMS / 1000
		switch {
		case !hasArrival:
		case basis == farInvalid:
			invalid = "sum.invalid_farther"
		case needsTemp && !tempOK:
			invalid = "sum.invalid_temperature"
		case !bandOK:
			invalid = "sum.invalid_band"
		case !cursorEmpty && !(cursorOK && cursor >= gLo && cursor <= gHi):
			invalid = "sum.invalid_cursor"
		case !levelOK:
			invalid = "sum.invalid_level"
		case !delayOK || delay < 0:
			invalid = "sum.invalid_delay"
		}
		if cursorEmpty {
			cursor = 0
		}
		if hasArrival && invalid == "" {
			if basis == farTime {
				arrival = basisSeconds
				fv, _ = logic.DistanceForDelay(arrival, temp.Base)
			} else {
				arrival, _ = logic.ArrivalDelay(basisMeters, temp.Base)
				fv = basisMeters
			}
			fv = logic.FromMeters(fv, distUnit.Selected)
			// Times must stay countable in samples, or they would show as Inf or endless digits
			switch {
			case !logic.SampleCountInRange(arrival, rate):
				invalid = "sum.invalid_farther"
			case !logic.SampleCountInRange(delay, rate):
				invalid = "sum.invalid_delay"
			}
			// The residual stays exact: rounding it would hide a comb that is there
			net = arrival - delay
			m = logic.TwoMicSum{Delay: net, Gain: logic.DBToGain(levelDB), Inverted: polarity.SelectedIndex() == 1}
			if invalid == "" && !m.Valid() {
				invalid = "sum.invalid_level" // the level is too extreme to convert
			}
		}
		if hasArrival && invalid == "" {
			curve := func(model logic.TwoMicSum, before bool) widgets.GraphCurve {
				fs := logic.SampleFrequencies(gLo, gHi, 600, model)
				ls := make([]float64, len(fs))
				for i, f := range fs {
					ls[i] = model.LevelDB(f)
				}
				return widgets.GraphCurve{Freqs: fs, Levels: ls, Before: before,
					Spacing: model.CombSpacing(), EnvLo: model.DipDB(), EnvHi: model.PeakDB()}
			}
			curves := []widgets.GraphCurve{}
			changed := delay != 0 || levelDB != 0 || m.Inverted
			if changed {
				curves = append(curves, curve(logic.TwoMicSum{Delay: arrival, Gain: 1}, true))
			}
			curves = append(curves, curve(m, false))
			if cursor > 0 {
				cursorDB = m.LevelDB(cursor)
			}
			graph.SetData(gLo, gHi, curves, lo, hi, cursor, cursorDB)
			lowDB, highDB = m.RangeDB(lo, hi)
			if !m.Aligned() {
				dipFreqs = m.DipsFrom(gLo, 3)
				peakFreqs = m.PeaksFrom(gLo, 3)
			}
		}
		modelAligned, dipDB, peakDB, combSpacing := m.Aligned(), m.DipDB(), m.PeakDB(), m.CombSpacing()
		renderResult = func() {
			defer relayout()

			// Setup card header
			title := presetSelect.Text()
			if title == "" {
				title = i18n.T("field.custom")
			}
			setup.SetTitle(title)
			rangeText := "–"
			if bandOK {
				rangeText = logic.FormatTrim(lo, 2) + "–" + logic.FormatTrim(hi, 2) + " Hz"
			}
			tempText := "–"
			if !temp.Empty() {
				tempText = minusSign(logic.FormatTrim(logic.FromCelsius(temp.Base, tempUnit.Selected), 1)) + " " + tempUnit.Selected
			}
			setup.SetSummary(rangeText + " · " + tempText + " · " + logic.FormatTrim(float64(rate)/1000, 1) + " kHz")

			// Arrival difference
			if !hasArrival || invalid != "" {
				message := i18n.T("sum.enter_distance")
				switch invalid {
				case "":
				case "sum.invalid_cursor":
					message = i18n.T(invalid, logic.FormatTrim(gLo, 1), logic.FormatTrim(gHi, 1))
				default:
					message = i18n.T(invalid)
				}
				widgets.SetStyledText(arrivalTitle, "–", "")
				widgets.SetStyledText(arrivalText, message, "")
				alignButton.Hide()
				for _, c := range []fyne.CanvasObject{graphCard, combCard, suggestCard} {
					c.Hide()
				}
				return
			}
			for _, c := range []fyne.CanvasObject{graphCard, combCard, suggestCard} {
				c.Show()
			}
			samples := func(s float64) string { return i18n.T("sum.samples", math.Abs(s)*float64(rate)) }
			laterOrEarlier := func(s float64) string {
				if s < 0 {
					return i18n.T("sum.earlier")
				}
				return i18n.T("sum.later")
			}
			// Delays are entered to 0.01 ms: within half of that Mic B counts as lined up, but the
			// response keeps the residual
			aligned := delay != 0 && math.Abs(net) < 0.000005
			switch {
			case aligned:
				widgets.SetStyledText(arrivalTitle, i18n.T("sum.aligned"), theme.ColorNameSuccess)
				detail := i18n.T("sum.aligned_detail", ms(arrival), ms(delay))
				if net != 0 {
					detail += i18n.T("sum.aligned_residual", fmt.Sprintf("%.2f", math.Abs(net)*1e6), laterOrEarlier(net))
				}
				widgets.SetStyledText(arrivalText, detail, "")
			case delay != 0:
				widgets.SetStyledText(arrivalTitle, i18n.T("sum.net", ms(net), laterOrEarlier(net)), theme.ColorNameForeground)
				widgets.SetStyledText(arrivalText, i18n.T("sum.delayed_detail", ms(arrival), laterOrEarlier(arrival), ms(delay)), "")
			case net == 0:
				widgets.SetStyledText(arrivalTitle, i18n.T("sum.together"), theme.ColorNameForeground)
				widgets.SetStyledText(arrivalText, i18n.T("sum.no_difference"), "")
			default:
				widgets.SetStyledText(arrivalTitle, i18n.T("sum.arrival", ms(arrival), laterOrEarlier(arrival)), theme.ColorNameForeground)
				detail := i18n.T("sum.sample_detail", samples(arrival), logic.FormatTrim(float64(rate)/1000, 1))
				if basis == farDistance {
					detail += i18n.T("sum.distance_detail", logic.FormatTrim(math.Abs(fv), 2), distUnit.Selected, map[bool]string{true: i18n.T("sum.farther_word"), false: i18n.T("sum.closer_word")}[fv >= 0], tempText)
				}
				widgets.SetStyledText(arrivalText, detail, "")
			}
			// Align fills in the delay that lines Mic B up; Reset takes it away again
			alignButton.Show()
			if delay != 0 {
				alignButton.SetText(i18n.T("action.reset"))
				alignButton.OnTapped = func() { delayA.SetText("0") }
				alignButton.Enable()
			} else {
				alignButton.SetText(i18n.T("action.align"))
				alignButton.OnTapped = func() { delayA.SetText(fmt.Sprintf("%.2f", arrival*1000)) }
				// A delay that rounds to 0.00 ms would change nothing
				if arrival > 0 && fmt.Sprintf("%.2f", arrival*1000) != "0.00" {
					alignButton.Enable()
				} else {
					alignButton.Disable() // Mic B is first: the delay belongs on Mic B
				}
			}

			// The model's floor stands for complete cancellation; down to −60 dB is "nearly"
			reading := ""
			if cursor > 0 {
				switch {
				case cursorDB <= logic.MinLevelDB:
					reading = i18n.T("sum.cursor_cancel", logic.FormatTrim(cursor, 2))
				case cursorDB <= -60:
					reading = i18n.T("sum.cursor_near", logic.FormatTrim(cursor, 2))
				default:
					reading = i18n.T("sum.cursor", logic.FormatTrim(cursor, 2), signedDB(cursorDB))
				}
			}
			{
				band := rangeText + ": "
				switch {
				case highDB <= logic.MinLevelDB:
					band += i18n.T("sum.band_cancel_all")
				case highDB-lowDB < 0.05:
					band += i18n.T("sum.band_constant", signedDB(highDB))
				case lowDB <= logic.MinLevelDB:
					band += i18n.T("sum.band_cancel", signedDB(highDB))
				case lowDB <= -60:
					band += i18n.T("sum.band_near", signedDB(highDB))
				default:
					band += i18n.T("sum.band_range", signedDB(lowDB), signedDB(highDB))
				}
				if reading != "" {
					reading += " · "
				}
				reading += band
			}
			widgets.SetStyledText(readout, reading, "")

			// Predicted cancellation and reinforcement
			switch {
			case modelAligned && !m.Inverted:
				setLines(combLines, i18n.T("sum.aligned_comb", signedDB(peakDB)))
			case modelAligned:
				if dipDB <= logic.MinLevelDB {
					setLines(combLines, i18n.T("sum.cancel_comb"))
				} else {
					// A louder inverted Mic B gives a boost, so the level is signed
					setLines(combLines, i18n.T("sum.reduced_comb", signedDB(dipDB)))
				}
			default:
				dips, peaks := i18n.T("sum.dips"), i18n.T("sum.peaks")
				if dipDB > -60 {
					dips = i18n.T("sum.dips_level", signedDB(dipDB))
					peaks = i18n.T("sum.peaks_level", signedDB(peakDB))
				}
				// List the comb from the graph's range on; a dense comb gets its spacing instead of more values
				if m.Inverted && gLo < combSpacing {
					dips += i18n.T("sum.low_frequencies")
				}
				spacing := ""
				if (gHi-gLo)/combSpacing > 8 {
					spacing = i18n.T("sum.spacing", logic.FormatTrim(math.Round(combSpacing*10)/10, 1))
				}
				lines := []string{dips + freqList(audible(dipFreqs)) + spacing, peaks + freqList(audible(peakFreqs)) + spacing}
				if len(audible(dipFreqs)) == 0 && len(audible(peakFreqs)) == 0 {
					// A residual of microseconds puts the comb above hearing: say what is left of it
					low, high := m.RangeDB(20, 20000)
					lines = []string{i18n.T("sum.no_audible_comb", signedDB(low), signedDB(high))}
				}
				var notes []string
				if levelDB < -0.05 {
					notes = append(notes, i18n.T("sum.shallow", -levelDB))
				}
				if m.Inverted {
					notes = append(notes, i18n.T("sum.invert_note"))
				}
				if len(notes) > 0 {
					n := strings.Join(notes, "; ")
					lines = append(lines, strings.ToUpper(n[:1])+n[1:]+".")
				}
				setLines(combLines, lines...)
			}

			// Suggestion: an option to listen to, not a recommendation
			switch {
			case aligned:
				widgets.SetStyledText(suggestTitle, i18n.T("sum.preview_title"), "")
				widgets.SetStyledText(suggestText, i18n.T("sum.preview", ms(delay)), "")
			case net == 0:
				suggestCard.Hide()
			case net > 0:
				widgets.SetStyledText(suggestTitle, i18n.T("sum.lineup_title"), "")
				total := arrival
				widgets.SetStyledText(suggestText, i18n.T("sum.lineup", ms(total), samples(total)), "")
			default:
				widgets.SetStyledText(suggestTitle, i18n.T("sum.lineup_title"), "")
				widgets.SetStyledText(suggestText, i18n.T("sum.lineup_b", ms(net), samples(net)), "")
			}
		}
		renderResult()
		saveSession()
	}

	saveSession = func() {
		if restoring {
			return
		}
		s := twoMicSession{Lo: loEntry.Text, Hi: hiEntry.Text, Freq: freqEntry.Text, TempC: temp.Base, TempUnit: tempUnit.Selected,
			DistUnit: distUnit.Selected, SampleRate: sampleRate.Selected, Farther: farther.Text, FartherUnit: fartherUnit.Selected,
			Level: level.Text, Inverted: polarity.SelectedIndex() == 1, DelayA: delayA.Text, Open: setup.IsOpen()}
		switch basis {
		case farDistance:
			s.Basis, s.Meters = string(farDistance), &basisMeters
		case farTime:
			s.Basis, s.Seconds = string(farTime), &basisSeconds
		}
		if data, err := json.Marshal(s); err == nil {
			fyne.CurrentApp().Preferences().SetString(prefTwoMicSession, string(data))
		}
	}

	// Restore the last setup; the first start opens the card with a kick drum at 20 °C
	s := twoMicSession{Lo: "50", Hi: "100", Freq: "60", TempC: 20, TempUnit: logic.UnitCelsius, DistUnit: logic.UnitMeters,
		SampleRate: "48000", FartherUnit: logic.UnitMeters, Level: "0", DelayA: "0", Open: true}
	if data := fyne.CurrentApp().Preferences().String(prefTwoMicSession); data != "" {
		_ = json.Unmarshal([]byte(data), &s)
	}
	restoring = true
	if s.TempUnit != "" {
		tempUnit.SetSelected(s.TempUnit)
	}
	temp.setBase(s.TempC)
	if s.DistUnit != "" {
		distUnit.SetSelected(s.DistUnit)
	}
	if s.FartherUnit != "" {
		fartherUnit.Selected = s.FartherUnit // the typed value is in this unit: no conversion
		fartherUnit.Refresh()
	}
	if s.SampleRate != "" {
		sampleRate.SetSelected(s.SampleRate)
	}
	loEntry.SetText(s.Lo)
	hiEntry.SetText(s.Hi)
	freqEntry.SetText(s.Freq)
	matchPreset()
	farther.SetText(s.Farther) // sets the basis from the text, as typed
	restoreFartherBasis(s, &basis, &basisMeters, &basisSeconds)
	if s.Basis != "" && (basis == farDistance || basis == farTime) {
		showFarther() // the stored full-precision value wins: the entry must not contradict it
	}
	level.SetText(s.Level)
	if s.Inverted {
		polarity.SetSelected(i18n.T("polarity.inverted"))
	}
	delayA.SetText(s.DelayA)
	setup.SetOpen(s.Open)
	restoring = false
	refresh()

	content = container.NewBorder(
		container.New(layout.NewCustomPaddedLayout(2, 2, 2, 2), setup.Card),
		nil, nil, nil,
		results,
	)
	texts.add(func() {
		for i, g := range groups {
			for j, option := range g.Options {
				keys := strings.SplitN(option.Value, "|", 2)
				groups[i].Name = i18n.T("source." + keys[0])
				groups[i].Options[j].Label = i18n.T("source." + keys[1])
			}
		}
		presetSelect.Placeholder = i18n.T("field.custom")
		presetSelect.SetGroups(groups)
	})
	texts.add(func() {
		if renderResult != nil {
			renderResult()
		}
	})

	return texts.view(content)
}

// farBasis is what the "Farther by" value was entered as
type farBasis string

const (
	farEmpty    farBasis = ""
	farInvalid  farBasis = "invalid"
	farDistance farBasis = "distance"
	farTime     farBasis = "time"
)

// restoreFartherBasis replaces the basis read from the saved text by the saved full-precision
// value. A session without that value (from an older version) keeps the basis from the text. A
// partial or unusable value leaves the entry invalid rather than guessing.
func restoreFartherBasis(s twoMicSession, basis *farBasis, meters, seconds *float64) {
	finite := func(v *float64) bool { return v != nil && !math.IsNaN(*v) && !math.IsInf(*v, 0) }
	switch {
	case s.Basis == "" && s.Meters == nil && s.Seconds == nil:
	case s.Basis == string(farDistance) && finite(s.Meters) && s.Seconds == nil:
		*basis, *meters = farDistance, *s.Meters
	case s.Basis == string(farTime) && finite(s.Seconds) && s.Meters == nil:
		*basis, *seconds = farTime, *s.Seconds
	default:
		*basis = farInvalid
	}
}
