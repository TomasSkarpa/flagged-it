package facts

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Game struct {
	content        *fyne.Container
	backFunc       func()
	countries      []models.Country
	currentCountry *models.Country
	currentFact    int
	triesLeft      int
	usedFacts      map[int]bool
	factLabel      *widget.Label
	guessEntry     *widget.Entry
	statusLabel    *widget.Label
	triesLabel     *widget.Label
	guessBtn       *widget.Button
	newGameBtn     *widget.Button
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
	title := widget.NewLabel("Guess by Facts")
	title.TextStyle.Bold = true

	backBtn := widget.NewButton("Back to Dashboard", g.backFunc)

	g.factLabel = widget.NewLabel("")
	g.factLabel.Wrapping = fyne.TextWrapWord

	g.guessEntry = widget.NewEntry()
	g.guessEntry.SetPlaceHolder("Enter country name...")
	g.guessEntry.OnSubmitted = func(text string) { g.makeGuess() }

	g.guessBtn = widget.NewButton("Guess", g.makeGuess)
	g.statusLabel = widget.NewLabel("")
	g.triesLabel = widget.NewLabel("")
	g.newGameBtn = widget.NewButton("New Game", g.newGame)

	guessContainer := container.NewGridWithColumns(2, g.guessEntry, g.guessBtn)
	buttonContainer := container.NewHBox(g.newGameBtn, backBtn)

	g.content = container.NewVBox(
		title,
		widget.NewSeparator(),
		g.statusLabel,
		g.triesLabel,
		widget.NewSeparator(),
		g.factLabel,
		widget.NewSeparator(),
		guessContainer,
		buttonContainer,
	)
}

func (g *Game) newGame() {
	if len(g.countries) == 0 {
		g.statusLabel.SetText("Error loading countries data")
		return
	}

	rand.Seed(time.Now().UnixNano())
	g.currentCountry = &g.countries[rand.Intn(len(g.countries))]
	g.currentFact = 0
	g.triesLeft = 3
	g.usedFacts = make(map[int]bool)
	g.guessEntry.SetText("")
	g.guessEntry.Enable()
	g.guessBtn.Enable()

	g.showCurrentFact()
	g.updateStatus()
}

func (g *Game) showCurrentFact() {
	if g.currentCountry != nil && len(g.usedFacts) < len(g.currentCountry.Facts) {
		var factIndex int
		for {
			factIndex = rand.Intn(len(g.currentCountry.Facts))
			if !g.usedFacts[factIndex] {
				break
			}
		}
		g.usedFacts[factIndex] = true
		fact := g.currentCountry.Facts[factIndex]
		g.factLabel.SetText(fmt.Sprintf("Fact %d: %s", g.currentFact+1, fact))
	}
}

func (g *Game) updateStatus() {
	g.statusLabel.SetText("Guess the country based on the fact!")
	g.triesLabel.SetText(fmt.Sprintf("Tries left: %d", g.triesLeft))
}

func (g *Game) makeGuess() {
	guess := strings.TrimSpace(g.guessEntry.Text)
	if guess == "" {
		return
	}

	if strings.EqualFold(guess, g.currentCountry.CountryName) {
		g.statusLabel.SetText(fmt.Sprintf("Correct! It was %s!", g.currentCountry.CountryName))
		g.guessEntry.Disable()
		g.guessBtn.Disable()
		return
	}

	g.triesLeft--
	g.guessEntry.SetText("")

	if g.triesLeft == 0 {
		g.statusLabel.SetText(fmt.Sprintf("Game Over! It was %s", g.currentCountry.CountryName))
		g.guessEntry.Disable()
		g.guessBtn.Disable()
		return
	}

	g.currentFact++
	if g.currentFact < 3 && g.currentFact < len(g.currentCountry.Facts) {
		g.showCurrentFact()
		g.statusLabel.SetText("Wrong! Try again with the next fact.")
	} else {
		g.statusLabel.SetText("Wrong! No more facts available.")
	}
	g.updateStatus()
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
