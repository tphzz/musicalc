//go:build !windows

package displayscale

import "fyne.io/fyne/v2"

func fitWindowOnScreen(_ fyne.Window, requested fyne.Size) fyne.Size { return requested }
