package ui

import "fyne.io/fyne/v2"

// Tabs can offer extra actions in an overflow menu (⋮) on the right of the header, like the
// Alignment Delay presets. main.go shows the button only for tabs that registered a menu.
var tabMenus = map[fyne.CanvasObject]func() *fyne.Menu{}

// setTabMenu registers the overflow menu of a tab's content. The menu is built each time it
// opens, so enabled and disabled items are always current.
func setTabMenu(tab fyne.CanvasObject, menu func() *fyne.Menu) {
	tabMenus[tab] = menu
}

// HasTabMenu reports whether the tab has an overflow menu
func HasTabMenu(tab fyne.CanvasObject) bool {
	_, ok := tabMenus[tab]
	return ok
}

// TabMenu builds the tab's overflow menu, or returns nil when it has none
func TabMenu(tab fyne.CanvasObject) *fyne.Menu {
	if menu, ok := tabMenus[tab]; ok {
		return menu()
	}
	return nil
}
