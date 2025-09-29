package higher_lower

import (
	"encoding/json"
	"os"
	"fmt"
	"math/rand"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Country struct {
	CountryName	string `json:"country_name"`
	Population 	int    `json:"population"`
}

type Game struct {
	content  *fyne.Container
	backFunc func()

	firstCountry int
	secondCountry int

	countryOneNameLabel	*widget.Label
	countryTwoNameLabel	*widget.Label
	countryOnePopLabel 	*widget.Label
	countryTwoPopLabel 	*widget.Label
}

func NewGame(backFunc func()) *Game {
	g := &Game{backFunc: backFunc}
	g.setupUI()
	return g
}

func (g *Game) setupUI() {
	title := widget.NewLabel("Higher or Lower Game")

	backBtn := widget.NewButton("← Back to Dashboard", func() {
		g.backFunc()
	})

	gameDescription := widget.NewLabel("Try to guess which country has a higher population!")

	g.countryOneNameLabel = widget.NewLabel("")
	g.countryOnePopLabel = widget.NewLabel("")
	g.countryTwoNameLabel = widget.NewLabel("")
	g.countryTwoPopLabel = widget.NewLabel("")

	scoreLabel := widget.NewLabel("Score: 0")

	var higherBtn, lowerBtn, startBtn *widget.Button

	higherBtn = widget.NewButton("Higher", func() {
		guess(true, g.firstCountry, g.secondCountry, scoreLabel)
	})
	lowerBtn = widget.NewButton("Lower", func() {
		guess(false, g.firstCountry, g.secondCountry, scoreLabel)
	})

	higherBtn.Hide()
	lowerBtn.Hide()

	startBtn = widget.NewButton("Start Game", func() {
		g.Start()
		startBtn.Hide()
		higherBtn.Show()
		lowerBtn.Show()
	})

	g.content = container.NewVBox(
		backBtn,
		title,
		gameDescription,
		widget.NewSeparator(),
		startBtn,
		g.countryOneNameLabel,
		g.countryOnePopLabel,
		widget.NewSeparator(),
		g.countryTwoNameLabel,
		g.countryTwoPopLabel,
		higherBtn,
		lowerBtn,
		scoreLabel,
	)
}

func loadCountries(filename string) ([]Country, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var countries []Country
	err = json.NewDecoder(file).Decode(&countries)
	return countries, err
}

func guess(isHigher bool, countryOnePop int, countryTwoPop int, scoreLabel *widget.Label) {
	var score int = 0
	if(isHigher && countryTwoPop > countryOnePop) || (!isHigher && countryTwoPop < countryOnePop) {
		score++
	}
	scoreLabel.SetText(fmt.Sprintf("Score: %d", score))
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) Start() {
	countries, err := loadCountries("C:/laragon/www/flagged-it/internal/data/sources/countries.json")
	if err != nil{
		panic(err)
	}

	firstRandomCountry := countries[rand.Intn(len(countries))]
	secondRandomCountry := countries[rand.Intn(len(countries))]

	for firstRandomCountry == secondRandomCountry {
		secondRandomCountry = countries[rand.Intn(len(countries))]
	}

	g.firstCountry = firstRandomCountry.Population
	g.secondCountry = secondRandomCountry.Population

	g.countryOneNameLabel.SetText(firstRandomCountry.CountryName)
	g.countryOnePopLabel.SetText(fmt.Sprintf("Population: %d", firstRandomCountry.Population))
	g.countryTwoNameLabel.SetText(secondRandomCountry.CountryName) 
	g.countryTwoPopLabel.SetText("Population: ?")
}

func (g *Game) Reset() {}
