package ui

import (
	"strings"

	"musicalc/internal/logic"
	"musicalc/internal/ui/widgets"

	"fyne.io/fyne/v2/widget"
)

// unitInput is a numeric entry whose value is kept in a base unit (metres or °C). Switching the
// display unit converts the shown number but keeps the stored value, so the physical setup stays
// the same and repeated switching doesn't accumulate rounding.
type unitInput struct {
	Entry *widgets.NumericEntry
	Base  float64 // value in the base unit; 0 while the entry is empty

	unit       string
	toBase     func(v float64, unit string) float64
	fromBase   func(v float64, unit string) float64
	converting bool // the entry is being rewritten by a unit switch, not by the user
}

// newDistanceInput creates a distance entry stored in metres
func newDistanceInput(placeholder, unit string, onChanged func()) *unitInput {
	return newUnitInput(placeholder, unit, logic.ToMeters, logic.FromMeters, onChanged)
}

// newTemperatureInput creates a temperature entry stored in °C
func newTemperatureInput(placeholder, unit string, onChanged func()) *unitInput {
	return newUnitInput(placeholder, unit, logic.ToCelsius, logic.FromCelsius, onChanged)
}

func newUnitInput(placeholder, unit string, toBase, fromBase func(float64, string) float64, onChanged func()) *unitInput {
	in := &unitInput{Entry: widgets.NewNumericEntry(), unit: unit, toBase: toBase, fromBase: fromBase}
	in.Entry.SetPlaceHolder(placeholder)
	in.Entry.OnChanged = func(s string) {
		if in.converting {
			return
		}
		in.Base = toBase(logic.ParseFloat(s), in.unit)
		if onChanged != nil {
			onChanged()
		}
	}
	return in
}

// SetText sets the value in the current display unit, as if typed
func (in *unitInput) SetText(text string) {
	in.Entry.SetText(text)
}

// setBase sets the stored value exactly and shows it in the current unit, without the
// rounding that typing the shown text would cause; onChanged is not called
func (in *unitInput) setBase(v float64) {
	in.Base = v
	in.converting = true
	in.Entry.SetText(logic.FormatTrim(in.fromBase(v, in.unit), 2))
	in.converting = false
}

// Empty reports whether nothing has been entered
func (in *unitInput) Empty() bool {
	return strings.TrimSpace(in.Entry.Text) == ""
}

// setUnit switches the display unit and shows the stored value converted to it
func (in *unitInput) setUnit(unit string) {
	if unit == in.unit {
		return
	}
	in.unit = unit
	if in.Empty() {
		return
	}
	in.converting = true
	in.Entry.SetText(logic.FormatTrim(in.fromBase(in.Base, unit), 2))
	in.converting = false
}

// newUnitSelect creates a unit dropdown that converts the given inputs when the unit changes.
// The first unit is selected; onChanged runs after every switch, once the inputs are converted.
func newUnitSelect(units []string, inputs []*unitInput, onChanged func(unit string)) *widget.Select {
	sel := widget.NewSelect(units, func(unit string) {
		for _, in := range inputs {
			in.setUnit(unit)
		}
		if onChanged != nil {
			onChanged(unit)
		}
	})
	sel.Selected = units[0] // without calling OnChanged while the tab is still being built
	return sel
}
