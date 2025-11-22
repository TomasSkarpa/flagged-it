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
	flagBtn := widget.NewButtonWithIcon("Guess by Flag", theme.MailForwardIcon(), func() {
		d.navigateFunc("flag")
	})
	countryListBtn := widget.NewButtonWithIcon("List All Countries", theme.ListIcon(), func() {
		d.navigateFunc("list")
	})
	countryGuessBtn := widget.NewButtonWithIcon("Guess by Shape", theme.MediaRecordIcon(), func() {
		d.navigateFunc("shape")
	})
	hangmanBtn := widget.NewButtonWithIcon("Hangman", theme.AccountIcon(), func() {
		d.navigateFunc("hangman")
	})
	factGuessBtn := widget.NewButtonWithIcon("Guess by Facts", theme.InfoIcon(), func() {
		d.navigateFunc("facts")
	})
	higher_lowerBtn := widget.NewButtonWithIcon("Higher or Lower", theme.UploadIcon(), func() {
		d.navigateFunc("higher_lower")
	})

	d.content = container.NewVBox(
		header,
		widget.NewSeparator(),
		flagBtn,
		countryListBtn,
		countryGuessBtn,
		hangmanBtn,
		factGuessBtn,
		higher_lowerBtn,
	)
}

func (d *Dashboard) GetContent() *fyne.Container {
	return d.content
}
