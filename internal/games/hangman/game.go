package hangman

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

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
	currentWord    string
	guessedWord    []rune
	wrongGuesses   int
	maxWrongs      int
	guessedLetters map[rune]bool
	wordLabel      *widget.Label
	wrongLabel     *widget.Label
	statusLabel    *widget.Label
	newGameBtn     *widget.Button
	keyboard       *fyne.Container
	letterButtons  map[rune]*widget.Button
}

func NewGame(backFunc func()) *Game {
	g := &Game{
		backFunc:       backFunc,
		maxWrongs:      6,
		guessedLetters: make(map[rune]bool),
		letterButtons:  make(map[rune]*widget.Button),
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
	topBar := components.NewTopBar("Hangman Game", g.backFunc, g.newGame)

	g.wordLabel = widget.NewLabel("")
	g.wordLabel.TextStyle.Monospace = true

	g.wrongLabel = widget.NewLabel("")
	g.statusLabel = widget.NewLabel("Guess the country name!")

	g.setupKeyboard()

	g.content = container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
		g.statusLabel,
		g.wordLabel,
		g.wrongLabel,
		widget.NewSeparator(),
		g.keyboard,
	)
}

func (g *Game) newGame() {
	if len(g.countries) == 0 {
		g.statusLabel.SetText("Error loading countries data")
		return
	}

	rand.Seed(time.Now().UnixNano())
	country := g.countries[rand.Intn(len(g.countries))]
	g.currentWord = strings.ToUpper(country.CountryName)
	g.guessedWord = make([]rune, len(g.currentWord))
	g.wrongGuesses = 0
	g.guessedLetters = make(map[rune]bool)

	for i, char := range g.currentWord {
		if char == ' ' {
			g.guessedWord[i] = ' '
		} else {
			g.guessedWord[i] = '_'
		}
	}

	g.updateDisplay()
	g.statusLabel.SetText("Guess the country name!")
	for _, btn := range g.letterButtons {
		btn.Enable()
	}
}

func (g *Game) setupKeyboard() {
	rows := []string{"QWERTYUIOP", "ASDFGHJKL", "ZXCVBNM"}

	keyboardRows := make([]*fyne.Container, len(rows))
	for i, row := range rows {
		buttons := make([]fyne.CanvasObject, len(row))
		for j, letter := range row {
			letter := letter
			btn := widget.NewButton(string(letter), func() {
				g.makeGuess(letter)
			})
			g.letterButtons[letter] = btn
			buttons[j] = btn
		}
		keyboardRows[i] = container.NewHBox(buttons...)
	}

	g.keyboard = container.NewVBox(
		keyboardRows[0],
		keyboardRows[1],
		keyboardRows[2],
	)
}

func (g *Game) makeGuess(letter rune) {
	if g.guessedLetters[letter] {
		g.statusLabel.SetText("Already guessed that letter!")
		return
	}

	g.guessedLetters[letter] = true
	g.letterButtons[letter].Disable()
	found := false

	for i, char := range g.currentWord {
		if char == letter {
			g.guessedWord[i] = letter
			found = true
		}
	}

	if !found {
		g.wrongGuesses++
	}

	g.updateDisplay()
	g.checkGameEnd()
}

func (g *Game) updateDisplay() {
	g.wordLabel.SetText(string(g.guessedWord))
	g.wrongLabel.SetText(fmt.Sprintf("Wrong guesses: %d/%d", g.wrongGuesses, g.maxWrongs))
}

func (g *Game) checkGameEnd() {
	if g.wrongGuesses >= g.maxWrongs {
		g.statusLabel.SetText(fmt.Sprintf("Game Over! The word was: %s", g.currentWord))
		for _, btn := range g.letterButtons {
			btn.Disable()
		}
		return
	}

	if !strings.Contains(string(g.guessedWord), "_") {
		g.statusLabel.SetText("Congratulations! You won!")
		for _, btn := range g.letterButtons {
			btn.Disable()
		}
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
