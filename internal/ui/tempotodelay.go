package ui

import (
	"fmt"
	"musicalc/internal/i18n"
	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// delayTempo is the tempo of the Tempo to Delay tab. It is shared so the Tap Tempo tab can send
// its measurement there, even before Tempo to Delay has been opened. It is bound to a variable
// with the default value, because setting a binding needs the running app.
var (
	delayTempoValue = "120"
	delayTempo      = binding.BindString(&delayTempoValue)
)

func NewTempoTab() *View {
	texts := newTextRefresh()
	bpm := delayTempo
	input := widgets.NewNumericEntry()
	input.Bind(bpm)
	input.Validator = nil // Bind adds a conversion validator, which would show a check mark
	texts.placeholder(input, "field.tempo")

	// Tap pad: each tap writes the measured tempo into the field, so the cards follow live
	isMobile := fyne.CurrentDevice().IsMobile()
	idleKey, padStyle, titleKey := "tap.inline_hint", widgets.TapPadInline, "tap.inline"
	if isMobile {
		idleKey, padStyle, titleKey = "tap.compact_hint", widgets.TapPadCompact, "tap.compact"
	}
	var tapper logic.TapTempo
	var lastTap logic.TapResult
	var pad *widgets.TapPad
	pad = widgets.NewTapPad(padStyle, i18n.T(titleKey), i18n.T(idleKey), func() {
		r := tapper.Tap(time.Now())
		lastTap = r
		if r.Valid {
			_ = bpm.Set(fmt.Sprintf("%.1f", r.BPM))
			pad.SetStatus(i18n.T("tap.delay_status", r.Taps, r.BPM))
		} else {
			pad.SetStatus(i18n.T("tap.again"))
		}
		pad.SetActive(true)
		whenTapsStop(&tapper, func() {
			pad.SetActive(false)
			pad.SetStatus(i18n.T(idleKey))
		})
	})

	// One card per note value, with its straight, dotted and triplet lengths side by side
	groups := []struct {
		Name     string
		Subtitle string
		Mult     float64
	}{
		{"1/1", "note.whole", 4.0},
		{"1/2", "note.half", 2.0},
		{"1/4", "note.quarter", 1.0},
		{"1/8", "note.eighth", 0.5},
		{"1/16", "note.sixteenth", 0.25},
		{"1/32", "note.thirtysecond", 0.125},
		{"1/64", "note.sixtyfourth", 0.0625},
	}
	variants := []struct {
		Name   string
		Factor float64
	}{
		{"note.straight", 1.0},
		{"note.dotted", 1.5},
		{"note.triplet", 2.0 / 3.0},
	}

	// Variant labels share the width of the widest one so the delay values line up
	var labelWidth float32
	for _, v := range variants {
		labelWidth = fyne.Max(labelWidth, widgets.NewMutedText(i18n.T(v.Name)).MinSize().Width)
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
	var variantLabels []*widget.RichText
	var variantWraps []*fyne.Container
	for gi, g := range groups {
		header := widgets.NewBaselineRow(textGap,
			widgets.NewStyledText(g.Name, theme.ColorNameForeground, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true}),
			texts.muted(g.Subtitle),
		)
		content := container.New(layout.NewCustomPaddedVBoxLayout(rowGap), header)
		for _, v := range variants {
			name := texts.muted(v.Name)
			wrap := container.NewGridWrap(fyne.NewSize(labelWidth, name.MinSize().Height), name)
			variantLabels = append(variantLabels, name)
			variantWraps = append(variantWraps, wrap)
			row := valueRow{
				ms: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Monospace: true}),
				hz: widgets.NewMutedText(""),
			}
			rows[gi] = append(rows[gi], row)
			content.Add(container.NewBorder(nil, nil,
				wrap,
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
				if !res.Valid {
					// No tempo yet, or one that is not usable: no values rather than a default tempo
					rows[gi][vi].ms.SetText("–")
					widgets.SetStyledText(rows[gi][vi].hz, "–", "")
					continue
				}
				rows[gi][vi].ms.SetText(fmt.Sprintf("%.2f ms", res.DelayMS))
				widgets.SetStyledText(rows[gi][vi].hz, fmt.Sprintf("%.2f Hz", res.ModHz), "")
			}
		}
	}))

	content := container.NewBorder(
		container.NewVBox(
			input,
			pad,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		// A little inset keeps the card shadows from being clipped by the scroller
		container.NewVScroll(container.New(layout.NewCustomPaddedLayout(2, 4, 2, 2), cards)),
	)
	// Desktop: T taps while the tempo field isn't focused
	setTabKeyHandler(content, func(ev *fyne.KeyEvent) bool {
		if ev.Name == fyne.KeyT {
			pad.OnTap()
			return true
		}
		return false
	})
	texts.add(func() {
		pad.SetTitle(i18n.T(titleKey))
		switch {
		case !pad.Active():
			pad.SetStatus(i18n.T(idleKey))
		case lastTap.Valid:
			pad.SetStatus(i18n.T("tap.delay_status", lastTap.Taps, lastTap.BPM))
		default:
			pad.SetStatus(i18n.T("tap.again"))
		}
		var width float32
		for _, label := range variantLabels {
			width = fyne.Max(width, label.MinSize().Width)
		}
		for i, wrap := range variantWraps {
			wrap.Layout = layout.NewGridWrapLayout(fyne.NewSize(width, variantLabels[i].MinSize().Height))
			wrap.Refresh()
		}
	})

	return texts.view(content)
}
