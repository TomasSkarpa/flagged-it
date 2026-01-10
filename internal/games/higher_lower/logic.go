package higher_lower

import (
	"math/rand"

	"flagged-it/internal/data/models"
)

// State represents the current state of the higher/lower game
type State struct {
	CurrentCountry *models.Country
	NextCountry    *models.Country
	Score          int
	Total          int
	IsComplete     bool
	ComparisonType string // "population" or "area"
}

// Logic handles the game logic for higher/lower
type Logic struct {
	countries []models.Country
	state     *State
	maxRounds int
}

// NewLogic creates a new higher/lower game logic instance
func NewLogic(countries []models.Country) *Logic {
	return &Logic{
		countries: countries,
		state: &State{
			ComparisonType: "population",
		},
		maxRounds: 10,
	}
}

// SetComparisonType sets what to compare (population or area)
func (l *Logic) SetComparisonType(compType string) {
	if compType == "population" || compType == "area" {
		l.state.ComparisonType = compType
	}
}

// GetState returns the current game state
func (l *Logic) GetState() *State {
	return l.state
}

// NewRound starts a new round
func (l *Logic) NewRound() error {
	if l.state.Total >= l.maxRounds {
		l.state.IsComplete = true
		return ErrGameComplete
	}

	if len(l.countries) < 2 {
		return ErrNoCountriesAvailable
	}

	// Server selects two random countries
	if l.state.CurrentCountry == nil {
		// First round
		l.state.CurrentCountry = &l.countries[rand.Intn(len(l.countries))]
	}
	
	l.state.NextCountry = &l.countries[rand.Intn(len(l.countries))]
	// Ensure next country is different
	for l.state.NextCountry.CCA2 == l.state.CurrentCountry.CCA2 {
		l.state.NextCountry = &l.countries[rand.Intn(len(l.countries))]
	}

	return nil
}

// MakeGuess processes a guess (server-side validation)
func (l *Logic) MakeGuess(guess string) (*GuessResult, error) {
	if l.state.IsComplete {
		return nil, ErrGameComplete
	}

	if l.state.CurrentCountry == nil || l.state.NextCountry == nil {
		return nil, ErrGameNotStarted
	}

	// Server validates: check if guess is correct
	var currentValue, nextValue float64
	if l.state.ComparisonType == "population" {
		currentValue = float64(l.state.CurrentCountry.Population)
		nextValue = float64(l.state.NextCountry.Population)
	} else {
		currentValue = l.state.CurrentCountry.Area
		nextValue = l.state.NextCountry.Area
	}

	var isCorrect bool
	if guess == "higher" {
		isCorrect = nextValue > currentValue
	} else if guess == "lower" {
		isCorrect = nextValue < currentValue
	} else {
		return &GuessResult{
			IsValidGuess: false,
			Error:        "Invalid guess. Must be 'higher' or 'lower'",
		}, nil
	}

	l.state.Total++
	if isCorrect {
		l.state.Score++
	}

	result := &GuessResult{
		IsValidGuess:   true,
		IsCorrect:      isCorrect,
		CurrentCountry: l.state.CurrentCountry,
		NextCountry:    l.state.NextCountry,
		Score:          l.state.Score,
		Total:          l.state.Total,
		ComparisonType: l.state.ComparisonType,
		IsComplete:     l.state.Total >= l.maxRounds,
	}

	// Move to next round
	l.state.CurrentCountry = l.state.NextCountry

	if result.IsComplete {
		l.state.IsComplete = true
	}

	return result, nil
}

// Reset resets the game state
func (l *Logic) Reset() {
	l.state.Score = 0
	l.state.Total = 0
	l.state.CurrentCountry = nil
	l.state.NextCountry = nil
	l.state.IsComplete = false
}

// GuessResult represents the result of a guess (server response)
type GuessResult struct {
	IsValidGuess   bool
	IsCorrect      bool
	Error          string
	CurrentCountry *models.Country
	NextCountry    *models.Country
	Score          int
	Total          int
	ComparisonType string
	IsComplete     bool
}

// Errors
var (
	ErrNoCountriesAvailable = &GameError{Message: "no countries available"}
	ErrGameComplete         = &GameError{Message: "game is complete"}
	ErrGameNotStarted       = &GameError{Message: "game not started"}
)

type GameError struct {
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}


