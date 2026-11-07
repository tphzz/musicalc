package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/logic"
)

type threeControls struct {
	root                 fyne.CanvasObject
	own, neighbour, gain *widget.Entry
	target, unit         *widget.Select
}

func openThree(t *testing.T) threeControls {
	t.Helper()
	root := openEnglishTab(t, NewThreeToOneTab)
	e := numericEntries(root)
	if len(e) != 3 {
		t.Fatalf("got %d numeric entries, want own, neighbour and gain", len(e))
	}
	c := threeControls{root: root, own: &e[0].Entry, neighbour: &e[1].Entry, gain: &e[2].Entry}
	visitUI(root, func(o fyne.CanvasObject) {
		if s, ok := o.(*widget.Select); ok {
			switch len(s.Options) {
			case 3:
				c.target = s
			case 2:
				c.unit = s
			}
		}
	})
	if c.target == nil || c.unit == nil {
		t.Fatal("missing 3:1 selects")
	}
	return c
}

// TestThreeToOneMinimumAdviceMeetsTarget: own 0.02 m, target 15 dB needs at least 0.112468 m,
// which must be advised as 0.12 m, and following that advice must meet the target
func TestThreeToOneMinimumAdviceMeetsTarget(t *testing.T) {
	c := openThree(t)
	c.target.SetSelectedIndex(2) // 15 dB below
	c.own.SetText("0.02")
	c.neighbour.SetText("0.04")
	text := visibleText(c.root)
	if !strings.Contains(text, "at least 0.12 m") || strings.Contains(text, "at least 0.11 m") {
		t.Fatalf("minimum advice:\n%s", text)
	}
	if r := logic.SpillCheck(0.02, 0.12, 0, 15); !r.MeetsTarget {
		t.Errorf("following the advice gives %.3f dB, still short of 15 dB", r.SpillDB)
	}
	// In feet the advice also rounds up
	c.unit.SetSelected(logic.UnitFeet)
	// 0.112468 m is 0.36899 ft: at least 0.37 ft
	if text := visibleText(c.root); !strings.Contains(text, "at least 0.37 ft") {
		t.Errorf("feet advice, want at least 0.37 ft:\n%s", text)
	}
	if r := logic.SpillCheck(0.02, logic.ToMeters(0.37, logic.UnitFeet), 0, 15); !r.MeetsTarget {
		t.Errorf("following the feet advice fails the target: %.3f dB", r.SpillDB)
	}
}

// TestThreeToOneLoudSpillDoesNotCancel: with own 1 m, neighbour 3 m and +12 dB the spill is louder
// than the own mic, which never cancels completely
func TestThreeToOneLoudSpillDoesNotCancel(t *testing.T) {
	c := openThree(t)
	c.own.SetText("1")
	c.neighbour.SetText("3")
	c.gain.SetText("12")
	text := visibleText(c.root)
	if strings.Contains(text, "cancel completely") || !strings.Contains(text, "cut by up to −9.7 dB") {
		t.Errorf("+12 dB spill:\n%s", text)
	}
	c.neighbour.SetText("1")
	c.gain.SetText("0")
	if text := visibleText(c.root); !strings.Contains(text, "cancel completely") {
		t.Errorf("equal levels should be able to cancel completely:\n%s", text)
	}
	c.gain.SetText("20")
	c.neighbour.SetText("1")
	if text := visibleText(c.root); !strings.Contains(text, "boost by +19.1 to +20.8 dB") {
		t.Errorf("spill 10 times as loud never dips below the own mic:\n%s", text)
	}
}

// TestThreeToOneInvalidInput checks that unusable values give a message instead of NaN results
func TestThreeToOneInvalidInput(t *testing.T) {
	for _, tc := range []struct{ own, neighbour, gain string }{
		{"NaN", "3", "0"}, {"1", "Inf", "0"}, {"-1", "3", "0"}, {"0", "3", "0"}, {"1", "3", "NaN"}, {"1", "3", "1e6"}, {"1", "3", "abc"},
	} {
		c := openThree(t)
		c.own.SetText(tc.own)
		c.neighbour.SetText(tc.neighbour)
		c.gain.SetText(tc.gain)
		text := visibleText(c.root)
		if !strings.Contains(text, "one of the values cannot be used") || strings.Contains(text, "NaN dB") || strings.Contains(text, "Estimated spill") {
			t.Errorf("%+v:\n%s", tc, text)
		}
	}
	c := openThree(t)
	c.own.SetText("1")
	if text := visibleText(c.root); !strings.Contains(text, "Enter the distances") {
		t.Errorf("a missing distance asks for it:\n%s", text)
	}
}
