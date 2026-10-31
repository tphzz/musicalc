package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"musicalc/internal/logic"
	"musicalc/internal/platform/filedialog"
	"musicalc/internal/ui/widgets"
)

// Preference keys of the Alignment Delay tab
const (
	prefAlignmentPresets = "alignmentDelay.presets"
	prefAlignmentSession = "alignmentDelay.session"
)

// presetFileFilter limits the import and export dialogs to preset files
var presetFileFilter = filedialog.Filter{Name: "MusiCalc presets", Extensions: []string{"json"}, MIME: "application/json"}

// presetController keeps the stored Alignment Delay presets and runs the preset menu and dialogs
type presetController struct {
	presets []logic.Preset
	loaded  string // name of the preset the setup came from; "" when none

	current func() logic.Preset  // the setup on screen, without a name
	apply   func(p logic.Preset) // puts a preset's setup on screen (the temperature stays)
	changed func()               // the loaded preset or the list changed: update the header and session
}

// newPresetController loads the stored presets
func newPresetController(current func() logic.Preset, apply func(logic.Preset), changed func()) *presetController {
	c := &presetController{current: current, apply: apply, changed: changed}
	if data := fyne.CurrentApp().Preferences().String(prefAlignmentPresets); data != "" {
		_ = json.Unmarshal([]byte(data), &c.presets)
	}
	return c
}

// store saves the preset list
func (c *presetController) store() {
	data, err := json.Marshal(c.presets)
	if err == nil {
		fyne.CurrentApp().Preferences().SetString(prefAlignmentPresets, string(data))
	}
	c.changed()
}

// loadedIndex returns the index of the loaded preset, or -1
func (c *presetController) loadedIndex() int {
	if c.loaded == "" {
		return -1
	}
	return logic.FindPreset(c.presets, c.loaded)
}

// dirty reports whether the setup on screen differs from the loaded preset
func (c *presetController) dirty() bool {
	i := c.loadedIndex()
	return i >= 0 && !logic.SameSetup(c.presets[i], c.current())
}

// title is the setup card's title: the loaded preset's name, with * when changed
func (c *presetController) title() string {
	i := c.loadedIndex()
	if i < 0 {
		return "Setup"
	}
	if c.dirty() {
		return c.presets[i].Name + "*"
	}
	return c.presets[i].Name
}

// menuItems builds the preset entries of the overflow menu
func (c *presetController) menuItems(win fyne.Window) []*fyne.MenuItem {
	empty := len(c.presets) == 0
	save := fyne.NewMenuItem("Save", c.saveLoaded)
	if i := c.loadedIndex(); i >= 0 {
		save.Label = "Save “" + c.presets[i].Name + "”"
	}
	save.Disabled = !c.dirty()
	load := fyne.NewMenuItem("Load preset…", func() { c.showLoad(win) })
	load.Disabled = empty
	manage := fyne.NewMenuItem("Manage presets…", func() { c.showManage(win) })
	manage.Disabled = empty
	export := fyne.NewMenuItem("Export presets…", func() { c.exportPresets(win) })
	export.Disabled = empty
	return []*fyne.MenuItem{
		save,
		fyne.NewMenuItem("Save as…", func() { c.showSaveAs(win) }),
		load,
		manage,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Import presets…", func() { c.importPresets(win) }),
		export,
	}
}

// saveLoaded overwrites the loaded preset with the setup on screen
func (c *presetController) saveLoaded() {
	i := c.loadedIndex()
	if i < 0 {
		return
	}
	p := c.current()
	p.Name = c.presets[i].Name
	c.presets[i] = p
	c.store()
}

// saveAs stores the setup on screen under name, replacing a preset of that name
func (c *presetController) saveAs(name string) {
	p := c.current()
	p.Name = strings.TrimSpace(name)
	if i := logic.FindPreset(c.presets, p.Name); i >= 0 {
		c.presets[i] = p
	} else {
		c.presets = append(c.presets, p)
	}
	c.loaded = p.Name
	c.store()
}

// dialogWidth fits a dialog on narrow phone screens
func dialogWidth(win fyne.Window) float32 {
	return fyne.Min(360, win.Canvas().Size().Width-32)
}

