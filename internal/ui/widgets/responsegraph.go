package widgets

import (
	"image"
	"image/color"
	"math"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/math/fixed"
)

// GraphCurve is one level curve: levels in dB at the given frequencies (sorted). A Before curve is
// drawn dashed and muted, to compare a change with the state before it.
//
// A comb repeats every Spacing Hz between EnvLo and EnvHi dB. On a log scale its cycles get
// narrower towards high frequencies; where a cycle is narrower than a few pixels, the curve is drawn
// as a band between EnvLo and EnvHi instead of a line that could only scribble. Spacing 0 means no comb.
type GraphCurve struct {
	Freqs  []float64
	Levels []float64
	Before bool

	Spacing      float64
	EnvLo, EnvHi float64
}

// Response graph scale, size and strokes
const (
	graphTopDB    = 9.0
	graphBottomDB = -24.0
	graphHeight   = 150

	graphLineWidth   = 2.5 // the current curve
	graphBeforeWidth = 1.25
	graphDash        = 4 // dashed before-curve: drawn and skipped lengths
	graphGap         = 3

	// graphMinCyclePx is the narrowest comb cycle, in pixels, still drawn as a line; a dashed
	// before-curve needs wider cycles to stay readable, so it turns into its band earlier
	graphMinCyclePx       = 6
	graphMinBeforeCyclePx = 14
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
	r := &responseGraphRenderer{g: g}
	r.raster = canvas.NewRaster(r.drawCurves)
	return r
}

// graphPolyline is a curve's line part in graph coordinates, drawn into the raster
type graphPolyline struct {
	points []fyne.Position
	color  color.Color
	width  float32
	dashed bool
}

type responseGraphRenderer struct {
	g       *ResponseGraph
	objects []fyne.CanvasObject

	// The curves are drawn anti-aliased into a raster at the device's pixel density; grid, bands and
	// labels stay canvas objects, which are crisp anyway
	raster    *canvas.Raster
	mu        sync.Mutex
	lines     []graphPolyline
	plotWidth float32
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
	r.raster.Refresh()
	canvas.Refresh(r.g)
}

func (r *responseGraphRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *responseGraphRenderer) Destroy() {}

// drawCurves renders the curves' line parts at the raster's pixel size
func (r *responseGraphRenderer) drawCurves(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	r.mu.Lock()
	lines, plotWidth := r.lines, r.plotWidth
	r.mu.Unlock()
	if plotWidth <= 0 || w <= 0 || h <= 0 {
		return img
	}
	scale := float64(w) / float64(plotWidth) // device pixels per graph unit
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	dasher := rasterx.NewDasher(w, h, scanner)
	for _, l := range lines {
		if len(l.points) < 2 {
			continue
		}
		var dashes []float64
		if l.dashed {
			dashes = []float64{graphDash * scale, graphGap * scale}
		}
		width := fixed.Int26_6(float64(l.width) * scale * 64)
		dasher.SetStroke(width, 4*64, rasterx.RoundCap, rasterx.RoundCap, rasterx.RoundGap, rasterx.Round, dashes, 0)
		dasher.SetColor(l.color)
		p := l.points[0]
		dasher.Start(rasterx.ToFixedP(float64(p.X)*scale, float64(p.Y)*scale))
		for _, p := range l.points[1:] {
			dasher.Line(rasterx.ToFixedP(float64(p.X)*scale, float64(p.Y)*scale))
		}
		dasher.Stop(false)
		dasher.Draw()
		dasher.Clear()
	}
	return img
}

