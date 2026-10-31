package ui

import (
	"fmt"
	"math"

	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewThreeToOneTab creates the 3:1 rule calculator: is another source far enough from a mic,
// compared with the mic's own source, to keep its bleed low?
func NewThreeToOneTab() fyne.CanvasObject {
	var update func()
	refresh := func() {
		if update != nil {
			update()
		}
	}

	own := newDistanceInput("Distance", logic.UnitMeters, refresh)
	other := newDistanceInput("Distance", logic.UnitMeters, refresh)
	ownSuffix, otherSuffix := widgets.NewMutedText(logic.UnitMeters), widgets.NewMutedText(logic.UnitMeters)
	distUnit := newUnitSelect([]string{logic.UnitMeters, logic.UnitFeet}, []*unitInput{own, other}, func(unit string) {
		widgets.SetStyledText(ownSuffix, unit, "")
		widgets.SetStyledText(otherSuffix, unit, "")
		refresh()
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

	// Readout: the ratio in large type
	ratioValue := widgets.NewStyledText("–", theme.ColorNamePlaceHolder, widgets.SizeNameDisplayText, fyne.TextStyle{})
	ratioUnit := widgets.NewStyledText(": 1", theme.ColorNamePlaceHolder, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true})
	caption := widgets.NewMutedText("")
	readout := container.NewVBox(
		container.NewCenter(widgets.NewBaselineRow(textGap, ratioValue, ratioUnit)),
		container.NewCenter(caption),
	)

	// Status: meets the rule or not
	statusIcon := widget.NewIcon(nil)
	statusTitle := widgets.NewStyledText("", theme.ColorNameSuccess, theme.SizeNameText, fyne.TextStyle{Bold: true})
	statusText := wrapped(widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{}))
	statusCard := widgets.NewCard(vbox(container.NewHBox(statusIcon, statusTitle), statusText))

	// How much margin there is, or what to change
	adviceTitle := widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{Bold: true})
	advice1 := wrapped(widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{}))
	advice2 := wrapped(widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{}))
	adviceCard := widgets.NewCard(vbox(adviceTitle, advice1, advice2))

	infoCard := widgets.NewCard(wrapped(widgets.NewMutedText(
		"3:1 rule: keep other sources at least three times farther from a mic than its own source. " +
			"The bleed is then at least 9.5 dB lower, which keeps comb filtering shallow " +
			"(about +2.5 / −3.5 dB) when the mics are mixed.")))

	results, relayout := newResultsScroll(readout, statusCard, adviceCard, infoCard)
	update = func() {
		defer relayout()
		unit := distUnit.Selected
		dist := func(meters float64) string {
			return logic.FormatTrim(logic.FromMeters(meters, unit), 2) + " " + unit
		}
		r := logic.ThreeToOne(own.Base, other.Base)
		if !r.Valid {
			widgets.SetStyledText(ratioValue, "–", theme.ColorNamePlaceHolder)
			widgets.SetStyledText(caption, "Enter both distances to check the 3:1 rule", "")
			statusCard.Hide()
			adviceCard.Hide()
			return
		}
		statusCard.Show()
		adviceCard.Show()

		color := theme.ColorNameSuccess
		if !r.Meets {
			color = theme.ColorNameError
		}
		widgets.SetStyledText(ratioValue, fmt.Sprintf("%.1f", r.Ratio), color)
		widgets.SetStyledText(caption, "distance ratio, other source : own source", "")

		statusCard.SetOutline(color)
		if r.Meets {
			statusIcon.SetResource(theme.NewSuccessThemedResource(theme.ConfirmIcon()))
			widgets.SetStyledText(statusTitle, "Meets the 3:1 rule", color)
		} else {
			statusIcon.SetResource(theme.NewErrorThemedResource(theme.CancelIcon()))
			widgets.SetStyledText(statusTitle, "Below 3:1", color)
		}
		relation := "below"
		if r.BleedDB < 0 {
			relation = "above" // the other source is closer than the mic's own source
		}
		widgets.SetStyledText(statusText, fmt.Sprintf("The other source reaches this mic %.1f dB %s its own source.",
			math.Abs(r.BleedDB), relation), "")

		if r.Meets {
			widgets.SetStyledText(adviceTitle, "Margin", "")
			widgets.SetStyledText(advice1, "The other source may come as close as "+dist(r.MinOther)+".", "")
			widgets.SetStyledText(advice2, "This mic may move up to "+dist(r.MaxOwn)+" from its source.", "")
		} else {
			widgets.SetStyledText(adviceTitle, "To fix it", "")
			widgets.SetStyledText(advice1, "Move the other source to at least "+dist(r.MinOther)+",", "")
			widgets.SetStyledText(advice2, "or the mic to within "+dist(r.MaxOwn)+" of its source.", "")
		}
	}
	refresh()

	return container.NewBorder(
		container.NewVBox(
			labeledRow("Mic to own source", withSuffix(own.Entry, ownSuffix)),
			labeledRow("Mic to other source", withSuffix(other.Entry, otherSuffix)),
			labeledRow("Units", distUnit),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		results,
	)
}
