package adapters

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/hangman"
	"flagged-it/internal/games/round"
)

func init() {
	countries := data.LoadCountries()
	round.Register(round.GameTypeHangman, func(opts round.RoundOptions) (round.RoundRunner, error) {
		return newHangmanRunner(countries, opts)
	})
}

type hangmanRunner struct {
	logic         *hangman.Logic
	opts          round.RoundOptions
	roundAnswered bool
}

func newHangmanRunner(countries []models.Country, opts round.RoundOptions) (*hangmanRunner, error) {
	logic := hangman.NewLogic(countries)
	return &hangmanRunner{logic: logic, opts: opts}, nil
}

func (r *hangmanRunner) StartRound(opts round.RoundOptions) error {
	r.roundAnswered = false
	return r.logic.NewRound()
}

func (r *hangmanRunner) Payload(locale string) (*round.RoundPayload, error) {
	st := r.logic.GetState()
	if st.CurrentWord == "" {
		return nil, nil
	}
	data, err := json.Marshal(map[string]interface{}{
		"guessedWord":  st.GuessedWord,
		"wrongGuesses": st.WrongGuesses,
		"maxWrongGuesses": st.MaxWrongGuesses,
	})
	if err != nil {
		return nil, err
	}
	return &round.RoundPayload{GameType: round.GameTypeHangman, Data: data}, nil
}

func (r *hangmanRunner) Submit(input *round.SubmitInput) (*round.RoundResult, error) {
	if input == nil || len(input.Data) == 0 {
		return &round.RoundResult{}, nil
	}
	var req struct {
		Letter string `json:"letter"`
	}
	if err := json.Unmarshal(input.Data, &req); err != nil || len(req.Letter) == 0 {
		return &round.RoundResult{}, nil
	}
	runes := []rune(req.Letter)
	if len(runes) != 1 {
		return &round.RoundResult{}, nil
	}
	res, err := r.logic.MakeGuess(runes[0])
	if err != nil {
		return nil, err
	}
	if res.IsWon || res.IsGameOver {
		r.roundAnswered = true
	}
	delta := 0
	if res.IsWon {
		delta = 1
	}
	var nextPayload *round.RoundPayload
	if !res.IsWon && !res.IsGameOver {
		payload, _ := r.Payload(r.opts.Locale)
		nextPayload = payload
	}
	var revealed interface{}
	if res.RevealedWord != "" {
		revealed = map[string]string{"revealedWord": res.RevealedWord}
	}
	return &round.RoundResult{
		Correct:       res.IsWon,
		ScoreDelta:    delta,
		RoundComplete: res.IsWon || res.IsGameOver,
		NextPayload:   nextPayload,
		RevealedAnswer: revealed,
	}, nil
}

func (r *hangmanRunner) IsRoundComplete() bool {
	return r.roundAnswered || r.logic.GetState().IsComplete
}
