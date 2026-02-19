package adapters

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/guessing"
	"flagged-it/internal/games/round"
)

func init() {
	countries := data.LoadCountries()
	round.Register(round.GameTypeGuessing, func(opts round.RoundOptions) (round.RoundRunner, error) {
		return newGuessingRunner(countries, opts)
	})
}

type guessingRunner struct {
	logic *guessing.Logic
	opts  round.RoundOptions
}

func newGuessingRunner(countries []models.Country, opts round.RoundOptions) (*guessingRunner, error) {
	logic := guessing.NewLogic(countries)
	return &guessingRunner{logic: logic, opts: opts}, nil
}

func (r *guessingRunner) StartRound(opts round.RoundOptions) error {
	return r.logic.NewGame()
}

func (r *guessingRunner) Payload(locale string) (*round.RoundPayload, error) {
	st := r.logic.GetState()
	if st.CurrentCountry == nil {
		return nil, nil
	}
	data, err := json.Marshal(map[string]interface{}{
		"guesses": st.Guesses,
	})
	if err != nil {
		return nil, err
	}
	return &round.RoundPayload{GameType: round.GameTypeGuessing, Data: data}, nil
}

func (r *guessingRunner) Submit(input *round.SubmitInput) (*round.RoundResult, error) {
	if input == nil || len(input.Data) == 0 {
		return &round.RoundResult{}, nil
	}
	var req struct {
		Guess string `json:"guess"`
	}
	if err := json.Unmarshal(input.Data, &req); err != nil {
		return &round.RoundResult{}, nil
	}
	locale := r.opts.Locale
	if locale == "" {
		locale = "en"
	}
	res, err := r.logic.MakeGuess(req.Guess, locale)
	if err != nil {
		return nil, err
	}
	delta := 0
	if res.IsCorrect {
		delta = 1
	}
	var nextPayload *round.RoundPayload
	if !res.IsCorrect && res.IsValidGuess {
		payload, _ := r.Payload(locale)
		nextPayload = payload
	}
	var revealed interface{}
	if res.CorrectCountry != nil {
		revealed = map[string]string{
			"correctCca2": res.CorrectCountry.CCA2,
			"correctName": res.CorrectCountry.Name.Common,
		}
	}
	return &round.RoundResult{
		Correct:       res.IsCorrect,
		ScoreDelta:    delta,
		RoundComplete: res.IsCorrect,
		NextPayload:   nextPayload,
		RevealedAnswer: revealed,
	}, nil
}

func (r *guessingRunner) IsRoundComplete() bool {
	return r.logic.GetState().IsComplete
}
