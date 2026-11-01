package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"musicalc/internal/ui/widgets"
)

// languageChoices keeps required radio selection on controls that own their key
// handling. Fyne's stock radio items cannot forward mobile Escape/Back events.
type languageOption struct{ Value, Label string }

type languageChoices struct {
	widget.BaseWidget
	Options   []languageOption
	Selected  string
	OnChanged func(string)
	dialog    *PreferencesDialog
	items     []*languageChoice
}

func newLanguageChoices(dialog *PreferencesDialog, options []languageOption) *languageChoices {
	group := &languageChoices{Options: options, dialog: dialog}
	group.ExtendBaseWidget(group)
	for i, option := range options {
		button := &languageChoice{group: group, index: i}
		button.Text = option.Label
		button.Alignment = widget.ButtonAlignLeading
		button.Importance = widget.LowImportance
		button.OnTapped = func() { group.SetSelected(group.Options[i].Value) }
		button.ExtendBaseWidget(button)
		group.items = append(group.items, button)
	}
	return group
}
func (g *languageChoices) SetSelected(option string) {
	if option == g.Selected {
		return
	}
	valid := false
	for _, candidate := range g.Options {
		if option == candidate.Value {
			valid = true
			break
		}
	}
	if !valid {
		return
	}
	g.Selected = option
	for i, button := range g.items {
		if g.Options[i].Value == option {
			button.Icon = theme.NewPrimaryThemedResource(theme.RadioButtonCheckedIcon())
		} else {
			button.Icon = theme.NewThemedResource(theme.RadioButtonIcon())
		}
	}
	g.Refresh()
	if g.OnChanged != nil {
		g.OnChanged(option)
	}
}
func (g *languageChoices) setLabels(labels []string) {
	for i, label := range labels {
		g.Options[i].Label = label
		g.items[i].SetText(label)
	}
}

func (g *languageChoices) CreateRenderer() fyne.WidgetRenderer {
	rows := make([]fyne.CanvasObject, len(g.items))
	for i, button := range g.items {
		rows[i] = container.NewThemeOverride(button, languageChoiceTheme{Theme: g.Theme()})
	}
	return widget.NewSimpleRenderer(container.NewVBox(rows...))
}

type languageChoice struct {
	widgets.Button
	group *languageChoices
	index int
}

func (b *languageChoice) TypedKey(event *fyne.KeyEvent) {
	if b.group.dialog.closeKey(event) {
		return
	}
	next := b.index
	switch event.Name {
	case fyne.KeyUp, fyne.KeyLeft:
		next = (next + len(b.group.items) - 1) % len(b.group.items)
	case fyne.KeyDown, fyne.KeyRight:
		next = (next + 1) % len(b.group.items)
	default:
		b.Button.TypedKey(event)
		return
	}
	b.group.SetSelected(b.group.Options[next].Value)
	b.group.dialog.canvas.Focus(b.group.items[next])
}
func (b *languageChoice) AccessibilityLabel() string {
	if b.group.Selected == b.group.Options[b.index].Value {
		return i18n.T("language.choice_selected", b.Text)
	}
	return b.Text
}

// Keep the approved regular-weight radio labels while reusing native button
// focus, pointer and touch behavior.
type languageChoiceTheme struct{ fyne.Theme }

func (t languageChoiceTheme) Font(style fyne.TextStyle) fyne.Resource {
	style.Bold = false
	return t.Theme.Font(style)
}