// fitDialog gives a dialog the dialog width and the height its content needs
func fitDialog(d dialog.Dialog, win fyne.Window) {
	d.Resize(fyne.NewSize(dialogWidth(win), 1)) // the dialog keeps at least its content height
}

// showFitted shows a dialog fitted to its content. Wrapped texts only know their height once they
// have a width, so the dialog is fitted again after it is shown.
func showFitted(d dialog.Dialog, win fyne.Window) {
	fitDialog(d, win)
	d.Show()
	fitDialog(d, win)
}

// wrappedMuted is muted text that wraps, for explanations in dialogs
func wrappedMuted(text string) *widget.RichText {
	rt := widgets.NewMutedText(text)
	rt.Wrapping = fyne.TextWrapWord
	return rt
}

// micCount describes the number of mics, e.g. "the 6 mics"
func micCount(n int) string {
	if n == 1 {
		return "the mic"
	}
	return fmt.Sprintf("the %d mics", n)
}

// showNameDialog asks for a preset name: a full-width field, an optional explanation, and a
// confirm button that stays disabled while validate rejects the name (the reason shows below the field)
func showNameDialog(win fyne.Window, title, confirm, initial, info string, validate func(string) error, ok func(name string)) {
	name := widget.NewEntry()
	name.SetPlaceHolder("Preset name")
	name.SetText(initial)
	problem := widgets.NewStyledText("", theme.ColorNameError, theme.SizeNameText, fyne.TextStyle{})
	problem.Wrapping = fyne.TextWrapWord
	items := []fyne.CanvasObject{name, problem}
	if info != "" {
		items = append(items, wrappedMuted(info))
	}
	var d *dialog.CustomDialog
	confirmBtn := widget.NewButtonWithIcon(confirm, theme.ConfirmIcon(), func() {
		if validate(name.Text) != nil {
			return
		}
		d.Hide()
		ok(strings.TrimSpace(name.Text))
	})
	confirmBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() { d.Hide() })
	check := func(s string) {
		err := validate(s)
		msg := ""
		// An empty field needs no explanation; the disabled button says enough
		if err != nil && strings.TrimSpace(s) != "" {
			msg = sentence(err).Error()
		}
		if changed := (msg == "") != problem.Hidden || msg != problem.String(); changed {
			widgets.SetStyledText(problem, msg, "")
			problem.Hidden = msg == ""
			if d != nil {
				fitDialog(d, win)
			}
		}
		if err != nil {
			confirmBtn.Disable()
		} else {
			confirmBtn.Enable()
		}
	}
	name.OnChanged = check
	name.OnSubmitted = func(string) { confirmBtn.OnTapped() }
	check(initial)
	d = dialog.NewCustomWithoutButtons(title, container.NewVBox(items...), win)
	d.SetButtons([]fyne.CanvasObject{cancelBtn, confirmBtn})
	showFitted(d, win)
	win.Canvas().Focus(name)
	// Select the suggested name, so typing replaces it
	name.TypedShortcut(&fyne.ShortcutSelectAll{})
}

// showSaveAs asks for a name and saves the setup on screen as a new preset
func (c *presetController) showSaveAs(win fyne.Window) {
	cur := c.current()
	source := "the source"
	if s := strings.TrimSpace(cur.Source); s != "" {
		source = "the source (" + s + ")"
	}
	info := "Saves " + source + ", reference distance, units, sample rate and " +
		micCount(len(cur.Mics)) + ". The temperature is not saved."
	validate := func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New("enter a name")
		}
		return nil
	}
	showNameDialog(win, "Save preset", "Save", c.loaded, info, validate, func(n string) {
		if i := logic.FindPreset(c.presets, n); i >= 0 && !strings.EqualFold(n, c.loaded) {
			confirm := dialog.NewConfirm("Replace preset?", "A preset called “"+c.presets[i].Name+"” already exists. Replace it?",
				func(replace bool) {
					if replace {
						c.saveAs(c.presets[i].Name)
					}
				}, win)
			confirm.SetConfirmText("Replace")
			confirm.SetDismissText("Cancel")
			confirm.Show()
			return
		}
		c.saveAs(n)
	})
}

