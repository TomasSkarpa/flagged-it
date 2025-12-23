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

	// 1 column on mobile, 2 on desktop
	columns := 2
	if utils.IsMobile() {
		columns = 1
	}
	buttonGrid := container.NewGridWithColumns(columns)
	for _, region := range regions {
		region := region
		translatedRegion := utils.TranslateRegion(region)
		buttonGrid.Add(NewButton(translatedRegion, func() {
			onRegionSelected(region)
		}))
	}

	return &RegionSelector{
		container: container.NewVBox(
			titleLabel,
			descLabel,
			buttonGrid,
		),
	}
}

func (r *RegionSelector) GetContainer() *fyne.Container {
	return r.container
}
