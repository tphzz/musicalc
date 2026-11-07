package main

import (
	_ "embed"
	"fmt"
	"io"
	"musicalc/internal/i18n"
	"musicalc/internal/platform/displayscale"
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

// appIcon is the window and taskbar icon. It is embedded because desktop builds are plain
// go builds without packaged metadata, and the app may start from any working directory.
//
//go:embed icons/appicon.png
var appIcon []byte

const (
	windowWidth  = 450
	windowHeight = 650
)

// CategoryInfo represents a category with tab indices
type CategoryInfo struct {
	Name       string
	TabIndices []int
}

// tabInfo describes one calculator tab; its content is built the first time it is shown
type tabInfo struct {
	text    string // message key for the desktop tab; mobile shows icons only
	heading string // message key for the full calculator heading
	icon    fyne.Resource
	build   func() *ui.View
	help    string // help page, see internal/ui/help
}

// createCompactHeader creates a compact header with: hamburger + "category / tab heading", and on
// the right the overflow button (⋮), always shown: the tab's own actions, if any, then Help
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
	if err := displayscale.Configure(); err != nil {
		fyne.LogError("Could not configure desktop scaling", err)
	}
	myApp := app.NewWithID("com.musicalc")
	setAppIcon(myApp)
	ui.InitializeLanguage(myApp.Preferences())
	myApp.Settings().SetTheme(ui.NewCustomTheme())

	var scale *displayscale.Controller
	if !fyne.CurrentDevice().IsMobile() {
		var err error
		scale, err = displayscale.NewController(myApp)
		if err != nil {
			fyne.LogError("Could not configure desktop zoom", err)
		}
	}
	window := buildMainWindow(myApp, scale)
	stopScaleUpdates := func() {}
	if scale != nil {
		stopScaleUpdates = scale.Follow(window, fyne.NewSize(windowWidth, windowHeight))
		myApp.Lifecycle().SetOnEnteredForeground(func() { fyne.Do(scale.Refresh) })
	}
	window.SetOnClosed(stopScaleUpdates)
	defer stopScaleUpdates()
	window.ShowAndRun()
}

// setAppIcon gives the app its embedded icon; windows without their own icon use it
func setAppIcon(a fyne.App) {
	a.SetIcon(fyne.NewStaticResource("appicon.png", appIcon))
}

