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
	title.Alignment = fyne.TextAlignCenter

	backBtn := NewButton(lang.X("button.dashboard", "Dashboard"), backFunc)
	
	// Use Stack to truly center the title
	centeredTitle := container.NewCenter(title)
	
	var topBar *fyne.Container
	if resetFunc != nil {
		resetBtn := NewButton(lang.X("button.new_game", "New Game"), resetFunc)
		topBar = container.NewStack(
			centeredTitle,
			container.NewBorder(nil, nil, backBtn, container.NewHBox(resetBtn)),
		)
	} else {
		topBar = container.NewStack(
			centeredTitle,
			container.NewBorder(nil, nil, backBtn, nil),
		)
	}

	return &TopBar{
		container: topBar,
	}
}

func (tb *TopBar) GetContainer() *fyne.Container {
	return tb.container
}
