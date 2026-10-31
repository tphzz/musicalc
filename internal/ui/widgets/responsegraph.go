package widgets

import (
	"image/color"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// GraphCurve is one level curve: levels in dB at the given frequencies (sorted). A Before curve is
// drawn dashed and muted, to compare a change with the state before it.
type GraphCurve struct {
	Freqs  []float64
	Levels []float64
	Before bool
}

// Response graph scale and size
const (
	graphTopDB    = 9.0
	graphBottomDB = -24.0
	graphHeight   = 150
	graphDash     = 4 // dashed curves: drawn and skipped lengths
	graphGap      = 3
)

// graphGridDB are the level lines with labels
var graphGridDB = []float64{6, 0, -12}

// graphTicks are the frequencies the x axis may label
var graphTicks = []float64{10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000, 20000}

// ResponseGraph plots combined levels over frequency on a log scale, with a shaded range of
// interest and a cursor at one frequency
type ResponseGraph struct {
	widget.BaseWidget

	minF, maxF     float64
	curves         []GraphCurve
	bandLo, bandHi float64
	cursor         float64
	cursorDB       float64
}

// NewResponseGraph creates an empty response graph
func NewResponseGraph() *ResponseGraph {
	g := &ResponseGraph{minF: 20, maxF: 1000}
	g.ExtendBaseWidget(g)
	return g
}

// SetData shows curves between minF and maxF Hz, shades bandLo–bandHi and puts the cursor at
// cursor Hz on the level cursorDB
func (g *ResponseGraph) SetData(minF, maxF float64, curves []GraphCurve, bandLo, bandHi, cursor, cursorDB float64) {
	g.minF, g.maxF, g.curves = minF, maxF, curves
	g.bandLo, g.bandHi, g.cursor, g.cursorDB = bandLo, bandHi, cursor, cursorDB
	g.Refresh()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (g *ResponseGraph) CreateRenderer() fyne.WidgetRenderer {
	return &responseGraphRenderer{g: g}
}

type responseGraphRenderer struct {
	g       *ResponseGraph
	objects []fyne.CanvasObject
}

func (r *responseGraphRenderer) labelSize() float32 {
	return theme.Size(theme.SizeNameCaptionText)
}

func (r *responseGraphRenderer) MinSize() fyne.Size {
	return fyne.NewSize(120, graphHeight+r.labelSize()+4)
}

func (r *responseGraphRenderer) Layout(size fyne.Size) { r.build(size) }

func (r *responseGraphRenderer) Refresh() {
	r.build(r.g.Size())
	canvas.Refresh(r.g)
}

func (r *responseGraphRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *responseGraphRenderer) Destroy() {}

// build creates the graph's lines and labels for the given size
func (r *responseGraphRenderer) build(size fyne.Size) {
	g := r.g
	r.objects = r.objects[:0]
	if size.Width <= 0 || g.minF <= 0 || g.maxF <= g.minF {
		return
	}
	th := g.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()
	primary := th.Color(theme.ColorNamePrimary, v)
	muted := th.Color(theme.ColorNamePlaceHolder, v)
	fg := th.Color(theme.ColorNameForeground, v)
	withAlpha := func(c color.Color, a uint8) color.Color {
		r, gg, b, _ := c.RGBA()
		return color.NRGBA{R: uint8(r >> 8), G: uint8(gg >> 8), B: uint8(b >> 8), A: a}
	}

	w, h := size.Width, float32(graphHeight)
	x := func(f float64) float32 {
		return w * float32(math.Log(f/g.minF)/math.Log(g.maxF/g.minF))
	}
	y := func(db float64) float32 {
		db = math.Max(graphBottomDB, math.Min(graphTopDB, db))
		return h * float32((graphTopDB-db)/(graphTopDB-graphBottomDB))
	}
	line := func(x1, y1, x2, y2 float32, c color.Color, width float32) {
		l := canvas.NewLine(c)
		l.StrokeWidth = width
		l.Position1, l.Position2 = fyne.NewPos(x1, y1), fyne.NewPos(x2, y2)
		r.objects = append(r.objects, l)
	}
	text := func(s string, c color.Color, px, py float32, alignRight bool) {
		t := canvas.NewText(s, c)
		t.TextSize = r.labelSize()
		ts := t.MinSize()
		if alignRight {
			px -= ts.Width
		}
		px = float32(math.Max(0, math.Min(float64(px), float64(w-ts.Width))))
		t.Move(fyne.NewPos(px, py))
		t.Resize(ts)
		r.objects = append(r.objects, t)
	}

	// Range of interest
	if g.bandHi > g.bandLo && g.bandLo > 0 {
		lo, hi := x(math.Max(g.bandLo, g.minF)), x(math.Min(g.bandHi, g.maxF))
		if hi > lo {
			band := canvas.NewRectangle(withAlpha(primary, 0x38))
			band.Move(fyne.NewPos(lo, 0))
			band.Resize(fyne.NewSize(hi-lo, h))
			r.objects = append(r.objects, band)
		}
	}
	// Grid: level lines and frequency lines
	for _, db := range graphGridDB {
		alpha := uint8(0x50)
		if db == 0 {
			alpha = 0xa0
		}
		line(0, y(db), w, y(db), withAlpha(muted, alpha), 1)
	}
	var ticks []float64
	for _, f := range graphTicks {
		if f >= g.minF && f <= g.maxF {
			ticks = append(ticks, f)
		}
	}
	for _, f := range ticks {
		line(x(f), 0, x(f), h, withAlpha(muted, 0x30), 1)
	}

	// Curves: the before curve first, so the current one is drawn on top
	for _, before := range []bool{true, false} {
		for _, c := range g.curves {
			if c.Before != before || len(c.Freqs) < 2 || len(c.Freqs) != len(c.Levels) {
				continue
			}
			col, width := primary, float32(2)
			if before {
				col, width = withAlpha(muted, 0xc0), 1.5
			}
			// Several samples in one pixel column (a dense comb) are drawn as one vertical line from
			// their lowest to their highest level, so the graph shows how far the level swings there
			type column struct{ x, first, last, min, max float32 }
			var cols []column
			for i := range c.Freqs {
				px, py := float32(math.Round(float64(x(c.Freqs[i])))), y(c.Levels[i])
				if n := len(cols); n > 0 && cols[n-1].x == px {
					cl := &cols[n-1]
					cl.last = py
					cl.min, cl.max = float32(math.Min(float64(cl.min), float64(py))), float32(math.Max(float64(cl.max), float64(py)))
					continue
				}
				cols = append(cols, column{x: px, first: py, last: py, min: py, max: py})
			}
			var travelled float32 // along the curve, for the dash pattern
			for i, cl := range cols {
				if cl.max-cl.min > 0.5 {
					band := col
					if before {
						band = withAlpha(muted, 0x60) // keep a dense before-curve in the background
					}
					line(cl.x, cl.min, cl.x, cl.max, band, width/2+0.5)
				}
				if i == 0 {
					continue
				}
				x1, y1, x2, y2 := cols[i-1].x, cols[i-1].last, cl.x, cl.first
				if !before {
					line(x1, y1, x2, y2, col, width)
					continue
				}
				// Dashed: split the segment where the pattern switches between dash and gap
				seg := float32(math.Hypot(float64(x2-x1), float64(y2-y1)))
				for done := float32(0); done < seg; {
					phase := float32(math.Mod(float64(travelled), graphDash+graphGap))
					step := float32(graphDash) - phase
					on := phase < graphDash
					if !on {
						step = graphDash + graphGap - phase
					}
					step = float32(math.Min(float64(step), float64(seg-done)))
					if on {
						t1, t2 := done/seg, (done+step)/seg
						line(x1+(x2-x1)*t1, y1+(y2-y1)*t1, x1+(x2-x1)*t2, y1+(y2-y1)*t2, col, width)
					}
					done += step
					travelled += step
				}
			}
		}
	}

	// Cursor
	if g.cursor >= g.minF && g.cursor <= g.maxF {
		cx := x(g.cursor)
		line(cx, 0, cx, h, fg, 1.5)
		dot := canvas.NewCircle(fg)
		dot.Move(fyne.NewPos(cx-4, y(g.cursorDB)-4))
		dot.Resize(fyne.NewSize(8, 8))
		r.objects = append(r.objects, dot)
	}

	// Labels: levels inside the plot on the left, frequencies below
	for _, db := range graphGridDB {
		s := strconv.FormatFloat(math.Abs(db), 'f', -1, 64)
		switch {
		case db > 0:
			s = "+" + s
		case db < 0:
			s = "−" + s
		}
		text(s, muted, 2, y(db)-r.labelSize()-2, false)
	}
	for i, f := range ticks {
		s := strconv.FormatFloat(f, 'f', -1, 64)
		if f >= 1000 {
			s = strconv.FormatFloat(f/1000, 'f', -1, 64) + "k"
		}
		if i == len(ticks)-1 {
			s += " Hz"
		}
		text(s, muted, x(f)-6, h+2, false)
	}
}
