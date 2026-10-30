package ui

import (
	"fmt"
	"time"

	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// noTempo is shown instead of a value until two taps have been measured
const noTempo = "– – –"

// whenTapsStop runs onIdle once the measurement has ended, TapResetAfter after the last tap.
// Every tap schedules a check; only the one after the final tap finds the measurement inactive.
func whenTapsStop(t *logic.TapTempo, onIdle func()) {
	time.AfterFunc(logic.TapResetAfter+50*time.Millisecond, func() {
		fyne.Do(func() {
			if !t.Active(time.Now()) {
				onIdle()
			}
		})
	})
}

// NewTapTempoTab creates the Tap Tempo calculator: tap the large pad along with the beat to
// measure the tempo. showDelay switches to Tempo to Delay after the tempo has been sent there.
func NewTapTempoTab(showDelay func()) fyne.CanvasObject {
	isMobile := fyne.CurrentDevice().IsMobile()
	idleHint, tappingHint := "Tap along with the beat", "Pause 2 s to start over"
	if !isMobile {
		idleHint, tappingHint = "Click or press Space along with the beat", "Pause 2 s to start over · Esc resets"
	}
	const idleTitle, tappingTitle = "Tap here to find the tempo", "Keep tapping…"

	var tapper logic.TapTempo
	var last logic.TapResult
	round := false

	// Readout: large value with a BPM unit next to it, and a status line below
	value := widgets.NewStyledText(noTempo, theme.ColorNamePlaceHolder, widgets.SizeNameDisplayText, fyne.TextStyle{})
	unit := widgets.NewStyledText("BPM", theme.ColorNamePlaceHolder, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true})
	caption := widgets.NewMutedText("Tap at least twice to measure")
	textGap := theme.Size(theme.SizeNamePadding) - 2*theme.Size(theme.SizeNameInnerPadding)
	readout := container.NewVBox(
		container.NewCenter(widgets.NewBaselineRow(textGap, value, unit)),
		container.NewCenter(caption),
	)

	formatTempo := func(bpm float64) string {
		if round {
			return fmt.Sprintf("%.0f", bpm)
		}
		return fmt.Sprintf("%.2f", bpm)
	}

	var delayBtn *widgets.Button
	showResult := func() {
		switch {
		case last.Valid:
			widgets.SetStyledText(value, formatTempo(last.BPM), theme.ColorNamePrimary)
			widgets.SetStyledText(caption, fmt.Sprintf("%d taps · ±%.1f BPM · quarter = %.1f ms",
				last.Taps, last.Spread, 60000/last.BPM), "")
			delayBtn.Enable()
		case last.Taps == 1:
			widgets.SetStyledText(value, noTempo, theme.ColorNamePlaceHolder)
			widgets.SetStyledText(caption, "Tap again…", "")
		default:
			widgets.SetStyledText(value, noTempo, theme.ColorNamePlaceHolder)
			widgets.SetStyledText(caption, "Tap at least twice to measure", "")
			delayBtn.Disable()
		}
	}

	pad := widgets.NewTapPad(widgets.TapPadLarge, idleTitle, idleHint, nil)
	tap := func() {
		last = tapper.Tap(time.Now())
		showResult()
		pad.SetActive(true)
		pad.SetTitle(tappingTitle)
		pad.SetStatus(tappingHint)
		pad.Flash()
		whenTapsStop(&tapper, func() {
			pad.SetActive(false)
			pad.SetTitle(idleTitle)
			pad.SetStatus(idleHint)
		})
	}
	pad.OnTap = tap

	reset := func() {
		tapper.Reset()
		last = logic.TapResult{}
		showResult()
		pad.SetActive(false)
		pad.SetTitle(idleTitle)
		pad.SetStatus(idleHint)
	}

	var roundBtn *widgets.Button
	roundBtn = widgets.NewButton("Round", func() {
		round = !round
		// A filled button shows that rounding is on
		if round {
			roundBtn.Importance = widget.HighImportance
		} else {
			roundBtn.Importance = widget.MediumImportance
		}
		roundBtn.Refresh()
		showResult()
	})
	delayBtn = widgets.NewButtonWithIcon("Delay", theme.ContentRedoIcon(), func() {
		if !last.Valid {
			return
		}
		_ = delayTempo.Set(formatTempo(last.BPM))
		if showDelay != nil {
			showDelay()
		}
	})
	resetBtn := widgets.NewButtonWithIcon("Reset", theme.ViewRefreshIcon(), reset)
	showResult()

	content := container.NewBorder(
		container.NewVBox(readout, container.NewGridWithColumns(3, roundBtn, delayBtn, resetBtn)),
		nil, nil, nil,
		pad,
	)
	// Desktop: Space or T taps, Esc resets (only while no field has focus)
	setTabKeyHandler(content, func(ev *fyne.KeyEvent) bool {
		switch ev.Name {
		case fyne.KeySpace, fyne.KeyT:
			tap()
		case fyne.KeyEscape:
			reset()
		default:
			return false
		}
		return true
	})
	return content
}
