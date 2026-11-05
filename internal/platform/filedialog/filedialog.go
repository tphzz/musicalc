// Package filedialog opens and saves files with the platform's own file dialog: the Explorer
// dialog on Windows, NSOpenPanel/NSSavePanel on macOS, the desktop portal (GNOME, KDE, …) on
// Linux and BSD, and the system picker on Android and iOS. Where no native dialog is available,
// it falls back to Fyne's file dialog.
package filedialog

import (
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

// Filter limits the dialog to one kind of file
type Filter struct {
	Name       string   // shown in the dialog, e.g. "MusiCalc presets"
	Extensions []string // without the dot, e.g. "json"; the first one is the default for saving
	MIME       string   // e.g. "application/json"; used by the mobile pickers and the portal
}

// OpenFile asks for a file and reads it. done runs on the UI thread with the file's contents,
// or with nil data and a nil error when the user cancelled.
func OpenFile(win fyne.Window, title string, f Filter, done func(data []byte, err error)) {
	deliver := func(data []byte, err error) { fyne.Do(func() { done(data, err) }) }
	if nativeOpen(win, title, f, deliver) {
		return
	}
	fyneOpen(win, f, deliver)
}

// SaveFile asks where to save data, suggesting suggestedName, and writes it. done runs on the
// UI thread; saved is false when the user cancelled.
func SaveFile(win fyne.Window, title, suggestedName string, f Filter, data []byte, done func(saved bool, err error)) {
	deliver := func(saved bool, err error) { fyne.Do(func() { done(saved, err) }) }
	if nativeSave(win, title, suggestedName, f, data, deliver) {
		return
	}
	fyneSave(win, suggestedName, f, data, deliver)
}

// fyneFilter is the Fyne version of f. The mobile pickers only understand MIME types.
func fyneFilter(f Filter) storage.FileFilter {
	if fyne.CurrentDevice().IsMobile() && f.MIME != "" {
		return storage.NewMimeTypeFileFilter([]string{f.MIME})
	}
	exts := make([]string, len(f.Extensions))
	for i, e := range f.Extensions {
		exts[i] = "." + e
	}
	return storage.NewExtensionFileFilter(exts)
}

// fyneOpen uses Fyne's file dialog, which is the system picker on Android and iOS
func fyneOpen(win fyne.Window, f Filter, deliver func([]byte, error)) {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			deliver(nil, err)
			return
		}
		defer r.Close()
		data, err := io.ReadAll(r)
		deliver(data, err)
	}, win)
	d.SetFilter(fyneFilter(f))
	d.Show()
}

// fyneSave uses Fyne's file dialog, which is the system picker on Android and iOS
func fyneSave(win fyne.Window, suggestedName string, f Filter, data []byte, deliver func(bool, error)) {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil || w == nil {
			deliver(false, err)
			return
		}
		_, err = w.Write(data)
		if cerr := w.Close(); err == nil {
			err = cerr
		}
		deliver(err == nil, err)
	}, win)
	d.SetFilter(fyneFilter(f))
	d.SetFileName(suggestedName)
	d.Show()
}
