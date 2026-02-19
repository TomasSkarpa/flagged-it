package round

// RoundRunner is the common interface for all game rounds.
// Any consumer (per-game API, quiz, random one question, multiplayer) uses only this interface.
type RoundRunner interface {
	// StartRound starts or advances to the current round. Options may include region, locale, etc.
	StartRound(opts RoundOptions) error

	// Payload returns what to send to the client for this round (game type + opaque data).
	Payload(locale string) (*RoundPayload, error)

	// Submit processes the user's answer/guess. Returns result and optionally next payload for multi-step rounds.
	Submit(input *SubmitInput) (*RoundResult, error)

	// IsRoundComplete returns true when the current round is finished (consumer can advance or end session).
	IsRoundComplete() bool
}
