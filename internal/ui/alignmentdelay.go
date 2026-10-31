package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"
)

// NewAlignmentDelayTab creates the Multi-Mic Alignment Delay calculator. The setup (temperature,
// reference mic, sample rate) sits in a card that folds into one summary line, so the mics get
// most of the screen; each mic is a compact two-line card.
func NewAlignmentDelayTab() fyne.CanvasObject {
	calc := logic.NewAlignmentDelayCalculator()
	var refreshMics func()
	refresh := func() {
		if refreshMics != nil {
			refreshMics()
		}
	}

	// Temperature and reference distance keep their values in °C and metres, so switching a
	// unit converts the shown number instead of re-interpreting it
	temp := newTemperatureInput("Temperature", logic.UnitCelsius, refresh)
	tempUnitSelect := newUnitSelect([]string{logic.UnitCelsius, logic.UnitFahrenheit}, []*unitInput{temp}, func(string) { refresh() })

	ref := newDistanceInput("Distance", logic.UnitMeters, refresh)

	sampleRateSelect := widget.NewSelect([]string{"44100", "48000", "88200", "96000", "192000"}, nil)
	sampleRateSelect.Selected = "48000"

	// Common microphone positions, grouped into submenus of the name dropdown
	micGroup := func(name string, positions ...string) widgets.MenuGroup {
		g := widgets.MenuGroup{Name: name}
		for _, p := range positions {
			g.Options = append(g.Options, widgets.MenuOption{Label: p, Value: p})
		}
		return g
	}
	commonMicPositions := []widgets.MenuGroup{
		micGroup("Drums", "Kick In", "Kick Out", "Snare Top", "Snare Bottom",
			"Hi-Hat", "Tom 1", "Tom 2", "Floor Tom"),
		micGroup("Overheads & Room", "Overhead L", "Overhead R", "Room L", "Room R", "Ambience"),
		micGroup("Instruments & Vocals", "Vocal", "Guitar Amp", "Bass Amp",
			"Piano L", "Piano R", "Strings", "Brass"),
	}

	// Mic name: free text, plus a drop-down button that picks a common position
	micNameSelect := widget.NewEntry()
	micNameSelect.SetPlaceHolder("Name")
	micMenuButton := widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), func() {
		c := fyne.CurrentApp().Driver().CanvasForObject(micNameSelect)
		if c == nil {
			return
		}
		menu := widgets.NewGroupedMenu(commonMicPositions, micNameSelect.Text, func(_ widgets.MenuGroup, o widgets.MenuOption) {
			micNameSelect.SetText(o.Value)
		})
		widget.ShowPopUpMenuAtRelativePosition(menu, c, fyne.NewPos(0, micNameSelect.Size().Height), micNameSelect)
	})
	micMenuButton.Importance = widget.LowImportance // as in widget.SelectEntry
	micNameSelect.ActionItem = micMenuButton

	// New mic distance, entered in the calculator's distance unit
	targetDistEntry := widgets.NewNumericEntry()
	targetDistEntry.SetPlaceHolder("Distance")

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

	// Setup card: a header with all values in one line, and the labelled settings when open.
	// The labels stay visible next to the default values, so the form explains itself.
	setupTitle := widgets.NewStyledText("Setup", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{Bold: true})
	setupSummary := widgets.NewMutedText("")
	setupSummary.Truncation = fyne.TextTruncateEllipsis
	setupChevron := widget.NewIcon(theme.MenuDropDownIcon())
	setupHint := widgets.NewMutedText("The reference mic is the farthest one, usually a room mic. " +
		"Every mic is delayed to line up with it.")
	setupHint.Wrapping = fyne.TextWrapWord
	setupBody := container.NewVBox(
		labeledRow("Temperature", container.NewBorder(nil, nil, nil, fixedWrap(tempUnitSelect, 70), temp.Entry)),
		labeledRow("Reference mic", container.NewBorder(nil, nil, nil, fixedWrap(distUnitSelect, 70), ref.Entry)),
		labeledRow("Sample rate", sampleRateSelect),
		setupHint,
	)
	var content *fyne.Container
	setOpen := func(open bool) {
		setupBody.Hidden = !open
		setupSummary.Hidden = open
		if open {
			setupChevron.SetResource(theme.MenuDropUpIcon())
		} else {
			setupChevron.SetResource(theme.MenuDropDownIcon())
		}
		if content != nil {
			content.Refresh() // the card's height changed
		}
	}
	setupHeader := widgets.NewTapArea(
		container.NewBorder(nil, nil, setupTitle, setupChevron, setupSummary),
		func() { setOpen(setupBody.Hidden) },
	)
	setupCard := widgets.NewCard(container.NewVBox(setupHeader, setupBody))

	updateSummary := func() {
		value := func(in *unitInput, unit string) string {
			if in.Empty() {
				return "–"
			}
			return logic.FormatTrim(in.fromBase(in.Base, unit), 2) + " " + unit
		}
		rate := logic.FormatTrim(logic.ParseFloat(sampleRateSelect.Selected)/1000, 1) + " kHz"
		widgets.SetStyledText(setupSummary,
			value(temp, tempUnitSelect.Selected)+" · ref "+value(ref, distUnitSelect.Selected)+" · "+rate, "")
	}

	// Each microphone is a compact card: name and distance on the left, delay and samples on the
	// right. The list is rebuilt on every change; it only holds a few mics.
	micList := container.NewVBox()
	emptyHint := widgets.NewMutedText("Add each close mic with its distance to the source; its delay appears here.")
	emptyHint.Wrapping = fyne.TextWrapWord

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
		if mic.IsBeyondReference {
			delayText, samplesText = "N/A", "farther than ref."
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

	refreshMics = func() {
		calc.SetTemperature(temp.Base, logic.UnitCelsius)
		calc.SetSampleRateLabel(sampleRateSelect.Selected)
		calc.SetReferenceDistance(ref.Base, logic.UnitMeters)
		calc.Recalculate()
		updateSummary()

		cards := make([]fyne.CanvasObject, 0, len(calc.Mics))
		for i, mic := range calc.Mics {
			cards = append(cards, newMicCard(i, mic))
		}
		if len(cards) == 0 {
			cards = append(cards, emptyHint)
		}
		micList.Objects = cards
		micList.Refresh()
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

	// Useful defaults: 20 °C is the usual reference temperature, 3 m a typical room mic distance
	temp.SetText("20")
	ref.SetText("3")
	setOpen(true)
	refresh()

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
			container.New(layout.NewCustomPaddedLayout(2, 2, 2, 2), setupCard),
			micRow,
		),
		nil, nil, nil,
		container.NewVScroll(container.New(layout.NewCustomPaddedLayout(2, 4, 2, 2), micList)),
	)
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
