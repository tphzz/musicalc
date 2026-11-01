package help

import (
	"bytes"
	"image"
	_ "image/png" // help images are PNGs
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// imageMaxWidth is the widest a help image is shown, so it fits a phone screen
const imageMaxWidth = 320

// Extra space above a section heading and around an image, on top of the usual line spacing
const (
	sectionGap = 10
	imageGap   = 8
)

// NewPage renders a help page in a vertical scroll view
func NewPage(md string) *container.Scroll {
	var scroll *container.Scroll
	rt := newRichText(md, func() {
		// A table changed height for a new width: lay the page out again
		if scroll != nil {
			scroll.Content.Refresh()
			scroll.Refresh()
		}
	})
	scroll = container.NewVScroll(container.New(layout.NewCustomPaddedLayout(4, 16, 8, 8), rt))
	return scroll
}

// RefreshPage replaces translated prose while retaining the scroll container and offset.
func RefreshPage(page *container.Scroll, md string) {
	offset := page.Offset
	text := newRichText(md, func() { page.Content.Refresh(); page.Refresh() })
	content := page.Content.(*fyne.Container)
	content.Objects = []fyne.CanvasObject{text}
	content.Refresh()
	page.Offset = offset
	page.Refresh()
}

// newRichText renders a help page's Markdown. Fyne's own segments are swapped where they don't
// suit the help: images come from the embedded files (Fyne would look for them on disk), tables
// wrap their cells to fit a phone screen, and section headings get some room above them.
// relayout is called when a table's height changes with the page width.
func newRichText(md string, relayout func()) *widget.RichText {
	rt := widget.NewRichTextFromMarkdown(md)
	rt.Wrapping = fyne.TextWrapWord
	segs := make([]widget.RichTextSegment, 0, len(rt.Segments))
	for i, s := range rt.Segments {
		switch seg := s.(type) {
		case *widget.ImageSegment:
			if seg.Source != nil {
				if res := Image(seg.Source.Name()); res != nil {
					s = &imageSegment{res: res}
				}
			}
		case *widget.TableSegment:
			s = &tableSegment{src: seg, relayout: relayout}
		case *widget.TextSegment:
			if i > 0 && isHeading(seg) {
				segs = append(segs, &gapSegment{height: sectionGap})
			}
		}
		segs = append(segs, s)
	}
	rt.Segments = segs
	rt.Refresh()
	return rt
}

func isHeading(seg *widget.TextSegment) bool {
	return seg.Style.SizeName == theme.SizeNameHeadingText || seg.Style.SizeName == theme.SizeNameSubHeadingText
}

// blockSegment holds what the help's own block segments have in common: no text, no selection
type blockSegment struct{}

func (blockSegment) Inline() bool              { return false }
func (blockSegment) Textual() string           { return "" }
func (blockSegment) Update(fyne.CanvasObject)  {}
func (blockSegment) Select(_, _ fyne.Position) {}
func (blockSegment) SelectedText() string      { return "" }
func (blockSegment) Unselect()                 {}

// gapSegment is empty vertical space
type gapSegment struct {
	blockSegment
	height float32
}

func (s *gapSegment) Visual() fyne.CanvasObject {
	gap := canvas.NewRectangle(nil)
	gap.SetMinSize(fyne.NewSize(0, s.height))
	return gap
}

// imageSegment shows an embedded image as a block. Images are made at twice the pixel density,
// so they are shown at half their pixel size, and never wider than imageMaxWidth.
type imageSegment struct {
	blockSegment
	res fyne.Resource
}

func (s *imageSegment) Visual() fyne.CanvasObject {
	img := canvas.NewImageFromResource(s.res)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(imageSize(s.res))
	return container.New(layout.NewCustomPaddedLayout(imageGap, imageGap, 0, 0), container.NewCenter(img))
}

// imageSize is the size an image is shown at: half its pixels, scaled down to imageMaxWidth
func imageSize(res fyne.Resource) fyne.Size {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(res.Content()))
	if err != nil || cfg.Width == 0 {
		return fyne.NewSize(imageMaxWidth, imageMaxWidth/2)
	}
	w, h := float32(cfg.Width)/2, float32(cfg.Height)/2
	if w > imageMaxWidth {
		h = h * imageMaxWidth / w
		w = imageMaxWidth
	}
	return fyne.NewSize(w, h)
}

