package help

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestTableColumnWidths(t *testing.T) {
	test.NewTempApp(t)
	// One row of three columns: two short ones and a long one, as in the Alignment Delay example
	table := &fitTable{cols: 3, cells: make([]*widget.RichText, 3), natural: []float32{110, 70, 400}, longest: []float32{90, 50, 80}}
	sum := func(w []float32) (s float32) {
		for _, v := range w {
			s += v
		}
		return s
	}
	gaps := float32(4) * 1 // four grid lines of the default 1 px

	wide := table.columnWidths(700)
	if wide[2] <= 400 || sum(wide) > 700-gaps+0.01 {
		t.Errorf("with room to spare every column should get its natural width and share the rest: %v", wide)
	}

	narrow := table.columnWidths(360)
	if narrow[0] != 110 || narrow[1] != 70 {
		t.Errorf("the short columns should keep their natural width: %v", narrow)
	}
	if narrow[2] < 80 || sum(narrow) > 360-gaps+0.01 {
		t.Errorf("the long column should wrap within the rest of the width: %v", narrow)
	}

	tiny := table.columnWidths(150)
	for c, w := range tiny {
		if w < table.longest[c] {
			t.Errorf("column %d is narrower than its longest word: %v", c, tiny)
		}
	}
}
