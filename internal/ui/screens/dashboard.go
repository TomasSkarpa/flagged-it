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
	countryGuessBtn := widget.NewButtonWithIcon("Guess by Shape", theme.VisibilityIcon(), func() {
		d.navigateFunc("shape")
	})

	countryListBtn := widget.NewButtonWithIcon("List All Countries", theme.ListIcon(), func() {
		d.navigateFunc("list")
	})

	hangmanBtn := widget.NewButtonWithIcon("Hangman", theme.HelpIcon(), func() {
		d.navigateFunc("hangman")
	})

	factGuessBtn := widget.NewButtonWithIcon("Guess by Facts", theme.InfoIcon(), func() {
		d.navigateFunc("facts")
	})

	higher_lowerBtn := widget.NewButtonWithIcon("Higher or Lower", theme.ContentRedoIcon(), func() {
		d.navigateFunc("higher_lower")
	})

	flagBtn := widget.NewButtonWithIcon("Guess by Flag", theme.ColorPaletteIcon(), func() {
		d.navigateFunc("flag")
	})

	guessingBtn := widget.NewButtonWithIcon("What Country is This", theme.GridIcon(), func(){
		d.navigateFunc("guessing")
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
		guessingBtn,
	)
}

func (d *Dashboard) GetContent() *fyne.Container {
	return d.content
}
