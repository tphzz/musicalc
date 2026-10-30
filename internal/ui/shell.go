package ui

import (
	"fyne.io/fyne/v2"
)

// keyboardInsetThreshold is how much of the bottom of the screen system UI must cover before we
// treat it as the on-screen keyboard. Navigation bars take up to ~48 dp, phone keyboards 250 dp or more.
const keyboardInsetThreshold = 150

// NewShell lays out the app: header at the top, footer (the tab bar) at the bottom and content in
// between. On mobile, Fyne shrinks the usable area by the keyboard height, which would push the tab
// bar up on top of the keyboard; like Android's bottom navigation, the footer hides instead while
// the keyboard is open.
func NewShell(header, content, footer fyne.CanvasObject) fyne.CanvasObject {
	return &fyne.Container{
		Layout:  &shellLayout{isMobile: fyne.CurrentDevice().IsMobile()},
		Objects: []fyne.CanvasObject{header, content, footer},
	}
}

type shellLayout struct {
	isMobile bool
}

// keyboardOpen compares the canvas with its interactive area: the keyboard is reported as part of
// the bottom inset. This doesn't depend on layout history, which Fyne can briefly disturb by laying
// content out at its minimum size.
func (l *shellLayout) keyboardOpen(obj fyne.CanvasObject) bool {
	if !l.isMobile {
		return false
	}
	c := fyne.CurrentApp().Driver().CanvasForObject(obj)
	if c == nil {
		return false
	}
	pos, area := c.InteractiveArea()
	bottomInset := c.Size().Height - pos.Y - area.Height
	return bottomInset > keyboardInsetThreshold
}

func (l *shellLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	header, content, footer := objects[0], objects[1], objects[2]

	headerHeight := header.MinSize().Height
	header.Move(fyne.NewPos(0, 0))
	header.Resize(fyne.NewSize(size.Width, headerHeight))

	var footerHeight float32
	if l.keyboardOpen(header) {
		footer.Hide()
	} else {
		footer.Show()
		footerHeight = footer.MinSize().Height
		footer.Move(fyne.NewPos(0, size.Height-footerHeight))
		footer.Resize(fyne.NewSize(size.Width, footerHeight))
	}

	content.Move(fyne.NewPos(0, headerHeight))
	content.Resize(fyne.NewSize(size.Width, size.Height-headerHeight-footerHeight))
}

func (l *shellLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	min := fyne.NewSize(0, 0)
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		m := o.MinSize()
		min.Width = fyne.Max(min.Width, m.Width)
		min.Height += m.Height
	}
	return min
}
