package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"musicalc/internal/i18n"
	"musicalc/internal/ui/help"
	"musicalc/internal/ui/widgets"
)

// NewHelpView shows a calculator's help page: a title bar with a close button (✕) and the
// page, rendered from its Markdown, in a scroll view. onClose returns to the calculator.
func NewHelpView(pageID, titleKey string, onClose func()) *View {
	heading := widget.NewLabelWithStyle(i18n.T("help.heading", i18n.T(titleKey)), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	heading.Truncation = fyne.TextTruncateEllipsis
	closeButton := widgets.NewButtonWithIcon("", theme.CancelIcon(), onClose)
	closeButton.Importance = widget.LowImportance
	// Spacers center the title vertically next to the (taller) button, like the app header
	bar := container.NewBorder(nil, nil, nil, closeButton,
		container.NewVBox(layout.NewSpacer(), heading, layout.NewSpacer()))

	body := help.NewPage(help.Markdown(pageID))

	content := container.NewBorder(container.NewVBox(bar, widget.NewSeparator()), nil, nil, nil, body)
	return &View{Content: content, refreshText: func() {
		heading.SetText(i18n.T("help.heading", i18n.T(titleKey)))
		help.RefreshPage(body, help.Markdown(pageID))
	}}
}
