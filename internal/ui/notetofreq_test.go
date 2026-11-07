package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"musicalc/internal/i18n"
	sclres "musicalc/internal/logic/scl"
	"musicalc/internal/ui/widgets"
)

// Bundled tunings used below, by their names in the catalogue
const (
	tuningKirnberger3 = "Kirnberger 3: 1/4 synt. comma (1744)"
	tuningPtolemy     = "Ptolemy's Intense Diatonic Syntonon, also Zarlino's scale"
	tuningHuzam       = "Arabic Huzam with perde segah on E by Dr. Oz."
)

type diapasonControls struct {
	reference  *widget.Entry
	frequency  *widget.Entry
	convention *widget.RadioGroup
	table      *widget.Table
	reset      *widgets.Button
	tuning     *widgets.MenuSelect
	root       *widget.Select
	columns    *ResponsiveTable
}

func findDiapasonControls(t *testing.T, root fyne.CanvasObject) diapasonControls {
	t.Helper()
	var c diapasonControls
	visitUI(root, func(object fyne.CanvasObject) {
		switch x := object.(type) {
		case *widget.Entry:
			if c.reference == nil {
				c.reference = x
			}
		case *widget.RadioGroup:
			c.convention = x
		case *widget.Table:
			c.table = x
		case *widgets.MenuSelect:
			c.tuning = x
		case *ResponsiveTable:
			c.columns = x
		case *widget.Select:
			if len(x.Options) == 12 {
				c.root = x
			}
		}
	})
	entries := numericEntries(root)
	if c.reference == nil || len(entries) != 1 || c.convention == nil || c.table == nil || c.tuning == nil || c.root == nil || c.columns == nil {
		t.Fatal("missing Note to Frequency controls")
	}
	c.frequency = &entries[0].Entry
	c.reset = buttonNamed(t, root, "Reset")
	return c
}

// pickTuning chooses a tuning the way the dropdown does when the user picks it
func (c diapasonControls) pickTuning(name string) {
	c.tuning.SetSelected(name)
	c.tuning.OnChanged(name)
}

// cellLabel renders one table cell the way the table does
func cellLabel(table *widget.Table, row, col int) *widget.Label {
	cell := table.CreateCell()
	table.UpdateCell(widget.TableCellID{Row: row, Col: col}, cell)
	return cell.(*fyne.Container).Objects[1].(*widget.Label)
}

func cellText(table *widget.Table, row, col int) string { return cellLabel(table, row, col).Text }

// headerText renders one column header
func headerText(table *widget.Table, col int) string {
	h := table.CreateHeader()
	table.UpdateHeader(widget.TableCellID{Row: -1, Col: col}, h)
	return h.(*widget.Label).Text
}

// TestDiapasonResetFromC4 checks that Reset restores A3 in C3 naming, not A2: switching the
// convention back renames the reference, so it must happen before A3 is restored
func TestDiapasonResetFromC4(t *testing.T) {
	c := findDiapasonControls(t, openEnglishTab(t, NewDiapasonTab))
	c.convention.SetSelected("C4")
	if c.reference.Text != "A4" {
		t.Fatalf("switching to C4 renamed A3 to %q, want A4", c.reference.Text)
	}
	c.frequency.SetText("432")
	press(c.reset)
	if c.reference.Text != "A3" || c.convention.Selected != "C3" || c.frequency.Text != "440" {
		t.Fatalf("after Reset: %q, %s, %q Hz; want A3, C3, 440", c.reference.Text, c.convention.Selected, c.frequency.Text)
	}
	if got := cellText(c.table, 69, 1); got != "440.00 Hz" {
		t.Errorf("after Reset MIDI 69 shows %q, want 440.00 Hz", got)
	}
}

// TestDiapasonReferenceMidiZero checks that C-2 (MIDI note 0) is a usable reference
func TestDiapasonReferenceMidiZero(t *testing.T) {
	c := findDiapasonControls(t, openEnglishTab(t, NewDiapasonTab))
	c.reference.SetText("C-2")
	c.frequency.SetText("10")
	if got := cellText(c.table, 0, 1); got != "10.00 Hz" {
		t.Errorf("C-2 = 10 Hz: MIDI 0 shows %q", got)
	}
	if got := cellText(c.table, 12, 1); got != "20.00 Hz" {
		t.Errorf("C-2 = 10 Hz: MIDI 12 shows %q, want 20.00 Hz", got)
	}
}

