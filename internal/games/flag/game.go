package flag

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Game struct {
	content        *fyne.Container
	backFunc       func()
	countries      []models.Country
	currentCountry *models.Country
	options        []models.Country
	flagImage      *canvas.Image
	statusLabel    *widget.Label
	buttons        []*widget.Button
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
	topBar := components.NewTopBar("Guess by Flag", g.backFunc, g.newGame)

	g.flagImage = canvas.NewImageFromResource(nil)
	g.flagImage.FillMode = canvas.ImageFillContain
	g.flagImage.SetMinSize(fyne.NewSize(338, 225))
	g.statusLabel = widget.NewLabel("Which country does this flag belong to?")

	g.content = container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
		g.statusLabel,
		g.flagImage,
		widget.NewSeparator(),
	)
}

func (g *Game) newGame() {
	if len(g.countries) == 0 {
		g.statusLabel.SetText("Error loading countries data")
		return
	}

	rand.Seed(time.Now().UnixNano())
	g.currentCountry = &g.countries[rand.Intn(len(g.countries))]

	g.options = []models.Country{*g.currentCountry}
	for len(g.options) < 4 {
		option := g.countries[rand.Intn(len(g.countries))]
		if option.CCA2 != g.currentCountry.CCA2 {
			g.options = append(g.options, option)
		}
	}

	rand.Shuffle(len(g.options), func(i, j int) {
		g.options[i], g.options[j] = g.options[j], g.options[i]
	})

	g.displayFlag()
	g.createButtons()
	g.statusLabel.SetText("Which country does this flag belong to?")
}

func (g *Game) displayFlag() {
	var flagResource fyne.Resource
	var err error
	if runtime.GOOS == "js" {
		flagURL := fmt.Sprintf("assets/twemoji_flags_cca2/%s.svg", g.currentCountry.CCA2)
		flagResource, err = fyne.LoadResourceFromURLString(flagURL)
	} else {
		flagPath := fmt.Sprintf("assets/twemoji_flags_cca2/%s.svg", g.currentCountry.CCA2)
		flagResource, err = fyne.LoadResourceFromPath(flagPath)
	}
	if err == nil {
		g.flagImage.Resource = flagResource
	}
	g.flagImage.Refresh()
}

func (g *Game) createButtons() {
	if len(g.buttons) > 0 {
		for i := len(g.content.Objects) - 1; i >= 0; i-- {
			if _, ok := g.content.Objects[i].(*widget.Button); ok {
				g.content.Objects = append(g.content.Objects[:i], g.content.Objects[i+1:]...)
			}
		}
	}

	g.buttons = make([]*widget.Button, 4)
	for i, country := range g.options {
		country := country
		btn := widget.NewButton(country.Name.Common, func() {
			g.makeGuess(country)
		})
		g.buttons[i] = btn
		g.content.Add(btn)
	}
	g.content.Refresh()
}

func (g *Game) makeGuess(guessed models.Country) {
	if guessed.CCA2 == g.currentCountry.CCA2 {
		g.statusLabel.SetText(fmt.Sprintf("Correct! It's %s!", g.currentCountry.Name.Common))
	} else {
		g.statusLabel.SetText(fmt.Sprintf("Wrong! It was %s", g.currentCountry.Name.Common))
	}

	for _, btn := range g.buttons {
		btn.Disable()
	}
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
