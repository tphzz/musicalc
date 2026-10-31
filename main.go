package main

import (
	"fmt"
	"io"
	"musicalc/internal/ui"
	"musicalc/internal/ui/widgets"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
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
	help    string // help page, see internal/ui/help
}

// createCompactHeader creates a compact header with: hamburger + "category / tab heading", and on
// the right the overflow button (⋮) for tabs with extra actions
func createCompactHeader(showMenu func(), headerLabel *widget.Label, actionsButton fyne.CanvasObject) *fyne.Container {
	menuButton := widgets.NewButtonWithIcon("", theme.MenuIcon(), showMenu)
	menuButton.Importance = widget.LowImportance
	// Truncate long "category / tab" headings instead of overflowing narrow phone screens
	headerLabel.Truncation = fyne.TextTruncateEllipsis
	// Spacers center the label vertically next to the (taller) menu button while it keeps the full width
	return container.NewBorder(nil, nil, menuButton, actionsButton,
		container.NewVBox(layout.NewSpacer(), headerLabel, layout.NewSpacer()))
}

// showTabMenu opens the shown tab's overflow menu under the button, right-aligned with it
func showTabMenu(window fyne.Window, button fyne.CanvasObject, menu *fyne.Menu) {
	if menu == nil {
		return
	}
	popUp := widget.NewPopUpMenu(menu, window.Canvas())
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(button)
	x := pos.X + button.Size().Width - popUp.MinSize().Width
	popUp.ShowAtPosition(fyne.NewPos(fyne.Max(x, 0), pos.Y+button.Size().Height))
}

func writeVersion(args []string, out io.Writer) bool {
	if len(args) != 1 || args[0] != "--version" {
		return false
	}
	fmt.Fprintln(out, ui.DisplayVersion(version, fyne.AppMetadata{}, false))
	return true
}

