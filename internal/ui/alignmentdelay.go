package ui

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"
)

// NewAlignmentDelayTab creates the Multi-Mic Alignment Delay calculator. The setup (source,
// temperature, reference mic, sample rate) sits in a card that folds into one summary line, so the
// mics get most of the screen; each mic is a compact two-line card. The setup is restored when the
// app starts again, and can be stored as presets through the header's overflow menu.
func NewAlignmentDelayTab() fyne.CanvasObject {
	calc := logic.NewAlignmentDelayCalculator()
	var refreshMics func()
	restoring := false // a preset or the last session is being put on screen
	refresh := func() {
		if refreshMics != nil && !restoring {
			refreshMics()
		}
	}

	// Temperature and reference distance keep their values in °C and metres, so switching a
	// unit converts the shown number instead of re-interpreting it
	temp := newTemperatureInput("Temperature", logic.UnitCelsius, refresh)
	tempUnitSelect := newUnitSelect([]string{logic.UnitCelsius, logic.UnitFahrenheit}, []*unitInput{temp}, func(string) { refresh() })

	ref := newDistanceInput("Distance", logic.UnitMeters, refresh)

	sampleRates := []string{"44100", "48000", "88200", "96000", "192000"}
	sampleRateSelect := widget.NewSelect(sampleRates, nil)
	sampleRateSelect.Selected = "48000"

	menuGroup := func(name string, options ...string) widgets.MenuGroup {
		g := widgets.MenuGroup{Name: name}
		for _, o := range options {
			g.Options = append(g.Options, widgets.MenuOption{Label: o, Value: o})
		}
		return g
	}
	// entryWithMenu adds a drop-down button to an entry that picks one of the options
	entryWithMenu := func(entry *widget.Entry, groups []widgets.MenuGroup) {
		button := widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), func() {
			c := fyne.CurrentApp().Driver().CanvasForObject(entry)
			if c == nil {
				return
			}
			menu := widgets.NewGroupedMenu(groups, entry.Text, func(_ widgets.MenuGroup, o widgets.MenuOption) {
				entry.SetText(o.Value)
			})
			widget.ShowPopUpMenuAtRelativePosition(menu, c, fyne.NewPos(0, entry.Size().Height), entry)
		})
		button.Importance = widget.LowImportance // as in widget.SelectEntry
		entry.ActionItem = button
	}

	// The sound source all distances are measured from: free text or a common source
	sourceEntry := widget.NewEntry()
	sourceEntry.SetPlaceHolder("Source")
	entryWithMenu(sourceEntry, []widgets.MenuGroup{
		menuGroup("Drums", "Snare", "Kick", "Hi-Hat", "Toms"),
		menuGroup("Instruments", "Guitar Amp", "Bass Amp", "Piano", "Vocal"),
	})
	sourceEntry.OnChanged = func(string) { refresh() }

	// Mic name: free text, or a common position
	micNameSelect := widget.NewEntry()
	micNameSelect.SetPlaceHolder("Name")
	entryWithMenu(micNameSelect, []widgets.MenuGroup{
		menuGroup("Drums", "Kick In", "Kick Out", "Snare Top", "Snare Bottom",
			"Hi-Hat", "Tom 1", "Tom 2", "Floor Tom"),
		menuGroup("Overheads & Room", "Overhead L", "Overhead R", "Room L", "Room R", "Ambience"),
		menuGroup("Instruments & Vocals", "Vocal", "Guitar Amp", "Bass Amp",
			"Piano L", "Piano R", "Strings", "Brass"),
	})

	// New mic distance, entered in the calculator's distance unit
	targetDistEntry := widgets.NewNumericEntry()

	// One distance unit for the whole calculator: the reference, new mics and the mic cards.
	// The mic row shows it as a suffix instead of having a dropdown of its own.
	unitSuffix := widgets.NewMutedText(logic.UnitMeters)
	distUnitSelect := newUnitSelect([]string{logic.UnitMeters, logic.UnitFeet}, []*unitInput{ref}, func(unit string) {
		widgets.SetStyledText(unitSuffix, unit, "")
		refresh()
	})

	// fixedWrap gives obj a fixed width; its height follows the row, so selects and buttons
	// line up with the (taller) entries next to them
	fixedWrap := func(obj fyne.CanvasObject, w float32) fyne.CanvasObject {
		return container.New(fixedWidthLayout{width: w}, obj)
	}
	// textLines stacks texts with their usual line spacing instead of the larger widget padding
	textGap := theme.Size(theme.SizeNameLineSpacing) - 2*theme.Size(theme.SizeNameInnerPadding)
	textLines := func(objects ...fyne.CanvasObject) *fyne.Container {
		return container.New(layout.NewCustomPaddedVBoxLayout(textGap), objects...)
	}

	// current is the setup on screen as a preset (without a name and the temperature)
	current := func() logic.Preset {
		p := logic.Preset{
			Source:          strings.TrimSpace(sourceEntry.Text),
			ReferenceMeters: ref.Base,
			DistanceUnit:    distUnitSelect.Selected,
			SampleRate:      logic.ParseSampleRate(sampleRateSelect.Selected),
		}
		for _, m := range calc.Mics {
			p.Mics = append(p.Mics, logic.PresetMic{Name: m.Name, DistanceMeters: m.DistanceMeters})
		}
		return p
	}
	// apply puts a preset's setup on screen; the temperature stays as it is
	var foldSetup func() // set once the setup card exists
	apply := func(p logic.Preset) {
		restoring = true
		if p.DistanceUnit != "" {
			distUnitSelect.SetSelected(p.DistanceUnit)
		}
		ref.setBase(p.ReferenceMeters)
		sourceEntry.SetText(p.Source)
		rate := strconv.Itoa(p.SampleRate)
		if !slices.Contains(sampleRateSelect.Options, rate) {
			sampleRateSelect.Options = append(sampleRateSelect.Options, rate)
		}
		sampleRateSelect.SetSelected(rate)
		calc.Mics = nil
		for _, m := range p.Mics {
			calc.Mics = append(calc.Mics, logic.AlignmentMic{Name: m.Name, DistanceMeters: m.DistanceMeters})
		}
		restoring = false
		refresh()
		// A loaded setup with mics is ready to use: make room for the mics
		if foldSetup != nil && len(calc.Mics) > 0 {
			foldSetup()
		}
	}
	presets := newPresetController(current, apply, refresh)

	// Setup card: a header with all values in one line, and the labelled settings when open.
	// The labels stay visible next to the values, so the form explains itself.
	var content *fyne.Container
	setup := newSetupCard(container.NewVBox(
		labeledRow("Source", sourceEntry),
		labeledRow("Temperature", container.NewBorder(nil, nil, nil, fixedWrap(tempUnitSelect, 70), temp.Entry)),
		labeledRow("Reference mic", container.NewBorder(nil, nil, nil, fixedWrap(distUnitSelect, 70), ref.Entry)),
		labeledRow("Sample rate", sampleRateSelect),
		wrappedMuted("All distances are measured from the source. The reference mic is the farthest one, "+
			"usually a room mic; the other mics get delay to line up with it."),
	), func(bool) {
		if content != nil {
			content.Refresh() // the card's height changed
		}
	})
	setOpen := setup.SetOpen
	foldSetup = func() { setOpen(false) }

	updateSetupHeader := func() {
		value := func(in *unitInput, unit string) string {
			if in.Empty() {
				return "–"
			}
			return logic.FormatTrim(in.fromBase(in.Base, unit), 2) + " " + unit
		}
		parts := []string{}
		if s := strings.TrimSpace(sourceEntry.Text); s != "" {
			parts = append(parts, s)
		}
		parts = append(parts,
			"ref "+value(ref, distUnitSelect.Selected),
			value(temp, tempUnitSelect.Selected),
			logic.FormatTrim(logic.ParseFloat(sampleRateSelect.Selected)/1000, 1)+" kHz")
		setup.SetSummary(strings.Join(parts, " · "))
		setup.SetTitle(presets.title())
	}

	// saveSession stores the setup on screen, so it is back when the app starts again
	saveSession := func() {
		s := logic.SessionState{Preset: current(), TemperatureC: temp.Base,
			TemperatureUnit: tempUnitSelect.Selected, LoadedPreset: presets.loaded}
		if data, err := json.Marshal(s); err == nil {
			fyne.CurrentApp().Preferences().SetString(prefAlignmentSession, string(data))
		}
	}

	// Each microphone is a compact card: name and distance on the left, delay and samples on the
	// right. The list is rebuilt on every change; it only holds a few mics.
	micList := container.NewVBox()
	listHeader := widgets.NewMutedText("Delay to add · estimate, check by ear")
	emptyHint := wrappedMuted("")

	newMicCard := func(index int, mic logic.AlignmentMic) fyne.CanvasObject {
		name := widgets.NewStyledText(mic.Name, theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{Bold: true})
		name.Truncation = fyne.TextTruncateEllipsis
		unit := distUnitSelect.Selected
		distance := widgets.NewMutedText(logic.FormatTrim(logic.FromMeters(mic.DistanceMeters, unit), 2) + " " + unit)
		// A plain button: the touch-responsive widgets.Button would stop drags on it from scrolling the list
		removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			calc.RemoveMicAt(index)
			refresh()
		})
		removeBtn.Importance = widget.LowImportance

		delayText, samplesText := logic.FormatTrim(mic.DelayMS, 2)+" ms", fmt.Sprintf("%d samples", mic.DelaySamples)
		delayColor, samplesColor := theme.ColorNameForeground, theme.ColorNamePlaceHolder
		switch {
		case mic.IsReference:
			delayText, samplesText = "0 ms", "reference"
			delayColor = theme.ColorNamePlaceHolder
		case mic.IsBeyondReference:
			// It arrives later than the reference: it would have to move earlier, not be delayed
			delayText = minusSign(logic.FormatTrim(mic.DelayMS, 2)) + " ms"
			samplesText = fmt.Sprintf("%d samples earlier", -mic.DelaySamples)
			if mic.DelaySamples == -1 {
				samplesText = "1 sample earlier"
			}
			delayColor, samplesColor = theme.ColorNameWarning, theme.ColorNameWarning
		}
		alignRight := func(o fyne.CanvasObject) fyne.CanvasObject { return container.NewBorder(nil, nil, nil, o) }
		result := textLines(
			alignRight(widgets.NewStyledText(delayText, delayColor, theme.SizeNameText, fyne.TextStyle{Bold: true})),
			alignRight(widgets.NewStyledText(samplesText, samplesColor, theme.SizeNameText, fyne.TextStyle{})),
		)
		return widgets.NewCard(container.NewBorder(nil, nil, nil,
			container.NewHBox(result, container.NewCenter(removeBtn)),
			textLines(name, distance),
		))
	}

	// newPairHint warns that a stereo pair gets different delays
	newPairHint := func(p logic.PairDiff) fyne.CanvasObject {
		card := widgets.NewCard(wrappedMuted(fmt.Sprintf("%s L/R differ by %s ms. Separate delays shift the "+
			"stereo image; use one delay for both to keep it.", p.Base, logic.FormatTrim(p.DiffMS, 2))))
		card.SetOutline(theme.ColorNameWarning)
		return card
	}

	refreshMics = func() {
		calc.Source = strings.TrimSpace(sourceEntry.Text)
		calc.SetTemperature(temp.Base, logic.UnitCelsius)
		calc.SetSampleRateLabel(sampleRateSelect.Selected)
		calc.SetReferenceDistance(ref.Base, logic.UnitMeters)
		calc.Recalculate()

		// Name the source where distances are entered
		from := "the source"
		if calc.Source != "" {
			from = "the " + strings.ToLower(calc.Source)
			targetDistEntry.SetPlaceHolder("From " + strings.ToLower(calc.Source))
		} else {
			targetDistEntry.SetPlaceHolder("Distance")
		}
		widgets.SetStyledText(emptyHint, "Add each mic with its distance from "+from+". The delay to add appears here.", "")
		updateSetupHeader()

		// Pair hints follow the second mic of their pair
		hints := map[int]logic.PairDiff{}
		for _, p := range logic.StereoPairs(calc.Mics) {
			hints[max(p.Left, p.Right)] = p
		}
		cards := make([]fyne.CanvasObject, 0, len(calc.Mics)+len(hints)+1)
		if len(calc.Mics) == 0 {
			cards = append(cards, emptyHint)
		} else {
			cards = append(cards, listHeader)
		}
		for i, mic := range calc.Mics {
			cards = append(cards, newMicCard(i, mic))
			if p, ok := hints[i]; ok {
				cards = append(cards, newPairHint(p))
			}
		}
		micList.Objects = cards
		micList.Refresh()
		saveSession()
	}

	addButton := widgets.NewButton("+", func() {
		name := strings.TrimSpace(micNameSelect.Text)
		if name == "" {
			name = fmt.Sprintf("Mic %d", len(calc.Mics)+1)
		}

		dist := logic.ParseFloat(targetDistEntry.Text)
		if dist == 0 {
			return
		}
		calc.AddMic(name, dist, distUnitSelect.Selected)

		micNameSelect.SetText("")
		targetDistEntry.SetText("")
		// The setup is usually done once the first mic is in: fold it away to make room
		if len(calc.Mics) == 1 {
			setOpen(false)
		}
		refresh()
	})
	addButton.Importance = widget.HighImportance

	sampleRateSelect.OnChanged = func(string) { refresh() }

	// Restore the last session, or start with useful defaults: 20 °C is the usual reference
	// temperature, 3 m a typical room mic distance
	var session logic.SessionState
	if data := fyne.CurrentApp().Preferences().String(prefAlignmentSession); data == "" ||
		json.Unmarshal([]byte(data), &session) != nil {
		session = logic.SessionState{
			Preset:       logic.Preset{Source: "Snare", ReferenceMeters: 3, DistanceUnit: logic.UnitMeters, SampleRate: 48000},
			TemperatureC: 20, TemperatureUnit: logic.UnitCelsius,
		}
	}
	restoring = true
	if session.TemperatureUnit != "" {
		tempUnitSelect.SetSelected(session.TemperatureUnit)
	}
	temp.setBase(session.TemperatureC)
	restoring = false
	apply(session.Preset)
	if logic.FindPreset(presets.presets, session.LoadedPreset) >= 0 {
		presets.loaded = session.LoadedPreset
	}
	refresh()
	setOpen(len(calc.Mics) == 0)

	micRow := container.NewBorder(
		nil, nil, nil,
		fixedWrap(addButton, 50),
		container.NewBorder(nil, nil, nil, unitSuffix,
			container.NewGridWithColumns(2, micNameSelect, targetDistEntry),
		),
	)

	// The setup and the mic row stay on top; the mic cards fill and scroll in the remaining space.
	// A little inset keeps the card shadows from being clipped.
	content = container.NewBorder(
		container.NewVBox(
			container.New(layout.NewCustomPaddedLayout(2, 2, 2, 2), setup.Card),
			micRow,
		),
		nil, nil, nil,
		container.NewVScroll(container.New(layout.NewCustomPaddedLayout(2, 4, 2, 2), micList)),
	)
	setTabMenu(content, func() []*fyne.MenuItem {
		windows := fyne.CurrentApp().Driver().AllWindows()
		if len(windows) == 0 {
			return nil
		}
		return presets.menuItems(windows[0])
	})
	return content
}

// fixedWidthLayout sizes its object to a fixed width and the full height it is given
type fixedWidthLayout struct {
	width float32
}

func (l fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(size)
	}
}

func (l fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var height float32
	for _, o := range objects {
		height = fyne.Max(height, o.MinSize().Height)
	}
	return fyne.NewSize(l.width, height)
}
