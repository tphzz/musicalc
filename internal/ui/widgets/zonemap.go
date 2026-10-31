package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	zoneBarHeight    = 18.0
	zoneMarkerWidth  = 3.0
	zoneMarkerExtend = 3.0 // the marker sticks out above and below the bar
	zoneLabelGap     = 4.0
)

// ZoneBand is a coloured range along a ZoneMap
type ZoneBand struct {
	Start, End float64
	Color      fyne.ThemeColorName
}

// ZoneLabel is a value written below a ZoneMap at a position
type ZoneLabel struct {
	At   float64
	Text string
}

// ZoneMap draws coloured bands along a range from 0 to Max, with labels below it and an
// optional marker, e.g. the phase zones over a range of distances
type ZoneMap struct {
	widget.BaseWidget

	max    float64
	bands  []ZoneBand
	labels []ZoneLabel
	marker float64 // position of the marker, negative for none
}

// NewZoneMap creates an empty zone map; fill it with SetZones
func NewZoneMap() *ZoneMap {
	m := &ZoneMap{marker: -1}
	m.ExtendBaseWidget(m)
	return m
}

// SetZones replaces the range, bands, labels and marker (negative for no marker)
func (m *ZoneMap) SetZones(max float64, bands []ZoneBand, labels []ZoneLabel, marker float64) {
	m.max, m.bands, m.labels, m.marker = max, bands, labels, marker
	m.Refresh()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (m *ZoneMap) CreateRenderer() fyne.WidgetRenderer {
	r := &zoneMapRenderer{zoneMap: m}
	r.rebuild()
	return r
}

type zoneMapRenderer struct {
	zoneMap *ZoneMap
	bands   []*canvas.Rectangle
	labels  []*canvas.Text
	marker  *canvas.Rectangle
	objects []fyne.CanvasObject
}

func (r *zoneMapRenderer) rebuild() {
	m := r.zoneMap
	th := m.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	r.bands = r.bands[:0]
	r.labels = r.labels[:0]
	r.objects = r.objects[:0]
	for _, b := range m.bands {
		c := th.Color(b.Color, v)
		// slightly translucent, so the bands blend with the card like in the design
		if rgba, ok := c.(color.NRGBA); ok {
			rgba.A = 0xC0
			c = rgba
		}
		rect := canvas.NewRectangle(c)
		r.bands = append(r.bands, rect)
		r.objects = append(r.objects, rect)
	}
	r.marker = canvas.NewRectangle(th.Color(theme.ColorNameForeground, v))
	if m.marker < 0 || m.max <= 0 {
		r.marker.Hide()
	}
	r.objects = append(r.objects, r.marker)
	for _, l := range m.labels {
		t := canvas.NewText(l.Text, th.Color(theme.ColorNamePlaceHolder, v))
		t.TextSize = th.Size(theme.SizeNameCaptionText)
		r.labels = append(r.labels, t)
		r.objects = append(r.objects, t)
	}
}

func (r *zoneMapRenderer) Layout(size fyne.Size) {
	m := r.zoneMap
	if m.max <= 0 {
		return
	}
	scale := size.Width / float32(m.max)
	top := float32(zoneMarkerExtend)
	for i, b := range m.bands {
		x0, x1 := float32(b.Start)*scale, float32(b.End)*scale
		r.bands[i].Move(fyne.NewPos(x0, top))
		r.bands[i].Resize(fyne.NewSize(x1-x0, zoneBarHeight))
	}
	r.marker.Move(fyne.NewPos(float32(m.marker)*scale-zoneMarkerWidth/2, 0))
	r.marker.Resize(fyne.NewSize(zoneMarkerWidth, zoneBarHeight+2*zoneMarkerExtend))
	for i, l := range m.labels {
		t := r.labels[i]
		w := t.MinSize().Width
		x := float32(l.At)*scale - w/2
		x = fyne.Max(0, fyne.Min(x, size.Width-w)) // keep the first and last label inside
		t.Move(fyne.NewPos(x, top+zoneBarHeight+zoneLabelGap))
		t.Resize(t.MinSize())
	}
}

func (r *zoneMapRenderer) MinSize() fyne.Size {
	labelHeight := float32(0)
	if len(r.labels) > 0 {
		labelHeight = r.labels[0].MinSize().Height + zoneLabelGap
	}
	return fyne.NewSize(100, zoneBarHeight+2*zoneMarkerExtend+labelHeight)
}

func (r *zoneMapRenderer) Refresh() {
	r.rebuild()
	r.Layout(r.zoneMap.Size())
	canvas.Refresh(r.zoneMap)
}

func (r *zoneMapRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *zoneMapRenderer) Destroy() {}
