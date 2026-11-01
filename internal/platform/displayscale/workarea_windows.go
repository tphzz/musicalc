package displayscale

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"golang.org/x/sys/windows"
	"math"
	"unsafe"
)

var (
	user32            = windows.NewLazySystemDLL("user32.dll")
	monitorFromWindow = user32.NewProc("MonitorFromWindow")
	getMonitorInfo    = user32.NewProc("GetMonitorInfoW")
	getWindowRect     = user32.NewProc("GetWindowRect")
	getClientRect     = user32.NewProc("GetClientRect")
	setWindowPos      = user32.NewProc("SetWindowPos")
)

type windowRect struct{ left, top, right, bottom int32 }
type monitorInfo struct {
	size          uint32
	monitor, work windowRect
	flags         uint32
}

// Keep zoomed windows on their current monitor, including the title bar.
func fitWindowOnScreen(window fyne.Window, requested fyne.Size) fyne.Size {
	native, ok := window.(driver.NativeWindow)
	if !ok {
		return requested
	}
	size := requested
	native.RunNative(func(context any) {
		contextWindow, ok := context.(driver.WindowsWindowContext)
		if !ok {
			return
		}
		handle := contextWindow.HWND
		monitor, _, _ := monitorFromWindow.Call(handle, 2) // MONITOR_DEFAULTTONEAREST.
		info := monitorInfo{}
		info.size = uint32(unsafe.Sizeof(info))
		outer, client := windowRect{}, windowRect{}
		success, _, _ := getMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info)))
		if success == 0 {
			return
		}
		success, _, _ = getWindowRect.Call(handle, uintptr(unsafe.Pointer(&outer)))
		if success == 0 {
			return
		}
		success, _, _ = getClientRect.Call(handle, uintptr(unsafe.Pointer(&client)))
		if success == 0 {
			return
		}
		borderWidth := max(int32(0), outer.right-outer.left-(client.right-client.left))
		borderHeight := max(int32(0), outer.bottom-outer.top-(client.bottom-client.top))
		available := fyne.NewSize(float32(info.work.right-info.work.left-borderWidth), float32(info.work.bottom-info.work.top-borderHeight))
		scale := window.Canvas().Scale()
		size = fitToAvailable(requested, scale, available)
		width := int32(math.Ceil(float64(size.Width)*float64(scale))) + borderWidth
		height := int32(math.Ceil(float64(size.Height)*float64(scale))) + borderHeight
		left := max(info.work.left, min(outer.left, info.work.right-width))
		top := max(info.work.top, min(outer.top, info.work.bottom-height))
		if left != outer.left || top != outer.top {
			// SWP_NOSIZE | SWP_NOZORDER | SWP_NOACTIVATE: preserve focus and stacking.
			success, _, err := setWindowPos.Call(handle, 0, uintptr(left), uintptr(top), 0, 0, 0x15)
			if success == 0 {
				fyne.LogError("Could not keep the window within the monitor", err)
			}
		}
	})
	return size
}
