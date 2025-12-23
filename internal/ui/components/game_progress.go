package components

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

// GameProgress displays comprehensive game progress information
type GameProgress struct {
	container      *fyne.Container
	progressBar    *ProgressBar
	roundLabel     *widget.Label
	percentLabel   *canvas.Text
	showPercentage bool
}

// GameProgressConfig configures the game progress display
type GameProgressConfig struct {
	ShowRounds     bool
	ShowPercentage bool
	ShowProgressBar bool
}

// NewGameProgress creates a new game progress display
func NewGameProgress(config GameProgressConfig) *GameProgress {
	gp := &GameProgress{
		showPercentage: config.ShowPercentage,
	}

	var elements []fyne.CanvasObject

	// Round indicator (e.g., "Round 3/10")
	if config.ShowRounds {
		gp.roundLabel = widget.NewLabel("")
		gp.roundLabel.Alignment = fyne.TextAlignCenter
		gp.roundLabel.TextStyle.Bold = true
		elements = append(elements, gp.roundLabel)
	}

	// Progress bar
	if config.ShowProgressBar {
		gp.progressBar = NewProgressBar()
		elements = append(elements, gp.progressBar)
	}

	// Percentage indicator (colored based on performance)
	if config.ShowPercentage {
		gp.percentLabel = canvas.NewText("0%", color.RGBA{107, 114, 128, 255}) // gray-500
		gp.percentLabel.TextSize = 20
		gp.percentLabel.TextStyle.Bold = true
		gp.percentLabel.Alignment = fyne.TextAlignTrailing
		
		percentContainer := container.NewBorder(nil, nil, nil, gp.percentLabel, container.NewMax())
		elements = append(elements, percentContainer)
	}

	gp.container = container.NewVBox(elements...)
	return gp
}

// UpdateProgress updates all progress indicators
func (gp *GameProgress) UpdateProgress(current, total, score int) {
	// Update round label
	if gp.roundLabel != nil {
		if total > 0 {
			gp.roundLabel.SetText(fmt.Sprintf(lang.X("game.round_progress", "Round %d/%d"), current, total))
		}
	}

	// Update progress bar (based on rounds completed)
	if gp.progressBar != nil && total > 0 {
		progress := float64(current) / float64(total)
		gp.progressBar.SetValue(progress)
		
		// Color based on progress
		if progress >= 0.8 {
			gp.progressBar.SetColor(color.RGBA{34, 197, 94, 255}) // green-500
		} else if progress >= 0.5 {
			gp.progressBar.SetColor(color.RGBA{59, 130, 246, 255}) // blue-500
		} else {
			gp.progressBar.SetColor(color.RGBA{249, 115, 22, 255}) // orange-500
		}
	}

	// Update percentage with color coding (based on score performance)
	if gp.percentLabel != nil && total > 0 {
		percentage := float64(score) / float64(total) * 100
		gp.percentLabel.Text = fmt.Sprintf("%.0f%%", percentage)
		
		// Color based on performance
		if percentage >= 80 {
			gp.percentLabel.Color = color.RGBA{34, 197, 94, 255} // green-500 - Excellent
		} else if percentage >= 60 {
			gp.percentLabel.Color = color.RGBA{59, 130, 246, 255} // blue-500 - Good
		} else if percentage >= 40 {
			gp.percentLabel.Color = color.RGBA{249, 115, 22, 255} // orange-500 - Fair
		} else {
			gp.percentLabel.Color = color.RGBA{239, 68, 68, 255} // red-500 - Needs improvement
		}
		gp.percentLabel.Refresh()
	}
}

// GetContainer returns the container with all progress elements
func (gp *GameProgress) GetContainer() *fyne.Container {
	return gp.container
}

// Reset resets all progress indicators
func (gp *GameProgress) Reset() {
	if gp.roundLabel != nil {
		gp.roundLabel.SetText("")
	}
	if gp.progressBar != nil {
		gp.progressBar.SetValue(0)
	}
	if gp.percentLabel != nil {
		gp.percentLabel.Text = "0%"
		gp.percentLabel.Color = color.RGBA{107, 114, 128, 255} // gray-500
		gp.percentLabel.Refresh()
	}
}

