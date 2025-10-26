package screens

import (
	"flagged-it/internal/utils"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Dashboard struct {
	content      *fyne.Container
	navigateFunc func(string)
	debugFunc    func()
	window       fyne.Window
	debugManager *utils.DebugManager
}

func NewDashboard(navigateFunc func(string), debugFunc func(), window fyne.Window) *Dashboard {
	d := &Dashboard{
		navigateFunc: navigateFunc,
		debugFunc:    debugFunc,
		window:       window,
		debugManager: utils.NewDebugManager(),
	}
	d.setupUI()
	return d
}

func (d *Dashboard) setupUI() {
	title := widget.NewLabel("Choose Your Game Mode")
	title.TextStyle = fyne.TextStyle{Bold: true}

	// Header with title and optional settings button
	var header *fyne.Container
	if d.debugManager.IsDebugEnabled() {
		settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), d.debugFunc)
		header = container.NewBorder(
			nil, nil,
			nil, settingsBtn,
			container.NewCenter(title),
		)
	} else {
		header = container.NewCenter(title)
	}

	// Game buttons
	countryGuessBtn := widget.NewButton("Guess by Shape", func() {
		d.navigateFunc("shape")
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

	flagBtn := widget.NewButton("Guess by Flag", func() {
		d.navigateFunc("flag")
	})

	d.content = container.NewVBox(
		header,
		widget.NewSeparator(),
		countryGuessBtn,
		countryListBtn,
		hangmanBtn,
		factGuessBtn,
		higher_lowerBtn,
		flagBtn,
	)
}

func (d *Dashboard) GetContent() *fyne.Container {
	return d.content
}
