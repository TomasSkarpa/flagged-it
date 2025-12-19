package components

import (
	"flagged-it/internal/utils"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type RegionSelector struct {
	container *fyne.Container
}

func NewRegionSelector(title, description string, regions []string, onRegionSelected func(string)) *RegionSelector {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle.Bold = true

	descLabel := widget.NewLabel(description)
	descLabel.Wrapping = fyne.TextWrapWord

	// Sort regions with World first
	sort.Slice(regions, func(i, j int) bool {
		if regions[i] == "World" {
			return true
		}
		if regions[j] == "World" {
			return false
		}
		return regions[i] < regions[j]
	})

	buttonGrid := container.NewGridWithColumns(2)
	for _, region := range regions {
		region := region
		translatedRegion := utils.TranslateRegion(region)
		buttonGrid.Add(widget.NewButton(translatedRegion, func() {
			onRegionSelected(region)
		}))
	}

	return &RegionSelector{
		container: container.NewVBox(
			titleLabel,
			descLabel,
			widget.NewSeparator(),
			buttonGrid,
		),
	}
}

func (r *RegionSelector) GetContainer() *fyne.Container {
	return r.container
}
