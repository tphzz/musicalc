package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/ui/widgets"
)

func TestTimecodeCompactPhoneHistory(t *testing.T) {
	app := test.NewTempApp(t)
	app.Settings().SetTheme(customTheme{isMobile: true})
	w := app.NewWindow("Timecode")
	defer w.Close()
	content := NewTimecodeTab().Content
	w.SetContent(content)
	w.Resize(fyne.NewSize(375, 647))
	var entries []*widgets.TimecodeEntry
	var add, clear, reset *widgets.Button
	var history *container.Scroll
	var visit func(fyne.CanvasObject)
	visit = func(obj fyne.CanvasObject) {
		switch x := obj.(type) {
		case *fyne.Container:
			for _, child := range x.Objects {
				visit(child)
			}
		case *widgets.TimecodeEntry:
			entries = append(entries, x)
		case *widgets.Button:
			switch x.Text {
			case "+":
				add = x
			case "Clear Hist.":
				clear = x
			case "Reset":
				reset = x
			}
		case *container.Scroll:
			history = x
		}
	}
	visit(content)
	if len(entries) != 2 || add == nil || clear == nil || reset == nil || history == nil {
		t.Fatal("missing Timecode controls")
	}
	for i := 0; i < 20; i++ {
		entries[0].SetComponents(0, 0, 10, 0)
		entries[1].SetComponents(0, 0, 5, 15)
		add.TouchDown(&mobile.TouchEvent{})
		add.TouchUp(&mobile.TouchEvent{})
	}
	text := history.Content.(*widget.Label).Text
	if !strings.Contains(text, "00:00:15:15 (465f) @30") {
		t.Fatalf("incorrect Timecode result: %s", text)
	}
	if history.Size().Height <= 0 || history.Content.MinSize().Height <= history.Size().Height {
		t.Fatal("long history should scroll inside the compact phone")
	}
	history.ScrollToBottom()
	if history.Offset.Y <= 0 {
		t.Fatal("history could not scroll")
	}
	about := NewAboutDialog("0.9.0 (42)", w)
	about.Show()
	about.Hide()
	if history.Content.(*widget.Label).Text != text || entries[0].Text != "00:00:15:15" {
		t.Fatal("About changed calculator inputs/history")
	}
	clear.TouchDown(&mobile.TouchEvent{})
	clear.TouchUp(&mobile.TouchEvent{})
	if history.Content.(*widget.Label).Text != "" || history.Offset.Y != 0 {
		t.Fatal("Clear Hist. left stale text or scroll offset")
	}
	entries[0].SetComponents(0, 0, 10, 0)
	entries[1].SetComponents(0, 0, 5, 15)
	add.TouchDown(&mobile.TouchEvent{})
	add.TouchUp(&mobile.TouchEvent{})
	reset.TouchDown(&mobile.TouchEvent{})
	reset.TouchUp(&mobile.TouchEvent{})
	if history.Content.(*widget.Label).Text != "" || history.Offset.Y != 0 || entries[0].Text != "" || entries[1].Text != "" {
		t.Fatal("Reset left stale history, offset or inputs")
	}

}
