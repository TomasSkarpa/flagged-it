package fact_guess

import (
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
	title := widget.NewLabel("Guess by Facts")

	backBtn := widget.NewButton("← Back to Dashboard", func() {
		g.backFunc()
	})

	gameContent := widget.NewLabel("Fact guessing game implementation")

	g.content = container.NewVBox(
		backBtn,
		title,
		widget.NewSeparator(),
		gameContent,
	)
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) Start() {}

func (g *Game) Reset() {}
