package capital

import (
	"math/rand"
	"time"

	"flagged-it/internal/data/models"
)

// State represents the current state of the capital guessing game
type State struct {
	CurrentCountry   *models.Country
	Options          []string
	Score            int
	Total            int
	UsedCountries    map[string]bool
	SelectedRegion   string
	IsComplete       bool
}

// Logic handles the game logic for capital guessing
type Logic struct {
	allCountries []models.Country
	countries    []models.Country
	state        *State
	maxRounds    int
}

// NewLogic creates a new capital game logic instance
func NewLogic(allCountries []models.Country) *Logic {
	return &Logic{
		allCountries: allCountries,
		countries:    filterWithCapital(allCountries),
		state: &State{
			UsedCountries: make(map[string]bool),
			SelectedRegion: "",
		},
		maxRounds: 10,
	}
}

// SetRegion filters countries by region (only those with capital)
func (l *Logic) SetRegion(region string) {
	l.state.SelectedRegion = region
	if region == "" || region == "World" {
		l.countries = filterWithCapital(l.allCountries)
	} else {
		var filtered []models.Country
		for _, c := range l.allCountries {
			if c.Region == region && len(c.Capital) > 0 && c.Capital[0] != "" {
				filtered = append(filtered, c)
			}
		}
		l.countries = filtered
	}
	l.Reset()
}

func filterWithCapital(list []models.Country) []models.Country {
	var out []models.Country
	for _, c := range list {
		if len(c.Capital) > 0 && c.Capital[0] != "" {
			out = append(out, c)
		}
	}
	return out
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
		c := &l.countries[rand.Intn(len(l.countries))]
		if !l.state.UsedCountries[c.CCA2] {
			newCountry = c
			break
		}
		if len(l.state.UsedCountries) >= len(l.countries) {
			l.state.UsedCountries = make(map[string]bool)
		}
	}
	l.state.UsedCountries[newCountry.CCA2] = true
	l.state.CurrentCountry = newCountry

	correctCapital := newCountry.Capital[0]
	l.state.Options = []string{correctCapital}
	usedCapitals := map[string]bool{correctCapital: true}
	for len(l.state.Options) < 4 {
		opt := l.countries[rand.Intn(len(l.countries))]
		if len(opt.Capital) > 0 && !usedCapitals[opt.Capital[0]] {
			l.state.Options = append(l.state.Options, opt.Capital[0])
			usedCapitals[opt.Capital[0]] = true
		}
	}
	rand.Shuffle(len(l.state.Options), func(i, j int) {
		l.state.Options[i], l.state.Options[j] = l.state.Options[j], l.state.Options[i]
	})
	return nil
}

// MakeGuess processes a guess (selected capital string)
func (l *Logic) MakeGuess(answer string) (*GuessResult, error) {
	if l.state.IsComplete {
		return nil, ErrGameComplete
	}
	if l.state.CurrentCountry == nil {
		return nil, ErrGameNotStarted
	}

	l.state.Total++
	correct := answer == l.state.CurrentCountry.Capital[0]
	if correct {
		l.state.Score++
	}

	result := &GuessResult{
		IsCorrect:      correct,
		CorrectCapital: l.state.CurrentCountry.Capital[0],
		CorrectCountry: l.state.CurrentCountry,
		Score:          l.state.Score,
		Total:          l.state.Total,
		IsComplete:     l.state.Total >= l.maxRounds,
	}
	if result.IsComplete {
		l.state.IsComplete = true
	}
	return result, nil
}

// GuessResult represents the result of a guess
type GuessResult struct {
	IsCorrect      bool
	CorrectCapital string
	CorrectCountry *models.Country
	Score          int
	Total          int
	IsComplete     bool
}

// Reset resets the game state
func (l *Logic) Reset() {
	l.state.Score = 0
	l.state.Total = 0
	l.state.UsedCountries = make(map[string]bool)
	l.state.IsComplete = false
	l.state.CurrentCountry = nil
	l.state.Options = nil
}

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
