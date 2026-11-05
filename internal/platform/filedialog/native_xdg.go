//go:build (linux || freebsd || openbsd || netbsd) && !android

package filedialog

import (
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
	"github.com/rymdport/portal"
	"github.com/rymdport/portal/filechooser"
)

// On Linux and the BSDs the file dialog comes from the XDG desktop portal, so it is the desktop's
// own dialog (GNOME, KDE, …). Without a running portal, Fyne's dialog is used instead.

// portalParent identifies the app window to the portal, so the dialog is modal to it. Wayland
// would need the xdg-foreign protocol; there the dialog simply isn't attached.
func portalParent(win fyne.Window) string {
	parent := ""
	if nw, ok := win.(driver.NativeWindow); ok {
		nw.RunNative(func(ctx any) {
			if c, ok := ctx.(driver.X11WindowContext); ok {
				parent = portal.FormatX11WindowHandle(c.WindowHandle)
			}
		})
	}
	return parent
}

// portalFilter converts f into the portal's filter: glob patterns in both cases, plus the MIME type
func portalFilter(f Filter) *filechooser.Filter {
	var rules []filechooser.Rule
	for _, e := range f.Extensions {
		rules = append(rules,
			filechooser.Rule{Type: filechooser.GlobPattern, Pattern: "*." + strings.ToLower(e)},
			filechooser.Rule{Type: filechooser.GlobPattern, Pattern: "*." + strings.ToUpper(e)})
	}
	if f.MIME != "" {
		rules = append(rules, filechooser.Rule{Type: filechooser.MIMEType, Pattern: f.MIME})
	}
	return &filechooser.Filter{Name: f.Name, Rules: rules}
}

// pathFromURI turns the portal's file:// URI into a local path. The portal client has already
// unescaped it, so it must not be parsed as a URL again.
func pathFromURI(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}

func nativeOpen(win fyne.Window, title string, f Filter, deliver func([]byte, error)) bool {
	parent := portalParent(win)
	filter := portalFilter(f)
	go func() {
		uris, err := filechooser.OpenFile(parent, title, &filechooser.OpenFileOptions{
			Filters: []*filechooser.Filter{filter}, CurrentFilter: filter,
		})
		if err != nil {
			// No portal on this desktop: fall back to Fyne's dialog
			fyne.Do(func() { fyneOpen(win, f, deliver) })
			return
		}
		if len(uris) == 0 {
			deliver(nil, nil) // cancelled
			return
		}
		deliver(os.ReadFile(pathFromURI(uris[0])))
	}()
	return true
}

func nativeSave(win fyne.Window, title, suggestedName string, f Filter, data []byte, deliver func(bool, error)) bool {
	parent := portalParent(win)
	filter := portalFilter(f)
	go func() {
		uris, err := filechooser.SaveFile(parent, title, &filechooser.SaveFileOptions{
			Filters: []*filechooser.Filter{filter}, CurrentFilter: filter, CurrentName: suggestedName,
		})
		if err != nil {
			fyne.Do(func() { fyneSave(win, suggestedName, f, data, deliver) })
			return
		}
		if len(uris) == 0 {
			deliver(false, nil)
			return
		}
		err = os.WriteFile(pathFromURI(uris[0]), data, 0o644)
		deliver(err == nil, err)
	}()
	return true
}
