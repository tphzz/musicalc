package ui

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ResponsiveTable wraps a table and automatically resizes columns proportionally
type ResponsiveTable struct {
	widget.BaseWidget
	table             *widget.Table
	columnProportions []float32 // Proportions for each column (should sum to 1.0)
	minWidth          float32   // Minimum total width before applying proportions
	padding           float32   // Extra space to leave free, beyond the gaps between columns

	// OnFirstLayout is called once, when the table first gets a size (e.g. to scroll to a row)
	OnFirstLayout func()
	laidOut       bool
}

// NewResponsiveTable creates a responsive table with proportional column widths
// columnProportions: array of proportions for each column (e.g., [0.25, 0.5, 0.25] for 25%/50%/25%)
// minWidth: minimum total width before scaling kicks in (default 300)
// padding: extra space to leave free beyond the column gaps, which are subtracted automatically
func NewResponsiveTable(table *widget.Table, columnProportions []float32, minWidth, padding float32) *ResponsiveTable {
	if minWidth <= 0 {
		minWidth = 300
	}
	if padding < 0 {
		padding = 20
	}

	r := &ResponsiveTable{
		table:             table,
		columnProportions: columnProportions,
		minWidth:          minWidth,
		padding:           padding,
	}
	r.ExtendBaseWidget(r)
	return r
}

func (r *ResponsiveTable) CreateRenderer() fyne.WidgetRenderer {
	return &responsiveTableRenderer{
		responsive: r,
		table:      r.table,
	}
}

type responsiveTableRenderer struct {
	responsive *ResponsiveTable
	table      *widget.Table
}

func (rr *responsiveTableRenderer) Layout(size fyne.Size) {
	rr.table.Resize(size)
	rr.responsive.updateColumnWidths(size.Width)

	r := rr.responsive
	if !r.laidOut && size.Width > 0 && size.Height > 0 {
		r.laidOut = true
		if r.OnFirstLayout != nil {
			r.OnFirstLayout()
		}
	}
}

func (rr *responsiveTableRenderer) MinSize() fyne.Size {
	return rr.table.MinSize()
}

func (rr *responsiveTableRenderer) Refresh() {
	rr.table.Refresh()
}

func (rr *responsiveTableRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{rr.table}
}

func (rr *responsiveTableRenderer) Destroy() {}

// SetProportions changes the column proportions and applies them at the current width
func (r *ResponsiveTable) SetProportions(columnProportions []float32) {
	r.columnProportions = columnProportions
	r.updateColumnWidths(r.Size().Width)
}

// updateColumnWidths recalculates and applies proportional column widths
func (r *ResponsiveTable) updateColumnWidths(width float32) {
	if width < r.minWidth {
		return
	}

	// The table puts a theme padding between columns; any width beyond ours makes it scroll sideways
	gaps := float32(len(r.columnProportions)-1) * r.table.Theme().Size(theme.SizeNamePadding)
	availableWidth := width - gaps - r.padding
	if availableWidth < 0 {
		availableWidth = 0
	}
	// Don't enforce minWidth - allow columns to shrink to fit available space
	// This prevents horizontal scrolling when container is narrower than minWidth

	for i, proportion := range r.columnProportions {
		// Round down so the columns never add up to more than the available width
		columnWidth := float32(math.Floor(float64(availableWidth * proportion)))
		r.table.SetColumnWidth(i, columnWidth)
	}
}
