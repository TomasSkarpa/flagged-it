package multiplayer

import (
	"flagged-it/internal/data/models"
	"time"
)

// RoomStatus represents the current status of a game room
type RoomStatus string

const (
	RoomStatusWaiting RoomStatus = "waiting"  // Waiting for players, not started
	RoomStatusPlaying RoomStatus = "playing"  // Game in progress
	RoomStatusFinished RoomStatus = "finished" // Game completed
)

// GameMode represents the type of game being played
type GameMode string

const (
	GameModeFlag      GameMode = "flag"      // Flag guessing
	GameModeShape     GameMode = "shape"      // Shape guessing
	GameModeCapital   GameMode = "capital"    // Capital guessing
	GameModeHigherLower GameMode = "higherlower" // Higher/lower
	GameModeFacts     GameMode = "facts"      // Facts guessing
	GameModeWorldle   GameMode = "worldle"    // Worldle guessing
)

// DifficultyLevel represents the difficulty setting
type DifficultyLevel string

const (
	DifficultyEasy   DifficultyLevel = "easy"
	DifficultyMedium DifficultyLevel = "medium"
	DifficultyHard   DifficultyLevel = "hard"
)

// RoomConfig holds configuration for a game room
type RoomConfig struct {
	GameMode      GameMode       `json:"gameMode"`
	NumQuestions  int            `json:"numQuestions"`
	Difficulty    DifficultyLevel `json:"difficulty"`
	TimeLimit     int            `json:"timeLimit"` // seconds per question, 0 = unlimited
	MaxPlayers    int            `json:"maxPlayers"`
	Region        string         `json:"region"` // empty = all regions
	Categories    []string       `json:"categories"` // game-specific categories
	IsPublic      bool           `json:"isPublic"`
	Password      string         `json:"password,omitempty"` // empty = no password
}

// DefaultRoomConfig returns a default room configuration
func DefaultRoomConfig() RoomConfig {
	return RoomConfig{
		GameMode:     GameModeFlag,
		NumQuestions: 10,
		Difficulty:   DifficultyMedium,
		TimeLimit:    0, // unlimited
		MaxPlayers:   10,
		Region:       "",
		Categories:   []string{},
		IsPublic:     false,
		Password:     "",
	}
}

// Player represents a player in a room
type Player struct {
	ID        string    `json:"id"`        // Unique player ID
	Name      string    `json:"name"`      // Display name
	IsHost    bool      `json:"isHost"`    // Is the room host
	IsReady   bool      `json:"isReady"`   // Ready to start
	Score     int       `json:"score"`     // Current score
	Streak    int       `json:"streak"`    // Current streak
	JoinedAt  time.Time `json:"joinedAt"` // When they joined
	LastAnswer *AnswerSubmission `json:"lastAnswer,omitempty"` // Last answer submitted
}

// AnswerSubmission represents a player's answer to a question
type AnswerSubmission struct {
	PlayerID    string    `json:"playerId"`
	QuestionID  string    `json:"questionId"`
	Answer      string    `json:"answer"`      // Answer value (country CCA2, etc.)
	IsCorrect   bool      `json:"isCorrect"`
	Points      int       `json:"points"`      // Points earned (with time bonus)
	TimeTaken   int       `json:"timeTaken"`   // Milliseconds taken
	SubmittedAt time.Time `json:"submittedAt"`
}

// Question represents a game question
type Question struct {
	ID          string           `json:"id"`
	Type        string           `json:"type"`        // "flag", "shape", etc.
	Data        interface{}      `json:"data"`        // Question-specific data
	Options     []models.Country `json:"options,omitempty"` // Multiple choice options
	CorrectAnswer string         `json:"-"`          // Server-only, not sent to clients
	TimeLimit   int              `json:"timeLimit"`   // seconds
	StartedAt   time.Time        `json:"startedAt"`
}

// Room represents a game room
type Room struct {
	ID            string                 `json:"id"`
	Code          string                 `json:"code"`          // Short code for joining
	HostID        string                 `json:"hostId"`
	Config        RoomConfig             `json:"config"`
	Status        RoomStatus             `json:"status"`
	Players       map[string]*Player     `json:"players"`       // playerID -> Player
	CurrentQuestion *Question            `json:"currentQuestion,omitempty"`
	QuestionIndex int                    `json:"questionIndex"` // Current question number (0-indexed)
	Answers       map[string]map[string]*AnswerSubmission `json:"answers,omitempty"` // questionID -> playerID -> Answer
	CreatedAt     time.Time              `json:"createdAt"`
	StartedAt     *time.Time             `json:"startedAt,omitempty"`
	FinishedAt    *time.Time             `json:"finishedAt,omitempty"`
	gameEngine    GameEngine             // Internal game engine instance
}

// PublicRoomInfo represents public information about a room (for listing)
type PublicRoomInfo struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	GameMode    GameMode  `json:"gameMode"`
	NumQuestions int      `json:"numQuestions"`
	Difficulty  DifficultyLevel `json:"difficulty"`
	PlayerCount int       `json:"playerCount"`
	MaxPlayers  int       `json:"maxPlayers"`
	Status      RoomStatus `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}
