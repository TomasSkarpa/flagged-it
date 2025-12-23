package hangman

import (
	"fmt"
	"image/color"
	"math/rand"
	"strings"
	"time"

	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
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
	hintLabel      *widget.Label
	wrongLabel     *widget.Label
	statusLabel    *widget.Label
	newGameBtn     *widget.Button
	keyboard       *fyne.Container
	letterButtons  map[rune]*components.Button
	score          int
	total          int
	scoreLabel     *widget.Label
	scoreManager   *utils.ScoreManager
}

func NewGame(backFunc func(), scoreManager *utils.ScoreManager) *Game {
	g := &Game{
		backFunc:       backFunc,
		maxWrongs:      6,
		guessedLetters: make(map[rune]bool),
		letterButtons:  make(map[rune]*components.Button),
		scoreManager:   scoreManager,
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
	topBar := components.NewTopBar(lang.X("game.hangman.title", "Hangman"), g.backFunc, g.Reset)

	g.wordLabel = widget.NewLabel("")
	g.wordLabel.TextStyle.Monospace = true

	g.hintLabel = widget.NewLabel("")
	g.wrongLabel = widget.NewLabel("")
	g.statusLabel = widget.NewLabel(lang.X("game.hangman.guess_country", "Guess the country name!"))
	g.scoreLabel = widget.NewLabel(fmt.Sprintf(lang.X("game.hangman.score", "Score: %d/5"), 0))

	g.setupKeyboard()

	g.content = container.NewVBox(
		topBar.GetContainer(),
		components.NewDashedSeparator(color.RGBA{200, 200, 200, 255}, 5),
		g.scoreLabel,
		g.statusLabel,
		g.wordLabel,
		g.hintLabel,
		g.wrongLabel,
		components.NewDashedSeparator(color.RGBA{200, 200, 200, 255}, 5),
		g.keyboard,
	)

}

func (g *Game) newGame() {
	if len(g.countries) == 0 {
		g.statusLabel.SetText(lang.X("error.loading_countries", "Error loading countries data"))
		return
	}

	if g.total >= 5 {
		g.scoreManager.SetTotal("hangman", 5)
		g.scoreManager.UpdateScore("hangman", g.score)
		g.statusLabel.SetText(fmt.Sprintf(lang.X("game.complete", "Game Complete! Final Score: %d/10 (%.0f%%)"), g.score, float64(g.score)/5*100))
		for _, btn := range g.letterButtons {
			btn.Disable()
		}
		return
	}

	rand.Seed(time.Now().UnixNano())
	country := g.countries[rand.Intn(len(g.countries))]
	g.currentWord = strings.ToUpper(country.Name.Common)
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
	g.statusLabel.SetText(lang.X("game.hangman.guess_country", "Guess the country name!"))
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
			btn := components.NewButton(string(letter), func() {
				g.makeGuess(letter)
			})
			g.letterButtons[letter] = btn
			buttons[j] = btn
		}
		keyboardRows[i] = container.NewHBox(buttons...)
	}

	g.keyboard = container.NewVBox(
		container.NewCenter(keyboardRows[0]),
		container.NewCenter(keyboardRows[1]),
		container.NewCenter(keyboardRows[2]),
	)
}

func (g *Game) makeGuess(letter rune) {
	if g.guessedLetters[letter] {
		g.statusLabel.SetText(lang.X("game.hangman.already_guessed", "Already guessed that letter!"))
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
	var displayWord strings.Builder
	for i, char := range g.guessedWord {
		if char == ' ' {
			displayWord.WriteString("   ")
		} else {
			displayWord.WriteRune(char)
			if i < len(g.guessedWord)-1 && g.guessedWord[i+1] != ' ' {
				displayWord.WriteString(" ")
			}
		}
	}

	letterCount := 0
	wordCount := 1
	for _, char := range g.currentWord {
		if char == ' ' {
			wordCount++
		} else {
			letterCount++
		}
	}

	g.wordLabel.SetText(displayWord.String())
	wordText := lang.X("game.hangman.word", "word")
	if wordCount != 1 {
		wordText = lang.X("game.hangman.words", "words")
	}
	g.hintLabel.SetText(fmt.Sprintf(lang.X("game.hangman.letters_words", "%d letters, %d %s"), letterCount, wordCount, wordText))
	g.wrongLabel.SetText(fmt.Sprintf(lang.X("game.hangman.wrong_guesses", "Wrong guesses: %d/%d"), g.wrongGuesses, g.maxWrongs))
}

func (g *Game) checkGameEnd() {
	if g.wrongGuesses >= g.maxWrongs {
		g.statusLabel.SetText(fmt.Sprintf(lang.X("game.hangman.game_over", "Game Over! The word was: %s"), g.currentWord))
		for _, btn := range g.letterButtons {
			btn.Disable()
		}
		g.total++
		g.scoreLabel.SetText(fmt.Sprintf(lang.X("game.hangman.score", "Score: %d/5"), g.score))
		time.AfterFunc(1500*time.Millisecond, func() {
			fyne.Do(func() {
				g.newGame()
			})
		})
		return
	}

	if !strings.Contains(string(g.guessedWord), "_") {
		g.statusLabel.SetText(lang.X("game.hangman.congratulations", "Congratulations! You won!"))
		for _, btn := range g.letterButtons {
			btn.Disable()
		}
		g.total++
		g.score++
		g.scoreLabel.SetText(fmt.Sprintf(lang.X("game.hangman.score", "Score: %d/5"), g.score))
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
	g.scoreLabel.SetText(fmt.Sprintf(lang.X("game.hangman.score", "Score: %d/5"), 0))
	g.newGame()
}

func (g *Game) TypedKey(key *fyne.KeyEvent) {
	if len(string(key.Name)) == 1 {
		letter := rune(strings.ToUpper(string(key.Name))[0])
		if letter >= 'A' && letter <= 'Z' {
			g.makeGuess(letter)
		}
	}
}
