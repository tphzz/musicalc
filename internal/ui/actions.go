package ui

import "fyne.io/fyne/v2"

// Every tab has an overflow menu (⋮) on the right of the header. Tabs can add their own actions
// to it, like the Alignment Delay presets; main.go adds Help as the last entry for every tab.
var tabMenus = map[fyne.CanvasObject]func() []*fyne.MenuItem{}

// setTabMenu registers a tab's own menu items. They are built each time the menu opens, so
// enabled and disabled items are always current.
func setTabMenu(tab fyne.CanvasObject, items func() []*fyne.MenuItem) {
	tabMenus[tab] = items
}

// TabMenuItems returns the tab's own menu items, or none
func TabMenuItems(tab fyne.CanvasObject) []*fyne.MenuItem {
	if items, ok := tabMenus[tab]; ok {
		return items()
	}
	return nil
}
