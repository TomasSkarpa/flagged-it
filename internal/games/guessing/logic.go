package guessing

import (
	"math"
	"math/rand"

	"flagged-it/internal/data/models"
	"flagged-it/internal/utils"
)

// State represents the current state of the guessing game
type State struct {
	CurrentCountry *models.Country
	Guesses        []GuessEntry
	IsComplete     bool
}

// GuessEntry represents a single guess with feedback
type GuessEntry struct {
	Country          models.Country
	IsCorrect        bool
	Continent        string // The guessed country's continent/region
	ContinentCorrect bool   // Whether the continent matches the target
	Population       ComparisonResult
	Area             ComparisonResult
}

// ComparisonResult indicates if guess is higher/lower/correct
type ComparisonResult struct {
	Value     float64
	Direction string // "higher", "lower", "correct"
	Proximity string // "very_close", "close", "far", "correct" (for color coding)
}

// Logic handles the game logic for country guessing
type Logic struct {
	countries []models.Country
	state     *State
}

// NewLogic creates a new guessing game logic instance
func NewLogic(countries []models.Country) *Logic {
	return &Logic{
		countries: countries,
		state: &State{
			Guesses: []GuessEntry{},
		},
	}
}

// GetState returns the current game state (without revealing answer)
func (l *Logic) GetState() *State {
	return l.state
}

// NewGame starts a new game with a random country
func (l *Logic) NewGame() error {
	if len(l.countries) == 0 {
		return ErrNoCountriesAvailable
	}

	l.state.CurrentCountry = &l.countries[rand.Intn(len(l.countries))]
	l.state.Guesses = []GuessEntry{}
	l.state.IsComplete = false

	return nil
}

// MakeGuess processes a guess and returns feedback (server-side validation)
func (l *Logic) MakeGuess(guessCountryName string) (*GuessResult, error) {
	if l.state.IsComplete {
		return nil, ErrGameComplete
	}

	if l.state.CurrentCountry == nil {
		return nil, ErrGameNotStarted
	}

	// Find the guessed country (server validates it exists)
	guessedCountry := l.findCountry(guessCountryName)
	if guessedCountry == nil {
		return &GuessResult{
			IsValidGuess: false,
			Error:        "Country not found",
		}, nil
	}

	// Server-side validation: check if correct
	isCorrect := guessedCountry.CCA2 == l.state.CurrentCountry.CCA2

	// Compare continent (correct if same region)
	continentCorrect := guessedCountry.Region == l.state.CurrentCountry.Region
	
	// Create comparison feedback (server calculates all hints)
	guessEntry := GuessEntry{
		Country:   *guessedCountry,
		IsCorrect: isCorrect,
		Continent: guessedCountry.Region,
		ContinentCorrect: continentCorrect,
		Population: l.compareValue(
			float64(guessedCountry.Population),
			float64(l.state.CurrentCountry.Population),
		),
		Area: l.compareValue(
			guessedCountry.Area,
			l.state.CurrentCountry.Area,
		),
	}

	l.state.Guesses = append(l.state.Guesses, guessEntry)

	result := &GuessResult{
		IsValidGuess: true,
		IsCorrect:    isCorrect,
		GuessEntry:   guessEntry,
		GuessCount:   len(l.state.Guesses),
		CorrectCountry: l.state.CurrentCountry, // Only revealed if correct or game ends
	}

	if isCorrect {
		l.state.IsComplete = true
	}

	return result, nil
}

// Reset resets the game state
func (l *Logic) Reset() {
	l.state.Guesses = []GuessEntry{}
	l.state.IsComplete = false
}

// compareValue compares two values and returns feedback
func (l *Logic) compareValue(guess, target float64) ComparisonResult {
	var direction string
	if guess > target {
		direction = "higher"
	} else if guess < target {
		direction = "lower"
	} else {
		direction = "correct"
	}

	proximity := l.calculateProximity(guess, target)

	return ComparisonResult{
		Value:     guess,
		Direction: direction,
		Proximity: proximity,
	}
}

// calculateProximity determines how close the guess is (for color coding)
func (l *Logic) calculateProximity(guess, target float64) string {
	if guess == target {
		return "correct"
	}

	var percentDiff float64
	if target != 0 {
		percentDiff = math.Abs(guess-target) / target * 100
	} else {
		percentDiff = math.Abs(guess - target)
	}

	if percentDiff <= 10 {
		return "very_close"
	} else if percentDiff <= 25 {
		return "close"
	} else if percentDiff <= 50 {
		return "somewhat_close"
	}
	return "far"
}


// findCountry finds a country by name (server-side lookup)
func (l *Logic) findCountry(name string) *models.Country {
	for _, country := range l.countries {
		if utils.MatchCountry(name, country, utils.MatchAll) {
			return &country
		}
	}
	return nil
}

// GuessResult represents the result of a guess (server response)
type GuessResult struct {
	IsValidGuess   bool
	IsCorrect      bool
	Error          string
	GuessEntry     GuessEntry
	GuessCount     int
	CorrectCountry *models.Country // Only set if correct or game ends
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


