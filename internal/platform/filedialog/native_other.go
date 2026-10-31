//go:build !windows && !(darwin && !ios && cgo) && !((linux || freebsd || openbsd || netbsd) && !android)

package filedialog

import "fyne.io/fyne/v2"

// On Android and iOS, Fyne's file dialog already opens the system picker; elsewhere it is the
// only dialog there is.

func nativeOpen(fyne.Window, string, Filter, func([]byte, error)) bool { return false }

func nativeSave(fyne.Window, string, string, Filter, []byte, func(bool, error)) bool { return false }