// TestDiapasonInvalidReference checks that an invalid reference shows no values instead of a
// table tuned to something the user did not enter
func TestDiapasonInvalidReference(t *testing.T) {
	c := findDiapasonControls(t, openEnglishTab(t, NewDiapasonTab))
	for _, tc := range []struct{ note, hz string }{{"H3", "440"}, {"A3", "0"}, {"A3", "abc"}, {"", "440"}} {
		c.reference.SetText(tc.note)
		c.frequency.SetText(tc.hz)
		if got := cellText(c.table, 69, 1); got != noValue {
			t.Errorf("reference %q at %q Hz: MIDI 69 shows %q, want %q", tc.note, tc.hz, got, noValue)
		}
		if got := cellText(c.table, 69, 0); got != "A3" {
			t.Errorf("reference %q: the note column shows %q, want A3", tc.note, got)
		}
	}
	c.reference.SetText("A3")
	c.frequency.SetText("440")
	if got := cellText(c.table, 69, 1); got != "440.00 Hz" {
		t.Errorf("after correcting the reference MIDI 69 shows %q", got)
	}
}

// TestDiapasonScaleRootDefaultsToTonic checks that a tuning starts on its own tonic: Kirnberger III
// on C shows its pure C–E third, while the reference A3 stays at 440 Hz
func TestDiapasonScaleRootDefaultsToTonic(t *testing.T) {
	c := findDiapasonControls(t, openEnglishTab(t, NewDiapasonTab))
	if c.root.Selected != "C (scale default)" {
		t.Errorf("12-TET root shows %q, want C (scale default)", c.root.Selected)
	}
	c.pickTuning(tuningKirnberger3)
	if c.root.Selected != "C (scale default)" {
		t.Errorf("Kirnberger III root shows %q", c.root.Selected)
	}
	for _, tc := range []struct {
		row, col int
		want     string
	}{
		{60, 1, "263.18 Hz"}, {60, 2, "+10.26"}, {64, 1, "328.98 Hz"}, {67, 1, "393.55 Hz"},
		{69, 1, "440.00 Hz"}, {69, 2, "+0.00"}, {72, 0, "C4"}, {72, 1, "526.36 Hz"},
	} {
		if got := cellText(c.table, tc.row, tc.col); got != tc.want {
			t.Errorf("Kirnberger III on C: row %d col %d = %q, want %q", tc.row, tc.col, got, tc.want)
		}
	}
	if !cellLabel(c.table, 72, 0).TextStyle.Bold || cellLabel(c.table, 71, 0).TextStyle.Bold {
		t.Error("only the root rows (C) are bold in the note column")
	}
	if got := headerText(c.table, 0); got != "Note" {
		t.Errorf("12-step header %q, want Note", got)
	}

	// Another root moves the 1/1; only the tonic keeps the "scale default" marker
	c.root.SetSelectedIndex(9)
	if c.root.Selected != "A" {
		t.Errorf("root A shows %q", c.root.Selected)
	}
	if got := cellText(c.table, 60, 1); got != "260.74 Hz" {
		t.Errorf("Kirnberger III on A: C3 = %q, want 260.74 Hz", got)
	}
	if !cellLabel(c.table, 57, 0).TextStyle.Bold || cellLabel(c.table, 72, 0).TextStyle.Bold {
		t.Error("on A the A rows are the bold roots")
	}

	// Picking a tuning again puts the root back on its tonic: Huzam is on E
	c.pickTuning(tuningHuzam)
	if c.root.Selected != "E (scale default)" || cellText(c.table, 64, 1) != "329.25 Hz" {
		t.Errorf("Huzam: root %q, E3 %q; want E (scale default), 329.25 Hz", c.root.Selected, cellText(c.table, 64, 1))
	}
	if got := headerText(c.table, 0); got != "Note" {
		t.Errorf("Huzam has 12 steps and keeps note names, header %q", got)
	}
}

