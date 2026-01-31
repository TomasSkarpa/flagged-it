package multiplayer

import (
	"flagged-it/internal/data/models"
)

// GameEngine interface for different game types
type GameEngine interface {
	// Initialize prepares the game with questions
	Initialize(config RoomConfig, countries []models.Country) error

	// GetNextQuestion returns the next question
	GetNextQuestion() (*Question, error)

	// SubmitAnswer validates an answer and returns result
	SubmitAnswer(questionID string, answer string, timeTaken int) (*AnswerResult, error)

	// IsComplete checks if all questions have been answered
	IsComplete() bool

	// GetQuestionCount returns total number of questions
	GetQuestionCount() int

	// GetCurrentIndex returns current question index (0-based)
	GetCurrentIndex() int
}

// AnswerResult represents the result of submitting an answer
type AnswerResult struct {
	IsCorrect    bool   `json:"isCorrect"`
	Points       int    `json:"points"`
	CorrectAnswer string `json:"correctAnswer"`
	Message      string `json:"message,omitempty"`
}

// BaseGameEngine provides common functionality for game engines
type BaseGameEngine struct {
	Config         RoomConfig
	Countries      []models.Country
	CurrentIdx     int
	TotalQuestions int
}


// CalculatePoints calculates points based on correctness and time taken
func CalculatePoints(isCorrect bool, timeLimit int, timeTaken int, difficulty DifficultyLevel) int {
	if !isCorrect {
		return 0
	}

	basePoints := 100
	switch difficulty {
	case DifficultyEasy:
		basePoints = 50
	case DifficultyMedium:
		basePoints = 100
	case DifficultyHard:
		basePoints = 200
	}

	// Time bonus: faster answers get more points
	if timeLimit > 0 && timeTaken > 0 {
		timeRatio := float64(timeTaken) / float64(timeLimit*1000) // Convert to ratio
		if timeRatio < 0.2 {
			basePoints = int(float64(basePoints) * 1.5) // 50% bonus for very fast
		} else if timeRatio < 0.5 {
			basePoints = int(float64(basePoints) * 1.25) // 25% bonus for fast
		} else if timeRatio < 0.8 {
			// No bonus
		} else {
			basePoints = int(float64(basePoints) * 0.75) // Penalty for slow
		}
	}

	return basePoints
}

type GameError struct {
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}
