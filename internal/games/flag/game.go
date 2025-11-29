package flag

import (
	"fmt"
	"image/color"
	"math/rand"
	"runtime"
	"time"

	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type coloredButton struct {
	widget.BaseWidget
	button *widget.Button
	rect   *canvas.Rectangle
}

func (c *coloredButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(
		c.rect,
		c.button,
	))
}

func (c *coloredButton) SetBgColor(col color.Color) {
	c.rect.FillColor = col
	c.Refresh()
}

type Game struct {
	content        *fyne.Container
	backFunc       func()
	countries      []models.Country
	currentCountry *models.Country
	previousCountry *models.Country
	options        []models.Country
	flagImage      *canvas.Image
	statusLabel    *widget.Label
	buttons        []*widget.Button
	coloredButtons []*coloredButton
	buttonGrid     *fyne.Container
	score          int
	total          int
	scoreLabel     *widget.Label
	scoreManager   *utils.ScoreManager
}

func NewGame(backFunc func(), scoreManager *utils.ScoreManager) *Game {
	g := &Game{
		backFunc:     backFunc,
		scoreManager: scoreManager,
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
	topBar := components.NewTopBar("Guess by Flag", g.backFunc, g.Reset)

	g.flagImage = canvas.NewImageFromResource(nil)
	g.flagImage.FillMode = canvas.ImageFillContain
	g.flagImage.SetMinSize(fyne.NewSize(400, 250))
	g.statusLabel = widget.NewLabel("Which country does this flag belong to?")
	g.statusLabel.Wrapping = fyne.TextWrapWord
	g.scoreLabel = widget.NewLabel("Score: 0/10")

	// Create responsive button grid
	g.buttonGrid = container.NewGridWithColumns(2)

	g.content = container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
		g.scoreLabel,
		g.statusLabel,
		container.NewCenter(g.flagImage),
		widget.NewSeparator(),
		g.buttonGrid,
	)
}

func (g *Game) newGame() {
	if len(g.countries) == 0 {
		g.statusLabel.SetText("Error loading countries data")
		return
	}

	rand.Seed(time.Now().UnixNano())
	var newCountry *models.Country
	for {
		newCountry = &g.countries[rand.Intn(len(g.countries))]
		if g.previousCountry == nil || newCountry.CCA2 != g.previousCountry.CCA2 {
			break
		}
	}
	g.previousCountry = g.currentCountry
	g.currentCountry = newCountry

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
	g.buttonGrid.RemoveAll()

	g.buttons = make([]*widget.Button, 4)
	g.coloredButtons = make([]*coloredButton, 4)
	for i, country := range g.options {
		country := country
		btn := widget.NewButtonWithIcon(country.Name.Common, nil, func() {
			g.makeGuess(country)
		})
		btn.Importance = widget.LowImportance
		g.buttons[i] = btn

		rect := canvas.NewRectangle(color.Transparent)

		colored := &coloredButton{button: btn, rect: rect}
		colored.ExtendBaseWidget(colored)
		g.coloredButtons[i] = colored
		g.buttonGrid.Add(colored)
	}
	g.buttonGrid.Refresh()
}

func (g *Game) makeGuess(guessed models.Country) {
	g.total++
	isCorrect := guessed.CCA2 == g.currentCountry.CCA2
	if isCorrect {
		g.score++
		g.statusLabel.SetText(fmt.Sprintf("Correct! It's %s!", g.currentCountry.Name.Common))
	} else {
		g.statusLabel.SetText(fmt.Sprintf("Wrong! It was %s", g.currentCountry.Name.Common))
	}

	g.scoreLabel.SetText(fmt.Sprintf("Score: %d/10", g.score))

	for i, btn := range g.buttons {
		var bgColor color.Color
		if btn.Text == g.currentCountry.Name.Common {
			bgColor = color.RGBA{30, 180, 80, 255}
		} else {
			bgColor = color.RGBA{255, 99, 71, 255}
		}
		g.coloredButtons[i].SetBgColor(bgColor)
		btn.Disable()
	}

	if g.total >= 10 {
		g.scoreManager.SetTotal("flag", 10)
		g.scoreManager.UpdateScore("flag", g.score)
		time.AfterFunc(1500*time.Millisecond, func() {
			fyne.Do(func() {
				g.statusLabel.SetText(fmt.Sprintf("Game Complete! Final Score: %d/10 (%.0f%%)", g.score, float64(g.score)/10*100))
			})
		})
	} else {
		time.AfterFunc(1500*time.Millisecond, func() {
			fyne.Do(func() {
				g.newGame()
			})
		})
	}
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) Start() {
	g.newGame()
}

func (g *Game) Reset() {
	g.score = 0
	g.total = 0
	g.scoreLabel.SetText("Score: 0/10")
	g.newGame()
}
