package guessing

import (
	"strings"
	"fmt"
	"math/rand"

	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Game struct {
	content        *fyne.Container
	backFunc       func()
	countries      []models.Country
	currentCountry *models.Country
	guessEntry     *widget.Entry
	statusLabel    *widget.Label
	popLabel	   *widget.Label
	continentLabel *widget.Label
	guessBtn       *widget.Button
}

func NewGame(backFunc func()) *Game {
	g := &Game{
		backFunc: backFunc,
	}
	g.loadCountries()
	g.setupUI()
	g.newGame()
	return g
}

func (g *Game) loadCountries() {
	g.countries = data.LoadCountries()
}

func (g *Game) setupUI() {
	topBar := components.NewTopBar("What country is this?", g.backFunc, g.newGame)

	g.continentLabel = widget.NewLabel("Continent: ")
	g.popLabel = widget.NewLabel("Population: ")
	g.statusLabel = widget.NewLabel("Make a guess!")

	g.guessEntry = widget.NewEntry()
	g.guessEntry.SetPlaceHolder("Enter country name...")
	g.guessEntry.OnSubmitted = func(text string) { g.makeGuess() }

	g.guessBtn = widget.NewButton("Guess", g.makeGuess)

	guessContainer := container.NewGridWithColumns(2, g.guessEntry, g.guessBtn)

	g.content = container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
		g.statusLabel,
		widget.NewSeparator(),
		widget.NewSeparator(),
		guessContainer,
		g.continentLabel,
		g.popLabel,
	)
}

func (g *Game) newGame() {
	if len(g.countries) == 0{
		g.statusLabel.SetText("Error loading countries data")
		return
	}

	g.currentCountry = &g.countries[rand.Intn(len(g.countries))]
}

func (g *Game) updateLabels(country *models.Country) {
	if country == nil{
		g.statusLabel.SetText("Enter a valid country name!")
	}else{
		g.popLabel.SetText(fmt.Sprintf("Population: %d", country.Population))
		g.continentLabel.SetText(fmt.Sprintf("Continent: %s", country.Region))
	}
}

func (g *Game) getCountry(countries []models.Country, name string) *models.Country{
	for _, country := range countries{
		if strings.EqualFold(country.Name.Common, name){
			return &country
		}
	}
	return nil
}

func (g *Game) makeGuess() {
	guess := strings.TrimSpace(g.guessEntry.Text)
	if guess == "" {
		return
	}

	if strings.EqualFold(guess, g.currentCountry.Name.Common) {
		g.statusLabel.SetText(fmt.Sprintf("Correct! It was %s!", g.currentCountry.Name.Common))
		g.guessEntry.Disable()
		g.guessBtn.Disable()
		return
	}

	g.guessEntry.SetText("")
	g.statusLabel.SetText("")

	g.updateLabels(g.getCountry(g.countries, guess))
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) Start() {
	g.newGame()
}

func (g *Game) Reset() {
	g.newGame()
}