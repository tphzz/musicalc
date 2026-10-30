package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SizeNameCardPadding is the theme size for the visible space between a card's edge and its text
const SizeNameCardPadding fyne.ThemeSizeName = "cardPadding"

// SizeNameDisplayText is the theme size for a large result value, like the tapped tempo
const SizeNameDisplayText fyne.ThemeSizeName = "displayText"

// Card is a rounded, slightly raised surface that groups one result.
// Unlike widget.Card it uses the button color, so it stays visible on the dark background.
type Card struct {
	widget.BaseWidget
	Content     fyne.CanvasObject
	highlighted bool
}

// NewCard creates a card around content
func NewCard(content fyne.CanvasObject) *Card {
	c := &Card{Content: content}
	c.ExtendBaseWidget(c)
	return c
}

// SetHighlighted outlines the card in the primary color
func (c *Card) SetHighlighted(highlighted bool) {
	c.highlighted = highlighted
	c.Refresh()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (c *Card) CreateRenderer() fyne.WidgetRenderer {
	r := &cardRenderer{card: c, background: canvas.NewRectangle(color.Transparent)}
	r.applyTheme()
	return r
}

type cardRenderer struct {
	card       *Card
	background *canvas.Rectangle
}

// inset is the layout padding around the content. Labels already draw their own inner
// padding, so it is subtracted to keep the visible text inset at the card padding size.
func (r *cardRenderer) inset() float32 {
	th := r.card.Theme()
	p := th.Size(SizeNameCardPadding)
	if p == 0 {
		p = 2 * th.Size(theme.SizeNameInnerPadding)
	}
	return fyne.Max(p-th.Size(theme.SizeNameInnerPadding), 0)
}

func (r *cardRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	p := r.inset()
	r.card.Content.Move(fyne.NewPos(p, p))
	r.card.Content.Resize(size.Subtract(fyne.NewSquareSize(2 * p)))
}

func (r *cardRenderer) MinSize() fyne.Size {
	return r.card.Content.MinSize().Add(fyne.NewSquareSize(2 * r.inset()))
}

func (r *cardRenderer) Refresh() {
	r.applyTheme()
	r.background.Refresh()
	r.Layout(r.card.Size())
}

func (r *cardRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.card.Content}
}

func (r *cardRenderer) Destroy() {}

func (r *cardRenderer) applyTheme() {
	th := r.card.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	r.background.FillColor = th.Color(theme.ColorNameButton, v)
	r.background.CornerRadius = th.Size(theme.SizeNameCardRadius)
	// Same shadow as Fyne's own cards (widget.CardLevel)
	r.background.Shadow = canvas.Shadow{
		Color:      th.Color(theme.ColorNameShadow, v),
		BlurRadius: 8,
		Offset:     fyne.NewPos(0, 2),
		Variant:    canvas.DropShadow,
	}
	if r.card.highlighted {
		r.background.StrokeColor = th.Color(theme.ColorNamePrimary, v)
		r.background.StrokeWidth = 2
	} else {
		r.background.StrokeWidth = 0
	}
}

// NewStyledText creates single-line text with a theme color, size and style,
// for card values that need more emphasis or less than a plain Label
func NewStyledText(text string, colorName fyne.ThemeColorName, sizeName fyne.ThemeSizeName, style fyne.TextStyle) *widget.RichText {
	return widget.NewRichText(&widget.TextSegment{
		Text: text,
		Style: widget.RichTextStyle{
			ColorName: colorName,
			SizeName:  sizeName,
			TextStyle: style,
			Inline:    true,
		},
	})
}

// NewMutedText creates single-line text in the placeholder color, for secondary values.
// Label's LowImportance uses the disabled color, which is too faint for values people read.
func NewMutedText(text string) *widget.RichText {
	return NewStyledText(text, theme.ColorNamePlaceHolder, theme.SizeNameText, fyne.TextStyle{})
}

// SetStyledText changes the text of a NewStyledText or NewMutedText, and optionally its color
func SetStyledText(rt *widget.RichText, text string, colorName fyne.ThemeColorName) {
	seg := rt.Segments[0].(*widget.TextSegment)
	seg.Text = text
	if colorName != "" {
		seg.Style.ColorName = colorName
	}
	rt.Refresh()
}
