package ui

import (
	"fmt"
	"math"
	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// maxTransposeInput bounds typed semitones and cents before integer conversion
const maxTransposeInput = 1e6

func NewTempoChangeTab() *View {
	texts := newTextRefresh()
	// Tempo bindings
	originalTempo := binding.NewString()
	_ = originalTempo.Set("120")

	// Input/output fields
	originalTempoEntry := widgets.NewNumericEntry()
	originalTempoEntry.SetText("120")

	newTempoEntry := widgets.NewNumericEntry()
	newTempoEntry.SetText("100")

	timeStretchEntry := widgets.NewNumericEntry()
	texts.placeholder(timeStretchEntry, "field.stretch_placeholder")

	// Read-only output fields (using Labels)
	tempoDeltaLabel := widget.NewLabel("")

	// Editable transpose inputs (standard notation)
	semitonesEntry := widgets.NewNumericEntry()
	texts.placeholder(semitonesEntry, "field.semitones_input")
	centsEntry := widgets.NewNumericEntry()
	texts.placeholder(centsEntry, "field.cents_input")

	// Read-only 50-cent notation outputs
	semitones50Label := widget.NewLabel("")
	cents50Label := widget.NewLabel("")

	// Flag to prevent circular updates
	updating := false

	signed := func(v int) string {
		if v > 0 {
			return fmt.Sprintf("+%d", v)
		}
		return fmt.Sprintf("%d", v)
	}
	// clear empties derived fields, so invalid input does not leave earlier results looking current
	clear := func(fields ...interface{ SetText(string) }) {
		for _, f := range fields {
			f.SetText("")
		}
	}
	showDelta := func(res logic.TempoChangeResult) {
		sign := ""
		if res.TempoVariation > 0 {
			sign = "+"
		}
		tempoDeltaLabel.SetText(fmt.Sprintf("%s%.2f %%", sign, res.TempoVariation))
		semitones50Label.SetText(signed(res.Semitones50Cent))
		cents50Label.SetText(signed(res.Cents50Cent))
	}

	// Calculate from new tempo
	calcFromNewTempo := func() {
		if updating {
			return
		}
		updating = true
		defer func() { updating = false }()

		res := logic.CalculateFromNewTempo(logic.ParseFloat(originalTempoEntry.Text), logic.ParseFloat(newTempoEntry.Text))
		if !res.Valid {
			clear(timeStretchEntry, tempoDeltaLabel, semitonesEntry, centsEntry, semitones50Label, cents50Label)
			return
		}
		timeStretchEntry.SetText(logic.FormatTwoDecimals(res.TimeStretchPercent))
		semitonesEntry.SetText(signed(res.Semitones))
		centsEntry.SetText(signed(res.Cents))
		showDelta(res)
	}

	// Calculate from time stretch %
	calcFromTimeStretch := func() {
		if updating {
			return
		}
		updating = true
		defer func() { updating = false }()

		res := logic.CalculateFromTimeStretch(logic.ParseFloat(originalTempoEntry.Text), logic.ParseFloat(timeStretchEntry.Text))
		if !res.Valid {
			clear(newTempoEntry, tempoDeltaLabel, semitonesEntry, centsEntry, semitones50Label, cents50Label)
			return
		}
		newTempoEntry.SetText(logic.FormatTwoDecimals(res.NewTempo))
		semitonesEntry.SetText(signed(res.Semitones))
		centsEntry.SetText(signed(res.Cents))
		showDelta(res)
	}

	// Calculate from transpose (semitones/cents inputs)
	calcFromTranspose := func() {
		if updating {
			return
		}
		updating = true
		defer func() { updating = false }()

		// Allow incomplete input (e.g., just "-" or "+") without updating
		semiText := semitonesEntry.Text
		centsText := centsEntry.Text
		if semiText == "" || semiText == "-" || semiText == "+" ||
			centsText == "" || centsText == "-" || centsText == "+" {
			return
		}

		origTempo := logic.ParseFloat(originalTempoEntry.Text)
		semitonesValue := logic.ParseFloat(semiText)
		centsValue := logic.ParseFloat(centsText)
		// Far beyond any shift the tempo range allows, and safe to convert to int
		if !(math.Abs(semitonesValue) <= maxTransposeInput && math.Abs(centsValue) <= maxTransposeInput) {
			clear(newTempoEntry, timeStretchEntry, tempoDeltaLabel, semitones50Label, cents50Label)
			return
		}
		origSemitones, origCents := int(semitonesValue), int(centsValue)

		// Cents wrap into ±50 and the shift is limited to the supported tempo range
		semitones, cents := logic.NormalizeTranspose(origTempo, origSemitones, origCents)
		if semitones != origSemitones {
			semitonesEntry.SetText(signed(semitones))
		}
		if cents != origCents {
			centsEntry.SetText(signed(cents))
		}

		res := logic.CalculateFromTranspose(origTempo, semitones, cents)
		if !res.Valid {
			clear(newTempoEntry, timeStretchEntry, tempoDeltaLabel, semitones50Label, cents50Label)
			return
		}
		newTempoEntry.SetText(logic.FormatTwoDecimals(res.NewTempo))
		timeStretchEntry.SetText(logic.FormatTwoDecimals(res.TimeStretchPercent))
		showDelta(res)
	}

	// Reset function (defined after calculation functions so it can call them)
	resetToDefaults := func() {
		updating = true
		_ = originalTempo.Set("120")
		originalTempoEntry.SetText("120")
		newTempoEntry.SetText("100")
		updating = false
		// Trigger recalculation of all dependent fields
		calcFromNewTempo()
	}

	// Wire up change handlers (only for input fields)
	originalTempoEntry.OnChanged = func(s string) {
		_ = originalTempo.Set(s)
		calcFromNewTempo()
	}
	newTempoEntry.OnChanged = func(s string) { calcFromNewTempo() }
	timeStretchEntry.OnChanged = func(s string) { calcFromTimeStretch() }
	semitonesEntry.OnChanged = func(s string) { calcFromTranspose() }
	centsEntry.OnChanged = func(s string) { calcFromTranspose() }

	// Swap button to swap original and new tempo
	swapBtn := texts.buttonWithIcon("action.swap", theme.ContentRedoIcon(), func() {
		origText := originalTempoEntry.Text
		newText := newTempoEntry.Text
		originalTempoEntry.SetText(newText)
		newTempoEntry.SetText(origText)
	})

	// Reset button
	resetBtn := texts.buttonWithIcon("action.reset", theme.ViewRefreshIcon(), func() {
		resetToDefaults()
	})

	// Initialize calculated fields on startup
	calcFromNewTempo()

	// The full 50-cent labels are wider than half a phone screen; use the
	// same short form as the Frequency to Note tab there
	semis50Key, cents50Key := "field.transpose_semis50", "field.transpose_cents50"
	if fyne.CurrentDevice().IsMobile() {
		semis50Key, cents50Key = "field.semis50", "field.cents50"
	}

	return texts.view(container.NewVBox(
		texts.row("field.original_tempo", originalTempoEntry),
		texts.row("field.new_tempo", newTempoEntry),
		container.NewGridWithColumns(2,
			swapBtn,
			resetBtn,
		),
		widget.NewSeparator(),
		texts.row("field.stretch", timeStretchEntry),
		texts.row("field.tempo_delta", tempoDeltaLabel),
		widget.NewSeparator(),
		texts.row("field.transpose_semis", semitonesEntry),
		texts.row("field.transpose_cents", centsEntry),
		widget.NewSeparator(),
		texts.row(semis50Key, semitones50Label),
		texts.row(cents50Key, cents50Label),
	))
}
