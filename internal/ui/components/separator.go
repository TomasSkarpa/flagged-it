package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// Separator is a styled horizontal line
type Separator struct {
	widget.BaseWidget
	line   *canvas.Rectangle
	dashed bool
	height float32
}

// NewSeparator creates a default separator (light gray, 1px)
func NewSeparator() *Separator {
	return NewStyledSeparator(color.RGBA{200, 200, 200, 255}, 1)
}

// NewStyledSeparator creates a separator with custom color and height
func NewStyledSeparator(col color.Color, height float32) *Separator {
	s := &Separator{
		line:   canvas.NewRectangle(col),
		dashed: false,
		height: height,
	}
	s.line.SetMinSize(fyne.NewSize(0, height))
	s.ExtendBaseWidget(s)
	return s
}

// NewDashedSeparator creates a dashed separator
func NewDashedSeparator(col color.Color, height float32) *Separator {
	s := &Separator{
		line:   canvas.NewRectangle(col),
		dashed: true,
		height: height,
	}
	s.line.SetMinSize(fyne.NewSize(0, height))
	s.ExtendBaseWidget(s)
	return s
}

// NewDarkSeparator creates a darker separator (for dark themes)
func NewDarkSeparator() *Separator {
	return NewStyledSeparator(color.RGBA{60, 60, 60, 255}, 1)
}

// NewThickSeparator creates a thicker separator
func NewThickSeparator() *Separator {
	return NewStyledSeparator(color.RGBA{200, 200, 200, 255}, 2)
}

func (s *Separator) CreateRenderer() fyne.WidgetRenderer {
	if s.dashed {
		return &dashedSeparatorRenderer{separator: s}
	}
	return widget.NewSimpleRenderer(s.line)
}

func (s *Separator) MinSize() fyne.Size {
	return s.line.MinSize()
}

// SetColor changes the separator color
func (s *Separator) SetColor(col color.Color) {
	s.line.FillColor = col
	s.line.Refresh()
}

// SetHeight changes the separator height
func (s *Separator) SetHeight(height float32) {
	s.height = height
	s.line.SetMinSize(fyne.NewSize(0, height))
	s.Refresh()
}

// Dashed separator renderer
type dashedSeparatorRenderer struct {
	separator *Separator
	dashes    []*canvas.Rectangle
}

func (r *dashedSeparatorRenderer) Layout(size fyne.Size) {
	// Clear existing dashes
	r.dashes = nil

	// Create dashed pattern: 8px dash, 4px gap
	dashWidth := float32(8)
	gapWidth := float32(4)
	x := float32(0)

	for x < size.Width {
		if x+dashWidth > size.Width {
			dashWidth = size.Width - x
		}
		dash := canvas.NewRectangle(r.separator.line.FillColor)
		dash.Resize(fyne.NewSize(dashWidth, r.separator.height))
		dash.Move(fyne.NewPos(x, 0))
		r.dashes = append(r.dashes, dash)
		x += dashWidth + gapWidth
	}
}

func (r *dashedSeparatorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, r.separator.height)
}

func (r *dashedSeparatorRenderer) Refresh() {
	for _, dash := range r.dashes {
		dash.FillColor = r.separator.line.FillColor
		dash.Refresh()
	}
}

func (r *dashedSeparatorRenderer) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, len(r.dashes))
	for i, dash := range r.dashes {
		objects[i] = dash
	}
	return objects
}

func (r *dashedSeparatorRenderer) Destroy() {}

