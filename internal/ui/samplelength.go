package ui

import (
	"fmt"
	"math"
	"musicalc/internal/i18n"
	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewSampleLengthTab() *View {
	texts := newTextRefresh()
	// Sample rate binding
	sampleRate := binding.NewString()
	_ = sampleRate.Set("44100")

	// Beats binding
	beats := binding.NewString()
	_ = beats.Set("4")

	// Common sample rates (including lo-fi)
	sampleRates := []string{
		"8000",   // Lo-fi, telephone quality
		"11025",  // Lo-fi, quarter CD quality
		"16000",  // Wideband audio
		"22050",  // Half CD quality
		"32000",  // MiniDv, video
		"44100",  // CD quality (standard)
		"48000",  // DVD, professional audio
		"88200",  // High-res audio
		"96000",  // High-res audio, Blu-ray
		"192000", // Ultra high-res
	}

	// Sample rate selector (callback will be set after calcFromTempo is defined)
	sampleRateSelect := widget.NewSelectEntry(sampleRates)
	sampleRateSelect.SetText("44100")
	texts.add(func() { sampleRateSelect.PlaceHolder = i18n.T("field.sample_rate_title"); sampleRateSelect.Refresh() })

	// Beats input
	beatsInput := widgets.NewNumericEntry()
	texts.placeholder(beatsInput, "field.beats")

	// Bidirectional input/output fields
	samplesEntry := widgets.NewNumericEntry()
	texts.placeholder(samplesEntry, "field.samples")
	msEntry := widgets.NewNumericEntry()
	texts.placeholder(msEntry, "field.milliseconds")
	tempoEntry := widgets.NewNumericEntry()
	texts.placeholder(tempoEntry, "field.tempo")

	// Flag to prevent circular updates
	updating := false

	// Declare calculation functions as variables first
	var calcFromTempo func()
	var calcFromSamples func()
	var calcFromMS func()

	// Reset function
	resetToDefaults := func() {
		updating = true
		_ = sampleRate.Set("44100")
		_ = beats.Set("4")
		tempoEntry.SetText("120")
		updating = false
		calcFromTempo()
	}

	// Calculate from Tempo (when Tempo is input)
	calcFromTempo = func() {
		if updating {
			return
		}
		updating = true
		defer func() { updating = false }()

		srVal, _ := sampleRate.Get()
		sr := logic.ParseFloat(srVal)
		bt := logic.ParseFloat(beatsInput.Text)
		bpm := logic.ParseFloat(tempoEntry.Text)

		res := logic.GetSampleData(sr, bpm, bt)
		if !res.Valid {
			// Invalid input must not leave earlier results looking current
			samplesEntry.SetText("")
			msEntry.SetText("")
			return
		}
		samplesEntry.SetText(fmt.Sprintf("%d", res.Samples))
		msEntry.SetText(logic.FormatTwoDecimals(res.MS))
	}

	// Calculate from Samples (when Samples is input)
	calcFromSamples = func() {
		if updating {
			return
		}
		updating = true
		defer func() { updating = false }()

		srVal, _ := sampleRate.Get()
		sr := logic.ParseFloat(srVal)
		bt := logic.ParseFloat(beatsInput.Text)
		samples := logic.ParseFloat(samplesEntry.Text)

		ms := samples / sr * 1000.0
		bpm := 60000.0 * bt / ms
		if !(samples > 0 && sr > 0 && bt > 0) || math.IsInf(ms, 0) || math.IsInf(bpm, 0) || math.IsNaN(bpm) {
			msEntry.SetText("")
			tempoEntry.SetText("")
			return
		}
		msEntry.SetText(logic.FormatTwoDecimals(ms))
		tempoEntry.SetText(logic.FormatTwoDecimals(bpm))
	}

	// Calculate from MS (when MS is input)
	calcFromMS = func() {
		if updating {
			return
		}
		updating = true
		defer func() { updating = false }()

		srVal, _ := sampleRate.Get()
		sr := logic.ParseFloat(srVal)
		bt := logic.ParseFloat(beatsInput.Text)
		ms := logic.ParseFloat(msEntry.Text)

		samples, ok := logic.SamplesIn(ms, sr)
		bpm := 60000.0 * bt / ms
		if !ok || !(bt > 0) || math.IsInf(bpm, 0) || math.IsNaN(bpm) {
			samplesEntry.SetText("")
			tempoEntry.SetText("")
			return
		}
		samplesEntry.SetText(fmt.Sprintf("%d", samples))
		tempoEntry.SetText(logic.FormatTwoDecimals(bpm))
	}

	// Wire up change handlers
	tempoEntry.OnChanged = func(s string) { calcFromTempo() }
	samplesEntry.OnChanged = func(s string) { calcFromSamples() }
	msEntry.OnChanged = func(s string) { calcFromMS() }

	// Set sample rate and beats callbacks now that functions are defined
	sampleRateSelect.OnChanged = func(s string) {
		_ = sampleRate.Set(s)
		calcFromTempo()
	}
	beatsInput.SetText("4")
	beatsInput.OnChanged = func(s string) {
		_ = beats.Set(s)
		calcFromTempo()
	}

	// Initialize values on startup
	tempoEntry.SetText("120")
	calcFromTempo()

	// Reset button
	resetBtn := texts.buttonWithIcon("action.reset", theme.ViewRefreshIcon(), func() {
		resetToDefaults()
		sampleRateSelect.SetText("44100")
		beatsInput.SetText("4")
	})

	return texts.view(container.NewVBox(
		container.NewGridWithColumns(2,
			texts.label("field.sample_rate_title"),
			sampleRateSelect,
		),
		container.NewGridWithColumns(2,
			texts.label("field.beats"),
			beatsInput,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel(""),
			resetBtn,
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			texts.label("field.samples"),
			samplesEntry,
		),
		container.NewGridWithColumns(2,
			texts.label("field.milliseconds"),
			msEntry,
		),
		container.NewGridWithColumns(2,
			texts.label("field.tempo"),
			tempoEntry,
		),
	))
}
