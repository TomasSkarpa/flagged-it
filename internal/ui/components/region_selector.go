package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type RegionSelector struct {
	container *fyne.Container
}

func NewRegionSelector(title, description string, onRegionSelected func(string)) *RegionSelector {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle.Bold = true

	descLabel := widget.NewLabel(description)

	buttons := []struct {
		text   string
		region string
	}{
		{"World (All Countries)", "World"},
		{"Europe", "Europe"},
		{"Americas", "Americas"},
		{"Asia", "Asia"},
		{"Africa", "Africa"},
		{"Oceania", "Oceania"},
	}

	var buttonWidgets []fyne.CanvasObject
	buttonWidgets = append(buttonWidgets, titleLabel, descLabel, widget.NewSeparator())

	for _, btn := range buttons {
		region := btn.region
		buttonWidgets = append(buttonWidgets, widget.NewButton(btn.text, func() {
			onRegionSelected(region)
		}))
	}

	return &RegionSelector{
		container: container.NewVBox(buttonWidgets...),
	}
}

func (r *RegionSelector) GetContainer() *fyne.Container {
	return r.container
}
