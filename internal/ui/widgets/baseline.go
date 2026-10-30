package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// NewBaselineRow places single-line texts of different sizes side by side so that their
// baselines line up, like "115.83 BPM". Aligning the text boxes' bottoms instead puts small text
// too low, because a large font reserves more room below its baseline. gap is the extra space
// between the texts, on top of their own inner padding (it may be negative to overlap it).
func NewBaselineRow(gap float32, texts ...*widget.RichText) fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, len(texts))
	for i, t := range texts {
		objects[i] = t
	}
	return &fyne.Container{Layout: &baselineLayout{gap: gap, texts: texts}, Objects: objects}
}

type baselineLayout struct {
	gap   float32
	texts []*widget.RichText
}

// baseline returns the distance from the top of a text's box to its baseline. All texts share
// the same inner padding, so it is left out.
func (l *baselineLayout) baseline(t *widget.RichText) float32 {
	if len(t.Segments) == 0 {
		return 0
	}
	seg, ok := t.Segments[0].(*widget.TextSegment)
	if !ok {
		return 0
	}
	th := t.Theme()
	sizeName := seg.Style.SizeName
	if sizeName == "" {
		sizeName = theme.SizeNameText
	}
	_, base := fyne.CurrentApp().Driver().RenderedTextSize(seg.Text, th.Size(sizeName), seg.Style.TextStyle, nil)
	return base
}

// offsets returns each text's y position and the row height
func (l *baselineLayout) offsets() ([]float32, float32) {
	maxBase := float32(0)
	bases := make([]float32, len(l.texts))
	for i, t := range l.texts {
		bases[i] = l.baseline(t)
		maxBase = fyne.Max(maxBase, bases[i])
	}
	ys := make([]float32, len(l.texts))
	height := float32(0)
	for i, t := range l.texts {
		ys[i] = maxBase - bases[i]
		height = fyne.Max(height, ys[i]+t.MinSize().Height)
	}
	return ys, height
}

func (l *baselineLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	ys, _ := l.offsets()
	x := float32(0)
	for i, t := range l.texts {
		min := t.MinSize()
		t.Move(fyne.NewPos(x, ys[i]))
		t.Resize(min)
		x += min.Width + l.gap
	}
}

func (l *baselineLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	_, height := l.offsets()
	width := float32(0)
	for i, t := range l.texts {
		if i > 0 {
			width += l.gap
		}
		width += t.MinSize().Width
	}
	return fyne.NewSize(width, height)
}
