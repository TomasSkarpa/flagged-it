package hangman

import (
	"math/rand"
	"strings"
	"unicode"

	"flagged-it/internal/data/models"
	"flagged-it/internal/utils"
)

// State represents the current state of the hangman game
type State struct {
	CurrentWord     string   // The word to guess (country name)
	GuessedWord     []string // Current state of guessed word (e.g., ["_", "_", "U", "_"])
	GuessedLetters  map[rune]bool
	WrongGuesses    int
	MaxWrongGuesses int
	Score           int
	Total           int
	IsComplete      bool
	IsWon           bool
}

// Logic handles the game logic for hangman
type Logic struct {
	countries []models.Country
	state     *State
	maxRounds int
}

// NewLogic creates a new hangman game logic instance
func NewLogic(countries []models.Country) *Logic {
	return &Logic{
		countries: countries,
		state: &State{
			GuessedLetters: make(map[rune]bool),
			MaxWrongGuesses: 6,
		},
		maxRounds: 5,
	}
}

// GetState returns the current game state (without revealing answer until game ends)
func (l *Logic) GetState() *State {
	return l.state
}

// NewRound starts a new round with a random country name
func (l *Logic) NewRound() error {
	if l.state.Total >= l.maxRounds {
		l.state.IsComplete = true
		return ErrGameComplete
	}

	if len(l.countries) == 0 {
		return ErrNoCountriesAvailable
	}

	// Server selects random country
	country := l.countries[rand.Intn(len(l.countries))]
	l.state.CurrentWord = strings.ToUpper(country.Name.Common)
	
	// Initialize guessed word with underscores
	l.state.GuessedWord = make([]string, len(l.state.CurrentWord))
	for i := range l.state.GuessedWord {
		if l.state.CurrentWord[i] == ' ' {
			l.state.GuessedWord[i] = " "
		} else {
			l.state.GuessedWord[i] = "_"
		}
	}

	l.state.GuessedLetters = make(map[rune]bool)
	l.state.WrongGuesses = 0
	l.state.IsWon = false

	return nil
}

// MakeGuess processes a letter guess (server-side validation)
func (l *Logic) MakeGuess(letter rune) (*GuessResult, error) {
	if l.state.IsComplete || l.state.CurrentWord == "" {
		return nil, ErrGameComplete
	}

	letter = unicode.ToUpper(letter)

	// Server validates: check if already guessed
	if l.state.GuessedLetters[letter] {
		return &GuessResult{
			IsValidGuess: false,
			Error:        "Letter already guessed",
		}, nil
	}

	// Server validates: check if letter is in word
	l.state.GuessedLetters[letter] = true
	isInWord := strings.ContainsRune(l.state.CurrentWord, letter)

	if isInWord {
		// Reveal all instances of the letter
		for i, r := range l.state.CurrentWord {
			if r == letter {
				l.state.GuessedWord[i] = string(letter)
			}
		}

		// Check if word is complete (server-side check)
		isWon := !strings.Contains(strings.Join(l.state.GuessedWord, ""), "_")
		
		result := &GuessResult{
			IsValidGuess: true,
			IsInWord:     true,
			GuessedWord:  l.state.GuessedWord,
			WrongGuesses: l.state.WrongGuesses,
			IsWon:        isWon,
		}

		if isWon {
			l.state.IsWon = true
			l.state.Total++
			l.state.Score++
			result.Score = l.state.Score
			result.Total = l.state.Total
			result.IsComplete = l.state.Total >= l.maxRounds
			if result.IsComplete {
				l.state.IsComplete = true
			}
			result.RevealedWord = l.state.CurrentWord
		}

		return result, nil
	}

	// Wrong guess
	l.state.WrongGuesses++
	isGameOver := l.state.WrongGuesses >= l.state.MaxWrongGuesses

	result := &GuessResult{
		IsValidGuess: true,
		IsInWord:     false,
		GuessedWord:  l.state.GuessedWord,
		WrongGuesses: l.state.WrongGuesses,
		IsGameOver:   isGameOver,
	}

	if isGameOver {
		l.state.Total++
		result.Total = l.state.Total
		result.IsComplete = l.state.Total >= l.maxRounds
		result.RevealedWord = l.state.CurrentWord // Reveal on loss
		if result.IsComplete {
			l.state.IsComplete = true
		}
	}

	return result, nil
}

// Reset resets the game state
func (l *Logic) Reset() {
	l.state.Score = 0
	l.state.Total = 0
	l.state.GuessedLetters = make(map[rune]bool)
	l.state.IsComplete = false
	l.state.IsWon = false
}

// GuessResult represents the result of a guess (server response)
type GuessResult struct {
	IsValidGuess bool
	IsInWord     bool
	Error        string
	GuessedWord  []string
	WrongGuesses int
	IsWon        bool
	IsGameOver   bool
	Score        int
	Total        int
	IsComplete   bool
	RevealedWord string // Only revealed when game ends (win or lose)
}

// Errors
var (
	ErrNoCountriesAvailable = &GameError{Message: "no countries available"}
	ErrGameComplete         = &GameError{Message: "game is complete"}
)

type GameError struct {
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}


