package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"musicalc/internal/ui/widgets"
)

// View keeps a calculator's content and its presentation refresh together.
// RefreshText never rebuilds the calculator or changes its input state.
type View struct {
	Content     fyne.CanvasObject
	refreshText func()
}

func (v *View) RefreshText() {
	if v != nil && v.refreshText != nil {
		v.refreshText()
		v.Content.Refresh()
	}
}

type textRefresh struct{ updates []func() }

func newTextRefresh() *textRefresh { return &textRefresh{} }

func (r *textRefresh) add(update func()) {
	r.updates = append(r.updates, update)
	update()
}
func (r *textRefresh) refresh() {
	for _, update := range r.updates {
		update()
	}
}
func (r *textRefresh) view(content fyne.CanvasObject) *View {
	return &View{Content: content, refreshText: r.refresh}
}
func (r *textRefresh) label(key string, args ...any) *widget.Label {
	label := widget.NewLabel("")
	r.add(func() { label.SetText(i18n.T(key, args...)) })
	return label
}
func (r *textRefresh) row(key string, control fyne.CanvasObject) fyne.CanvasObject {
	label := r.label(key)
	label.Wrapping = fyne.TextWrapWord
	return container.New(&formRowLayout{}, label, control)
}
func (r *textRefresh) button(key string, tapped func()) *widgets.Button {
	button := widgets.NewButton("", tapped)
	r.add(func() { button.SetText(i18n.T(key)) })
	return button
}
func (r *textRefresh) buttonWithIcon(key string, icon fyne.Resource, tapped func()) *widgets.Button {
	button := widgets.NewButtonWithIcon("", icon, tapped)
	r.add(func() { button.SetText(i18n.T(key)) })
	return button
}
func (r *textRefresh) muted(key string, args ...any) *widget.RichText {
	text := widgets.NewMutedText("")
	r.add(func() { widgets.SetStyledText(text, i18n.T(key, args...), "") })
	return text
}
func (r *textRefresh) wrappedMuted(key string) *widget.RichText {
	text := r.muted(key)
	text.Wrapping = fyne.TextWrapWord
	return text
}
func (r *textRefresh) placeholder(entry interface{ SetPlaceHolder(string) }, key string) {
	r.add(func() { entry.SetPlaceHolder(i18n.T(key)) })
}
func (r *textRefresh) selectOptions(selectWidget *widget.Select, keys []string) {
	r.add(func() {
		index := selectWidget.SelectedIndex()
		options := make([]string, len(keys))
		for i, key := range keys {
			options[i] = i18n.T(key)
		}
		selectWidget.Options = options
		selectWidget.Selected = ""
		if index >= 0 && index < len(options) {
			selectWidget.Selected = options[index]
		}
		selectWidget.Refresh()
	})
}
