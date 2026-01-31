package adapters

import (
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/shape"
	"flagged-it/internal/multiplayer"
	"math/rand"
	"time"
)

// ShapeGameEngine wraps shape game logic for multiplayer
type ShapeGameEngine struct {
	multiplayer.BaseGameEngine
	logic     *shape.Logic
	questions []*multiplayer.Question
}

// NewShapeGameEngine creates a new shape game engine
func NewShapeGameEngine(config multiplayer.RoomConfig, countries []models.Country) (multiplayer.GameEngine, error) {
	// Filter countries by region if specified
	filteredCountries := countries
	if config.Region != "" && config.Region != "World" {
		filteredCountries = []models.Country{}
		for _, country := range countries {
			if country.Region == config.Region {
				filteredCountries = append(filteredCountries, country)
			}
		}
	}

	if len(filteredCountries) == 0 {
		return nil, &multiplayer.GameError{Message: "no countries available for selected region"}
	}

	logic := shape.NewLogic(filteredCountries)
	logic.SetRegion(config.Region)

	engine := &ShapeGameEngine{
		BaseGameEngine: multiplayer.BaseGameEngine{
			Config:         config,
			Countries:      filteredCountries,
			CurrentIdx:     0,
			TotalQuestions: config.NumQuestions,
		},
		logic:     logic,
		questions: make([]*multiplayer.Question, 0, config.NumQuestions),
	}

	return engine, nil
}

func (e *ShapeGameEngine) Initialize(config multiplayer.RoomConfig, countries []models.Country) error {
	return nil
}

func (e *ShapeGameEngine) GetNextQuestion() (*multiplayer.Question, error) {
	if e.IsComplete() {
		return nil, &multiplayer.GameError{Message: "game is complete"}
	}

	if err := e.logic.NewRound(); err != nil {
		return nil, err
	}

	state := e.logic.GetState()
	if state.CurrentCountry == nil {
		return nil, &multiplayer.GameError{Message: "failed to generate question"}
	}

	// Generate 4 options (1 correct + 3 wrong)
	options := []models.Country{*state.CurrentCountry}
	usedOptions := make(map[string]bool)
	usedOptions[state.CurrentCountry.CCA2] = true
	
	// Get available countries for options
	availableCountries := e.Countries
	if len(availableCountries) < 4 {
		// If we have fewer than 4 countries, use what we have
		availableCountries = state.Countries
	}
	
	// Add 3 random wrong options
	for len(options) < 4 && len(availableCountries) > 1 {
		option := availableCountries[rand.Intn(len(availableCountries))]
		if !usedOptions[option.CCA2] {
			options = append(options, option)
			usedOptions[option.CCA2] = true
		}
		// Prevent infinite loop if we don't have enough countries
		if len(usedOptions) >= len(availableCountries) {
			break
		}
	}

	// Shuffle options
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	questionID := generateQuestionID()
	question := &multiplayer.Question{
		ID:          questionID,
		Type:        "shape",
		Data: map[string]interface{}{
			"country": state.CurrentCountry.CCA2,
		},
		Options:      options,
		CorrectAnswer: state.CurrentCountry.CCA2,
		TimeLimit:    e.Config.TimeLimit,
		StartedAt:    time.Now(),
	}

	e.questions = append(e.questions, question)
	e.CurrentIdx++

	return question, nil
}

func (e *ShapeGameEngine) SubmitAnswer(questionID string, answer string, timeTaken int) (*multiplayer.AnswerResult, error) {
	var question *multiplayer.Question
	for _, q := range e.questions {
		if q.ID == questionID {
			question = q
			break
		}
	}

	if question == nil {
		return nil, &multiplayer.GameError{Message: "question not found"}
	}

	// Find the country that matches the answer (answer is CCA2)
	var guessedCountry *models.Country
	for _, country := range question.Options {
		if country.CCA2 == answer {
			guessedCountry = &country
			break
		}
	}

	if guessedCountry == nil {
		return &multiplayer.AnswerResult{
			IsCorrect: false,
			Points:    0,
		}, nil
	}

	// Use game logic to validate (MakeGuess expects country name string)
	result, err := e.logic.MakeGuess(guessedCountry.Name.Common)
	if err != nil {
		return nil, err
	}

	// Handle invalid guess (country not found by name matching)
	if !result.IsValidGuess {
		return &multiplayer.AnswerResult{
			IsCorrect:    false,
			Points:       0,
			CorrectAnswer: question.CorrectAnswer,
		}, nil
	}

	points := multiplayer.CalculatePoints(result.IsCorrect, e.Config.TimeLimit, timeTaken, e.Config.Difficulty)

	return &multiplayer.AnswerResult{
		IsCorrect:    result.IsCorrect,
		Points:       points,
		CorrectAnswer: question.CorrectAnswer,
	}, nil
}

func (e *ShapeGameEngine) IsComplete() bool {
	return e.CurrentIdx >= e.TotalQuestions
}

func (e *ShapeGameEngine) GetQuestionCount() int {
	return e.TotalQuestions
}

func (e *ShapeGameEngine) GetCurrentIndex() int {
	return e.CurrentIdx
}
