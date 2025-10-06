package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Dashboard struct {
	content      *fyne.Container
	navigateFunc func(string)
}

func NewDashboard(navigateFunc func(string)) *Dashboard {
	d := &Dashboard{navigateFunc: navigateFunc}
	d.setupUI()
	return d
}

func (d *Dashboard) setupUI() {
	title := widget.NewLabel("Flagged It - Choose Your Game")
	title.TextStyle = fyne.TextStyle{Bold: true}

	countryGuessBtn := widget.NewButton("Country Guessing", func() {
		d.navigateFunc("shapes")
	})

	countryListBtn := widget.NewButton("List All Countries", func() {
		d.navigateFunc("list")
	})

	hangmanBtn := widget.NewButton("Hangman", func() {
		d.navigateFunc("hangman")
	})

	factGuessBtn := widget.NewButton("Guess by Facts", func() {
		d.navigateFunc("facts")
	})

	higher_lowerBtn := widget.NewButton("Higher or Lower", func() {
		d.navigateFunc("higher_lower")
	})

	d.content = container.NewVBox(
		title,
		widget.NewSeparator(),
		countryGuessBtn,
		countryListBtn,
		hangmanBtn,
		factGuessBtn,
		higher_lowerBtn,
	)
}

func (d *Dashboard) GetContent() *fyne.Container {
	return d.content
}
