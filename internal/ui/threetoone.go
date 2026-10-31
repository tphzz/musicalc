package ui

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// prefThreeToOneSession stores the 3:1 setup, so it is back when the app starts again
const prefThreeToOneSession = "threeToOne.session"

// threeToOneSession is the 3:1 setup on screen
type threeToOneSession struct {
	Source          string  `json:"source"`
	OwnMeters       float64 `json:"ownMeters"`
	NeighbourMeters float64 `json:"neighbourMeters"`
	HasOwn          bool    `json:"hasOwn"`
	HasNeighbour    bool    `json:"hasNeighbour"`
	GainDB          float64 `json:"gainDB"`
	Target          int     `json:"target"`
	Unit            string  `json:"unit"`
	Open            bool    `json:"open"`
}

// NewThreeToOneTab creates the 3:1 rule calculator: one source reaches its own mic and a
// neighbouring mic; how strongly does it come through the neighbouring one, and what can be changed?
func NewThreeToOneTab() fyne.CanvasObject {
	var update func()
	restoring := false
	refresh := func() {
		if update != nil && !restoring {
			update()
		}
	}

	// Setup: the source, its distance to both mics, the neighbouring channel's extra level and the target
	source := widget.NewEntry()
	source.SetPlaceHolder("Source")
	source.OnChanged = func(string) { refresh() }
	own := newDistanceInput("Distance", logic.UnitMeters, refresh)
	neighbour := newDistanceInput("Distance", logic.UnitMeters, refresh)
	ownSuffix, neighbourSuffix := widgets.NewMutedText(logic.UnitMeters), widgets.NewMutedText(logic.UnitMeters)
	distUnit := newUnitSelect([]string{logic.UnitMeters, logic.UnitFeet}, []*unitInput{own, neighbour}, func(unit string) {
		widgets.SetStyledText(ownSuffix, unit, "")
		widgets.SetStyledText(neighbourSuffix, unit, "")
		refresh()
	})
	gain := widgets.NewNumericEntry()
	gain.SetPlaceHolder("0")
	gain.OnChanged = func(string) { refresh() }
	targetNames := make([]string, len(logic.SpillTargets))
	for i, t := range logic.SpillTargets {
		targetNames[i] = t.Name
	}
	target := widget.NewSelect(targetNames, func(string) { refresh() })
	target.Selected = targetNames[0]

	var content *fyne.Container
	var saveSession func()
	setup := newSetupCard(container.NewVBox(
		labeledRow("Source", source),
		labeledRow("To its mic", withSuffix(own.Entry, ownSuffix)),
		labeledRow("To neighbour mic", withSuffix(neighbour.Entry, neighbourSuffix)),
		labeledRow("Neighbour gain", withSuffix(gain, widgets.NewMutedText("dB"))),
		labeledRow("Target", target),
		labeledRow("Units", distUnit),
		wrappedMuted("Check one source at a time: its distance to its own mic and to the neighbouring mic. "+
			"Neighbour gain is how much louder the neighbouring channel is (gain, sensitivity, pickup)."),
	), func(bool) {
		if content != nil {
			content.Refresh() // the card's height changed
		}
		if saveSession != nil {
			saveSession()
		}
	})

	textGap := theme.Size(theme.SizeNamePadding) - 2*theme.Size(theme.SizeNameInnerPadding)
	rowGap := theme.Size(theme.SizeNameLineSpacing) - 2*theme.Size(theme.SizeNameInnerPadding)
	vbox := func(objects ...fyne.CanvasObject) *fyne.Container {
		return container.New(layout.NewCustomPaddedVBoxLayout(rowGap), objects...)
	}
	wrapped := func(rt *widget.RichText) *widget.RichText {
		rt.Wrapping = fyne.TextWrapWord
		return rt
	}
	plain := func() *widget.RichText {
		return wrapped(widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{}))
	}
	bold := func(color fyne.ThemeColorName) *widget.RichText {
		return widgets.NewStyledText("", color, theme.SizeNameText, fyne.TextStyle{Bold: true})
	}

	// Readout: the distance ratio in large type
	ratioValue := widgets.NewStyledText("–", theme.ColorNamePlaceHolder, widgets.SizeNameDisplayText, fyne.TextStyle{})
	ratioUnit := widgets.NewStyledText(": 1", theme.ColorNamePlaceHolder, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true})
	caption := wrapped(widgets.NewMutedText(""))
	caption.Segments[0].(*widget.TextSegment).Style.Alignment = fyne.TextAlignCenter
	readout := container.NewVBox(
		container.NewCenter(widgets.NewBaselineRow(textGap, ratioValue, ratioUnit)),
		caption,
	)

	// Distance check: the geometry alone
	geoIcon := widget.NewIcon(nil)
	geoTitle := bold(theme.ColorNameSuccess)
	geoText := plain()
	geoCard := widgets.NewCard(vbox(container.NewHBox(geoIcon, geoTitle), geoText))

	// Estimated spill: the level, including the neighbouring channel's gain, against the target
	spillTitle, spillValue := bold(theme.ColorNameSuccess), bold(theme.ColorNameSuccess)
	widgets.SetStyledText(spillTitle, "Estimated spill", "")
	spillText := plain()
	spillCard := widgets.NewCard(vbox(
		container.NewBorder(nil, nil, nil, spillValue, spillTitle),
		spillText,
		wrapped(widgets.NewMutedText("Direct sound only, assuming matched mics and the level difference above.")),
	))

	// What to change, or how much room there is
	fixTitle := bold(theme.ColorNameForeground)
	fixLines := []*widget.RichText{plain(), plain(), plain()}
	fixCard := widgets.NewCard(vbox(fixTitle, fixLines[0], fixLines[1], fixLines[2]))

	combText := wrapped(widgets.NewMutedText(""))
	combCard := widgets.NewCard(combText)

	results, relayout := newResultsScroll(readout, geoCard, spillCard, fixCard, combCard)

	// dB shows a level difference in result texts, always with one decimal like the values
	dB := func(v float64) string { return fmt.Sprintf("%.1f", math.Abs(v)) }
	// signedDB is the short form for the folded line, e.g. "+6" or "−1.5"
	signedDB := func(v float64) string {
		short := logic.FormatTrim(math.Round(math.Abs(v)*10)/10, 1)
		if v < 0 {
			return "−" + short
		}
		return "+" + short
	}

	update = func() {
		defer relayout()
		defer saveSession()
		unit := distUnit.Selected
		dist := func(meters float64) string {
			return logic.FormatTrim(logic.FromMeters(meters, unit), 2) + " " + unit
		}
		name := strings.TrimSpace(source.Text)
		srcName, srcSentence := "the source", "The source"
		if name != "" {
			srcName, srcSentence = name, name
			setup.SetTitle(name)
		} else {
			setup.SetTitle("Setup")
		}
		gainDB := logic.ParseFloat(gain.Text)
		t := logic.SpillTargets[max(target.SelectedIndex(), 0)]
		targetShown := logic.FormatTrim(math.Round(t.DB*10)/10, 1)

		// Folded line: distances, gain, target
		value := func(in *unitInput) string {
			if in.Empty() {
				return "–"
			}
			return logic.FormatTrim(logic.FromMeters(in.Base, unit), 2)
		}
		setup.SetSummary(value(own) + " / " + value(neighbour) + " " + unit + " · " + signedDB(gainDB) + " dB · ≥ " + targetShown + " dB")

		r := logic.SpillCheck(own.Base, neighbour.Base, gainDB, t.DB)
		if !r.Valid {
			widgets.SetStyledText(ratioValue, "–", theme.ColorNamePlaceHolder)
			widgets.SetStyledText(caption, "Enter the distances from "+srcName+" to both mics", "")
			for _, c := range []fyne.CanvasObject{geoCard, spillCard, fixCard, combCard} {
				c.Hide()
			}
			return
		}
		for _, c := range []fyne.CanvasObject{geoCard, spillCard, fixCard, combCard} {
			c.Show()
		}

		// Ratio and distance check
		times := logic.FormatTrim(math.Round(r.Ratio*10)/10, 1)
		var geoColor fyne.ThemeColorName
		switch r.Geometry {
		case logic.GeometryMeets:
			geoColor = theme.ColorNameSuccess
			geoIcon.SetResource(theme.NewSuccessThemedResource(theme.ConfirmIcon()))
			widgets.SetStyledText(geoTitle, "Meets the 3:1 guideline", geoColor)
			widgets.SetStyledText(geoText, fmt.Sprintf("The neighbouring mic is %s times as far from %s as its own mic.", times, srcName), "")
		case logic.GeometryThreshold:
			geoColor = theme.ColorNameWarning
			geoIcon.SetResource(theme.NewWarningThemedResource(theme.WarningIcon()))
			widgets.SetStyledText(geoTitle, "At the 3:1 threshold", geoColor)
			widgets.SetStyledText(geoText, "Meets the guideline with no room to spare: any closer and it falls below 3:1.", "")
		default:
			geoColor = theme.ColorNameError
			geoIcon.SetResource(theme.NewErrorThemedResource(theme.CancelIcon()))
			widgets.SetStyledText(geoTitle, "Below 3:1", geoColor)
			widgets.SetStyledText(geoText, fmt.Sprintf("The neighbouring mic is only %s times as far from %s as its own mic.", times, srcName), "")
		}
		widgets.SetStyledText(ratioValue, fmt.Sprintf("%.1f", r.Ratio), geoColor)
		widgets.SetStyledText(caption, "distance ratio, neighbour mic : its mic", "")

		// Spill against the target
		spillColor := theme.ColorNameSuccess
		if !r.MeetsTarget {
			spillColor = theme.ColorNameError
		}
		spillCard.SetOutline(spillColor)
		widgets.SetStyledText(spillTitle, "Estimated spill", spillColor)
		widgets.SetStyledText(spillValue, minusSign(fmt.Sprintf("%+.1f dB", r.SpillDB)), spillColor)
		level := dB(r.SpillDB) + " dB below its own mic"
		switch {
		case math.Abs(r.SpillDB) < 0.05:
			level = "as loud as its own mic"
		case r.SpillDB > 0:
			level = dB(r.SpillDB) + " dB above its own mic"
		}
		if r.SpillDB <= -0.05 && !r.MeetsTarget {
			level = "only " + level // below, but not far enough
		}
		var vsTarget string
		switch {
		case math.Abs(r.MarginDB) < 0.05:
			vsTarget = "on the " + targetShown + " dB target"
		case r.MarginDB > 0:
			vsTarget = dB(r.MarginDB) + " dB better than the " + targetShown + " dB target"
		default:
			vsTarget = dB(r.MarginDB) + " dB short of the " + targetShown + " dB target"
			if gainDB >= 0.05 {
				vsTarget += ", because the neighbouring channel is " + dB(gainDB) + " dB louder"
			}
		}
		widgets.SetStyledText(spillText, fmt.Sprintf("%s reaches the neighbouring mic %s: %s.", srcSentence, level, vsTarget), "")

		// Fixes, or the room there is
		setLines := func(lines ...string) {
			for i, l := range fixLines {
				if i < len(lines) {
					widgets.SetStyledText(l, lines[i], "")
					l.Show()
				} else {
					l.Hide()
				}
			}
		}
		closest := dist(r.MinNeighbour) // the distances went in as metres
		if r.MeetsTarget {
			if r.MarginDB < 0.05 {
				widgets.SetStyledText(fixTitle, "On the target", "")
				setLines(fmt.Sprintf("No room to spare: keep the neighbouring mic at least %s from %s.", closest, srcName),
					"Now check each other source the same way.")
			} else {
				widgets.SetStyledText(fixTitle, "Room to spare", "")
				setLines(fmt.Sprintf("The neighbouring mic may come as close as %s to %s, with the other settings unchanged.", closest, srcName),
					"Now check each other source the same way.")
			}
		} else {
			widgets.SetStyledText(fixTitle, "To reach −"+targetShown+" dB", "")
			turnDown := "• Turn the neighbouring channel down by " + dB(r.MarginDB) + " dB. This also lowers its own source."
			move := fmt.Sprintf("• Move the neighbouring mic to at least %s from %s.", closest, srcName)
			aim := "• Or aim its rejection at " + srcName + ", then measure again."
			if r.Geometry == logic.GeometryBelow {
				setLines(move, "• Or t"+strings.TrimPrefix(turnDown, "• T"), aim)
			} else {
				setLines(turnDown, "• Or m"+strings.TrimPrefix(move, "• M"), aim)
			}
		}

		// Comb filtering, stated as the condition it depends on
		comb := "If both contributions were otherwise identical, mixing them could boost by up to +" + dB(r.CombBoostDB) + " dB and "
		if r.CanCancel {
			comb += "cancel completely at some frequencies."
		} else {
			comb += "cut by up to −" + dB(r.CombCutDB) + " dB, depending on frequency."
		}
		widgets.SetStyledText(combText, comb+" Listen to both mics summed in mono.", "")
	}

	saveSession = func() {
		if restoring {
			return
		}
		s := threeToOneSession{
			Source: strings.TrimSpace(source.Text), OwnMeters: own.Base, NeighbourMeters: neighbour.Base,
			HasOwn: !own.Empty(), HasNeighbour: !neighbour.Empty(), GainDB: logic.ParseFloat(gain.Text),
			Target: max(target.SelectedIndex(), 0), Unit: distUnit.Selected, Open: setup.IsOpen(),
		}
		if data, err := json.Marshal(s); err == nil {
			fyne.CurrentApp().Preferences().SetString(prefThreeToOneSession, string(data))
		}
	}

	// Restore the last setup; the first start opens the card with the defaults
	s := threeToOneSession{Open: true, Unit: logic.UnitMeters}
	if data := fyne.CurrentApp().Preferences().String(prefThreeToOneSession); data != "" {
		_ = json.Unmarshal([]byte(data), &s)
	}
	restoring = true
	if s.Unit != "" {
		distUnit.SetSelected(s.Unit)
	}
	if s.HasOwn {
		own.setBase(s.OwnMeters)
	}
	if s.HasNeighbour {
		neighbour.setBase(s.NeighbourMeters)
	}
	source.SetText(s.Source)
	gain.SetText(logic.FormatTrim(s.GainDB, 1))
	if s.Target >= 0 && s.Target < len(targetNames) {
		target.SetSelectedIndex(s.Target)
	}
	setup.SetOpen(s.Open)
	restoring = false
	refresh()

	content = container.NewBorder(
		container.New(layout.NewCustomPaddedLayout(2, 2, 2, 2), setup.Card),
		nil, nil, nil,
		results,
	)
	return content
}
