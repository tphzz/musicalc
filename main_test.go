package main

import (
	"testing"
)

func TestCategoryMenuActions(t *testing.T) {
	categories := []CategoryInfo{{Name: "First"}, {Name: "Second"}}
	selected, aboutCalls := -1, 0
	menu := categoryMenu(categories, func(i int) { selected = i }, func() { aboutCalls++ })
	if len(menu.Items) != 3 || menu.Items[2].Label != "About MusiCalc" {
		t.Fatal("About must follow the category entries")
	}
	menu.Items[0].Action()
	if selected != 0 {
		t.Fatalf("first category selected %d", selected)
	}
	menu.Items[1].Action()
	if selected != 1 {
		t.Fatalf("second category selected %d", selected)
	}
	menu.Items[2].Action()
	if aboutCalls != 1 || selected != 1 {
		t.Fatal("About changed the selected category")
	}

	nativeMacMenu := categoryMenu(categories, func(int) {}, nil)
	if len(nativeMacMenu.Items) != 2 {
		t.Fatal("native macOS About must not be duplicated in the category menu")
	}
}
