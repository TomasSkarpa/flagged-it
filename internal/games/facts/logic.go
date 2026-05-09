package facts

import (
	"math/rand"
	"strings"
	"time"

	"flagged-it/internal/data/models"
	"flagged-it/internal/utils"
)

// State represents the current state of the facts guessing game
type State struct {
	CurrentCountry  *models.Country
	CurrentFacts    []string
	CurrentFact     int
	CurrentFactText string // Track the currently displayed fact text
	TriesLeft       int
	UsedFacts       map[int]bool
	GuessHistory    []GuessHistoryEntry
	Score           int
	Total           int
	IsComplete      bool
}

// GuessHistoryEntry represents a single guess in the history
type GuessHistoryEntry struct {
	Guess string
	Fact  string
}

// Logic handles the game logic for facts guessing
type Logic struct {
	countries  []models.Country
	factsData  map[string]models.CountryFacts
	state      *State
	roundCount int
}

// NewLogic creates a new facts game logic instance
func NewLogic(countries []models.Country, factsData map[string]models.CountryFacts, maxRounds int) *Logic {
	return &Logic{
		countries: countries,
		factsData: factsData,
		state: &State{
			UsedFacts: make(map[int]bool),
		},
		roundCount: maxRounds,
	}
}

// GetState returns the current game state
func (l *Logic) GetState() *State {
	return l.state
}

// NewRound starts a new round of the game
func (l *Logic) NewRound() error {
	if l.state.Total >= l.roundCount {
		l.state.IsComplete = true
		return ErrGameComplete
	}

	// Find countries that have facts available
	var availableCountries []models.Country
	for _, country := range l.countries {
		if _, hasFacts := l.factsData[country.CCA2]; hasFacts {
			availableCountries = append(availableCountries, country)
		}
	}

	if len(availableCountries) == 0 {
		return ErrNoCountriesAvailable
	}

	rand.Seed(time.Now().UnixNano())
	l.state.CurrentCountry = &availableCountries[rand.Intn(len(availableCountries))]
	l.state.CurrentFacts = l.factsData[l.state.CurrentCountry.CCA2].Facts
	l.state.CurrentFact = 0
	l.state.CurrentFactText = ""
	l.state.TriesLeft = 3
	l.state.UsedFacts = make(map[int]bool)
	l.state.GuessHistory = []GuessHistoryEntry{}

	return nil
}

// GetCurrentFact returns the current fact to display
func (l *Logic) GetCurrentFact() (string, error) {
	if l.state.CurrentCountry == nil || len(l.state.UsedFacts) >= len(l.state.CurrentFacts) {
		return "", ErrNoMoreFacts
	}

	var factIndex int
	for {
		factIndex = rand.Intn(len(l.state.CurrentFacts))
		if !l.state.UsedFacts[factIndex] {
			break
		}
	}
	l.state.UsedFacts[factIndex] = true
	fact := l.state.CurrentFacts[factIndex]
	l.state.CurrentFact++
	l.state.CurrentFactText = fact // Store the current fact text

	return fact, nil
}

// MakeGuess processes a guess and returns the result.
func (l *Logic) MakeGuess(guess string, locale string) (*GuessResult, error) {
	if l.state.IsComplete {
		return nil, ErrGameComplete
	}

	// Get current fact text for history - use the stored current fact text
	currentFactText := l.state.CurrentFactText

	guess = strings.TrimSpace(guess)
	if guess == "" {
		return nil, ErrEmptyGuess
	}

	// Validate that the guessed country exists
	guessedCountry := l.findCountry(guess, locale)
	if guessedCountry == nil {
		return &GuessResult{
			IsValidGuess: false,
			Error:        "Country not found",
		}, nil
	}

	// Check if guess matches the correct country
	isCorrect := guessedCountry.CCA2 == l.state.CurrentCountry.CCA2

	l.state.GuessHistory = append(l.state.GuessHistory, GuessHistoryEntry{
		Guess: guess,
		Fact:  currentFactText,
	})

	result := &GuessResult{
		IsValidGuess:   true,
		IsCorrect:      isCorrect,
		CorrectCountry: l.state.CurrentCountry,
		TriesLeft:      l.state.TriesLeft,
		Score:          l.state.Score,
		Total:          l.state.Total,
		GuessHistory:   l.state.GuessHistory,
	}

	if isCorrect {
		l.state.Total++
		l.state.Score++
		result.Score = l.state.Score
		result.Total = l.state.Total
		result.IsComplete = l.state.Total >= l.roundCount
		if result.IsComplete {
			l.state.IsComplete = true
		}
		return result, nil
	}

	// Wrong guess
	l.state.TriesLeft--
	result.TriesLeft = l.state.TriesLeft

	if l.state.TriesLeft == 0 {
		l.state.Total++
		result.Total = l.state.Total
		result.IsComplete = l.state.Total >= l.roundCount
		if result.IsComplete {
			l.state.IsComplete = true
		}
		return result, nil
	}

	// More tries available
	return result, nil
}

// Skip skips the current round without guessing
func (l *Logic) Skip() (*GuessResult, error) {
	if l.state.IsComplete {
		return nil, ErrGameComplete
	}

	// Get current fact text for history - use the stored current fact text
	currentFactText := l.state.CurrentFactText

	// Add skip entry to history
	l.state.GuessHistory = append(l.state.GuessHistory, GuessHistoryEntry{
		Guess: "Skip",
		Fact:  currentFactText,
	})

	// Mark round as complete (like running out of tries)
	l.state.TriesLeft = 0
	l.state.Total++

	result := &GuessResult{
		IsCorrect:      false,
		CorrectCountry: l.state.CurrentCountry,
		TriesLeft:      0,
		Score:          l.state.Score,
		Total:          l.state.Total,
		IsComplete:     l.state.Total >= l.roundCount,
		GuessHistory:   l.state.GuessHistory,
	}

	if result.IsComplete {
		l.state.IsComplete = true
	}

	return result, nil
}

// Reset resets the game state
func (l *Logic) Reset() {
	l.state.Score = 0
	l.state.Total = 0
	l.state.CurrentFact = 0
	l.state.CurrentFactText = ""
	l.state.UsedFacts = make(map[int]bool)
	l.state.GuessHistory = []GuessHistoryEntry{}
	l.state.IsComplete = false
}

// findCountry finds a country by name (server-side lookup).
func (l *Logic) findCountry(name string, locale string) *models.Country {
	return utils.FindCountryByGuess(name, locale, l.countries)
}

// GuessResult represents the result of a guess
type GuessResult struct {
	IsValidGuess   bool
	IsCorrect      bool
	Error          string
	CorrectCountry *models.Country
	TriesLeft      int
	Score          int
	Total          int
	IsComplete     bool
	GuessHistory   []GuessHistoryEntry
}

// Errors
var (
	ErrNoCountriesAvailable = &GameError{Message: "no countries available"}
	ErrGameComplete         = &GameError{Message: "game is complete"}
	ErrNoMoreFacts          = &GameError{Message: "no more facts available"}
	ErrEmptyGuess           = &GameError{Message: "empty guess"}
)

type GameError struct {
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}
