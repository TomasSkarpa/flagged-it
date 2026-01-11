package list

import (
	"math/rand"
	"sort"

	"flagged-it/internal/data/models"
)

// State represents the current state of the list game
type State struct {
	Countries  []models.Country // Countries to order
	UserOrder  []string         // User's ordered list of CCA2 codes
	Score      int
	Total      int
	IsComplete bool
	SortBy     string // "population", "area", "name"
}

// Logic handles the game logic for list ordering
type Logic struct {
	countries []models.Country
	state     *State
	maxRounds int
}

// NewLogic creates a new list game logic instance
func NewLogic(countries []models.Country) *Logic {
	return &Logic{
		countries: countries,
		state: &State{
			SortBy: "population",
		},
		maxRounds: 5,
	}
}

// SetSortBy sets what to sort by
func (l *Logic) SetSortBy(sortBy string) {
	if sortBy == "population" || sortBy == "area" || sortBy == "name" {
		l.state.SortBy = sortBy
	}
}

// GetState returns the current game state (without correct order)
func (l *Logic) GetState() *State {
	return l.state
}

// NewRound starts a new round with random countries
func (l *Logic) NewRound() error {
	if l.state.Total >= l.maxRounds {
		l.state.IsComplete = true
		return ErrGameComplete
	}

	if len(l.countries) < 3 {
		return ErrNoCountriesAvailable
	}

	// Server selects 5-8 random countries
	count := 5 + rand.Intn(4) // 5-8 countries
	selected := make([]models.Country, 0, count)
	used := make(map[string]bool)

	for len(selected) < count {
		c := l.countries[rand.Intn(len(l.countries))]
		if !used[c.CCA2] {
			selected = append(selected, c)
			used[c.CCA2] = true
		}
	}

	// Shuffle them for display
	shuffled := make([]models.Country, len(selected))
	copy(shuffled, selected)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	l.state.Countries = shuffled
	l.state.UserOrder = []string{}

	return nil
}

// SubmitOrder validates the user's ordering (server-side validation)
func (l *Logic) SubmitOrder(userOrder []string) (*OrderResult, error) {
	if l.state.IsComplete {
		return nil, ErrGameComplete
	}

	if len(userOrder) != len(l.state.Countries) {
		return &OrderResult{
			IsValid: false,
			Error:   "Order length doesn't match country count",
		}, nil
	}

	// Server calculates correct order
	correctOrder := l.getCorrectOrder()

	// Server validates: check if order is correct
	isCorrect := true
	for i, cca2 := range userOrder {
		if cca2 != correctOrder[i] {
			isCorrect = false
			break
		}
	}

	l.state.UserOrder = userOrder
	l.state.Total++

	result := &OrderResult{
		IsValid:      true,
		IsCorrect:    isCorrect,
		UserOrder:    userOrder,
		CorrectOrder: correctOrder, // Revealed after submission
		Score:        l.state.Score,
		Total:        l.state.Total,
	}

	if isCorrect {
		l.state.Score++
		result.Score = l.state.Score
	}

	result.IsComplete = l.state.Total >= l.maxRounds
	if result.IsComplete {
		l.state.IsComplete = true
	}

	return result, nil
}

// getCorrectOrder calculates the correct order (server-side only)
func (l *Logic) getCorrectOrder() []string {
	sorted := make([]models.Country, len(l.state.Countries))
	copy(sorted, l.state.Countries)

	switch l.state.SortBy {
	case "population":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Population > sorted[j].Population
		})
	case "area":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Area > sorted[j].Area
		})
	case "name":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name.Common < sorted[j].Name.Common
		})
	}

	result := make([]string, len(sorted))
	for i, c := range sorted {
		result[i] = c.CCA2
	}

	return result
}

// Reset resets the game state
func (l *Logic) Reset() {
	l.state.Score = 0
	l.state.Total = 0
	l.state.Countries = []models.Country{}
	l.state.UserOrder = []string{}
	l.state.IsComplete = false
}

// OrderResult represents the result of submitting an order (server response)
type OrderResult struct {
	IsValid      bool
	IsCorrect    bool
	Error        string
	UserOrder    []string
	CorrectOrder []string // Revealed after submission
	Score        int
	Total        int
	IsComplete   bool
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
