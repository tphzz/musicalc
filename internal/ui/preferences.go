package ui

import (
	"errors"
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"musicalc/internal/i18n"
	"musicalc/internal/platform/displayscale"
	"musicalc/internal/ui/widgets"
)

// PreferencesDialog applies language and desktop zoom without replacing views.
type PreferencesDialog struct {
	widget.BaseWidget
	canvas                               fyne.Canvas
	scale                                *displayscale.Controller
	onLanguageChanged                    func()
	choices                              *languageChoices
	title                                *canvas.Text
	languageLabel, scaleLabel, scaleHint *widget.RichText
	scaleSelect                          *preferencesScaleSelect
	close                                *preferencesCloseButton
}

func NewPreferencesDialog(preferences fyne.Preferences, parent fyne.Window, scale *displayscale.Controller, onLanguageChanged func()) *PreferencesDialog {
	d := &PreferencesDialog{canvas: parent.Canvas(), scale: scale, onLanguageChanged: onLanguageChanged}
	d.ExtendBaseWidget(d)
	d.BaseWidget.Hide()
	labels := d.languageLabels()
	d.choices = newLanguageChoices(d, []languageOption{{Value: "system", Label: labels[0]}, {Value: "en", Label: labels[1]}, {Value: "de", Label: labels[2]}})
	selected := preferences.String(prefLanguage)
	if selected != "en" && selected != "de" {
		selected = "system"
	}
	d.choices.SetSelected(selected)
	d.choices.OnChanged = func(code string) {
		preferences.SetString(prefLanguage, code)
		i18n.Use(i18n.ResolveLanguage(code, string(lang.SystemLocale())))
		d.RefreshText()
		if d.onLanguageChanged != nil {
			d.onLanguageChanged()
		}
	}
	if scale != nil {
		picker := &preferencesScaleSelect{dialog: d}
		picker.ExtendBaseWidget(picker)
		picker.OnChanged = func(option string) {
			for i, percent := range displayscale.Percents() {
				if i < len(picker.Options) && option == picker.Options[i] {
					if err := scale.SetPercent(percent); err != nil {
						dialog.ShowError(errors.New(i18n.T("preferences.scale_error", err)), parent)
					}
					d.RefreshText()
					return
				}
			}
		}
		d.scaleSelect = picker
	}
	d.RefreshText()
	return d
}

func (d *PreferencesDialog) languageLabels() []string {
	detected := i18n.ResolveLanguage("system", string(lang.SystemLocale()))
	name := i18n.T("language.english")
	if detected == "de" {
		name = i18n.T("language.german")
	}
	return []string{i18n.T("preferences.system", name), i18n.T("language.english"), i18n.T("language.german")}
}

func scaleLabel(percent int) string {
	if percent == 100 {
		return i18n.T("preferences.scale_default", percent)
	}
	return fmt.Sprintf("%d%%", percent)
}

func (d *PreferencesDialog) scaleHintText() string {
	if d.scale != nil && d.scale.Override() != "" {
		return i18n.T("preferences.scale_override")
	}
	return i18n.T("preferences.scale_hint", 100)
}

// RefreshText updates labels by value, without triggering another preference change.
func (d *PreferencesDialog) RefreshText() {
	d.choices.setLabels(d.languageLabels())
	if d.scaleSelect != nil {
		options := make([]string, 0, len(displayscale.Percents())+1)
		for _, percent := range displayscale.Percents() {
			options = append(options, scaleLabel(percent))
		}
		selected := scaleLabel(d.scale.Percent())
		if value := d.scale.Override(); value != "" {
			factor, _ := strconv.ParseFloat(value, 64)
			selected = strconv.FormatFloat(factor*100, 'f', -1, 64) + "%"
			found := false
			for _, option := range options {
				if option == selected {
					found = true
				}
			}
			if !found {
				options = append(options, selected)
			}
			d.scaleSelect.Disable()
		}
		d.scaleSelect.Options = options
		d.scaleSelect.Selected = selected
		d.scaleSelect.Refresh()
	}
	if d.title == nil {
		return
	}
	d.title.Text = i18n.T("preferences.title")
	d.title.Refresh()
	widgets.SetStyledText(d.languageLabel, i18n.T("preferences.language"), "")
	if d.scaleLabel != nil {
		widgets.SetStyledText(d.scaleLabel, i18n.T("preferences.scale"), "")
		widgets.SetStyledText(d.scaleHint, d.scaleHintText(), "")
	}
	d.close.SetText(i18n.T("action.close"))
	d.Refresh()
}

func (d *PreferencesDialog) Show() {
	if d.Visible() {
		return
	}
	d.RefreshText()
	d.BaseWidget.Show()
	d.canvas.Overlays().Add(d)
}
func (d *PreferencesDialog) Hide() {
	if d.Visible() {
		for d.dismissChildOverlay() {
		}
		d.canvas.Overlays().Remove(d)
	}
	d.BaseWidget.Hide()
}