// presetList is a scrolling list of presets, up to four rows high. resize fits the height to the
// number of presets again after some were deleted.
func presetList(c *presetController, width float32, trailing func() fyne.CanvasObject,
	update func(i int, trailing fyne.CanvasObject)) (list *widget.List, sized fyne.CanvasObject, resize func()) {
	newRow := func() fyne.CanvasObject {
		name := widgets.NewStyledText("", theme.ColorNameForeground, theme.SizeNameText, fyne.TextStyle{Bold: true})
		name.Truncation = fyne.TextTruncateEllipsis
		summary := widgets.NewMutedText("")
		summary.Truncation = fyne.TextTruncateEllipsis
		gap := theme.Size(theme.SizeNameLineSpacing) - 2*theme.Size(theme.SizeNameInnerPadding)
		texts := container.New(newTextLinesLayout(gap), name, summary)
		return container.NewBorder(nil, nil, nil, trailing(), texts)
	}
	list = widget.NewList(
		func() int { return len(c.presets) },
		newRow,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)
			texts := row.Objects[0].(*fyne.Container)
			widgets.SetStyledText(texts.Objects[0].(*widget.RichText), c.presets[i].Name, "")
			widgets.SetStyledText(texts.Objects[1].(*widget.RichText), logic.PresetSummary(c.presets[i]), "")
			update(i, row.Objects[1])
		},
	)
	rowHeight := newRow().MinSize().Height + theme.Size(theme.SizeNamePadding)
	sizer := canvas.NewRectangle(color.Transparent)
	resize = func() {
		rows := fyne.Min(float32(len(c.presets)), 4)
		sizer.SetMinSize(fyne.NewSize(width, rows*rowHeight))
	}
	resize()
	return list, container.NewStack(sizer, list), resize
}

// newTextLinesLayout stacks texts with their line spacing
func newTextLinesLayout(gap float32) fyne.Layout {
	return &textLinesLayout{gap: gap}
}

type textLinesLayout struct{ gap float32 }

func (l *textLinesLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for _, o := range objects {
		h := o.MinSize().Height
		o.Move(fyne.NewPos(0, y))
		o.Resize(fyne.NewSize(size.Width, h))
		y += h + l.gap
	}
}

func (l *textLinesLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var w, h float32
	for i, o := range objects {
		m := o.MinSize()
		w = fyne.Max(w, m.Width)
		h += m.Height
		if i > 0 {
			h += l.gap
		}
	}
	return fyne.NewSize(w, h)
}

// showLoad lists the presets; tapping one loads it
func (c *presetController) showLoad(win fyne.Window) {
	width := dialogWidth(win) - 2*theme.Padding()
	var d *dialog.CustomDialog
	list, sized, _ := presetList(c, width,
		func() fyne.CanvasObject { return widget.NewIcon(theme.ConfirmIcon()) },
		func(i int, trailing fyne.CanvasObject) {
			if strings.EqualFold(c.presets[i].Name, c.loaded) {
				trailing.Show()
			} else {
				trailing.Hide()
			}
		})
	list.OnSelected = func(i widget.ListItemID) {
		p := c.presets[i]
		c.loaded = p.Name
		c.apply(p)
		c.changed()
		d.Hide()
	}
	d = dialog.NewCustom("Load preset", "Cancel", container.NewVBox(
		sized,
		wrappedMuted("Loading replaces the source, reference, units, sample rate and mics. The temperature stays as set."),
	), win)
	showFitted(d, win)
}

// showManage lists the presets with rename and delete buttons
func (c *presetController) showManage(win fyne.Window) {
	width := dialogWidth(win) - 2*theme.Padding()
	var d *dialog.CustomDialog
	var list *widget.List
	var resize func()
	list, sized, resize := presetList(c, width,
		func() fyne.CanvasObject {
			// Plain buttons: they let drags through, so the list still scrolls on phones
			rename := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil)
			rename.Importance = widget.LowImportance
			del := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			del.Importance = widget.LowImportance
			return container.NewHBox(rename, del)
		},
		func(i int, trailing fyne.CanvasObject) {
			buttons := trailing.(*fyne.Container).Objects
			buttons[0].(*widget.Button).OnTapped = func() { c.showRename(win, i, func() { list.Refresh() }) }
			buttons[1].(*widget.Button).OnTapped = func() {
				c.confirmDelete(win, i, func() {
					if len(c.presets) == 0 {
						d.Hide()
						return
					}
					list.Refresh()
					resize()
				})
			}
		})
	list.OnSelected = func(i widget.ListItemID) { list.Unselect(i) }
	d = dialog.NewCustom("Manage presets", "Close", container.NewVBox(
		sized,
		wrappedMuted("Deleting a preset can't be undone; export your presets first to keep a copy."),
	), win)
	showFitted(d, win)
}

