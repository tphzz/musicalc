package ui

import (
	"fmt"
	"musicalc/internal/i18n"
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
func NewTapTempoTab(showDelay func()) *View {
	texts := newTextRefresh()
	isMobile := fyne.CurrentDevice().IsMobile()
	idleKey, tappingKey := "tap.idle_mobile", "tap.pause_mobile"
	if !isMobile {
		idleKey, tappingKey = "tap.idle_desktop", "tap.pause_desktop"
	}

	var tapper logic.TapTempo
	var last logic.TapResult
	round := false

	// Readout: large value with a BPM unit next to it, and a status line below
	value := widgets.NewStyledText(noTempo, theme.ColorNamePlaceHolder, widgets.SizeNameDisplayText, fyne.TextStyle{})
	unit := widgets.NewStyledText("BPM", theme.ColorNamePlaceHolder, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true})
	caption := texts.muted("tap.measure")
	caption.Wrapping = fyne.TextWrapWord
	caption.Segments[0].(*widget.TextSegment).Style.Alignment = fyne.TextAlignCenter
	textGap := theme.Size(theme.SizeNamePadding) - 2*theme.Size(theme.SizeNameInnerPadding)
	readout := container.NewVBox(
		container.NewCenter(widgets.NewBaselineRow(textGap, value, unit)),
		caption,
	)

	formatTempo := func(bpm float64) string {
		if round {
			return fmt.Sprintf("%.0f", bpm)
		}
		return fmt.Sprintf("%.2f", bpm)
	}

	var delayBtn *widgets.Button
	renderResult := func() {
		switch {
		case last.Valid:
			widgets.SetStyledText(value, formatTempo(last.BPM), theme.ColorNamePrimary)
			widgets.SetStyledText(caption, i18n.T("tap.status",
				last.Taps, last.Spread, 60000/last.BPM), "")
		case last.Taps == 1:
			widgets.SetStyledText(value, noTempo, theme.ColorNamePlaceHolder)
			widgets.SetStyledText(caption, i18n.T("tap.again"), "")
		default:
			widgets.SetStyledText(value, noTempo, theme.ColorNamePlaceHolder)
			widgets.SetStyledText(caption, i18n.T("tap.measure"), "")
		}
	}

	showResult := func() {
		renderResult()
		if last.Valid {
			delayBtn.Enable()
		} else if last.Taps != 1 {
			delayBtn.Disable()
		}
	}

	pad := widgets.NewTapPad(widgets.TapPadLarge, i18n.T("tap.title"), i18n.T(idleKey), nil)
	tap := func() {
		last = tapper.Tap(time.Now())
		showResult()
		pad.SetActive(true)
		pad.SetTitle(i18n.T("tap.keep_tapping"))
		pad.SetStatus(i18n.T(tappingKey))
		pad.Flash()
		whenTapsStop(&tapper, func() {
			pad.SetActive(false)
			pad.SetTitle(i18n.T("tap.title"))
			pad.SetStatus(i18n.T(idleKey))
		})
	}
	pad.OnTap = tap

	reset := func() {
		tapper.Reset()
		last = logic.TapResult{}
		showResult()
		pad.SetActive(false)
		pad.SetTitle(i18n.T("tap.title"))
		pad.SetStatus(i18n.T(idleKey))
	}

	var roundBtn *widgets.Button
	roundBtn = texts.button("action.round", func() {
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
	delayBtn = texts.buttonWithIcon("tab.delay", theme.ContentRedoIcon(), func() {
		if !last.Valid {
			return
		}
		_ = delayTempo.Set(formatTempo(last.BPM))
		if showDelay != nil {
			showDelay()
		}
	})
	resetBtn := texts.buttonWithIcon("action.reset", theme.ViewRefreshIcon(), reset)
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
	texts.add(func() {
		renderResult()
		if pad.Active() {
			pad.SetTitle(i18n.T("tap.keep_tapping"))
			pad.SetStatus(i18n.T(tappingKey))
		} else {
			pad.SetTitle(i18n.T("tap.title"))
			pad.SetStatus(i18n.T(idleKey))
		}
	})
	return texts.view(content)
}
