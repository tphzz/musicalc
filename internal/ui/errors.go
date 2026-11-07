package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// showErrorDialog shows err in an error dialog on the app window. It is for problems found outside
// a dialog flow that has its own window, such as an unreadable stored setup or a failed test tone.
// The app has one window; the most recently created one is used, as a driver may list others first.
func showErrorDialog(err error) {
	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) == 0 {
		return
	}
	dialog.ShowError(err, windows[len(windows)-1])
}