// showRename asks for a new name for the preset at index
func (c *presetController) showRename(win fyne.Window, index int, done func()) {
	if index < 0 || index >= len(c.presets) {
		return
	}
	old := c.presets[index].Name
	validate := func(s string) error {
		return logic.RenamePreset(append([]logic.Preset(nil), c.presets...), index, s)
	}
	showNameDialog(win, "Rename preset", "Rename", old, "", validate, func(n string) {
		if logic.RenamePreset(c.presets, index, n) != nil {
			return
		}
		if strings.EqualFold(c.loaded, old) {
			c.loaded = c.presets[index].Name
		}
		c.store()
		done()
	})
}

// confirmDelete asks before deleting the preset at index
func (c *presetController) confirmDelete(win fyne.Window, index int, done func()) {
	if index < 0 || index >= len(c.presets) {
		return
	}
	name := c.presets[index].Name
	d := dialog.NewConfirm("Delete preset?", "“"+name+"” will be deleted. This can't be undone.", func(ok bool) {
		if !ok {
			return
		}
		c.presets = logic.DeletePreset(c.presets, index)
		if strings.EqualFold(c.loaded, name) {
			c.loaded = "" // the setup stays on screen, as an unnamed setup
		}
		c.store()
		done()
	}, win)
	d.SetConfirmText("Delete")
	d.SetDismissText("Cancel")
	d.SetConfirmImportance(widget.DangerImportance)
	d.Show()
}

// importPresets reads presets from a file chosen in the platform's file dialog and adds them
func (c *presetController) importPresets(win fyne.Window) {
	filedialog.OpenFile(win, "Import presets", presetFileFilter, func(data []byte, err error) {
		if err != nil {
			dialog.ShowError(sentence(fmt.Errorf("could not read the file: %w", err)), win)
			return
		}
		if data == nil {
			return // cancelled
		}
		imported, err := logic.UnmarshalPresets(data)
		if err != nil {
			dialog.ShowError(sentence(err), win)
			return
		}
		if len(imported) == 0 {
			dialog.ShowInformation("Import presets", "The file contains no presets.", win)
			return
		}
		var renamed []string
		c.presets, renamed = logic.MergePresets(c.presets, imported)
		c.store()
		msg := fmt.Sprintf("Imported %d preset%s.", len(imported), plural(len(imported)))
		if len(renamed) > 0 {
			msg += "\nRenamed because the name was taken: " + strings.Join(renamed, ", ") + "."
		}
		dialog.ShowInformation("Presets imported", msg, win)
	})
}

// exportPresets writes all presets to a file chosen in the platform's file dialog
func (c *presetController) exportPresets(win fyne.Window) {
	data, err := logic.MarshalPresets(c.presets)
	if err != nil {
		dialog.ShowError(err, win)
		return
	}
	n := len(c.presets)
	filedialog.SaveFile(win, "Export presets", "musicalc-presets.json", presetFileFilter, data, func(saved bool, err error) {
		if err != nil {
			dialog.ShowError(sentence(fmt.Errorf("could not save the file: %w", err)), win)
			return
		}
		if saved {
			dialog.ShowInformation("Presets exported", fmt.Sprintf("Exported %d preset%s.", n, plural(n)), win)
		}
	})
}

// sentence capitalises an error message for display; Go errors start in lower case
func sentence(err error) error {
	msg := err.Error()
	if msg == "" {
		return err
	}
	r := []rune(msg)
	return errors.New(strings.ToUpper(string(r[0])) + string(r[1:]))
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