func buildMainWindow(myApp fyne.App, scale *displayscale.Controller) fyne.Window {
	window := myApp.NewWindow("MusiCalc")
	window.Resize(fyne.NewSize(windowWidth, windowHeight))
	window.SetFixedSize(true)

	tabs := []tabInfo{
		{"tab.timecode", "title.timecode", ui.ResourceTimecodeSvg, ui.NewTimecodeTab, "timecode"},
		{"tab.delay", "title.delay", ui.ResourceDelaySvg, ui.NewTempoTab, "tempo-to-delay"},
		{"tab.tempochange", "title.tempochange", ui.ResourceTempochangeSvg, ui.NewTempoChangeTab, "tempo-change"},
		{"tab.notefreq", "title.notefreq", ui.ResourceNote2freqSvg, ui.NewDiapasonTab, "note-to-frequency"},
		{"tab.freqnote", "title.freqnote", ui.ResourceFreq2noteSvg, ui.NewFrequencyToNoteTab, "frequency-to-note"},
		{"tab.samples", "title.samples", ui.ResourceSamplelengthSvg, ui.NewSampleLengthTab, "sample-length"},
		{"tab.alignment", "title.alignment", ui.ResourceAlignmentdelaySvg, ui.NewAlignmentDelayTab, "alignment-delay"},
		{"tab.tap", "title.tap", ui.ResourceTaptempoSvg, nil, "tap-tempo"}, // built below: it needs the tab bar
		{"tab.sum", "title.sum", ui.ResourcePhaseSvg, ui.NewTwoMicSumTab, "two-mic-sum"},
		{"tab.three", "title.three", ui.ResourceThreetooneSvg, ui.NewThreeToOneTab, "three-to-one"},
	}
	// Mobile shows icons only (larger, in fixed-width tabs)
	isMobile := fyne.CurrentDevice().IsMobile()
	about := ui.NewAboutDialog(ui.DisplayVersion(version, myApp.Metadata(), isMobile), window)
	var preferencesDialog *ui.PreferencesDialog
	showAbout := func() { preferencesDialog.Hide(); about.Show() }
	showPreferences := func() { about.Hide(); preferencesDialog.Show() }
	nativeAboutMenu := runtime.GOOS == "darwin" && !isMobile
	if nativeAboutMenu {
		// Fyne requires the literal About role ID; the OS owns the native menu title.
		window.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("MusiCalc", fyne.NewMenuItem("About", showAbout))))
	}

	categories := []CategoryInfo{
		{Name: i18n.T("category.time"), TabIndices: []int{0, 1, 2, 7}},
		{Name: i18n.T("category.pitch"), TabIndices: []int{3, 4}},
		{Name: i18n.T("category.analysis"), TabIndices: []int{5}},
		{Name: i18n.T("category.mics"), TabIndices: []int{6, 8, 9}},
	}

	// Always start with Time & Tempo category (Timecode Calculator)
	currentCategoryIndex := 0

	headerLabel := widget.NewLabel("")
	content := container.NewStack()
	tabBar := widgets.NewTabBar()
	built := make([]*ui.View, len(tabs))
	shown := 0 // global index of the tab on screen

	// Help replaces the whole window content until it is closed; the shell, with every built
	// tab and its values, comes back unchanged
	var shell fyne.CanvasObject
	helpOpen := false
	var helpView *ui.View
	closeHelp := func() {
		if helpOpen {
			helpOpen = false
			window.SetContent(shell)
		}
	}
	showHelp := func() {
		t := tabs[shown]
		helpOpen = true
		helpView = ui.NewHelpView(t.help, t.heading, closeHelp)
		window.SetContent(helpView.Content)
	}

	// Overflow button (⋮): the shown tab's own actions, then Help
	var actionsButton *widgets.Button
	actionsButton = widgets.NewButtonWithIcon("", theme.MoreVerticalIcon(), func() {
		items := ui.TabMenuItems(built[shown].Content)
		if len(items) > 0 {
			items = append(items, fyne.NewMenuItemSeparator())
		}
		items = append(items, fyne.NewMenuItem(i18n.T("menu.help"), showHelp))
		showTabMenu(window, actionsButton, fyne.NewMenu("", items...))
	})
	actionsButton.Importance = widget.LowImportance

	// Tap Tempo's Delay button sends the tempo to Tempo to Delay (index 1 in Time & Tempo) and opens it
	tabs[7].build = func() *ui.View {
		return ui.NewTapTempoTab(func() { tabBar.Select(1) })
	}

	// showTab shows a tab by its global index, building its content on first use.
	// Built tabs are kept, so their inputs survive switching away and back.
	showTab := func(index int) {
		if built[index] == nil {
			built[index] = tabs[index].build()
		}
		shown = index
		content.Objects = []fyne.CanvasObject{built[index].Content}
		content.Refresh()
		headerLabel.SetText(categories[currentCategoryIndex].Name + " / " + i18n.T(tabs[index].heading))
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
				items[i].Text = i18n.T(tabs[tabIdx].text)
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

	refreshLanguage := func() {
		for i, key := range []string{"category.time", "category.pitch", "category.analysis", "category.mics"} {
			categories[i].Name = i18n.T(key)
		}
		for _, view := range built {
			if view != nil {
				view.RefreshText()
			}
		}
		indices := categories[currentCategoryIndex].TabIndices
		labels := make([]string, len(indices))
		if !isMobile {
			for i, index := range indices {
				labels[i] = i18n.T(tabs[index].text)
			}
		}
		tabBar.SetLabels(labels)
		headerLabel.SetText(categories[currentCategoryIndex].Name + " / " + i18n.T(tabs[shown].heading))
		about.RefreshText()
		if helpOpen && helpView != nil {
			helpView.RefreshText()
		}
		shell.Refresh()
	}
	preferencesDialog = ui.NewPreferencesDialog(myApp.Preferences(), window, scale, refreshLanguage)

	var categoryAbout func()
	if !nativeAboutMenu {
		categoryAbout = showAbout
	}
	header := createCompactHeader(func() {
		showCategoryMenu(window, categories, switchCategory, showPreferences, categoryAbout)
	}, headerLabel, actionsButton)
	// Keys typed while no field has focus: F1 opens help, Esc and the Android back button close
	// it; other keys go to the shown tab (e.g. Space taps in Tap Tempo)
	window.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		switch {
		case about.HandleKey(ev):
		case preferencesDialog.HandleKey(ev):
		case helpOpen && (ev.Name == fyne.KeyEscape || ev.Name == mobile.KeyBack):
			closeHelp()
		case helpOpen:
		case ev.Name == fyne.KeyF1:
			showHelp()
		case ui.HandleTabKey(built[shown].Content, ev):
		case ev.Name == mobile.KeyBack:
			// What Android's back button does without our key handler: leave the app
			if d, ok := fyne.CurrentApp().Driver().(interface{ GoBack() }); ok {
				d.GoBack()
			}
		}
	})
	var calculatorBody fyne.CanvasObject = content
	var tabFooter fyne.CanvasObject = tabBar
	if !isMobile {
		calculatorBody = container.NewScroll(content)
		tabFooter = container.NewHScroll(tabBar)
	}
	shell = ui.NewShell(
		container.NewVBox(header, widget.NewSeparator()),
		calculatorBody,
		container.NewVBox(widget.NewSeparator(), tabFooter),
	)
	window.SetContent(shell)
	return window
}

// showCategoryMenu displays a popup menu for category selection
func showCategoryMenu(window fyne.Window, categories []CategoryInfo, onSelect func(int), showPreferences, showAbout func()) *widget.PopUpMenu {
	popupMenu := widget.NewPopUpMenu(categoryMenu(categories, onSelect, showPreferences, showAbout), window.Canvas())
	popupMenu.Resize(fyne.NewSize(fyne.Min(240, fyne.Max(0, window.Canvas().Size().Width-20)), popupMenu.MinSize().Height))
	popupMenu.ShowAtPosition(fyne.NewPos(10, 50))
	return popupMenu
}

func categoryMenu(categories []CategoryInfo, onSelect func(int), showPreferences, showAbout func()) *fyne.Menu {
	var menuItems []*fyne.MenuItem

	for i, category := range categories {
		index := i
		menuItems = append(menuItems, fyne.NewMenuItem(category.Name, func() {
			onSelect(index)
		}))
	}

	if len(categories) > 0 && (showPreferences != nil || showAbout != nil) {
		menuItems = append(menuItems, fyne.NewMenuItemSeparator())
	}

	if showPreferences != nil {
		menuItems = append(menuItems, fyne.NewMenuItem(i18n.T("menu.preferences"), showPreferences))
	}
	if showAbout != nil {
		menuItems = append(menuItems, fyne.NewMenuItem(i18n.T("menu.about"), showAbout))
	}
	return fyne.NewMenu(i18n.T("menu.categories"), menuItems...)
}
