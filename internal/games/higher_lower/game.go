package higher_lower

import (
	"flagged-it/internal/data"
	"flagged-it/internal/ui/components"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"math/rand"
)

type Game struct {
	content  *fyne.Container
	backFunc func()

	firstCountry  int
	secondCountry int
	score         int

	countryOneNameLabel *widget.Label
	countryTwoNameLabel *widget.Label
	countryOnePopLabel  *widget.Label
	countryTwoPopLabel  *widget.Label
	scoreLabel          *widget.Label
	nextBtn             *widget.Button
	higherBtn           *widget.Button
	lowerBtn            *widget.Button
}

func NewGame(backFunc func()) *Game {
	g := &Game{backFunc: backFunc}
	g.setupUI()
	return g
}

func (g *Game) setupUI() {
	topBar := components.NewTopBar("Higher or Lower Game", g.backFunc, g.Reset)

	gameDescription := widget.NewLabel("Try to guess which country has a higher population!")

	g.countryOneNameLabel = widget.NewLabel("")
	g.countryOnePopLabel = widget.NewLabel("")
	g.countryTwoNameLabel = widget.NewLabel("")
	g.countryTwoPopLabel = widget.NewLabel("")

	g.scoreLabel = widget.NewLabel("Score: 0")

	var startBtn *widget.Button

	g.higherBtn = widget.NewButton("Higher", func() {
		g.makeGuess(true)
	})
	g.lowerBtn = widget.NewButton("Lower", func() {
		g.makeGuess(false)
	})
	g.nextBtn = widget.NewButton("Next Round", func() {
		g.nextRound()
	})

	g.higherBtn.Hide()
	g.lowerBtn.Hide()
	g.nextBtn.Hide()

	startBtn = widget.NewButton("Start Game", func() {
		g.Start()
		startBtn.Hide()
		g.higherBtn.Show()
		g.lowerBtn.Show()
	})

	g.content = container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
		gameDescription,
		widget.NewSeparator(),
		startBtn,
		g.countryOneNameLabel,
		g.countryOnePopLabel,
		widget.NewSeparator(),
		g.countryTwoNameLabel,
		g.countryTwoPopLabel,
		g.higherBtn,
		g.lowerBtn,
		g.nextBtn,
		g.scoreLabel,
	)
}

func (g *Game) makeGuess(isHigher bool) {
	correct := (isHigher && g.secondCountry > g.firstCountry) || (!isHigher && g.secondCountry < g.firstCountry)
	if correct {
		g.score++
	}else{
		g.score = 0
	}
	g.scoreLabel.SetText(fmt.Sprintf("Score: %d", g.score))
	g.countryTwoPopLabel.SetText(fmt.Sprintf("Population: %d", g.secondCountry))
	g.higherBtn.Hide()
	g.lowerBtn.Hide()
	g.nextBtn.Show()
}

func (g *Game) nextRound() {
	countries := data.LoadCountries()
	newCountry := countries[rand.Intn(len(countries))]
	
	for newCountry.Name.Common == g.countryTwoNameLabel.Text{
		newCountry = countries[rand.Intn(len(countries))]
	}

	g.firstCountry = g.secondCountry
	g.countryOneNameLabel.SetText(g.countryTwoNameLabel.Text)
	g.countryOnePopLabel.SetText(fmt.Sprintf("Population: %d", g.firstCountry))

	g.secondCountry = newCountry.Population
	g.countryTwoNameLabel.SetText(newCountry.Name.Common)
	g.countryTwoPopLabel.SetText("Population: ?")

	g.nextBtn.Hide()
	g.higherBtn.Show()
	g.lowerBtn.Show()
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) Start() {
	countries := data.LoadCountries()

	firstRandomCountry := countries[rand.Intn(len(countries))]
	secondRandomCountry := countries[rand.Intn(len(countries))]

	for firstRandomCountry.Name.Common == secondRandomCountry.Name.Common {
		secondRandomCountry = countries[rand.Intn(len(countries))]
	}

	g.firstCountry = firstRandomCountry.Population
	g.secondCountry = secondRandomCountry.Population

	g.countryOneNameLabel.SetText(firstRandomCountry.Name.Common)
	g.countryOnePopLabel.SetText(fmt.Sprintf("Population: %d", firstRandomCountry.Population))
	g.countryTwoNameLabel.SetText(secondRandomCountry.Name.Common)
	g.countryTwoPopLabel.SetText("Population: ?")
}

func (g *Game) Reset() {
	g.score = 0
	g.scoreLabel.SetText("Score: 0")
	g.Start()
}
