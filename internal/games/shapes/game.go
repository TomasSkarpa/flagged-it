package shapes

import (
	"flagged-it/internal/ui/components"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Game struct {
	content  *fyne.Container
	backFunc func()
}

func NewGame(backFunc func()) *Game {
	g := &Game{backFunc: backFunc}
	g.setupUI()
	return g
}

func (g *Game) setupUI() {
	topBar := components.NewTopBar("Country Guessing Game", g.backFunc, g.Reset)

	gameContent := widget.NewLabel("Country guessing game implementation")

	g.content = container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
		gameContent,
	)
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) Start() {}

func (g *Game) Reset() {}
