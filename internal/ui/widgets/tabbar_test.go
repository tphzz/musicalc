package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"testing"
)

func TestTabBarRelabelRetainsButtonsAndSelection(t *testing.T) {
	test.NewTempApp(t)
	bar := NewTabBar()
	bar.SetItems([]TabBarItem{{Text: "First"}, {Text: "Second"}}, 1)
	buttons := append([]fyne.CanvasObject(nil), bar.tabs.Objects...)
	renderers := make([]*tabButtonRenderer, len(buttons))
	for i, button := range buttons {
		renderers[i] = test.WidgetRenderer(button.(*tabButton)).(*tabButtonRenderer)
	}
	calls := 0
	bar.OnSelected = func(int) { calls++ }
	bar.SetLabels([]string{"Erster", "Zweiter"})
	if bar.Selected() != 1 || calls != 0 {
		t.Fatal("relabel changed selection")
	}
	for i, button := range buttons {
		if bar.tabs.Objects[i] != button {
			t.Fatal("relabel replaced a tab button")
		}
	}
	if renderers[1].label.Text != "Zweiter" || test.WidgetRenderer(buttons[1].(*tabButton)) != renderers[1] {
		t.Fatal("existing rendered label was not updated")
	}
}
