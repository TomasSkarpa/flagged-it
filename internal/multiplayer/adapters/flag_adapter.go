package adapters

import (
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/flag"
	"flagged-it/internal/multiplayer"
	"time"
)

// FlagGameEngine wraps flag game logic for multiplayer
type FlagGameEngine struct {
	multiplayer.BaseGameEngine
	logic      *flag.Logic
	questions  []*multiplayer.Question
	usedCountries map[string]bool
}

// NewFlagGameEngine creates a new flag game engine
func NewFlagGameEngine(config multiplayer.RoomConfig, countries []models.Country) (multiplayer.GameEngine, error) {
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

	logic := flag.NewLogic(filteredCountries)
	logic.SetRegion(config.Region)

	engine := &FlagGameEngine{
		BaseGameEngine: multiplayer.BaseGameEngine{
			Config:         config,
			Countries:      filteredCountries,
			CurrentIdx:     0,
			TotalQuestions: config.NumQuestions,
		},
		logic:          logic,
		questions:      make([]*multiplayer.Question, 0, config.NumQuestions),
		usedCountries:  make(map[string]bool),
	}

	return engine, nil
}

// Initialize prepares questions for the game
func (e *FlagGameEngine) Initialize(config multiplayer.RoomConfig, countries []models.Country) error {
	// Questions are generated on-demand, so initialization is just validation
	return nil
}

// GetNextQuestion returns the next question
func (e *FlagGameEngine) GetNextQuestion() (*multiplayer.Question, error) {
	if e.IsComplete() {
		return nil, &multiplayer.GameError{Message: "game is complete"}
	}

	// Generate question using game logic
	if err := e.logic.NewRound(); err != nil {
		return nil, err
	}

	state := e.logic.GetState()
	if state.CurrentCountry == nil {
		return nil, &multiplayer.GameError{Message: "failed to generate question"}
	}

	// Create question
	questionID := generateQuestionID()
	question := &multiplayer.Question{
		ID:          questionID,
		Type:        "flag",
		Data: map[string]interface{}{
			"flagUrl": "/assets/twemoji_flags_cca2/" + state.CurrentCountry.CCA2 + ".svg",
		},
		Options:      state.Options,
		CorrectAnswer: state.CurrentCountry.CCA2,
		TimeLimit:    e.Config.TimeLimit,
		StartedAt:    time.Now(),
	}

	e.questions = append(e.questions, question)
	e.CurrentIdx++

	return question, nil
}

// SubmitAnswer validates an answer
func (e *FlagGameEngine) SubmitAnswer(questionID string, answer string, timeTaken int) (*multiplayer.AnswerResult, error) {
	// Find the question
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

	// Validate answer using game logic
	// Find the country that matches the answer
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

	// Use game logic to validate
	result, err := e.logic.MakeGuess(guessedCountry)
	if err != nil {
		return nil, err
	}

	isCorrect := result.IsCorrect
	points := multiplayer.CalculatePoints(isCorrect, e.Config.TimeLimit, timeTaken, e.Config.Difficulty)

	return &multiplayer.AnswerResult{
		IsCorrect:    isCorrect,
		Points:       points,
		CorrectAnswer: question.CorrectAnswer,
	}, nil
}

// IsComplete checks if all questions have been answered
func (e *FlagGameEngine) IsComplete() bool {
	return e.CurrentIdx >= e.TotalQuestions
}

// GetQuestionCount returns total number of questions
func (e *FlagGameEngine) GetQuestionCount() int {
	return e.TotalQuestions
}

// GetCurrentIndex returns current question index
func (e *FlagGameEngine) GetCurrentIndex() int {
	return e.CurrentIdx
}
