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
func NewThreeToOneTab() *View {
	texts := newTextRefresh()
	var update func()
	restoring := false
	refresh := func() {
		if update != nil && !restoring {
			update()
		}
	}

	// Setup: the source, its distance to both mics, the neighbouring channel's extra level and the target
	source := widget.NewEntry()
	texts.placeholder(source, "field.source")
	source.OnChanged = func(string) { refresh() }
	own := newDistanceInput(i18n.T("field.distance"), logic.UnitMeters, refresh)
	texts.placeholder(own.Entry, "field.distance")
	neighbour := newDistanceInput(i18n.T("field.distance"), logic.UnitMeters, refresh)
	texts.placeholder(neighbour.Entry, "field.distance")
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
	for i, key := range []string{"target.standard", "target.ten", "target.fifteen"} {
		targetNames[i] = i18n.T(key)
	}
	target := widget.NewSelect(targetNames, func(string) { refresh() })
	target.Selected = targetNames[0]
	texts.selectOptions(target, []string{"target.standard", "target.ten", "target.fifteen"})

	var content *fyne.Container
	var saveSession func()
	setup := newSetupCard(container.NewVBox(
		texts.row("field.source", source),
		texts.row("three.own", withSuffix(own.Entry, ownSuffix)),
		texts.row("three.neighbour", withSuffix(neighbour.Entry, neighbourSuffix)),
		texts.row("three.gain", withSuffix(gain, widgets.NewMutedText("dB"))),
		texts.row("three.target", target),
		texts.row("field.units", distUnit),
		texts.wrappedMuted("three.hint"),
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
	widgets.SetStyledText(spillTitle, i18n.T("three.spill"), "")
	spillText := plain()
	spillCard := widgets.NewCard(vbox(
		container.NewBorder(nil, nil, nil, spillValue, spillTitle),
		spillText,
		wrapped(texts.muted("three.model")),
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

	var renderResult func()
	update = func() {
		unit := distUnit.Selected
		name := strings.TrimSpace(source.Text)
		gainDB, gainOK := logic.ParseNumber(gain.Text)
		if strings.TrimSpace(gain.Text) == "" {
			gainDB, gainOK = 0, true
		}
		// usable: empty (still missing) or a positive finite distance
		usable := func(in *unitInput) bool {
			v, ok := logic.ParseNumber(in.Entry.Text)
			return in.Empty() || (ok && v > 0)
		}
		// Something was typed that cannot be used, as opposed to a value still missing
		inputInvalid := !gainOK || !usable(own) || !usable(neighbour)
		t := logic.SpillTargets[max(target.SelectedIndex(), 0)]
		r := logic.SpillCheck(own.Base, neighbour.Base, gainDB, t.DB)
		if inputInvalid {
			r = logic.SpillResult{}
		}
		renderResult = func() {
			defer relayout()
			// atLeast shows a minimum distance rounded up at the shown precision, so following
			// the advice meets the target
			atLeast := func(meters float64) string {
				return logic.FormatTrim(logic.CeilTo(logic.FromMeters(meters, unit), 2), 2) + " " + unit
			}
			srcName, srcSentence := i18n.T("source.generic"), i18n.T("source.sentence")
			if name != "" {
				srcName, srcSentence = name, name
				setup.SetTitle(name)
			} else {
				setup.SetTitle(i18n.T("setup.title"))
			}
			targetShown := logic.FormatTrim(math.Round(t.DB*10)/10, 1)

			// Folded line: distances, gain, target
			value := func(in *unitInput) string {
				if in.Empty() {
					return "–"
				}
				return logic.FormatTrim(logic.FromMeters(in.Base, unit), 2)
			}
			setup.SetSummary(value(own) + " / " + value(neighbour) + " " + unit + " · " + signedDB(gainDB) + " dB · ≥ " + targetShown + " dB")

			if !r.Valid {
				widgets.SetStyledText(ratioValue, "–", theme.ColorNamePlaceHolder)
				if inputInvalid || (!own.Empty() && !neighbour.Empty()) {
					// Both distances are there, but a value cannot be used (or the result overflows)
					widgets.SetStyledText(caption, i18n.T("three.invalid"), "")
				} else {
					widgets.SetStyledText(caption, i18n.T("three.enter", srcName), "")
				}
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
				widgets.SetStyledText(geoTitle, i18n.T("three.meets"), geoColor)
				widgets.SetStyledText(geoText, i18n.T("three.meets_detail", times, srcName), "")
			case logic.GeometryThreshold:
				geoColor = theme.ColorNameWarning
				geoIcon.SetResource(theme.NewWarningThemedResource(theme.WarningIcon()))
				widgets.SetStyledText(geoTitle, i18n.T("three.threshold"), geoColor)
				widgets.SetStyledText(geoText, i18n.T("three.threshold_detail"), "")
			default:
				geoColor = theme.ColorNameError
				geoIcon.SetResource(theme.NewErrorThemedResource(theme.CancelIcon()))
				widgets.SetStyledText(geoTitle, i18n.T("three.below"), geoColor)
				widgets.SetStyledText(geoText, i18n.T("three.below_detail", times, srcName), "")
			}
			widgets.SetStyledText(ratioValue, fmt.Sprintf("%.1f", r.Ratio), geoColor)
			widgets.SetStyledText(caption, i18n.T("three.ratio"), "")

			// Spill against the target
			spillColor := theme.ColorNameSuccess
			if !r.MeetsTarget {
				spillColor = theme.ColorNameError
			}
			spillCard.SetOutline(spillColor)
			widgets.SetStyledText(spillTitle, i18n.T("three.spill"), spillColor)
			widgets.SetStyledText(spillValue, minusSign(fmt.Sprintf("%+.1f dB", r.SpillDB)), spillColor)
			level := i18n.T("three.level_below", dB(r.SpillDB))
			switch {
			case math.Abs(r.SpillDB) < 0.05:
				level = i18n.T("three.level_equal")
			case r.SpillDB > 0:
				level = i18n.T("three.level_above", dB(r.SpillDB))
			}
			if r.SpillDB <= -0.05 && !r.MeetsTarget {
				level = i18n.T("three.only", level) // below, but not far enough
			}
			var vsTarget string
			switch {
			case math.Abs(r.MarginDB) < 0.05:
				vsTarget = i18n.T("three.on_target", targetShown)
			case r.MarginDB > 0:
				vsTarget = i18n.T("three.better", dB(r.MarginDB), targetShown)
			default:
				vsTarget = i18n.T("three.short", dB(r.MarginDB), targetShown)
				if gainDB >= 0.05 {
					vsTarget += i18n.T("three.because_louder", dB(gainDB))
				}
			}
			widgets.SetStyledText(spillText, i18n.T("three.result", srcSentence, level, vsTarget), "")

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
			closest := atLeast(r.MinNeighbour) // the distances went in as metres
			if r.MeetsTarget {
				if r.MarginDB < 0.05 {
					widgets.SetStyledText(fixTitle, i18n.T("three.on_target_title"), "")
					setLines(i18n.T("three.no_margin", closest, srcName),
						i18n.T("three.next_source"))
				} else {
					widgets.SetStyledText(fixTitle, i18n.T("three.margin"), "")
					setLines(i18n.T("three.margin_detail", closest, srcName),
						i18n.T("three.next_source"))
				}
			} else {
				widgets.SetStyledText(fixTitle, i18n.T("three.reach", targetShown), "")
				turnDown := i18n.T("three.turn_down", dB(r.MarginDB))
				move := i18n.T("three.move", closest, srcName)
				aim := i18n.T("three.aim", srcName)
				if r.Geometry == logic.GeometryBelow {
					setLines(move, i18n.T("three.or_turn_down", dB(r.MarginDB)), aim)
				} else {
					setLines(turnDown, i18n.T("three.or_move", closest, srcName), aim)
				}
			}

			// Comb filtering, stated as the condition it depends on. Only equal levels can cancel
			// completely; a spill twice as loud or more never dips below the own mic (at exactly
			// twice, the minimum is 0 dB, so a cut that would show as −0.0 dB counts as none).
			var comb string
			switch {
			case r.CanCancel:
				comb = i18n.T("three.comb_cancel", dB(r.CombBoostDB))
			case r.CombCutDB < -0.05:
				comb = i18n.T("three.comb", dB(r.CombBoostDB), dB(r.CombCutDB))
			default:
				comb = i18n.T("three.comb_boost", dB(r.CombCutDB), dB(r.CombBoostDB))
			}
			widgets.SetStyledText(combText, comb, "")
		}
		renderResult()
		saveSession()
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
	texts.add(func() {
		if renderResult != nil {
			renderResult()
		}
	})
	return texts.view(content)
}