// Use the child's backdrop dismissal so native controls clear their popup state too.
func (d *PreferencesDialog) dismissChildOverlay() bool {
	layers := d.canvas.Overlays().List()
	if len(layers) == 0 || layers[len(layers)-1] == d {
		return false
	}
	top := layers[len(layers)-1]
	if tappable, ok := top.(fyne.Tappable); ok {
		tappable.Tapped(&fyne.PointEvent{})
	}
	layers = d.canvas.Overlays().List()
	if len(layers) > 0 && layers[len(layers)-1] == top {
		top.Hide()
		d.canvas.Overlays().Remove(top)
	}
	return true
}
func (d *PreferencesDialog) Tapped(*fyne.PointEvent) {}
func (d *PreferencesDialog) closeKey(event *fyne.KeyEvent) bool {
	if !d.Visible() || (event.Name != fyne.KeyEscape && event.Name != mobile.KeyBack) {
		return false
	}
	if d.dismissChildOverlay() {
		return true
	}
	d.Hide()
	return true
}
func (d *PreferencesDialog) HandleKey(event *fyne.KeyEvent) bool {
	if !d.Visible() {
		return false
	}
	d.closeKey(event)
	return true
}

func (d *PreferencesDialog) CreateRenderer() fyne.WidgetRenderer {
	th := d.Theme()
	title := canvas.NewText(i18n.T("preferences.title"), th.Color(theme.ColorNameForeground, fyne.CurrentApp().Settings().ThemeVariant()))
	title.TextSize = th.Size(theme.SizeNameSubHeadingText)
	title.TextStyle.Bold = true
	d.title = title
	label := func(text string) (*widget.RichText, fyne.CanvasObject) {
		value := widgets.NewStyledText(text, theme.ColorNamePlaceHolder, theme.SizeNameText, fyne.TextStyle{Bold: true})
		return value, container.NewThemeOverride(value, modalTextTheme{Theme: th, size: th.Size(theme.SizeNameText)})
	}
	var languageLabel fyne.CanvasObject
	d.languageLabel, languageLabel = label(i18n.T("preferences.language"))
	body := container.New(layout.NewCustomPaddedVBoxLayout(12), languageLabel, d.choices)
	if d.scaleSelect != nil {
		var scaleLabel fyne.CanvasObject
		d.scaleLabel, scaleLabel = label(i18n.T("preferences.scale"))
		d.scaleHint = wrappedMuted(d.scaleHintText())
		hint := container.NewThemeOverride(d.scaleHint, modalTextTheme{Theme: th, size: 13})
		body.Add(container.New(layout.NewCustomPaddedVBoxLayout(8), scaleLabel, d.scaleSelect, hint))
	}
	scroll := container.NewVScroll(body)
	close := &preferencesCloseButton{dialog: d}
	close.Text = i18n.T("action.close")
	close.OnTapped = d.Hide
	close.ExtendBaseWidget(close)
	d.close = close
	content := container.New(&preferencesPanelLayout{body: body}, title, scroll, close)
	return newModalRenderer(&d.BaseWidget, title, content)
}

type preferencesCloseButton struct {
	widgets.Button
	dialog *PreferencesDialog
}

func (b *preferencesCloseButton) TypedKey(event *fyne.KeyEvent) {
	if !b.dialog.closeKey(event) {
		b.Button.TypedKey(event)
	}
}

type preferencesScaleSelect struct {
	widget.Select
	dialog *PreferencesDialog
}

func (s *preferencesScaleSelect) TypedKey(event *fyne.KeyEvent) {
	if !s.dialog.closeKey(event) {
		s.Select.TypedKey(event)
	}
}

type preferencesPanelLayout struct{ body fyne.CanvasObject }

func (l *preferencesPanelLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 2*modalPadding+24+objects[0].MinSize().Height+l.body.MinSize().Height+objects[2].MinSize().Height)
}
func (l *preferencesPanelLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	width := fyne.Max(0, size.Width-2*modalPadding)
	title, body, close := objects[0], objects[1], objects[2]
	title.Move(fyne.NewPos(modalPadding, modalPadding))
	title.Resize(fyne.NewSize(width, title.MinSize().Height))
	close.Resize(close.MinSize())
	close.Move(fyne.NewPos(size.Width-modalPadding-close.Size().Width, size.Height-modalPadding-close.Size().Height))
	bodyY := modalPadding + title.MinSize().Height + 12
	body.Move(fyne.NewPos(modalPadding, bodyY))
	body.Resize(fyne.NewSize(width, fyne.Max(0, close.Position().Y-12-bodyY)))
}
