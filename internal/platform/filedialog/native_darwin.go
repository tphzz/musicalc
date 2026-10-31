//go:build darwin && !ios && cgo

package filedialog

/*
#cgo LDFLAGS: -framework AppKit -weak_framework UniformTypeIdentifiers
#include <stdint.h>
#include <stdlib.h>

void musicalcOpenPanel(uintptr_t window, const char* title, const char* ext, int handle);
void musicalcSavePanel(uintptr_t window, const char* title, const char* name, const char* ext, int handle);
*/
import "C"

import (
	"os"
	"sync"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

// On macOS the dialogs are NSOpenPanel and NSSavePanel, shown as a sheet on the app window.
// native_darwin.m runs them on the main queue and reports the chosen path to panelDone.

var (
	panelMu      sync.Mutex
	panelNext    C.int
	panelPending = map[C.int]func(path string){}
)

// registerPanel stores the function that receives a panel's result and returns its handle
func registerPanel(done func(path string)) C.int {
	panelMu.Lock()
	defer panelMu.Unlock()
	panelNext++
	panelPending[panelNext] = done
	return panelNext
}

//export musicalcPanelDone
func musicalcPanelDone(handle C.int, path *C.char) {
	panelMu.Lock()
	done := panelPending[handle]
	delete(panelPending, handle)
	panelMu.Unlock()
	if done == nil {
		return
	}
	p := ""
	if path != nil {
		p = C.GoString(path)
	}
	// Leave the main thread before touching the file
	go done(p)
}

// nsWindow returns the app window's NSWindow, so the panel opens as a sheet on it
func nsWindow(win fyne.Window) C.uintptr_t {
	var w uintptr
	if nw, ok := win.(driver.NativeWindow); ok {
		nw.RunNative(func(ctx any) {
			if c, ok := ctx.(driver.MacWindowContext); ok {
				w = c.NSWindow
			}
		})
	}
	return C.uintptr_t(w)
}

func extension(f Filter) string {
	if len(f.Extensions) > 0 {
		return f.Extensions[0]
	}
	return ""
}

func nativeOpen(win fyne.Window, title string, f Filter, deliver func([]byte, error)) bool {
	handle := registerPanel(func(path string) {
		if path == "" {
			deliver(nil, nil) // cancelled
			return
		}
		deliver(os.ReadFile(path))
	})
	cTitle, cExt := C.CString(title), C.CString(extension(f))
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cExt))
	C.musicalcOpenPanel(nsWindow(win), cTitle, cExt, handle)
	return true
}

func nativeSave(win fyne.Window, title, suggestedName string, f Filter, data []byte, deliver func(bool, error)) bool {
	handle := registerPanel(func(path string) {
		if path == "" {
			deliver(false, nil)
			return
		}
		err := os.WriteFile(path, data, 0o644)
		deliver(err == nil, err)
	})
	cTitle, cName, cExt := C.CString(title), C.CString(suggestedName), C.CString(extension(f))
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cExt))
	C.musicalcSavePanel(nsWindow(win), cTitle, cName, cExt, handle)
	return true
}
