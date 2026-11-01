package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MenuOption is one choice in a grouped menu
type MenuOption struct {
	Label string // text shown in the submenu
	Value string // value reported when the option is picked
}

// MenuGroup is a named submenu of options
type MenuGroup struct {
	Name    string
	Options []MenuOption
}

// NewGroupedMenu builds a menu with one submenu per group. The option whose value is selected,
// and the group that contains it, are shown checked.
func NewGroupedMenu(groups []MenuGroup, selected string, onPick func(group MenuGroup, option MenuOption)) *fyne.Menu {
	items := make([]*fyne.MenuItem, len(groups))
	for i, g := range groups {
		children := make([]*fyne.MenuItem, len(g.Options))
		groupChecked := false
		for j, o := range g.Options {
			children[j] = fyne.NewMenuItem(o.Label, func() { onPick(g, o) })
			if o.Value == selected {
				children[j].Checked = true
				groupChecked = true
			}
		}
		items[i] = fyne.NewMenuItem(g.Name, nil)
		items[i].ChildMenu = fyne.NewMenu("", children...)
		items[i].Checked = groupChecked
	}
	return fyne.NewMenu("", items...)
}

// MenuSelect looks like a widget.Select, but its options are grouped into submenus, which keeps
// long lists short enough for a phone screen. It shows the selection as "Group · Label".
type MenuSelect struct {
	widget.BaseWidget
	OnChanged func(value string)
	// Placeholder is shown when the selected value matches no option, e.g. "Custom"
	Placeholder string

	groups   []MenuGroup
	selected string
	text     string
	hovered  bool
	pressed  bool // the current touch started on this widget
}

// NewMenuSelect creates a grouped select; onChanged receives the picked option's value
func NewMenuSelect(groups []MenuGroup, onChanged func(value string)) *MenuSelect {
	s := &MenuSelect{groups: groups, OnChanged: onChanged}
	s.ExtendBaseWidget(s)
	return s
}

// SetGroups refreshes labels while retaining the semantic value, without an edit callback.
func (s *MenuSelect) SetGroups(groups []MenuGroup) {
	selected := s.selected
	s.groups = groups
	s.SetSelected(selected)
}

// Selected returns the value of the selected option
func (s *MenuSelect) Selected() string {
	return s.selected
}

// Text returns what the select shows: "Group · Label", or the Placeholder
func (s *MenuSelect) Text() string {
	return s.text
}

// SetSelected selects the option with the given value without calling OnChanged.
// A value without an option clears the selection and shows the Placeholder.
func (s *MenuSelect) SetSelected(value string) {
	for _, g := range s.groups {
		for _, o := range g.Options {
			if o.Value == value {
				s.selected = value
				s.text = g.Name + " · " + o.Label
				s.Refresh()
				return
			}
		}
	}
	s.selected = ""
	s.text = s.Placeholder
	s.Refresh()
}

func (s *MenuSelect) showMenu() {
	c := fyne.CurrentApp().Driver().CanvasForObject(s)
	if c == nil {
		return
	}
	menu := NewGroupedMenu(s.groups, s.selected, func(_ MenuGroup, o MenuOption) {
		changed := o.Value != s.selected
		s.SetSelected(o.Value)
		if changed && s.OnChanged != nil {
			s.OnChanged(o.Value)
		}
	})
	// Open below the control, like widget.Select
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(s)
	pos = pos.Add(fyne.NewPos(0, s.Size().Height-s.Theme().Size(theme.SizeNameInputBorder)))
	widget.NewPopUpMenu(menu, c).ShowAtPosition(pos)
}

// Tapped opens the menu on desktop; on mobile it opens from TouchUp instead
func (s *MenuSelect) Tapped(*fyne.PointEvent) {
	if fyne.CurrentDevice().IsMobile() {
		return
	}
	s.showMenu()
}

// TouchDown starts a press; the menu opens on TouchUp
func (s *MenuSelect) TouchDown(*mobile.TouchEvent) { s.pressed = true }

// TouchUp opens the menu if the press started here, for any press length
func (s *MenuSelect) TouchUp(*mobile.TouchEvent) {
	if !s.pressed {
		return
	}
	s.pressed = false
	s.showMenu()
}

// TouchCancel ends the press without opening the menu
func (s *MenuSelect) TouchCancel(*mobile.TouchEvent) { s.pressed = false }

// MouseIn shows the hover state on desktop
func (s *MenuSelect) MouseIn(*desktop.MouseEvent) {
	s.hovered = true
	s.Refresh()
}

// MouseMoved is required by desktop.Hoverable
func (s *MenuSelect) MouseMoved(*desktop.MouseEvent) {}

// MouseOut clears the hover state
func (s *MenuSelect) MouseOut() {
	s.hovered = false
	s.Refresh()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (s *MenuSelect) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabel(s.text)
	label.Truncation = fyne.TextTruncateEllipsis
	r := &menuSelectRenderer{
		sel:        s,
		background: canvas.NewRectangle(nil),
		label:      label,
		icon:       widget.NewIcon(theme.MenuDropDownIcon()),
		// Theme sizes don't change at runtime, so measure the reference Select once
		selectHeight: widget.NewSelect(nil, nil).MinSize().Height,
	}
	r.applyTheme()
	return r
}

type menuSelectRenderer struct {
	sel          *MenuSelect
	background   *canvas.Rectangle
	label        *widget.Label
	icon         *widget.Icon
	selectHeight float32 // height of a widget.Select with the current theme
}

// Layout mirrors widget.Select: text at the left, drop-down icon at the right
func (r *menuSelectRenderer) Layout(size fyne.Size) {
	th := r.sel.Theme()
	innerPad := th.Size(theme.SizeNameInnerPadding)
	iconSize := th.Size(theme.SizeNameInlineIcon)

	r.background.Resize(size)
	iconPos := fyne.NewPos(size.Width-iconSize-innerPad, (size.Height-iconSize)/2)
	r.icon.Resize(fyne.NewSquareSize(iconSize))
	r.icon.Move(iconPos)

	// Label draws its own inner padding, which is where Select draws its text
	labelHeight := r.label.MinSize().Height
	r.label.Move(fyne.NewPos(0, (size.Height-labelHeight)/2))
	r.label.Resize(fyne.NewSize(iconPos.X, labelHeight))
}

// MinSize uses the height of a real widget.Select, so both line up in a row; the width only
// needs room for a few characters, as the label truncates
func (r *menuSelectRenderer) MinSize() fyne.Size {
	th := r.sel.Theme()
	pad := th.Size(theme.SizeNamePadding)
	innerPad := th.Size(theme.SizeNameInnerPadding)
	iconSize := th.Size(theme.SizeNameInlineIcon)
	minText := fyne.MeasureText("12-TET", th.Size(theme.SizeNameText), fyne.TextStyle{}).Width
	return fyne.NewSize(2*pad+minText+innerPad+iconSize+innerPad, r.selectHeight)
}

func (r *menuSelectRenderer) Refresh() {
	r.applyTheme()
	r.label.SetText(r.sel.text)
	r.Layout(r.sel.Size())
	canvas.Refresh(r.sel)
}

func (r *menuSelectRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.label, r.icon}
}

func (r *menuSelectRenderer) Destroy() {}

func (r *menuSelectRenderer) applyTheme() {
	th := r.sel.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	if r.sel.hovered || r.sel.pressed {
		r.background.FillColor = th.Color(theme.ColorNameHover, v)
	} else {
		r.background.FillColor = th.Color(theme.ColorNameInputBackground, v)
	}
	r.background.CornerRadius = th.Size(theme.SizeNameInputRadius)
	r.background.Refresh()
}