// build creates the graph's grid, bands and labels for the given size, and the curves for the raster
func (r *responseGraphRenderer) build(size fyne.Size) {
	g := r.g
	r.objects = r.objects[:0]
	if size.Width <= 0 || g.minF <= 0 || g.maxF <= g.minF {
		r.mu.Lock()
		r.lines = nil
		r.mu.Unlock()
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
	rect := func(x0, y0, x1, y1 float32, c color.Color) {
		rc := canvas.NewRectangle(c)
		rc.Move(fyne.NewPos(x0, y0))
		rc.Resize(fyne.NewSize(x1-x0, float32(math.Max(float64(y1-y0), 1))))
		r.objects = append(r.objects, rc)
	}
	text := func(s string, c color.Color, px, py float32) {
		t := canvas.NewText(s, c)
		t.TextSize = r.labelSize()
		ts := t.MinSize()
		px = float32(math.Max(0, math.Min(float64(px), float64(w-ts.Width))))
		t.Move(fyne.NewPos(px, py))
		t.Resize(ts)
		r.objects = append(r.objects, t)
	}

	// Range of interest: a neutral tint, so the blue curve stands out on it
	if g.bandHi > g.bandLo && g.bandLo > 0 {
		lo, hi := x(math.Max(g.bandLo, g.minF)), x(math.Min(g.bandHi, g.maxF))
		if hi > lo {
			rect(lo, 0, hi, h, withAlpha(fg, 0x16))
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

	// Curves: their line parts go into the raster, their dense parts become bands. The before
	// curve comes first, so the current one is drawn on top.
	var lines []graphPolyline
	var bands []fyne.CanvasObject
	for _, before := range []bool{true, false} {
		for _, c := range g.curves {
			if c.Before != before || len(c.Freqs) < 2 || len(c.Freqs) != len(c.Levels) {
				continue
			}
			// Above denseFrom the comb's cycles are too narrow to draw as a line
			denseFrom := math.Inf(1)
			if c.Spacing > 0 {
				minPx := float64(graphMinCyclePx)
				if before {
					minPx = graphMinBeforeCyclePx
				}
				denseFrom = c.Spacing / (math.Exp(minPx*math.Log(g.maxF/g.minF)/float64(w)) - 1)
			}
			pl := graphPolyline{color: primary, width: graphLineWidth}
			if before {
				pl = graphPolyline{color: withAlpha(muted, 0xa0), width: graphBeforeWidth, dashed: true}
			}
			for i, f := range c.Freqs {
				if i > 0 && c.Freqs[i-1] >= denseFrom {
					break
				}
				pl.points = append(pl.points, fyne.NewPos(x(f), y(c.Levels[i])))
			}
			lines = append(lines, pl)

			if denseFrom < g.maxF {
				// The band: a fill between the comb's lowest and highest level. The current one has
				// crisp edges; the before-curve stays a faint wash behind it.
				x0 := x(math.Max(denseFrom, g.minF))
				top, bottom := y(c.EnvHi), y(c.EnvLo)
				fill := withAlpha(primary, 0x70)
				if before {
					fill = withAlpha(muted, 0x0e)
				}
				band := canvas.NewRectangle(fill)
				band.Move(fyne.NewPos(x0, top))
				band.Resize(fyne.NewSize(w-x0, float32(math.Max(float64(bottom-top), 1))))
				bands = append(bands, band)
				if !before {
					for _, ey := range []float32{top, bottom} {
						l := canvas.NewLine(primary)
						l.StrokeWidth = 2
						l.Position1, l.Position2 = fyne.NewPos(x0, ey), fyne.NewPos(w, ey)
						bands = append(bands, l)
					}
				}
			}
		}
	}
	r.mu.Lock()
	r.lines, r.plotWidth = lines, w
	r.mu.Unlock()
	r.objects = append(r.objects, bands...)
	r.raster.Move(fyne.NewPos(0, 0))
	r.raster.Resize(fyne.NewSize(w, h))
	r.objects = append(r.objects, r.raster)

	// Cursor
	if g.cursor >= g.minF && g.cursor <= g.maxF {
		cx := x(g.cursor)
		line(cx, 0, cx, h, withAlpha(fg, 0xd0), 1.5)
		dot := canvas.NewCircle(fg)
		dot.StrokeColor = th.Color(theme.ColorNameBackground, v)
		dot.StrokeWidth = 1.5
		dot.Move(fyne.NewPos(cx-4.5, y(g.cursorDB)-4.5))
		dot.Resize(fyne.NewSize(9, 9))
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
		text(s, muted, 2, y(db)-r.labelSize()-2)
	}
	// Frequency labels, placed from the right so the last one, with the unit, always shows;
	// a label that would run into its right-hand neighbour is skipped
	nextStart := float32(math.Inf(1))
	for i := len(ticks) - 1; i >= 0; i-- {
		f := ticks[i]
		s := strconv.FormatFloat(f, 'f', -1, 64)
		if f >= 1000 {
			s = strconv.FormatFloat(f/1000, 'f', -1, 64) + "k"
		}
		if i == len(ticks)-1 {
			s += " Hz"
		}
		probe := canvas.NewText(s, muted)
		probe.TextSize = r.labelSize()
		tw := probe.MinSize().Width
		px := float32(math.Max(0, math.Min(float64(x(f)-6), float64(w-tw))))
		if px+tw > nextStart-6 {
			continue
		}
		text(s, muted, px, h+2)
		nextStart = px
	}
}
