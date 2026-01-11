package flag

import (
	"math/rand"
	"time"

	"flagged-it/internal/data/models"
)

// State represents the current state of the flag guessing game
type State struct {
	CurrentCountry *models.Country
	Options        []models.Country
	Score          int
	Total          int
	UsedCountries  map[string]bool
	SelectedRegion string
	IsComplete     bool
}

// Logic handles the game logic for flag guessing
type Logic struct {
	allCountries []models.Country
	countries    []models.Country
	state        *State
	maxRounds    int
}

// NewLogic creates a new flag game logic instance
func NewLogic(allCountries []models.Country) *Logic {
	return &Logic{
		allCountries: allCountries,
		countries:    allCountries,
		state: &State{
			UsedCountries:  make(map[string]bool),
			SelectedRegion: "",
		},
		maxRounds: 10,
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
	l.Reset()
}

// GetState returns the current game state
func (l *Logic) GetState() *State {
	return l.state
}

// NewRound starts a new round of the game
func (l *Logic) NewRound() error {
	if len(l.countries) == 0 {
		return ErrNoCountriesAvailable
	}

	if l.state.Total >= l.maxRounds {
		l.state.IsComplete = true
		return ErrGameComplete
	}

	rand.Seed(time.Now().UnixNano())
	var newCountry *models.Country
	for {
		newCountry = &l.countries[rand.Intn(len(l.countries))]
		if !l.state.UsedCountries[newCountry.CCA2] {
			break
		}
	}
	l.state.UsedCountries[newCountry.CCA2] = true
	l.state.CurrentCountry = newCountry

	// Generate 4 options (1 correct + 3 wrong)
	l.state.Options = []models.Country{*l.state.CurrentCountry}
	usedOptions := make(map[string]bool)
	usedOptions[l.state.CurrentCountry.CCA2] = true
	for len(l.state.Options) < 4 {
		option := l.countries[rand.Intn(len(l.countries))]
		if !usedOptions[option.CCA2] {
			l.state.Options = append(l.state.Options, option)
			usedOptions[option.CCA2] = true
		}
	}

	// Shuffle options
	rand.Shuffle(len(l.state.Options), func(i, j int) {
		l.state.Options[i], l.state.Options[j] = l.state.Options[j], l.state.Options[i]
	})

	return nil
}

// MakeGuess processes a guess and returns the result
func (l *Logic) MakeGuess(guessedCountry *models.Country) (*GuessResult, error) {
	if l.state.IsComplete {
		return nil, ErrGameComplete
	}

	l.state.Total++
	isCorrect := guessedCountry.CCA2 == l.state.CurrentCountry.CCA2

	result := &GuessResult{
		IsCorrect:      isCorrect,
		CorrectCountry: l.state.CurrentCountry,
		Score:          l.state.Score,
		Total:          l.state.Total,
		IsComplete:     l.state.Total >= l.maxRounds,
	}

	if isCorrect {
		l.state.Score++
		result.Score = l.state.Score
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
	l.state.UsedCountries = make(map[string]bool)
	l.state.IsComplete = false
}

// GuessResult represents the result of a guess
type GuessResult struct {
	IsCorrect      bool
	CorrectCountry *models.Country
	Score          int
	Total          int
	IsComplete     bool
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
