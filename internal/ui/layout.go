package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// minusSign formats negative numbers with a real minus sign instead of a hyphen
func minusSign(s string) string {
	return strings.Replace(s, "-", "−", 1)
}

// labeledRow is a two-column row: a label and its control, like the other calculators
func labeledRow(label string, control fyne.CanvasObject) fyne.CanvasObject {
	text := widget.NewLabel(label)
	text.Wrapping = fyne.TextWrapWord
	return container.New(&formRowLayout{}, text, control)
}

// formRowLayout keeps equal columns when they fit and gives controls their
// required width on compact screens; the label can wrap into the remaining space.
type formRowLayout struct{}

func (*formRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	left, right := objects[0].MinSize(), objects[1].MinSize()
	return fyne.NewSize(left.Width+right.Width+theme.Padding(), fyne.Max(left.Height, right.Height))
}
func (*formRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	gap := theme.Padding()
	leftWidth := fyne.Min((size.Width-gap)/2, fyne.Max(0, size.Width-gap-objects[1].MinSize().Width))
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(leftWidth, size.Height))
	objects[1].Move(fyne.NewPos(leftWidth+gap, 0))
	objects[1].Resize(fyne.NewSize(fyne.Max(0, size.Width-leftWidth-gap), size.Height))
}

// withSuffix puts a muted unit after a control
func withSuffix(control fyne.CanvasObject, suffix *widget.RichText) fyne.CanvasObject {
	return container.NewBorder(nil, nil, nil, suffix, control)
}

// newResultsScroll puts result cards in a scroll view and returns a function that lays them out
// again. A wrapped text can gain or lose a line without changing its width, which Fyne doesn't
// notice by itself, so call it after updating the cards' texts.
func newResultsScroll(objects ...fyne.CanvasObject) (fyne.CanvasObject, func()) {
	list := container.NewVBox(objects...)
	// A little inset keeps the card shadows from being clipped by the scroller
	scroll := container.NewVScroll(container.New(layout.NewCustomPaddedLayout(2, 4, 2, 2), list))
	return scroll, func() {
		list.Refresh()
		scroll.Refresh()
	}
}
