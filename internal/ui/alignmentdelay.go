package ui

import (
	"encoding/json"
	"errors"
	"musicalc/internal/i18n"
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
func NewAlignmentDelayTab() *View {
	texts := newTextRefresh()
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
	temp := newTemperatureInput(i18n.T("field.temperature"), logic.UnitCelsius, refresh)
	texts.placeholder(temp.Entry, "field.temperature")
	tempUnitSelect := newUnitSelect([]string{logic.UnitCelsius, logic.UnitFahrenheit}, []*unitInput{temp}, func(string) { refresh() })

	ref := newDistanceInput(i18n.T("field.distance"), logic.UnitMeters, refresh)
	texts.placeholder(ref.Entry, "field.distance")

	sampleRates := []string{"44100", "48000", "88200", "96000", "192000"}
	sampleRateSelect := widget.NewSelect(sampleRates, nil)
	sampleRateSelect.Selected = "48000"

	menuGroup := func(name string, options ...string) widgets.MenuGroup {
		g := widgets.MenuGroup{Name: i18n.T("source." + name)}
		for _, o := range options {
			g.Options = append(g.Options, widgets.MenuOption{Label: i18n.T("source." + o), Value: i18n.T("source." + o)})
		}
		return g
	}
	// entryWithMenu adds a drop-down button to an entry that picks one of the options
	entryWithMenu := func(entry *widget.Entry, groups func() []widgets.MenuGroup) {
		button := widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), func() {
			c := fyne.CurrentApp().Driver().CanvasForObject(entry)
			if c == nil {
				return
			}
			menu := widgets.NewGroupedMenu(groups(), entry.Text, func(_ widgets.MenuGroup, o widgets.MenuOption) {
				entry.SetText(o.Value)
			})
			widget.ShowPopUpMenuAtRelativePosition(menu, c, fyne.NewPos(0, entry.Size().Height), entry)
		})
		button.Importance = widget.LowImportance // as in widget.SelectEntry
		entry.ActionItem = button
	}

	// The sound source all distances are measured from: free text or a common source
	sourceEntry := widget.NewEntry()
	texts.placeholder(sourceEntry, "field.source")
	entryWithMenu(sourceEntry, func() []widgets.MenuGroup {
		return []widgets.MenuGroup{
			menuGroup("Drums", "Snare", "Kick", "Hi-Hat", "Toms"),
			menuGroup("Instruments", "Guitar Amp", "Bass Amp", "Piano", "Vocal"),
		}
	})
	sourceEntry.OnChanged = func(string) { refresh() }

	// Mic name: free text, or a common position
	micNameSelect := widget.NewEntry()
	texts.placeholder(micNameSelect, "field.name")
	entryWithMenu(micNameSelect, func() []widgets.MenuGroup {
		return []widgets.MenuGroup{
			menuGroup("Drums", "Kick In", "Kick Out", "Snare Top", "Snare Bottom",
				"Hi-Hat", "Tom 1", "Tom 2", "Floor Tom"),
			menuGroup("Overheads & Room", "Overhead L", "Overhead R", "Room L", "Room R", "Ambience"),
			menuGroup("Instruments & Vocals", "Vocal", "Guitar Amp", "Bass Amp",
				"Piano L", "Piano R", "Strings", "Brass"),
		}
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
		texts.row("field.source", sourceEntry),
		texts.row("field.temperature", container.NewBorder(nil, nil, nil, fixedWrap(tempUnitSelect, 70), temp.Entry)),
		texts.row("field.reference_mic", container.NewBorder(nil, nil, nil, fixedWrap(distUnitSelect, 70), ref.Entry)),
		texts.row("field.sample_rate", sampleRateSelect),
		texts.wrappedMuted("alignment.hint"),
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
			i18n.T("alignment.reference_summary", value(ref, distUnitSelect.Selected)),
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
	var micTextRefresh []func()
	listHeader := texts.muted("alignment.delay_title")
	emptyHint := wrappedMuted("")
	// The setup warning explains why no delays are shown, in the stereo hint's warning card
	setupWarningText := wrappedMuted("")
	setupWarning := widgets.NewCard(setupWarningText)
	setupWarning.SetOutline(theme.ColorNameWarning)
	var setupErr error   // why the setup cannot be calculated, or nil
	addRejected := false // the last Add had a distance that cannot be used

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

		delayText, samplesText := logic.FormatTrim(mic.DelayMS, 2)+" ms", alignmentSamplesText(mic)
		delayColor, samplesColor := theme.ColorNameForeground, theme.ColorNamePlaceHolder
		valid := setupErr == nil
		switch {
		case !valid:
			delayText, samplesText = "–", ""
			delayColor = theme.ColorNamePlaceHolder
		case mic.IsReference:
			delayText = "0 ms"
			delayColor = theme.ColorNamePlaceHolder
		case mic.IsBeyondReference:
			// It arrives later than the reference: it would have to move earlier, not be delayed
			delayText = minusSign(logic.FormatTrim(mic.DelayMS, 2)) + " ms"
			delayColor, samplesColor = theme.ColorNameWarning, theme.ColorNameWarning
		}
		alignRight := func(o fyne.CanvasObject) fyne.CanvasObject { return container.NewBorder(nil, nil, nil, o) }
		samplesDisplay := widgets.NewStyledText(samplesText, samplesColor, theme.SizeNameText, fyne.TextStyle{})
		if valid {
			micTextRefresh = append(micTextRefresh, func() { widgets.SetStyledText(samplesDisplay, alignmentSamplesText(mic), "") })
		}
		result := textLines(
			alignRight(widgets.NewStyledText(delayText, delayColor, theme.SizeNameText, fyne.TextStyle{Bold: true})),
			alignRight(samplesDisplay),
		)
		return widgets.NewCard(container.NewBorder(nil, nil, nil,
			container.NewHBox(result, container.NewCenter(removeBtn)),
			textLines(name, distance),
		))
	}

	// newPairHint warns that a stereo pair gets different delays
	newPairHint := func(p logic.PairDiff) fyne.CanvasObject {
		hint := wrappedMuted(i18n.T("alignment.stereo_warning", p.Base, logic.FormatTrim(p.DiffMS, 2)))
		micTextRefresh = append(micTextRefresh, func() {
			widgets.SetStyledText(hint, i18n.T("alignment.stereo_warning", p.Base, logic.FormatTrim(p.DiffMS, 2)), "")
		})
		card := widgets.NewCard(hint)
		card.SetOutline(theme.ColorNameWarning)
		return card
	}

	renderSetup := func() {
		// Name the source where distances are entered
		from := i18n.T("source.generic")
		if calc.Source != "" {
			name := calc.Source
			if i18n.Language() == "en" {
				name = strings.ToLower(name)
			}
			from = i18n.T("source.named", name)
			targetDistEntry.SetPlaceHolder(i18n.T("alignment.from", name))
		} else {
			targetDistEntry.SetPlaceHolder(i18n.T("field.distance"))
		}
		widgets.SetStyledText(emptyHint, i18n.T("alignment.add_hint", from), "")
		updateSetupHeader()

	}

	refreshMics = func() {
		calc.Source = strings.TrimSpace(sourceEntry.Text)
		calc.SetTemperature(temp.Base, logic.UnitCelsius)
		calc.SetSampleRateLabel(sampleRateSelect.Selected)
		calc.SetReferenceDistance(ref.Base, logic.UnitMeters)
		// Empty or unreadable temperature and reference text must not stand for 0
		switch {
		case !temp.Valid():
			setupErr = logic.ErrTemperatureRange
			calc.ClearDelays()
		case !ref.Valid():
			setupErr = logic.ErrReferenceInvalid
			calc.ClearDelays()
		default:
			setupErr = calc.Recalculate()
		}

		renderSetup()
		micTextRefresh = nil

		// Pair hints follow the second mic of their pair
		hints := map[int]logic.PairDiff{}
		for _, p := range logic.StereoPairs(calc.Mics) {
			hints[max(p.Left, p.Right)] = p
		}
		cards := make([]fyne.CanvasObject, 0, len(calc.Mics)+len(hints)+2)
		if setupErr != nil || addRejected {
			widgets.SetStyledText(setupWarningText, alignmentSetupMessage(setupErr, addRejected), "")
			cards = append(cards, setupWarning)
		}
		if len(calc.Mics) == 0 {
			cards = append(cards, emptyHint)
		} else {
			cards = append(cards, listHeader)
		}
		for i, mic := range calc.Mics {
			cards = append(cards, newMicCard(i, mic))
			if p, ok := hints[i]; ok && setupErr == nil {
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
			name = i18n.T("alignment.mic", len(calc.Mics)+1)
		}

		text := strings.TrimSpace(targetDistEntry.Text)
		dist, ok := logic.ParseNumber(text)
		if text == "" || (ok && dist == 0) {
			return
		}
		if !ok || dist < 0 {
			// A distance from the source cannot be negative; keep the text for correcting it
			addRejected = true
			refresh()
			return
		}
		addRejected = false
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
	session, sessionErr := restoreAlignmentSession(fyne.CurrentApp().Preferences().String(prefAlignmentSession))
	if sessionErr != nil {
		showErrorDialog(errors.New(i18n.T("alignment.session_error", alignmentErrorText(sessionErr))))
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
	texts.add(func() {
		renderSetup()
		for _, render := range micTextRefresh {
			render()
		}
	})

	return texts.view(content)
}

// fixedWidthLayout sizes its object to a fixed width and the full height it is given

func alignmentSamplesText(mic logic.AlignmentMic) string {
	switch {
	case mic.IsReference:
		return i18n.T("alignment.reference")
	case mic.IsBeyondReference && mic.DelaySamples == -1:
		return i18n.T("alignment.one_earlier")
	case mic.IsBeyondReference:
		return i18n.T("alignment.earlier", -mic.DelaySamples)
	default:
		return i18n.Plural("alignment.samples", mic.DelaySamples, mic.DelaySamples)
	}
}

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

// defaultAlignmentSession is the first-use setup: 20 °C is the usual reference temperature, 3 m a
// typical room mic distance
func defaultAlignmentSession() logic.SessionState {
	return logic.SessionState{
		Preset:       logic.Preset{Source: "Snare", ReferenceMeters: 3, DistanceUnit: logic.UnitMeters, SampleRate: 48000},
		TemperatureC: 20, TemperatureUnit: logic.UnitCelsius,
	}
}

// restoreAlignmentSession reads the stored session. Nothing stored gives the defaults; a session
// that cannot be read or is not a usable setup gives the defaults and the reason, so it is not
// silently replaced.
func restoreAlignmentSession(data string) (logic.SessionState, error) {
	if data == "" {
		return defaultAlignmentSession(), nil
	}
	var s logic.SessionState
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return defaultAlignmentSession(), errStoredUnreadable
	}
	if err := logic.ValidateSession(&s); err != nil {
		return defaultAlignmentSession(), err
	}
	return s, nil
}

// errStoredUnreadable is stored data that is not valid JSON of the expected shape
var errStoredUnreadable = errors.New("the stored data is damaged")

// alignmentErrorText describes why a setup, a stored session or stored presets cannot be used
func alignmentErrorText(err error) string {
	switch {
	case errors.Is(err, errStoredUnreadable):
		return i18n.T("stored.unreadable")
	case errors.Is(err, logic.ErrTemperatureRange):
		return i18n.T("alignment.invalid_temperature")
	case errors.Is(err, logic.ErrReferenceInvalid):
		return i18n.T("alignment.invalid_reference")
	case errors.Is(err, logic.ErrMicInvalid):
		return i18n.T("alignment.invalid_mic")
	case errors.Is(err, logic.ErrDelayTooLarge):
		return i18n.T("alignment.invalid_delay")
	case errors.Is(err, logic.ErrSampleRate):
		return i18n.T("alignment.invalid_rate")
	default:
		return presetErrorText(err)
	}
}

// alignmentSetupMessage is the setup warning: why no delays are shown, and a rejected distance
func alignmentSetupMessage(setupErr error, addRejected bool) string {
	var parts []string
	if setupErr != nil {
		parts = append(parts, i18n.T("alignment.no_delays", alignmentErrorText(setupErr)))
	}
	if addRejected {
		parts = append(parts, i18n.T("alignment.invalid_distance"))
	}
	return strings.Join(parts, " ")
}
