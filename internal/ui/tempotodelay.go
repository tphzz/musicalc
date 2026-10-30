package ui

import (
	"fmt"
	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewTempoTab() fyne.CanvasObject {
	bpm := binding.NewString()
	_ = bpm.Set("120")
	input := widgets.NewNumericEntry()
	input.SetText("120")
	input.PlaceHolder = "Tempo"
	input.OnChanged = func(s string) {
		_ = bpm.Set(s)
	}

	// One card per note value, with its straight, dotted and triplet lengths side by side
	groups := []struct {
		Name     string
		Subtitle string
		Mult     float64
	}{
		{"1/1", "Whole note", 4.0},
		{"1/2", "Half note", 2.0},
		{"1/4", "Quarter note", 1.0},
		{"1/8", "Eighth note", 0.5},
		{"1/16", "16th note", 0.25},
		{"1/32", "32nd note", 0.125},
		{"1/64", "64th note", 0.0625},
	}
	variants := []struct {
		Name   string
		Factor float64
	}{
		{"Straight", 1.0},
		{"Dotted", 1.5},
		{"Triplet", 2.0 / 3.0},
	}

	// Variant labels share the width of the widest one so the delay values line up
	var labelWidth float32
	for _, v := range variants {
		labelWidth = fyne.Max(labelWidth, widgets.NewMutedText(v.Name).MinSize().Width)
	}

	// Labels draw their own inner padding; overlapping it keeps the value rows at line spacing
	rowGap := theme.Size(theme.SizeNameLineSpacing) - 2*theme.Size(theme.SizeNameInnerPadding)
	// ...and keeps texts on one line a padding apart
	textGap := theme.Size(theme.SizeNamePadding) - 2*theme.Size(theme.SizeNameInnerPadding)

	type valueRow struct {
		ms *widget.Label
		hz *widget.RichText
	}
	rows := make([][]valueRow, len(groups))
	cards := container.NewVBox()
	for gi, g := range groups {
		// The subtitle sits at the bottom so it shares the larger title's baseline
		header := container.New(layout.NewCustomPaddedHBoxLayout(textGap),
			widgets.NewStyledText(g.Name, theme.ColorNameForeground, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true}),
			container.NewVBox(layout.NewSpacer(), widgets.NewMutedText(g.Subtitle)),
		)
		content := container.New(layout.NewCustomPaddedVBoxLayout(rowGap), header)
		for _, v := range variants {
			name := widgets.NewMutedText(v.Name)
			row := valueRow{
				ms: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true}),
				hz: widgets.NewMutedText(""),
			}
			rows[gi] = append(rows[gi], row)
			content.Add(container.NewBorder(nil, nil,
				container.NewGridWrap(fyne.NewSize(labelWidth, name.MinSize().Height), name),
				row.hz,
				row.ms,
			))
		}
		cards.Add(widgets.NewCard(content))
	}

	bpm.AddListener(binding.NewDataListener(func() {
		val, _ := bpm.Get()
		tempo := logic.ParseFloat(val)
		for gi, g := range groups {
			for vi, v := range variants {
				res := logic.GetTempoData(tempo, g.Mult*v.Factor)
				rows[gi][vi].ms.SetText(fmt.Sprintf("%.2f ms", res.DelayMS))
				widgets.SetStyledText(rows[gi][vi].hz, fmt.Sprintf("%.2f Hz", res.ModHz), "")
			}
		}
	}))

	return container.NewBorder(
		container.NewVBox(
			input,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		// A little inset keeps the card shadows from being clipped by the scroller
		container.NewVScroll(container.New(layout.NewCustomPaddedLayout(2, 4, 2, 2), cards)),
	)
}
