package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"musicalc/internal/ui/widgets"
)

// setupCard holds a calculator's settings. Open, it shows the labelled settings; folded, it is a
// single line with a title and all values, so the results get the screen. Tapping the header
// opens and folds it.
type setupCard struct {
	Card *widgets.Card

	title    *widget.RichText
	summary  *widget.RichText
	chevron  *widget.Icon
	body     fyne.CanvasObject
	onToggle func(open bool)
}

// newSetupCard creates a setup card around body, open. onToggle runs after the card opened or
// folded, e.g. to lay the tab out again for the card's new height.
func newSetupCard(body fyne.CanvasObject, onToggle func(open bool)) *setupCard {
	c := &setupCard{
		title:    widgets.NewStyledText("Setup", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{Bold: true}),
		summary:  widgets.NewMutedText(""),
		chevron:  widget.NewIcon(theme.MenuDropUpIcon()),
		body:     body,
		onToggle: onToggle,
	}
	c.summary.Truncation = fyne.TextTruncateEllipsis
	c.summary.Hide()
	header := widgets.NewTapArea(
		container.NewBorder(nil, nil, c.title, c.chevron, c.summary),
		func() { c.SetOpen(!c.IsOpen()) },
	)
	c.Card = widgets.NewCard(container.NewVBox(header, body))
	return c
}

// SetTitle sets the bold title, e.g. "Setup" or the name of a loaded preset
func (c *setupCard) SetTitle(title string) {
	widgets.SetStyledText(c.title, title, "")
}

// SetSummary sets the one-line summary of the values shown while folded
func (c *setupCard) SetSummary(summary string) {
	widgets.SetStyledText(c.summary, summary, "")
}

// IsOpen reports whether the settings are shown
func (c *setupCard) IsOpen() bool {
	return c.body.Visible()
}

// SetOpen shows or folds the settings
func (c *setupCard) SetOpen(open bool) {
	if open {
		c.body.Show()
		c.summary.Hide()
		c.chevron.SetResource(theme.MenuDropUpIcon())
	} else {
		c.body.Hide()
		c.summary.Show()
		c.chevron.SetResource(theme.MenuDropDownIcon())
	}
	if c.onToggle != nil {
		c.onToggle(open)
	}
}
