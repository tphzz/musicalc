package main

import (
	"musicalc/internal/ui"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var version = "dev"

// CategoryInfo represents a category with tab indices
type CategoryInfo struct {
	Name       string
	TabIndices []int
}

// createCompactHeader creates a compact header with: hamburger + "category / tab heading"
func createCompactHeader(window fyne.Window, categories []CategoryInfo, switchCategory func(int), headerLabel *widget.Label) *fyne.Container {
	menuButton := widget.NewButtonWithIcon("", theme.MenuIcon(), func() {
		showCategoryMenu(window, categories, switchCategory)
	})
	menuButton.Importance = widget.LowImportance
	// Truncate long "category / tab" headings instead of overflowing narrow phone screens
	headerLabel.Truncation = fyne.TextTruncateEllipsis
	// Spacers center the label vertically next to the (taller) menu button while it keeps the full width
	return container.NewBorder(nil, nil, menuButton, nil,
		container.NewVBox(layout.NewSpacer(), headerLabel, layout.NewSpacer()))
}

func main() {
	myApp := app.NewWithID("com.musicalc")
	myApp.Settings().SetTheme(ui.NewCustomTheme())

	ver := strings.TrimSpace(version)
	windowTitle := "MusiCalc"
	if ver != "" {
		windowTitle = "MusiCalc v" + ver
	}
	window := myApp.NewWindow(windowTitle)
	window.Resize(fyne.NewSize(450, 650))
	window.SetFixedSize(true)

	if icon, err := fyne.LoadResourceFromPath("icons/appicon.png"); err == nil {
		window.SetIcon(icon)
	}

	tabHeadings := map[string]string{
		"timecode":     "Timecode Calculator",
		"tempo":        "Tempo to Delay",
		"tempochange":  "Tempo Change",
		"note2freq":    "Note to Frequency",
		"freq2note":    "Frequency to Note",
		"samplelength": "Sample Length",
		"alignment":    "Alignment Delay",
	}

	var switchCategory func(int)

	// Determine tab text based on device type
	isMobile := fyne.CurrentDevice().IsMobile()
	var timecodeText, tempoText, tempoChangeText, note2freqText, freq2noteText, sampleLengthText, alignmentText string
	if !isMobile {
		timecodeText = "Timecode"
		tempoText = "Delay"
		tempoChangeText = "Tempo Chg"
		note2freqText = "Note→Freq"
		freq2noteText = "Freq→Note"
		sampleLengthText = "Sample Len"
		alignmentText = "Align Dly"
	}

	// Create all tab items with empty placeholders; content is built on first display
	// so startup only pays for the tab that is actually shown
	timecodeTab := container.NewTabItemWithIcon(timecodeText, ui.ResourceTimecodeSvg, container.NewStack())
	tempoTab := container.NewTabItemWithIcon(tempoText, ui.ResourceDelaySvg, container.NewStack())
	tempoChangeTab := container.NewTabItemWithIcon(tempoChangeText, ui.ResourceTempochangeSvg, container.NewStack())
	note2freqTab := container.NewTabItemWithIcon(note2freqText, ui.ResourceNote2freqSvg, container.NewStack())
	freq2noteTab := container.NewTabItemWithIcon(freq2noteText, ui.ResourceFreq2noteSvg, container.NewStack())
	sampleLengthTab := container.NewTabItemWithIcon(sampleLengthText, ui.ResourceSamplelengthSvg, container.NewStack())
	alignmentTab := container.NewTabItemWithIcon(alignmentText, ui.ResourceAlignmentdelaySvg, container.NewStack())

	// Create single AppTabs with ALL tabs (maintains left alignment)
	allTabs := []*container.TabItem{
		timecodeTab, tempoTab, tempoChangeTab,
		note2freqTab, freq2noteTab,
		sampleLengthTab,
		alignmentTab,
	}

	// Content builders for each global tab index
	tabBuilders := []func() fyne.CanvasObject{
		ui.NewTimecodeTab, ui.NewTempoTab, ui.NewTempoChangeTab,
		ui.NewDiapasonTab, ui.NewFrequencyToNoteTab,
		ui.NewSampleLengthTab,
		ui.NewAlignmentDelayTab,
	}
	tabBuilt := make([]bool, len(allTabs))

	// ensureBuilt creates a tab's content the first time it is shown; returns true if it was built now
	ensureBuilt := func(item *container.TabItem) bool {
		for i, tab := range allTabs {
			if tab == item {
				if tabBuilt[i] {
					return false
				}
				tabBuilt[i] = true
				tab.Content = tabBuilders[i]()
				return true
			}
		}
		return false
	}
	ensureBuilt(timecodeTab)

	tabs := container.NewAppTabs(allTabs...)
	tabs.SetTabLocation(container.TabLocationBottom)

	// Define categories with their tab indices
	categories := []CategoryInfo{
		{Name: "Time & Tempo", TabIndices: []int{0, 1, 2}},
		{Name: "Frequency & Pitch", TabIndices: []int{3, 4}},
		{Name: "Analysis", TabIndices: []int{5}},
		{Name: "Multi-Mic", TabIndices: []int{6}},
	}

	// Tab heading keys for each global tab index
	tabHeadingKeys := []string{
		"timecode", "tempo", "tempochange",
		"note2freq", "freq2note",
		"samplelength",
		"alignment",
	}

	// Always start with Time & Tempo category (Timecode Calculator)
	currentCategoryIndex := 0

	headerLabel := widget.NewLabel("")

	// Helper to show only tabs for a specific category
	showCategoryTabs := func(categoryIdx int) {
		if categoryIdx < 0 || categoryIdx >= len(categories) {
			return
		}

		cat := categories[categoryIdx]

		// Build list of visible tabs for this category
		var visibleTabs []*container.TabItem
		for _, tabIdx := range cat.TabIndices {
			if tabIdx >= 0 && tabIdx < len(allTabs) {
				visibleTabs = append(visibleTabs, allTabs[tabIdx])
			}
		}

		// Replace tabs.Items with visible tabs
		tabs.Items = visibleTabs

		// Select first tab in category
		if len(visibleTabs) > 0 {
			ensureBuilt(visibleTabs[0])
			tabs.Select(visibleTabs[0])
		}

		tabs.Refresh()
	}

	// Helper to update header text based on selected tab
	updateHeader := func() {
		selectedTab := tabs.Selected()
		if selectedTab == nil {
			return
		}

		// Find which tab is selected
		for i, tab := range allTabs {
			if tab == selectedTab {
				// Find category for this tab
				for _, cat := range categories {
					for _, tabIdx := range cat.TabIndices {
						if tabIdx == i {
							headerText := cat.Name + " / " + tabHeadings[tabHeadingKeys[i]]
							headerLabel.SetText(headerText)
							return
						}
					}
				}
			}
		}
	}

	// Set initial category tabs visibility
	showCategoryTabs(currentCategoryIndex)
	updateHeader()

	// Update header when tab changes
	tabs.OnSelected = func(item *container.TabItem) {
		if ensureBuilt(item) {
			tabs.Refresh()
		}
		updateHeader()
	}

	// Function to switch categories
	switchCategory = func(index int) {
		if index >= 0 && index < len(categories) {
			if index == currentCategoryIndex {
				return
			}

			currentCategoryIndex = index
			showCategoryTabs(index)
			updateHeader()

			myApp.Preferences().SetInt("lastCategory", index)
		}
	}

	// Create initial header and content
	header := createCompactHeader(window, categories, switchCategory, headerLabel)

	content := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		nil, nil, nil,
		tabs,
	)
	window.SetContent(content)
	window.ShowAndRun()
}

// showCategoryMenu displays a popup menu for category selection
func showCategoryMenu(window fyne.Window, categories []CategoryInfo, onSelect func(int)) {
	var menuItems []*fyne.MenuItem

	for i, category := range categories {
		index := i
		menuItems = append(menuItems, fyne.NewMenuItem(category.Name, func() {
			onSelect(index)
		}))
	}

	menu := fyne.NewMenu("Categories", menuItems...)
	popupMenu := widget.NewPopUpMenu(menu, window.Canvas())
	popupMenu.ShowAtPosition(fyne.NewPos(10, 50))
}
