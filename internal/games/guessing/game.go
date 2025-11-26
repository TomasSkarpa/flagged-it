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
	"fyne.io/fyne/v2/theme"
)

type Game struct {
	content        *fyne.Container
	popContent     *fyne.Container
	areaContent    *fyne.Container
	backFunc       func()
	countries      []models.Country
	currentCountry *models.Country
	guessEntry     *widget.Entry
	statusLabel    *widget.Label
	popLabel	   *widget.Label
	areaLabel	   *widget.Label
	continentLabel *widget.Label
	popIcon        *widget.Icon
	areaIcon	   *widget.Icon
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
	g.areaLabel = widget.NewLabel("Area: ")
	g.statusLabel = widget.NewLabel("Make a guess!")

	g.popIcon = widget.NewIcon(theme.MoveUpIcon())
	g.areaIcon = widget.NewIcon(theme.MoveUpIcon())
	g.popIcon.Hide()
	g.areaIcon.Hide()

	g.guessEntry = widget.NewEntry()
	g.guessEntry.SetPlaceHolder("Enter country name...")
	g.guessEntry.OnSubmitted = func(text string) { g.makeGuess() }

	g.guessBtn = widget.NewButton("Guess", g.makeGuess)

	guessContainer := container.NewGridWithColumns(2, g.guessEntry, g.guessBtn)
	g.popContent = container.NewHBox(g.popLabel, g.popIcon)
	g.areaContent = container.NewHBox(g.areaLabel, g.areaIcon)

	g.content = container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
		g.statusLabel,
		widget.NewSeparator(),
		widget.NewSeparator(),
		guessContainer,
		g.continentLabel,
		g.popContent,
		g.areaContent,
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
		if country.Population > g.currentCountry.Population{
			g.popIcon.SetResource(theme.MoveDownIcon())
		}else{
			g.popIcon.SetResource(theme.MoveUpIcon())
		}
		if country.Area > g.currentCountry.Area{
			g.areaIcon.SetResource(theme.MoveDownIcon())
		}else{
			g.areaIcon.SetResource(theme.MoveUpIcon())
		}
		g.popLabel.SetText(fmt.Sprintf("Population: %d", country.Population))
		g.continentLabel.SetText(fmt.Sprintf("Continent: %s", country.Region))
		g.areaLabel.SetText(fmt.Sprintf("Area: %.2f km²", country.Area)) 
		g.popContent = container.NewHBox(g.popLabel, g.popIcon)
		g.areaContent = container.NewHBox(g.areaLabel, g.areaIcon)
		g.popIcon.Show()
		g.areaIcon.Show()
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