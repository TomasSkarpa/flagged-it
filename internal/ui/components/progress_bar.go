package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// ProgressBar is a custom progress bar widget
type ProgressBar struct {
	widget.BaseWidget
	Value     float64 // 0.0 to 1.0
	bg        *canvas.Rectangle
	fill      *canvas.Rectangle
	FillColor color.Color
}

// NewProgressBar creates a new progress bar
func NewProgressBar() *ProgressBar {
	pb := &ProgressBar{
		Value:     0,
		FillColor: color.RGBA{34, 197, 94, 255}, // green-500
	}
	pb.ExtendBaseWidget(pb)
	return pb
}

// SetValue updates the progress bar value (0.0 to 1.0)
func (pb *ProgressBar) SetValue(value float64) {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	pb.Value = value
	pb.Refresh()
}

// SetColor updates the fill color
func (pb *ProgressBar) SetColor(c color.Color) {
	pb.FillColor = c
	pb.Refresh()
}

// CreateRenderer creates the renderer for the progress bar
func (pb *ProgressBar) CreateRenderer() fyne.WidgetRenderer {
	pb.bg = canvas.NewRectangle(color.RGBA{229, 231, 235, 255}) // gray-200
	pb.bg.CornerRadius = 4

	pb.fill = canvas.NewRectangle(pb.FillColor)
	pb.fill.CornerRadius = 4

	return &progressBarRenderer{
		progressBar: pb,
		objects:     []fyne.CanvasObject{pb.bg, pb.fill},
	}
}

type progressBarRenderer struct {
	progressBar *ProgressBar
	objects     []fyne.CanvasObject
}

func (r *progressBarRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 8)
}

func (r *progressBarRenderer) Layout(size fyne.Size) {
	r.progressBar.bg.Resize(size)
	r.progressBar.bg.Move(fyne.NewPos(0, 0))

	fillWidth := float32(r.progressBar.Value) * size.Width
	r.progressBar.fill.Resize(fyne.NewSize(fillWidth, size.Height))
	r.progressBar.fill.Move(fyne.NewPos(0, 0))
}

func (r *progressBarRenderer) Refresh() {
	// Update fill color
	r.progressBar.fill.FillColor = r.progressBar.FillColor

	// Update fill width based on value
	size := r.progressBar.Size()
	fillWidth := float32(r.progressBar.Value) * size.Width
	r.progressBar.fill.Resize(fyne.NewSize(fillWidth, size.Height))

	r.progressBar.bg.Refresh()
	r.progressBar.fill.Refresh()
	canvas.Refresh(r.progressBar)
}

func (r *progressBarRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *progressBarRenderer) Destroy() {}