func main() {
	if writeVersion(os.Args[1:], os.Stdout) {
		return
	}
	myApp := app.NewWithID("com.musicalc")
	myApp.Settings().SetTheme(ui.NewCustomTheme())

	window := myApp.NewWindow("MusiCalc")
	window.Resize(fyne.NewSize(450, 650))
	window.SetFixedSize(true)

	if icon, err := fyne.LoadResourceFromPath("icons/appicon.png"); err == nil {
		window.SetIcon(icon)
	}

	tabs := []tabInfo{
		{"Timecode", "Timecode Calculator", ui.ResourceTimecodeSvg, ui.NewTimecodeTab, "timecode"},
		{"Delay", "Tempo to Delay", ui.ResourceDelaySvg, ui.NewTempoTab, "tempo-to-delay"},
		{"Tempo Chg", "Tempo Change", ui.ResourceTempochangeSvg, ui.NewTempoChangeTab, "tempo-change"},
		{"Note→Freq", "Note to Frequency", ui.ResourceNote2freqSvg, ui.NewDiapasonTab, "note-to-frequency"},
		{"Freq→Note", "Frequency to Note", ui.ResourceFreq2noteSvg, ui.NewFrequencyToNoteTab, "frequency-to-note"},
		{"Sample Len", "Sample Length", ui.ResourceSamplelengthSvg, ui.NewSampleLengthTab, "sample-length"},
		{"Align Dly", "Alignment Delay", ui.ResourceAlignmentdelaySvg, ui.NewAlignmentDelayTab, "alignment-delay"},
		{"Tap", "Tap Tempo", ui.ResourceTaptempoSvg, nil, "tap-tempo"}, // built below: it needs the tab bar
		{"Sum", "Two-Mic Sum", ui.ResourcePhaseSvg, ui.NewTwoMicSumTab, "two-mic-sum"},
		{"3:1", "3:1 Rule", ui.ResourceThreetooneSvg, ui.NewThreeToOneTab, "three-to-one"},
	}
	// Mobile shows icons only (larger, in fixed-width tabs)
	isMobile := fyne.CurrentDevice().IsMobile()
	about := ui.NewAboutDialog(ui.DisplayVersion(version, myApp.Metadata(), isMobile), window)
	nativeAboutMenu := runtime.GOOS == "darwin" && !isMobile
	if nativeAboutMenu {
		// Fyne moves the About role into the native macOS application menu.
		window.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("MusiCalc", fyne.NewMenuItem("About", about.Show))))
	}

	categories := []CategoryInfo{
		{Name: "Time & Tempo", TabIndices: []int{0, 1, 2, 7}},
		{Name: "Frequency & Pitch", TabIndices: []int{3, 4}},
		{Name: "Analysis", TabIndices: []int{5}},
		{Name: "Multi-Mic", TabIndices: []int{6, 8, 9}},
	}

	// Always start with Time & Tempo category (Timecode Calculator)
	currentCategoryIndex := 0

	headerLabel := widget.NewLabel("")
	content := container.NewStack()
	tabBar := widgets.NewTabBar()
	built := make([]fyne.CanvasObject, len(tabs))
	shown := 0 // global index of the tab on screen

	// Help replaces the whole window content until it is closed; the shell, with every built
	// tab and its values, comes back unchanged
	var shell fyne.CanvasObject
	helpOpen := false
	closeHelp := func() {
		if helpOpen {
			helpOpen = false
			window.SetContent(shell)
		}
	}
	showHelp := func() {
		t := tabs[shown]
		helpOpen = true
		window.SetContent(ui.NewHelpView(t.help, t.heading, closeHelp))
	}

	// Overflow button (⋮): the shown tab's own actions, then Help
	var actionsButton *widgets.Button
	actionsButton = widgets.NewButtonWithIcon("", theme.MoreVerticalIcon(), func() {
		items := ui.TabMenuItems(built[shown])
		if len(items) > 0 {
			items = append(items, fyne.NewMenuItemSeparator())
		}
		items = append(items, fyne.NewMenuItem("Help", showHelp))
		showTabMenu(window, actionsButton, fyne.NewMenu("", items...))
	})
	actionsButton.Importance = widget.LowImportance

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

	var categoryAbout func()
	if !nativeAboutMenu {
		categoryAbout = about.Show
	}
	header := createCompactHeader(func() {
		showCategoryMenu(window, categories, switchCategory, categoryAbout)
	}, headerLabel, actionsButton)
	// Keys typed while no field has focus: F1 opens help, Esc and the Android back button close
	// it; other keys go to the shown tab (e.g. Space taps in Tap Tempo)
	window.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		switch {
		case about.HandleKey(ev):
		case helpOpen && (ev.Name == fyne.KeyEscape || ev.Name == mobile.KeyBack):
			closeHelp()
		case helpOpen:
		case ev.Name == fyne.KeyF1:
			showHelp()
		case ui.HandleTabKey(built[shown], ev):
		case ev.Name == mobile.KeyBack:
			// What Android's back button does without our key handler: leave the app
			if d, ok := fyne.CurrentApp().Driver().(interface{ GoBack() }); ok {
				d.GoBack()
			}
		}
	})
	shell = ui.NewShell(
		container.NewVBox(header, widget.NewSeparator()),
		content,
		container.NewVBox(widget.NewSeparator(), tabBar),
	)
	window.SetContent(shell)
	window.ShowAndRun()
}

// showCategoryMenu displays a popup menu for category selection
func showCategoryMenu(window fyne.Window, categories []CategoryInfo, onSelect func(int), showAbout func()) {
	popupMenu := widget.NewPopUpMenu(categoryMenu(categories, onSelect, showAbout), window.Canvas())
	popupMenu.ShowAtPosition(fyne.NewPos(10, 50))
}

func categoryMenu(categories []CategoryInfo, onSelect func(int), showAbout func()) *fyne.Menu {
	var menuItems []*fyne.MenuItem

	for i, category := range categories {
		index := i
		menuItems = append(menuItems, fyne.NewMenuItem(category.Name, func() {
			onSelect(index)
		}))
	}

	if showAbout != nil {
		menuItems = append(menuItems, fyne.NewMenuItem("About MusiCalc", showAbout))
	}
	return fyne.NewMenu("Categories", menuItems...)
}
