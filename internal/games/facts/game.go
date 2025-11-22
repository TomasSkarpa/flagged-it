package facts

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type GuessHistory struct {
	Guess string
	Fact  string
}

type Game struct {
	content        *fyne.Container
	backFunc       func()
	countries      []models.Country
	factsData      map[string]models.CountryFacts
	currentCountry *models.Country
	currentFacts   []string
	currentFact    int
	triesLeft      int
	usedFacts      map[int]bool
	guessHistory   []GuessHistory
	factLabel      *widget.Label
	guessEntry     *widget.Entry
	statusLabel    *widget.Label
	triesLabel     *widget.Label
	guessBtn       *widget.Button
	newGameBtn     *widget.Button
	historyContainer *fyne.Container
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
	g.factsData = data.LoadCountryFacts()
}

func (g *Game) setupUI() {
	topBar := components.NewTopBar("Guess by Facts", g.backFunc, g.newGame)

	g.factLabel = widget.NewLabel("")
	g.factLabel.Wrapping = fyne.TextWrapWord

	g.guessEntry = widget.NewEntry()
	g.guessEntry.SetPlaceHolder("Enter country name...")
	g.guessEntry.OnSubmitted = func(text string) { g.makeGuess() }

	g.guessBtn = widget.NewButton("Guess", g.makeGuess)
	g.statusLabel = widget.NewLabel("")
	g.triesLabel = widget.NewLabel("")

	guessContainer := container.NewGridWithColumns(2, g.guessEntry, g.guessBtn)
	g.historyContainer = container.NewVBox()

	g.content = container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
		g.statusLabel,
		g.triesLabel,
		widget.NewSeparator(),
		g.factLabel,
		widget.NewSeparator(),
		guessContainer,
		g.historyContainer,
	)
}

func (g *Game) newGame() {
	if len(g.countries) == 0 || len(g.factsData) == 0 {
		g.statusLabel.SetText("Error loading countries data")
		return
	}

	// Find countries that have facts available
	var availableCountries []models.Country
	for _, country := range g.countries {
		if _, hasFacts := g.factsData[country.CCA2]; hasFacts {
			availableCountries = append(availableCountries, country)
		}
	}

	if len(availableCountries) == 0 {
		g.statusLabel.SetText("No countries with facts available")
		return
	}

	rand.Seed(time.Now().UnixNano())
	g.currentCountry = &availableCountries[rand.Intn(len(availableCountries))]
	g.currentFacts = g.factsData[g.currentCountry.CCA2].Facts
	g.currentFact = 0
	g.triesLeft = 3
	g.usedFacts = make(map[int]bool)
	g.guessHistory = []GuessHistory{}
	g.guessEntry.SetText("")
	g.guessEntry.Enable()
	g.guessBtn.Enable()
	g.updateHistoryUI()

	g.showCurrentFact()
	g.updateStatus()
}

func (g *Game) showCurrentFact() {
	if g.currentCountry != nil && len(g.usedFacts) < len(g.currentFacts) {
		var factIndex int
		for {
			factIndex = rand.Intn(len(g.currentFacts))
			if !g.usedFacts[factIndex] {
				break
			}
		}
		g.usedFacts[factIndex] = true
		fact := g.currentFacts[factIndex]
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

	if utils.MatchesCountry(guess, *g.currentCountry) {
		currentFactText := g.factLabel.Text
		g.guessHistory = append(g.guessHistory, GuessHistory{
			Guess: fmt.Sprintf("%s ✅", guess),
			Fact:  currentFactText,
		})
		g.updateHistoryUI()
		
		g.statusLabel.SetText(fmt.Sprintf("Correct! It was %s!", g.currentCountry.Name.Common))
		g.guessEntry.Disable()
		g.guessBtn.Disable()
		return
	}

	currentFactText := g.factLabel.Text
	g.guessHistory = append(g.guessHistory, GuessHistory{
		Guess: guess,
		Fact:  currentFactText,
	})
	g.updateHistoryUI()

	g.triesLeft--
	g.guessEntry.SetText("")

	if g.triesLeft == 0 {
		g.statusLabel.SetText(fmt.Sprintf("Game Over! It was %s", g.currentCountry.Name.Common))
		g.guessEntry.Disable()
		g.guessBtn.Disable()
		return
	}

	g.currentFact++
	if g.currentFact < 3 && g.currentFact < len(g.currentFacts) {
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

func (g *Game) updateHistoryUI() {
	g.historyContainer.RemoveAll()
	
	if len(g.guessHistory) == 0 {
		return
	}

	historyTitle := widget.NewLabel("Previous Guesses:")
	historyTitle.TextStyle.Bold = true
	g.historyContainer.Add(historyTitle)
	g.historyContainer.Add(widget.NewSeparator())

	for i, history := range g.guessHistory {
		// Create guess header (emoji already included in history.Guess for correct answers)
		guessText := history.Guess
		if !strings.Contains(guessText, "✅") {
			guessText = fmt.Sprintf("%s ❌", guessText)
		}
		guessHeader := widget.NewLabel(fmt.Sprintf("Guess %d: %s", i+1, guessText))
		guessHeader.TextStyle.Bold = true
		
		// Create fact text (remove "Fact X:" prefix for cleaner display)
		factText := history.Fact
		if strings.Contains(factText, ": ") {
			parts := strings.SplitN(factText, ": ", 2)
			if len(parts) == 2 {
				factText = parts[1]
			}
		}
		factLabel := widget.NewLabel(fmt.Sprintf("Fact: %s", factText))
		factLabel.Wrapping = fyne.TextWrapWord
		
		// Create a card-like container for each guess
		guessCard := container.NewVBox(
			guessHeader,
			factLabel,
		)
		
		g.historyContainer.Add(guessCard)
		if i < len(g.guessHistory)-1 {
			g.historyContainer.Add(widget.NewSeparator())
		}
	}
	
	g.historyContainer.Refresh()
}

func (g *Game) Reset() {
	g.newGame()
}
