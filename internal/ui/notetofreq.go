package ui

import (
	"fmt"
	"image/color"
	"musicalc/internal/i18n"
	"musicalc/internal/logic"
	sclres "musicalc/internal/logic/scl"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// noValue marks an unavailable table value, as the MIDI column does above note 127
const noValue = "-"

// Column proportions of the table. A scale without 12 steps shows "degree (period)" and cents of up
// to "+1088.27" above the root, which need a wider Cents column than deviations from 12-TET.
var (
	noteColumns = []float32{0.18, 0.32, 0.22, 0.28}
	stepColumns = []float32{0.18, 0.30, 0.28, 0.24}
)

func NewDiapasonTab() *View {
	texts := newTextRefresh()
	// Reference frequency binding
	refFreq := binding.NewString()
	_ = refFreq.Set("440")

	// Reference note binding (default A3, MIDI note 69 in C3 naming)
	refNote := binding.NewString()
	_ = refNote.Set("A3")

	// The selected tuning and the pitch class its degree 0 is laid on (0 = C … 11 = B)
	tuningName := sclres.DefaultScaleName
	tonics := make(map[string]int)
	for _, t := range sclres.Tunings() {
		tonics[t.Name] = t.Tonic
	}
	root := tonics[tuningName]
	steps := 12 // degrees per period of the tuning, known even while the reference is invalid
	selectTuning := func(name string) {
		tuningName, root = name, tonics[name]
		n, err := logic.ScaleSteps(name)
		if err != nil {
			// Every bundled tuning parses (checked by the logic tests). Should one not, NewDiapason
			// fails the same way and the table shows "-" values under note names.
			n = 12
		}
		steps = n
	}
	selectTuning(tuningName)

	// Tuning options: one submenu per catalog group (Equal, Just, Historic, World), short names inside
	buildTuningGroups := func() []widgets.MenuGroup {
		var tuningGroups []widgets.MenuGroup
		lastGroup := ""
		for _, t := range sclres.Tunings() {
			if lastGroup != t.Group {
				lastGroup = t.Group
				tuningGroups = append(tuningGroups, widgets.MenuGroup{Name: i18n.T("tuning.group." + t.Group)})
			}
			g := &tuningGroups[len(tuningGroups)-1]
			g.Options = append(g.Options, widgets.MenuOption{Label: i18n.T("tuning.scale." + t.Short), Value: t.Name})
		}

		return tuningGroups
	}
	tuningGroups := buildTuningGroups()

	// Reference note selector (simple entry, no longer needed for note selection)
	refNoteEntry := widget.NewEntry()
	refNoteEntry.SetText("A3")
	refNoteEntry.OnChanged = func(s string) {
		_ = refNote.Set(s)
	}

	// Frequency input
	freqInput := widgets.NewNumericEntry()
	freqInput.SetText("440")
	texts.placeholder(freqInput, "field.frequency")
	freqInput.OnChanged = func(s string) {
		_ = refFreq.Set(s)
	}

	// Middle C convention selector (radio buttons)
	middleCRadio := widget.NewRadioGroup([]string{"C3", "C4"}, nil)
	middleCRadio.SetSelected("C3")
	middleCRadio.Horizontal = true
	middleCRadio.Required = true

	// Track previous Middle C selection to avoid adjusting on unchanged selections
	previousMiddleC := "C3"
	octaveOffset := func(convention string) int {
		if convention == "C4" {
			return logic.OctaveOffsetC4
		}
		return logic.OctaveOffsetC3
	}

	// Declare table variables first for use in the callbacks below
	var table *widget.Table
	var responsiveTableWidget *ResponsiveTable

	// The tuning laid out on the keys, rebuilt when an input changes rather than per table cell.
	// It is nil while the reference note or frequency is invalid; the table then shows no values.
	var diapason *logic.Diapason
	var cachedOctaveOffset int
	cachedRefMidi := -1 // -1 while the reference is invalid

	updateCache := func() {
		refFreqVal, _ := refFreq.Get()
		refNoteVal, _ := refNote.Get()

		cachedOctaveOffset = octaveOffset(middleCRadio.Selected)
		diapason, cachedRefMidi = nil, -1
		refMidi, err := logic.ParseNoteName(refNoteVal, cachedOctaveOffset)
		if err != nil {
			return
		}
		d, err := logic.NewDiapason(tuningName, root, refMidi, logic.ParseFloat(refFreqVal))
		if err != nil {
			return
		}
		diapason, cachedRefMidi = d, refMidi
	}

	// refreshTable recalculates the table and switches its columns between note names and degrees
	refreshTable := func() {
		updateCache()
		if table == nil || responsiveTableWidget == nil {
			return
		}
		if steps == 12 {
			responsiveTableWidget.SetProportions(noteColumns)
		} else {
			responsiveTableWidget.SetProportions(stepColumns)
		}
		table.Refresh()
	}

	// Scale root selector: C … B, the tuning's catalogue tonic marked as its default
	rootSelect := widget.NewSelect(nil, nil)
	rootOptions := func() []string {
		options := make([]string, 12)
		for pc := range options {
			options[pc] = logic.PitchClassName(pc)
			if pc == tonics[tuningName] {
				options[pc] = i18n.T("field.scale_default", options[pc])
			}
		}
		return options
	}
	// showRoot lists the options for the current tuning and selects root without a callback
	showRoot := func() {
		rootSelect.OnChanged = nil
		rootSelect.SetOptions(rootOptions())
		rootSelect.SetSelectedIndex(root)
		rootSelect.OnChanged = func(string) {
			root = rootSelect.SelectedIndex()
			refreshTable()
		}
	}
	showRoot()

	// Tuning selector (grouped dropdown with submenus; its values are the scale names). Picking a
	// tuning puts its root on the scale's own tonic.
	tuningSelect := widgets.NewMenuSelect(tuningGroups, func(name string) {
		selectTuning(name)
		showRoot()
		refreshTable()
	})
	tuningSelect.SetSelected(sclres.DefaultScaleName)

	// Initialize cache
	updateCache()

	// Reset button
	resetBtn := texts.buttonWithIcon("action.reset", theme.ViewRefreshIcon(), func() {
		// The convention goes first: switching it renames the current reference, which must not
		// shift the restored A3 by an octave
		middleCRadio.SetSelected("C3")
		selectTuning(sclres.DefaultScaleName)
		tuningSelect.SetSelected(tuningName)
		showRoot()
		_ = refNote.Set("A3")
		_ = refFreq.Set("440")
		refNoteEntry.SetText("A3")
		freqInput.SetText("440")
		refreshTable()
	})

	// Table displaying all notes from C-2 to B8 (132 rows; MIDI note numbers only go to 127)
	table = widget.NewTableWithHeaders(
		func() (int, int) { return 132, 4 },
		func() fyne.CanvasObject {
			// The rectangle highlights the reference note's row
			return container.NewStack(canvas.NewRectangle(color.Transparent), widget.NewLabel(""))
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			cell := o.(*fyne.Container)
			highlight := cell.Objects[0].(*canvas.Rectangle)
			l := cell.Objects[1].(*widget.Label)
			l.Alignment = fyne.TextAlignLeading

			// MIDI note starts at 0 (C-2)
			midiNote := id.Row

			var result logic.NoteFrequency
			if diapason != nil {
				result = diapason.Note(midiNote)
			}

			// Highlight the reference row; a bold first column marks each root (degree 0). In a
			// 12-step scale the root is a note name, so it is known without a valid reference.
			isReference := midiNote == cachedRefMidi
			if isReference {
				highlight.FillColor = theme.Color(theme.ColorNameSelection)
			} else {
				highlight.FillColor = color.Transparent
			}
			highlight.Refresh()
			isRoot := diapason != nil && result.Degree == 0
			if steps == 12 {
				isRoot = (midiNote-root)%12 == 0
			}
			l.TextStyle = fyne.TextStyle{Bold: isReference || (id.Col == 0 && isRoot)}

			switch id.Col {
			case 0:
				switch {
				case steps == 12:
					l.SetText(logic.NoteName(midiNote, cachedOctaveOffset))
				case diapason == nil:
					l.SetText(noValue) // the degree depends on the reference key
				default:
					l.SetText(fmt.Sprintf("%d (%d)", result.Degree, result.Period))
				}
			case 1:
				if diapason == nil {
					l.SetText(noValue)
					return
				}
				l.SetText(fmt.Sprintf("%.2f Hz", result.Frequency))
			case 2:
				if diapason == nil {
					l.SetText(noValue)
					return
				}
				// 12 steps: the deviation from 12-TET; otherwise the degree above its period's root
				cents := result.Cents
				if steps != 12 {
					cents = result.DegreeCents
				}
				if cents >= 0 {
					l.SetText(fmt.Sprintf("+%.2f", cents))
				} else {
					l.SetText(fmt.Sprintf("%.2f", cents))
				}
			case 3:
				// Show both MIDI conventions: standard (C4=60) / alternative (C3=60)
				// MIDI values only go 0-127, show "-" for invalid values
				midiStandard := midiNote
				midiAlternative := midiNote + 12

				standardStr := fmt.Sprintf("%d", midiStandard)
				if midiStandard > 127 {
					standardStr = noValue
				}

				alternativeStr := fmt.Sprintf("%d", midiAlternative)
				if midiAlternative > 127 {
					alternativeStr = noValue
				}

				l.SetText(fmt.Sprintf("%s / %s", standardStr, alternativeStr))
			}
		},
	)

	table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabel("")
	}
	table.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		l := o.(*widget.Label)

		// Hide row headers (left column)
		if id.Col == -1 {
			l.SetText("")
			return
		}

		l.TextStyle = fyne.TextStyle{Bold: true}
		l.Alignment = fyne.TextAlignLeading

		switch id.Col {
		case 0:
			if steps == 12 {
				l.SetText(i18n.T("field.note"))
			} else {
				l.SetText(i18n.T("field.step"))
			}
		case 1:
			l.SetText(i18n.T("field.frequency"))
		case 2:
			l.SetText(i18n.T("field.cents"))
		case 3:
			l.SetText(i18n.T("field.midi"))
		}
	}

	// Hide row header column
	table.ShowHeaderColumn = false

	// Wrap table in responsive container with proportional column widths.
	// minWidth stays below phone widths (~340) so the proportions also apply on Android
	responsiveTableWidget = NewResponsiveTable(table, noteColumns, 300, 0)

	// scrollToReference shows the reference note near the top with two rows of context above it.
	// Scrolling up to a row puts it at the top (behind the sticky header, hence 3 rows), so jump
	// to the bottom first.
	scrollToReference := func() {
		if cachedRefMidi < 0 {
			return // keep the position while the reference is being corrected
		}
		row := cachedRefMidi - 3
		if row < 0 {
			row = 0
		} else if row > 131 {
			row = 131
		}
		table.ScrollToBottom()
		table.ScrollTo(widget.TableCellID{Row: row, Col: 0})
	}
	responsiveTableWidget.OnFirstLayout = scrollToReference

	// Add listener for Middle C radio buttons
	middleCRadio.OnChanged = func(s string) {
		// Only adjust reference if Middle C selection actually changed
		if s == previousMiddleC {
			return
		}

		// Keep the same reference key and rename it in the new convention, e.g. A3 → A4.
		// An invalid reference is left as typed.
		if midi, err := logic.ParseNoteName(refNoteEntry.Text, octaveOffset(previousMiddleC)); err == nil {
			newRef := logic.NoteName(midi, octaveOffset(s))
			refNoteEntry.SetText(newRef)
			_ = refNote.Set(newRef)
		}

		// Update previous selection tracker
		previousMiddleC = s

		refreshTable()
		scrollToReference()
	}

	// Add listeners to refresh table when inputs change
	refFreq.AddListener(binding.NewDataListener(refreshTable))
	refNote.AddListener(binding.NewDataListener(func() {
		refreshTable()
		scrollToReference()
	}))

	// Build UI layout
	texts.add(func() { tuningSelect.SetGroups(buildTuningGroups()); showRoot(); table.Refresh() })

	return texts.view(container.NewBorder(
		container.NewVBox(
			container.NewGridWithColumns(2,
				texts.label("field.reference"),
				refNoteEntry,
			),
			container.NewGridWithColumns(2,
				texts.label("field.frequency_hz"),
				freqInput,
			),
			container.NewGridWithColumns(2,
				texts.label("field.middle_c"),
				middleCRadio,
			),
			// The tuning dropdown takes all the width Reset leaves, so its labels show in full
			container.NewBorder(nil, nil, nil, resetBtn, tuningSelect),
			container.NewGridWithColumns(2,
				texts.label("field.scale_root"),
				rootSelect,
			),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		responsiveTableWidget,
	))
}