// tableSegment shows a Markdown table whose cells wrap, so it fits the page width instead of
// scrolling sideways like Fyne's own table
type tableSegment struct {
	blockSegment
	src      *widget.TableSegment
	relayout func()
}

func (s *tableSegment) Textual() string      { return s.src.Textual() }
func (s *tableSegment) SelectedText() string { return s.src.Textual() }
func (s *tableSegment) Visual() fyne.CanvasObject {
	return newFitTable(s.src, s.relayout)
}

// fitTable is a table that fits its width: columns get their natural width when there is room,
// otherwise the space beyond each column's longest word is shared in proportion to how much
// text it holds, and cells wrap
type fitTable struct {
	widget.BaseWidget
	cols     int
	cells    []*widget.RichText // row-major, header row first
	header   bool
	natural  []float32 // per cell: unwrapped width
	longest  []float32 // per cell: widest word
	width    float32   // the width the table was last laid out at
	relayout func()
}

// fitTableDefaultWidth is assumed until the table is first laid out: a phone's page width
const fitTableDefaultWidth = 328

func newFitTable(src *widget.TableSegment, relayout func()) *fitTable {
	cols := 0
	for _, r := range append([][][]widget.RichTextSegment{src.Headers}, src.Rows...) {
		cols = max(cols, len(r))
	}
	t := &fitTable{cols: cols, header: src.Headers != nil, relayout: relayout}
	addRow := func(cells [][]widget.RichTextSegment, header bool) {
		for c := 0; c < cols; c++ {
			var segs []widget.RichTextSegment
			if c < len(cells) {
				segs = cells[c]
			}
			t.addCell(segs, header)
		}
	}
	if t.header {
		addRow(src.Headers, true)
	}
	for _, r := range src.Rows {
		addRow(r, false)
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *fitTable) addCell(segs []widget.RichTextSegment, header bool) {
	var text strings.Builder
	for _, s := range segs {
		if ts, ok := s.(*widget.TextSegment); ok && header {
			ts.Style.TextStyle.Bold = true
		}
		text.WriteString(s.Textual())
	}
	if len(segs) == 0 {
		segs = []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleInline, Text: " "}}
	}
	cell := widget.NewRichText(segs...)
	cell.Wrapping = fyne.TextWrapWord
	t.cells = append(t.cells, cell)

	style := fyne.TextStyle{Bold: header}
	size := theme.TextSize()
	pad := 2 * theme.InnerPadding()
	t.natural = append(t.natural, fyne.MeasureText(text.String(), size, style).Width+pad)
	var widest float32
	for _, w := range strings.Fields(text.String()) {
		widest = max(widest, fyne.MeasureText(w, size, style).Width)
	}
	t.longest = append(t.longest, widest+pad)
}

// columnWidths shares width among the columns
func (t *fitTable) columnWidths(width float32) []float32 {
	natural := make([]float32, t.cols)
	longest := make([]float32, t.cols)
	for i := range t.cells {
		c := i % t.cols
		natural[c] = max(natural[c], t.natural[i])
		longest[c] = max(longest[c], t.longest[i])
	}
	avail := width - float32(t.cols+1)*theme.SeparatorThicknessSize()
	widths := make([]float32, t.cols)
	var sumNatural float32
	for c := range natural {
		sumNatural += natural[c]
	}
	if sumNatural <= avail {
		for c := range widths {
			widths[c] = natural[c] + (avail-sumNatural)/float32(t.cols)
		}
		return widths
	}

	// Columns that fit an even share of what is left keep their natural width, so short
	// columns don't wrap; repeat, as each one that fits leaves more for the others
	fits := make([]bool, t.cols)
	left, open := avail, t.cols
	for changed := true; changed && open > 0; {
		changed = false
		for c := range natural {
			if !fits[c] && natural[c] <= left/float32(open) {
				fits[c], widths[c] = true, natural[c]
				left -= natural[c]
				open--
				changed = true
			}
		}
	}

	// The wrapping columns get their longest word, and share the rest in proportion to their text
	var sumNat, sumLongest float32
	for c := range natural {
		if !fits[c] {
			sumNat += natural[c]
			sumLongest += longest[c]
		}
	}
	for c := range widths {
		switch {
		case fits[c]:
		case sumLongest >= left || sumNat == sumLongest:
			widths[c] = longest[c]
		default:
			widths[c] = longest[c] + (natural[c]-longest[c])*(left-sumLongest)/(sumNat-sumLongest)
		}
	}
	return widths
}

