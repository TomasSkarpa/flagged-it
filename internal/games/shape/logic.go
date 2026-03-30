package shape

import (
	"math/rand"

	"flagged-it/internal/data/models"
	"flagged-it/internal/utils"
)

// State represents the current state of the shape guessing game
type State struct {
	CurrentCountry *models.Country
	Countries      []models.Country // Available countries for the region
	Score          int
	Total          int
	IsComplete     bool
	SelectedRegion string
}

// Logic handles the game logic for shape guessing
type Logic struct {
	allCountries []models.Country
	countries    []models.Country
	state        *State
	roundCount   int
}

// NewLogic creates a new shape game logic instance
func NewLogic(allCountries []models.Country, maxRounds int) *Logic {
	return &Logic{
		allCountries: allCountries,
		countries:    allCountries,
		state: &State{
			SelectedRegion: "",
		},
		roundCount: maxRounds,
	}
}

// SetRegion filters countries by region
func (l *Logic) SetRegion(region string) {
	l.state.SelectedRegion = region
	if region == "" || region == "World" {
		l.countries = l.allCountries
	} else {
		l.countries = []models.Country{}
		for _, country := range l.allCountries {
			if country.Region == region {
				l.countries = append(l.countries, country)
			}
		}
	}
	l.state.Countries = l.countries
	l.Reset()
}

// GetState returns the current game state
func (l *Logic) GetState() *State {
	return l.state
}

// NewRound starts a new round with a random country
func (l *Logic) NewRound() error {
	if l.state.Total >= l.roundCount {
		l.state.IsComplete = true
		return ErrGameComplete
	}

	if len(l.countries) == 0 {
		return ErrNoCountriesAvailable
	}

	// Server selects random country
	l.state.CurrentCountry = &l.countries[rand.Intn(len(l.countries))]

	return nil
}

// MakeGuess processes a guess (server-side validation)
func (l *Logic) MakeGuess(guessCountryName string) (*GuessResult, error) {
	if l.state.IsComplete {
		return nil, ErrGameComplete
	}

	if l.state.CurrentCountry == nil {
		return nil, ErrGameNotStarted
	}

	// Server validates: find guessed country
	guessedCountry := l.findCountry(guessCountryName)
	if guessedCountry == nil {
		return &GuessResult{
			IsValidGuess: false,
			Error:        "Country not found",
		}, nil
	}

	// Server validates: check if correct
	isCorrect := guessedCountry.CCA2 == l.state.CurrentCountry.CCA2

	l.state.Total++
	if isCorrect {
		l.state.Score++
	}

	result := &GuessResult{
		IsValidGuess:   true,
		IsCorrect:      isCorrect,
		CorrectCountry: l.state.CurrentCountry, // Revealed after guess
		Score:          l.state.Score,
		Total:          l.state.Total,
		IsComplete:     l.state.Total >= l.roundCount,
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
	l.state.CurrentCountry = nil
	l.state.IsComplete = false
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
	CorrectCountry *models.Country // Revealed after guess
	Score          int
	Total          int
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
