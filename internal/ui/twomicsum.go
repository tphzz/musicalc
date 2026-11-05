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
	Level        string
	Inverted     bool
	DelayA       string
	Open         bool
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
	converting := false
	farther.OnChanged = func(string) {
		if !converting {
			refresh()
		}
	}
	// convertFarther rewrites the "Farther by" value from one unit to another; a distance and a time
	// convert through the speed of sound
	convertFarther := func(from, to string) {
		if from == to || strings.TrimSpace(farther.Text) == "" {
			return
		}
		v := logic.ParseFloat(farther.Text)
		var out float64
		switch {
		case from != unitMS && to != unitMS:
			out = logic.ConvertDistance(v, from, to)
		case to == unitMS:
			out = logic.ArrivalDelay(logic.ToMeters(v, from), temp.Base) * 1000
		default:
			out = logic.FromMeters(v/1000*logic.SpeedOfSound(temp.Base), to)
		}
		converting = true
		farther.SetText(logic.FormatTrim(out, 2))
		converting = false
	}
	lastFartherUnit := fartherUnit.Selected
	fartherUnit.OnChanged = func(unit string) {
		convertFarther(lastFartherUnit, unit)
		lastFartherUnit = unit
		refresh()
	}
	// The distance unit for the whole tab; "Farther by" offers it and ms
	distUnit := widget.NewSelect([]string{logic.UnitMeters, logic.UnitFeet}, nil)
	distUnit.Selected = logic.UnitMeters
	distUnit.OnChanged = func(unit string) {
		fartherUnit.Options = []string{unit, unitMS}
		if fartherUnit.Selected != unitMS {
			fartherUnit.SetSelected(unit) // converts the value
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
		lo, hi := logic.ParseFloat(loEntry.Text), logic.ParseFloat(hiEntry.Text)
		cursor := logic.ParseFloat(freqEntry.Text)
		rate := logic.ParseSampleRate(sampleRate.Selected)
		hasArrival := strings.TrimSpace(farther.Text) != ""
		var fv, arrival, delay, net, levelDB, gLo, gHi, cursorDB, lowDB, highDB float64
		var m logic.TwoMicSum
		var dipFreqs, peakFreqs []float64
		if hasArrival {
			fv = logic.ParseFloat(farther.Text)
			arrival = fv / 1000
			if fartherUnit.Selected != unitMS {
				arrival = logic.ArrivalDelay(logic.ToMeters(fv, fartherUnit.Selected), temp.Base)
			}
			delay = logic.ParseFloat(delayA.Text) / 1000
			net = arrival - delay
			// Delays are shown to 0.01 ms: within half of that, Mic A and B count as lined up
			if math.Abs(net) < 0.000005 {
				net = 0
			}
			// Response
			levelDB = logic.ParseFloat(level.Text)
			m = logic.TwoMicSum{Delay: net, Gain: logic.DBToGain(levelDB), Inverted: polarity.SelectedIndex() == 1}
			gLo, gHi = logic.GraphRange(lo, hi)
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
			graph.SetData(gLo, gHi, curves, lo, hi, cursor, m.LevelDB(cursor))

			cursorDB = m.LevelDB(cursor)
			if lo > 0 && hi > lo {
				lowDB, highDB = m.RangeDB(lo, hi)
			}
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
			if lo > 0 && hi > lo {
				rangeText = logic.FormatTrim(lo, 2) + "–" + logic.FormatTrim(hi, 2) + " Hz"
			}
			tempText := "–"
			if !temp.Empty() {
				tempText = minusSign(logic.FormatTrim(logic.FromCelsius(temp.Base, tempUnit.Selected), 1)) + " " + tempUnit.Selected
			}
			setup.SetSummary(rangeText + " · " + tempText + " · " + logic.FormatTrim(float64(rate)/1000, 1) + " kHz")

			// Arrival difference
			if !hasArrival {
				widgets.SetStyledText(arrivalTitle, "–", "")
				widgets.SetStyledText(arrivalText, i18n.T("sum.enter_distance"), "")
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
			aligned := delay != 0 && net == 0
			switch {
			case aligned:
				widgets.SetStyledText(arrivalTitle, i18n.T("sum.aligned"), theme.ColorNameSuccess)
				widgets.SetStyledText(arrivalText, i18n.T("sum.aligned_detail", ms(arrival), ms(delay)), "")
			case delay != 0:
				widgets.SetStyledText(arrivalTitle, i18n.T("sum.net", ms(net), laterOrEarlier(net)), theme.ColorNameForeground)
				widgets.SetStyledText(arrivalText, i18n.T("sum.delayed_detail", ms(arrival), laterOrEarlier(arrival), ms(delay)), "")
			case net == 0:
				widgets.SetStyledText(arrivalTitle, i18n.T("sum.together"), theme.ColorNameForeground)
				widgets.SetStyledText(arrivalText, i18n.T("sum.no_difference"), "")
			default:
				widgets.SetStyledText(arrivalTitle, i18n.T("sum.arrival", ms(arrival), laterOrEarlier(arrival)), theme.ColorNameForeground)
				detail := i18n.T("sum.sample_detail", samples(arrival), logic.FormatTrim(float64(rate)/1000, 1))
				if fartherUnit.Selected != unitMS {
					detail += i18n.T("sum.distance_detail", logic.FormatTrim(math.Abs(fv), 2), fartherUnit.Selected, map[bool]string{true: i18n.T("sum.farther_word"), false: i18n.T("sum.closer_word")}[fv >= 0], tempText)
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
				if arrival > 0 {
					alignButton.Enable()
				} else {
					alignButton.Disable() // Mic B is first: the delay belongs on Mic B
				}
			}

			reading := ""
			if cursor > 0 {
				if cursorDB <= -60 {
					reading = i18n.T("sum.cursor_cancel", logic.FormatTrim(cursor, 2))
				} else {
					reading = i18n.T("sum.cursor", logic.FormatTrim(cursor, 2), signedDB(cursorDB))
				}
			}
			if lo > 0 && hi > lo {
				band := rangeText + ": "
				switch {
				case highDB-lowDB < 0.05:
					band += i18n.T("sum.band_constant", signedDB(highDB))
				case lowDB <= -60:
					// The model's floor stands for complete cancellation, not a level to read
					band += i18n.T("sum.band_cancel", signedDB(highDB))
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
				if dipDB <= -60 {
					setLines(combLines, i18n.T("sum.cancel_comb"))
				} else {
					setLines(combLines, i18n.T("sum.reduced_comb", -dipDB))
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
		fartherUnit.Selected = s.FartherUnit // the value is stored in this unit: no conversion
		lastFartherUnit = s.FartherUnit
		fartherUnit.Refresh()
	}
	if s.SampleRate != "" {
		sampleRate.SetSelected(s.SampleRate)
	}
	loEntry.SetText(s.Lo)
	hiEntry.SetText(s.Hi)
	freqEntry.SetText(s.Freq)
	matchPreset()
	farther.SetText(s.Farther)
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
