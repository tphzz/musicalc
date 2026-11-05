//go:build windows

package filedialog

import (
	"os"
	"runtime"
	"strings"
	"unicode/utf16"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"golang.org/x/sys/windows"
)

// The common file dialogs of comdlg32. Without a hook or template, Windows shows them as the
// modern Explorer-style Open and Save As dialogs.
var (
	comdlg32               = windows.NewLazySystemDLL("comdlg32.dll")
	procGetOpenFileNameW   = comdlg32.NewProc("GetOpenFileNameW")
	procGetSaveFileNameW   = comdlg32.NewProc("GetSaveFileNameW")
	procCommDlgExtendedErr = comdlg32.NewProc("CommDlgExtendedError")
)

const (
	ofnOverwritePrompt = 0x00000002
	ofnHideReadOnly    = 0x00000004
	ofnNoChangeDir     = 0x00000008
	ofnPathMustExist   = 0x00000800
	ofnFileMustExist   = 0x00001000
	ofnExplorer        = 0x00080000

	maxPath = 32768
)

// openFileName is OPENFILENAMEW
type openFileName struct {
	structSize    uint32
	owner         uintptr
	instance      uintptr
	filter        *uint16
	customFilter  *uint16
	maxCustFilter uint32
	filterIndex   uint32
	file          *uint16
	maxFile       uint32
	fileTitle     *uint16
	maxFileTitle  uint32
	initialDir    *uint16
	title         *uint16
	flags         uint32
	fileOffset    uint16
	fileExtension uint16
	defExt        *uint16
	custData      uintptr
	hook          uintptr
	templateName  *uint16
	reserved      uintptr
	reservedInt   uint32
	flagsEx       uint32
}

// ownerWindow returns the HWND of the app window, so the dialog is modal to it
func ownerWindow(win fyne.Window) uintptr {
	var hwnd uintptr
	if nw, ok := win.(driver.NativeWindow); ok {
		nw.RunNative(func(ctx any) {
			if c, ok := ctx.(driver.WindowsWindowContext); ok {
				hwnd = c.HWND
			}
		})
	}
	return hwnd
}

// filterString builds the double-NUL-terminated filter list, e.g. "MusiCalc presets (*.json)\0*.json\0…"
func filterString(f Filter) []uint16 {
	patterns := make([]string, len(f.Extensions))
	for i, e := range f.Extensions {
		patterns[i] = "*." + e
	}
	p := strings.Join(patterns, ";")
	s := f.Name + " (" + p + ")\x00" + p + "\x00All files (*.*)\x00*.*\x00\x00"
	return utf16.Encode([]rune(s))
}

// runDialog shows the Open or Save As dialog on its own OS thread, so the app window keeps
// drawing, and returns the chosen path, or "" when the user cancelled
func runDialog(proc *windows.LazyProc, owner uintptr, title, fileName string, f Filter, flags uint32) (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED); err == nil {
		defer windows.CoUninitialize()
	}

	file := make([]uint16, maxPath)
	copy(file, utf16.Encode([]rune(fileName)))
	filter := filterString(f)
	titlePtr, _ := windows.UTF16PtrFromString(title)
	var defExt *uint16
	if len(f.Extensions) > 0 {
		defExt, _ = windows.UTF16PtrFromString(f.Extensions[0])
	}
	ofn := openFileName{
		owner:       owner,
		filter:      &filter[0],
		filterIndex: 1,
		file:        &file[0],
		maxFile:     maxPath,
		title:       titlePtr,
		flags:       flags | ofnExplorer | ofnNoChangeDir | ofnHideReadOnly | ofnPathMustExist,
		defExt:      defExt,
	}
	ofn.structSize = uint32(unsafe.Sizeof(ofn))

	ok, _, _ := proc.Call(uintptr(unsafe.Pointer(&ofn)))
	if ok == 0 {
		if code, _, _ := procCommDlgExtendedErr.Call(); code != 0 {
			return "", windows.Errno(code)
		}
		return "", nil // cancelled
	}
	return windows.UTF16ToString(file), nil
}

func nativeOpen(win fyne.Window, title string, f Filter, deliver func([]byte, error)) bool {
	owner := ownerWindow(win)
	go func() {
		path, err := runDialog(procGetOpenFileNameW, owner, title, "", f, ofnFileMustExist)
		if err != nil || path == "" {
			deliver(nil, err)
			return
		}
		deliver(os.ReadFile(path))
	}()
	return true
}

func nativeSave(win fyne.Window, title, suggestedName string, f Filter, data []byte, deliver func(bool, error)) bool {
	owner := ownerWindow(win)
	go func() {
		path, err := runDialog(procGetSaveFileNameW, owner, title, suggestedName, f, ofnOverwritePrompt)
		if err != nil || path == "" {
			deliver(false, err)
			return
		}
		err = os.WriteFile(path, data, 0o644)
		deliver(err == nil, err)
	}()
	return true
}
