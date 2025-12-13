package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

type TopBar struct {
	container *fyne.Container
}

func NewTopBar(gameTitle string, backFunc func(), resetFunc func()) *TopBar {
	title := widget.NewLabel(gameTitle)
	title.TextStyle.Bold = true

	backBtn := widget.NewButton(lang.X("button.dashboard", "Dashboard"), backFunc)
	resetBtn := widget.NewButton(lang.X("button.new_game", "New Game"), resetFunc)

	topBar := container.NewBorder(
		nil, nil,
		backBtn,
		container.NewHBox(resetBtn),
		container.NewCenter(title),
	)

	return &TopBar{
		container: topBar,
	}
}

func (tb *TopBar) GetContainer() *fyne.Container {
	return tb.container
}
