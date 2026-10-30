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

// NewAlignmentDelayTab creates the Multi-Mic Alignment Delay calculator
func NewAlignmentDelayTab() fyne.CanvasObject {
	// Phones get shorter texts and a two-row header so inputs and columns aren't truncated
	isMobile := fyne.CurrentDevice().IsMobile()

	// Temperature input
	tempEntry := widget.NewEntry()
	//tempEntry.SetText("22")
	tempEntry.SetPlaceHolder("Room Temp.")
	if isMobile {
		tempEntry.SetPlaceHolder("Temp.")
	}

	tempUnitSelect := widget.NewSelect([]string{"C", "F"}, nil)
	tempUnitSelect.SetSelected("C")

	// Reference distance input
	refDistEntry := widget.NewEntry()
	//refDistEntry.SetText("5")
	refDistEntry.SetPlaceHolder("Ref. Dist.")

	refUnitSelect := widget.NewSelect([]string{"m", "ft"}, nil)
	refUnitSelect.SetSelected("m")

	// Sample rate dropdown
	sampleRateSelect := widget.NewSelect([]string{
		"44100",
		"48000",
		"88200",
		"96000",
		"192000",
	}, nil)
	sampleRateSelect.SetSelected("48000")

	// Common microphone positions for dropdown
	commonMicPositions := []string{
		"Kick In", "Kick Out", "Snare Top", "Snare Bottom",
		"Hi-Hat", "Tom 1", "Tom 2", "Floor Tom",
		"Overhead L", "Overhead R", "Room L", "Room R",
		"Vocal", "Guitar Amp", "Bass Amp",
		"Piano L", "Piano R", "Strings", "Brass", "Ambience",
	}

	// Mic name as Select (dropdown only for consistent height)
	micNameSelect := widget.NewSelectEntry(commonMicPositions)
	micNameSelect.SetPlaceHolder("Name")

	// Target distance input
	targetDistEntry := widget.NewEntry()
	targetDistEntry.SetPlaceHolder("Distance to Ref.")
	if isMobile {
		targetDistEntry.SetPlaceHolder("Distance")
	}

	targetUnitSelect := widget.NewSelect([]string{"m", "ft"}, nil)
	targetUnitSelect.SetSelected("m")
	targetUnitSelect.Resize(fyne.NewSize(40, targetUnitSelect.MinSize().Height))

	calc := logic.NewAlignmentDelayCalculator()

	formatTrimFloat := func(v float64, decimals int) string {
		s := fmt.Sprintf("%.*f", decimals, v)
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
		if s == "" || s == "-0" {
			return "0"
		}
		return s
	}

	// Each microphone is shown as a card. The list is rebuilt on every change; it only holds a few mics.
	micList := container.NewVBox()
	emptyHint := widgets.NewMutedText("Add a mic with a name and its distance to the reference.")
	emptyHint.Wrapping = fyne.TextWrapWord

	var refreshMics func()
	newMicCard := func(index int, mic logic.AlignmentMic) fyne.CanvasObject {
		name := widget.NewLabelWithStyle(mic.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		name.Truncation = fyne.TextTruncateEllipsis
		unit := targetUnitSelect.Selected
		distance := widgets.NewMutedText(formatTrimFloat(logic.FromMeters(mic.DistanceMeters, unit), 3) + " " + unit)
		removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			calc.RemoveMicAt(index)
			refreshMics()
		})
		removeBtn.Importance = widget.LowImportance

		delayText, samplesText := formatTrimFloat(mic.DelayMS, 2)+" ms", fmt.Sprintf("%d samples", mic.DelaySamples)
		delayColor, samplesColor := theme.ColorNameForeground, theme.ColorNamePlaceHolder
		if mic.IsBeyondReference {
			delayText, samplesText = "N/A", "Farther than the reference"
			delayColor, samplesColor = theme.ColorNameWarning, theme.ColorNameWarning
		}

		return widgets.NewCard(container.New(layout.NewCustomPaddedVBoxLayout(0),
			container.NewBorder(nil, nil, nil, container.NewHBox(distance, removeBtn), name),
			// Overlap the texts' inner padding so they sit a padding apart; the smaller text
			// sits at the bottom to share the delay's baseline
			container.New(layout.NewCustomPaddedHBoxLayout(theme.Padding()-2*theme.InnerPadding()),
				widgets.NewStyledText(delayText, delayColor, theme.SizeNameSubHeadingText, fyne.TextStyle{Bold: true}),
				container.NewVBox(layout.NewSpacer(),
					widgets.NewStyledText(samplesText, samplesColor, theme.SizeNameText, fyne.TextStyle{})),
			),
		))
	}

	refreshMics = func() {
		calc.SetTemperature(logic.ParseFloat(tempEntry.Text), tempUnitSelect.Selected)
		calc.SetSampleRateLabel(sampleRateSelect.Selected)
		calc.SetReferenceDistance(logic.ParseFloat(refDistEntry.Text), refUnitSelect.Selected)
		calc.Recalculate()

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

	// Add microphone button with emphasis
	addButton := widget.NewButton("+", func() {
		name := strings.TrimSpace(micNameSelect.Text)
		if name == "" {
			name = fmt.Sprintf("Mic %d", len(calc.Mics)+1)
		}

		dist := logic.ParseFloat(targetDistEntry.Text)
		if dist == 0 {
			return
		}
		calc.AddMic(name, dist, targetUnitSelect.Selected)

		micNameSelect.SetText("")
		targetDistEntry.SetText("")
		refreshMics()
	})

	// Recalculate when temperature, reference, sample rate or display unit change
	tempEntry.OnChanged = func(string) { refreshMics() }
	tempUnitSelect.OnChanged = func(string) { refreshMics() }
	refDistEntry.OnChanged = func(string) { refreshMics() }
	refUnitSelect.OnChanged = func(string) { refreshMics() }
	sampleRateSelect.OnChanged = func(string) { refreshMics() }
	targetUnitSelect.OnChanged = func(string) { refreshMics() }
	refreshMics()

	// Compact layout with minimal widths
	tempEntry.Resize(fyne.NewSize(80, tempEntry.MinSize().Height))
	refDistEntry.Resize(fyne.NewSize(80, refDistEntry.MinSize().Height))
	targetDistEntry.Resize(fyne.NewSize(80, targetDistEntry.MinSize().Height))

	fixedWrap := func(obj fyne.CanvasObject, w float32) fyne.CanvasObject {
		return container.NewGridWrap(fyne.NewSize(w, obj.MinSize().Height), obj)
	}

	// Top section: Temperature, Reference, and Sample Rate on same row

	tempGroup := container.NewBorder(
		nil, nil, nil,
		fixedWrap(tempUnitSelect, 60),
		tempEntry,
	)

	refGroup := container.NewBorder(
		nil, nil, nil,
		fixedWrap(refUnitSelect, 70),
		refDistEntry,
	)

	tempRefRow := container.NewGridWithColumns(2,
		tempGroup,
		refGroup,
	)

	var topRow fyne.CanvasObject = container.NewBorder(
		nil, nil,
		fixedWrap(sampleRateSelect, 100), nil,
		tempRefRow,
	)
	if isMobile {
		// Sample rate gets its own row, laid out like the Sample Length tab
		topRow = container.NewVBox(
			tempRefRow,
			container.NewGridWithColumns(2,
				widget.NewLabel("Sample Rate"),
				sampleRateSelect,
			),
		)
	}

	// Emphasize the add button and make it wider
	addButton.Importance = widget.HighImportance

	fixedAddButton := container.NewGridWrap(
		fyne.NewSize(50, micNameSelect.MinSize().Height),
		addButton,
	)

	micRow := container.NewBorder(
		nil, nil, nil,
		fixedAddButton,
		container.NewBorder(nil, nil, nil, fixedWrap(targetUnitSelect, 70),
			container.NewGridWithColumns(2,
				micNameSelect,
				targetDistEntry,
			),
		),
	)

	// The inputs stay on top; the mic cards fill and scroll in the remaining space.
	// A little inset keeps the card shadows from being clipped by the scroller.
	content := container.NewBorder(
		container.NewVBox(
			topRow,
			widget.NewSeparator(),
			micRow,
		),
		nil, nil, nil,
		container.NewVScroll(container.New(layout.NewCustomPaddedLayout(2, 4, 2, 2), micList)),
	)

	return content
}