// rowHeights wraps every cell to its column width and returns the tallest cell of each row
func (t *fitTable) rowHeights(widths []float32) []float32 {
	heights := make([]float32, (len(t.cells)+t.cols-1)/t.cols)
	for i, cell := range t.cells {
		cell.Resize(fyne.NewSize(widths[i%t.cols], cell.Size().Height))
		heights[i/t.cols] = max(heights[i/t.cols], cell.MinSize().Height)
	}
	return heights
}

func (t *fitTable) layoutWidth() float32 {
	if t.width > 0 {
		return t.width
	}
	return fitTableDefaultWidth
}

func (t *fitTable) MinSize() fyne.Size {
	widths := t.columnWidths(t.layoutWidth())
	gap := theme.SeparatorThicknessSize()
	h := gap
	for _, rh := range t.rowHeights(widths) {
		h += rh + gap
	}
	return fyne.NewSize(0, h)
}

func (t *fitTable) Resize(size fyne.Size) {
	changed := t.width != 0 && size.Width != t.width
	t.width = size.Width
	t.BaseWidget.Resize(size)
	if changed && t.relayout != nil {
		// The height depends on the width: lay the page out again once this layout pass is done
		fyne.Do(t.relayout)
	}
}

func (t *fitTable) CreateRenderer() fyne.WidgetRenderer {
	r := &fitTableRenderer{t: t, border: canvas.NewRectangle(theme.Color(theme.ColorNameInputBorder))}
	r.objects = append(r.objects, r.border)
	for i, cell := range t.cells {
		fill := theme.Color(theme.ColorNameBackground)
		if t.header && i < t.cols {
			fill = theme.Color(theme.ColorNameHeaderBackground)
		}
		bg := canvas.NewRectangle(fill)
		r.backgrounds = append(r.backgrounds, bg)
		r.objects = append(r.objects, bg, cell)
	}
	return r
}

type fitTableRenderer struct {
	t           *fitTable
	border      *canvas.Rectangle
	backgrounds []*canvas.Rectangle
	objects     []fyne.CanvasObject
}

// Layout draws the grid lines as the border showing through a one-pixel gap around each cell
func (r *fitTableRenderer) Layout(size fyne.Size) {
	t := r.t
	widths := t.columnWidths(size.Width)
	heights := t.rowHeights(widths)
	gap := theme.SeparatorThicknessSize()
	var width float32 = gap
	for _, w := range widths {
		width += w + gap
	}
	height := gap
	for _, h := range heights {
		height += h + gap
	}
	r.border.Resize(fyne.NewSize(width, height))
	y := gap
	for row, rh := range heights {
		x := gap
		for c, w := range widths {
			i := row*t.cols + c
			if i >= len(t.cells) {
				break
			}
			for _, o := range []fyne.CanvasObject{r.backgrounds[i], t.cells[i]} {
				o.Move(fyne.NewPos(x, y))
				o.Resize(fyne.NewSize(w, rh))
			}
			x += w + gap
		}
		y += rh + gap
	}
}

func (r *fitTableRenderer) MinSize() fyne.Size           { return r.t.MinSize() }
func (r *fitTableRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *fitTableRenderer) Destroy()                     {}

func (r *fitTableRenderer) Refresh() {
	r.border.FillColor = theme.Color(theme.ColorNameInputBorder)
	for i, bg := range r.backgrounds {
		bg.FillColor = theme.Color(theme.ColorNameBackground)
		if r.t.header && i < r.t.cols {
			bg.FillColor = theme.Color(theme.ColorNameHeaderBackground)
		}
	}
	r.Layout(r.t.Size())
	canvas.Refresh(r.t)
}
