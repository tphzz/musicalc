package ui

import "fyne.io/fyne/v2"

// Tabs can react to keys typed while no field has focus (the window canvas only receives keys
// then, so typing into a field never triggers them). main.go forwards the canvas's key events
// to the tab that is shown.
var tabKeyHandlers = map[fyne.CanvasObject]func(*fyne.KeyEvent) bool{}

// setTabKeyHandler registers the key handler of a tab's content
func setTabKeyHandler(tab fyne.CanvasObject, handler func(*fyne.KeyEvent) bool) {
	tabKeyHandlers[tab] = handler
}

// HandleTabKey passes a key event to the shown tab's handler; it reports whether it was used
func HandleTabKey(tab fyne.CanvasObject, ev *fyne.KeyEvent) bool {
	if handler, ok := tabKeyHandlers[tab]; ok {
		return handler(ev)
	}
	return false
}