// TestDiapasonDegreeLabels checks that a 7-step scale names keys by degree and period, with cents
// above the period's root, and that Reset returns to note names
func TestDiapasonDegreeLabels(t *testing.T) {
	c := findDiapasonControls(t, openEnglishTab(t, NewDiapasonTab))
	c.pickTuning(tuningPtolemy)
	// The wider Cents column fits "+1088.27" at phone width
	if got := c.columns.columnProportions; got[1] != stepColumns[1] || got[2] != stepColumns[2] {
		t.Errorf("7-step column shares %v, want %v", got, stepColumns)
	}
	if got := headerText(c.table, 0); got != "Step" {
		t.Errorf("7-step header %q, want Step", got)
	}
	if got := headerText(c.table, 3); got != "MIDI" {
		t.Errorf("MIDI header %q", got)
	}
	for _, tc := range []struct {
		row  int
		want [3]string
	}{
		{60, [3]string{"0 (0)", "176.00 Hz", "+0.00"}},
		{67, [3]string{"0 (1)", "352.00 Hz", "+0.00"}},
		{69, [3]string{"2 (1)", "440.00 Hz", "+386.31"}},
		{73, [3]string{"6 (1)", "660.00 Hz", "+1088.27"}},
		{59, [3]string{"6 (-1)", "165.00 Hz", "+1088.27"}},
	} {
		for col, want := range tc.want {
			if got := cellText(c.table, tc.row, col); got != want {
				t.Errorf("Ptolemy row %d col %d = %q, want %q", tc.row, col, got, want)
			}
		}
	}
	if !cellLabel(c.table, 67, 0).TextStyle.Bold || cellLabel(c.table, 68, 0).TextStyle.Bold {
		t.Error("degree 0 rows are bold, others are not")
	}
	// The degree depends on the reference key, so it is unknown while the reference is invalid
	c.reference.SetText("A")
	if got := cellText(c.table, 69, 0); got != noValue {
		t.Errorf("invalid reference: degree column shows %q, want %q", got, noValue)
	}
	c.reference.SetText("A3")

	c.root.SetSelectedIndex(2) // D
	press(c.reset)
	if c.tuning.Selected() != sclres.DefaultScaleName || c.root.Selected != "C (scale default)" || headerText(c.table, 0) != "Note" {
		t.Errorf("after Reset: tuning %q, root %q, header %q", c.tuning.Selected(), c.root.Selected, headerText(c.table, 0))
	}
	if got := cellText(c.table, 69, 1); got != "440.00 Hz" || cellText(c.table, 69, 0) != "A3" {
		t.Errorf("after Reset MIDI 69 = %q, %q", cellText(c.table, 69, 0), got)
	}
	if got := c.columns.columnProportions; got[1] != noteColumns[1] || got[2] != noteColumns[2] {
		t.Errorf("12-step column shares after Reset %v, want %v", got, noteColumns)
	}
}

// TestDiapasonGermanLabels checks the localised root select and step header, and that a language
// switch keeps the chosen root
func TestDiapasonGermanLabels(t *testing.T) {
	app := test.NewTempApp(t)
	app.Preferences().SetString(prefLanguage, "en")
	InitializeLanguage(app.Preferences())
	i18n.Use("en")
	t.Cleanup(func() { i18n.Use("en") })
	window := app.NewWindow("MusiCalc")
	t.Cleanup(window.Close)
	view := NewDiapasonTab()
	window.SetContent(view.Content)
	c := findDiapasonControls(t, view.Content)
	c.pickTuning(tuningPtolemy)
	c.root.SetSelectedIndex(2)

	i18n.Use("de")
	view.RefreshText()
	text := visibleText(view.Content)
	if !strings.Contains(text, "Grundton") || headerText(c.table, 0) != "Stufe" || headerText(c.table, 3) != "MIDI" {
		t.Errorf("German labels: header %q / %q\n%s", headerText(c.table, 0), headerText(c.table, 3), text)
	}
	if c.root.Selected != "D" || c.root.Options[0] != "C (Standard)" {
		t.Errorf("language switch: root %q, options %v", c.root.Selected, c.root.Options)
	}
	if got := cellText(c.table, 62, 0); got != "0 (0)" {
		t.Errorf("root D stays after the language switch: MIDI 62 = %q, want 0 (0)", got)
	}
}
