package main

import (
	"bytes"
	"strings"
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

func TestWriteVersion(t *testing.T) {
	previous := version
	t.Cleanup(func() { version = previous })
	for _, value := range []string{"0.9.0-snapshot.gdbc3c2c", "0.9.0-snapshot", "0.9.0"} {
		version = value
		var out bytes.Buffer
		if !writeVersion([]string{"--version"}, &out) || out.String() != value+"\n" {
			t.Fatalf("version output=%q, want %q", out.String(), value+"\n")
		}
	}
	for _, args := range [][]string{nil, {"other"}, {"--version", "other"}} {
		var out bytes.Buffer
		if writeVersion(args, &out) || strings.TrimSpace(out.String()) != "" {
			t.Fatalf("unexpected version handling for %v", args)
		}
	}
}
