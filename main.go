package main

import (
	"musicalc/internal/ui"
	"musicalc/internal/ui/widgets"
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

// tabInfo describes one calculator tab; its content is built the first time it is shown
type tabInfo struct {
	text    string // tab label on desktop; mobile shows icons only
	heading string
	icon    fyne.Resource
	build   func() fyne.CanvasObject
}

// createCompactHeader creates a compact header with: hamburger + "category / tab heading"
func createCompactHeader(window fyne.Window, categories []CategoryInfo, switchCategory func(int), headerLabel *widget.Label) *fyne.Container {
	menuButton := widgets.NewButtonWithIcon("", theme.MenuIcon(), func() {
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

	tabs := []tabInfo{
		{"Timecode", "Timecode Calculator", ui.ResourceTimecodeSvg, ui.NewTimecodeTab},
		{"Delay", "Tempo to Delay", ui.ResourceDelaySvg, ui.NewTempoTab},
		{"Tempo Chg", "Tempo Change", ui.ResourceTempochangeSvg, ui.NewTempoChangeTab},
		{"Note→Freq", "Note to Frequency", ui.ResourceNote2freqSvg, ui.NewDiapasonTab},
		{"Freq→Note", "Frequency to Note", ui.ResourceFreq2noteSvg, ui.NewFrequencyToNoteTab},
		{"Sample Len", "Sample Length", ui.ResourceSamplelengthSvg, ui.NewSampleLengthTab},
		{"Align Dly", "Alignment Delay", ui.ResourceAlignmentdelaySvg, ui.NewAlignmentDelayTab},
		{"Tap", "Tap Tempo", ui.ResourceTaptempoSvg, nil}, // built below: it needs the tab bar
	}
	// Mobile shows icons only (larger, in fixed-width tabs)
	isMobile := fyne.CurrentDevice().IsMobile()

	categories := []CategoryInfo{
		{Name: "Time & Tempo", TabIndices: []int{0, 1, 2, 7}},
		{Name: "Frequency & Pitch", TabIndices: []int{3, 4}},
		{Name: "Analysis", TabIndices: []int{5}},
		{Name: "Multi-Mic", TabIndices: []int{6}},
	}

	// Always start with Time & Tempo category (Timecode Calculator)
	currentCategoryIndex := 0

	headerLabel := widget.NewLabel("")
	content := container.NewStack()
	tabBar := widgets.NewTabBar()
	built := make([]fyne.CanvasObject, len(tabs))
	shown := 0 // global index of the tab on screen

	// Tap Tempo's Delay button sends the tempo to Tempo to Delay (index 1 in Time & Tempo) and opens it
	tabs[7].build = func() fyne.CanvasObject {
		return ui.NewTapTempoTab(func() { tabBar.Select(1) })
	}

	// showTab shows a tab by its global index, building its content on first use.
	// Built tabs are kept, so their inputs survive switching away and back.
	showTab := func(index int) {
		if built[index] == nil {
			built[index] = tabs[index].build()
		}
		shown = index
		content.Objects = []fyne.CanvasObject{built[index]}
		content.Refresh()
		headerLabel.SetText(categories[currentCategoryIndex].Name + " / " + tabs[index].heading)
	}

	tabBar.OnSelected = func(i int) {
		showTab(categories[currentCategoryIndex].TabIndices[i])
	}

	// showCategoryTabs shows the tabs of a category and selects its first tab
	showCategoryTabs := func(categoryIdx int) {
		cat := categories[categoryIdx]
		items := make([]widgets.TabBarItem, len(cat.TabIndices))
		for i, tabIdx := range cat.TabIndices {
			items[i] = widgets.TabBarItem{Icon: tabs[tabIdx].icon}
			if !isMobile {
				items[i].Text = tabs[tabIdx].text
			}
		}
		tabBar.SetItems(items, 0)
		showTab(cat.TabIndices[0])
	}

	showCategoryTabs(currentCategoryIndex)

	switchCategory := func(index int) {
		if index < 0 || index >= len(categories) || index == currentCategoryIndex {
			return
		}
		currentCategoryIndex = index
		showCategoryTabs(index)
		myApp.Preferences().SetInt("lastCategory", index)
	}

	header := createCompactHeader(window, categories, switchCategory, headerLabel)
	// Keys typed while no field has focus go to the shown tab (e.g. Space taps in Tap Tempo)
	window.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		ui.HandleTabKey(built[shown], ev)
	})
	window.SetContent(ui.NewShell(
		container.NewVBox(header, widget.NewSeparator()),
		content,
		container.NewVBox(widget.NewSeparator(), tabBar),
	))
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
