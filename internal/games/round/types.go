package round

import "encoding/json"

// GameType identifies a game for the registry and payloads.
// Values must match frontend GameType (e.g. "flag", "shape", "facts").
type GameType string

const (
	GameTypeFlag       GameType = "flag"
	GameTypeShape      GameType = "shape"
	GameTypeCapital    GameType = "capital"
	GameTypeHigherLower GameType = "higher_lower"
	GameTypeHangman    GameType = "hangman"
	GameTypeFacts      GameType = "facts"
	GameTypeGuessing   GameType = "guessing"
	GameTypeList       GameType = "list"
)

// RoundOptions holds options when starting a round or creating a runner.
// Consumer passes these; runner uses what it needs (e.g. region, locale).
type RoundOptions struct {
	Region          string            `json:"region,omitempty"`
	Locale          string            `json:"locale,omitempty"`
	ComparisonType  string            `json:"comparisonType,omitempty"` // e.g. "population", "area" for higher_lower
	Extra           map[string]string `json:"-"`                        // game-specific, not serialized
}

// RoundPayload is what the runner returns for the client to render.
// GameType tells the frontend which UI to use; Data is game-specific JSON.
type RoundPayload struct {
	GameType GameType       `json:"gameType"`
	Data     json.RawMessage `json:"data"`
}

// SubmitInput is the opaque submit payload from the client.
// Each game type interprets it (e.g. optionId, guess, direction).
type SubmitInput struct {
	Data json.RawMessage `json:"data"`
}

// RoundResult is returned after Submit.
// RoundComplete true means the consumer can advance to the next round.
// NextPayload is set for multi-step rounds (e.g. next fact, guess again).
type RoundResult struct {
	Correct       bool         `json:"correct"`
	ScoreDelta    int          `json:"scoreDelta"`
	RoundComplete bool         `json:"roundComplete"`
	NextPayload   *RoundPayload `json:"nextPayload,omitempty"`
	RevealedAnswer interface{} `json:"revealedAnswer,omitempty"` // optional, for end-of-round reveal
}
