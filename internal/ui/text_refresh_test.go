package ui

import (
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"musicalc/internal/ui/widgets"
	"testing"
)

func TestTextRefreshPreservesValuesAndOptionIdentity(t *testing.T) {
	test.NewTempApp(t)
	defer i18n.Use("en")
	i18n.Use("en")
	texts := newTextRefresh()
	label := texts.label("field.frequency_hz")
	button := texts.button("action.reset", func() {})
	entry := widget.NewEntry()
	entry.SetText("123.456789")
	texts.placeholder(entry, "field.frequency")
	options := widget.NewSelect([]string{i18n.T("polarity.normal"), i18n.T("polarity.inverted")}, nil)
	options.SetSelectedIndex(1)
	changes := 0
	options.OnChanged = func(string) { changes++ }
	texts.selectOptions(options, []string{"polarity.normal", "polarity.inverted"})
	i18n.Use("de")
	texts.refresh()
	if label.Text != i18n.T("field.frequency_hz") || button.Text != i18n.T("action.reset") || entry.PlaceHolder != i18n.T("field.frequency") {
		t.Fatal("static text not refreshed")
	}
	if entry.Text != "123.456789" || options.SelectedIndex() != 1 || changes != 0 {
		t.Fatal("text refresh changed user state")
	}
}

func TestMenuSelectRelabelsWithoutChangingSemanticSelection(t *testing.T) {
	test.NewTempApp(t)
	changes := 0
	selectWidget := widgets.NewMenuSelect([]widgets.MenuGroup{{Name: "Drums", Options: []widgets.MenuOption{{Label: "Kick", Value: "Drums|Kick"}}}}, func(string) { changes++ })
	selectWidget.SetSelected("Drums|Kick")
	selectWidget.SetGroups([]widgets.MenuGroup{{Name: "Schlagzeug", Options: []widgets.MenuOption{{Label: "Bassdrum", Value: "Drums|Kick"}}}})
	if selectWidget.Selected() != "Drums|Kick" || selectWidget.Text() != "Schlagzeug · Bassdrum" || changes != 0 {
		t.Fatal("relabel changed selection or fired an edit")
	}
}
