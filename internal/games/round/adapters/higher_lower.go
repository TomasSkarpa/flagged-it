package adapters

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/higher_lower"
	"flagged-it/internal/games/round"
)

func init() {
	countries := data.LoadCountries()
	round.Register(round.GameTypeHigherLower, func(opts round.RoundOptions) (round.RoundRunner, error) {
		return newHigherLowerRunner(countries, opts)
	})
}

type higherLowerRunner struct {
	logic         *higher_lower.Logic
	opts          round.RoundOptions
	roundAnswered bool
}

func newHigherLowerRunner(countries []models.Country, opts round.RoundOptions) (*higherLowerRunner, error) {
	logic := higher_lower.NewLogic(countries)
	if opts.ComparisonType != "" {
		logic.SetComparisonType(opts.ComparisonType)
	}
	return &higherLowerRunner{logic: logic, opts: opts}, nil
}

func (r *higherLowerRunner) StartRound(opts round.RoundOptions) error {
	if opts.ComparisonType != "" {
		r.logic.SetComparisonType(opts.ComparisonType)
	}
	r.roundAnswered = false
	return r.logic.NewRound()
}

func (r *higherLowerRunner) Payload(locale string) (*round.RoundPayload, error) {
	st := r.logic.GetState()
	if st.CurrentCountry == nil || st.NextCountry == nil {
		return nil, nil
	}
	cur := translateCountries([]models.Country{*st.CurrentCountry}, locale)[0]
	next := translateCountries([]models.Country{*st.NextCountry}, locale)[0]
	data, err := json.Marshal(map[string]interface{}{
		"currentCountry": cur,
		"nextCountry":    next,
		"comparisonType": st.ComparisonType,
	})
	if err != nil {
		return nil, err
	}
	return &round.RoundPayload{GameType: round.GameTypeHigherLower, Data: data}, nil
}

func (r *higherLowerRunner) Submit(input *round.SubmitInput) (*round.RoundResult, error) {
	if input == nil || len(input.Data) == 0 {
		return &round.RoundResult{RoundComplete: true}, nil
	}
	var req struct {
		Direction string `json:"direction"`
	}
	if err := json.Unmarshal(input.Data, &req); err != nil {
		return &round.RoundResult{RoundComplete: true}, nil
	}
	res, err := r.logic.MakeGuess(req.Direction)
	if err != nil {
		return nil, err
	}
	r.roundAnswered = true
	delta := 0
	if res.IsCorrect {
		delta = 1
	}
	return &round.RoundResult{
		Correct:       res.IsCorrect,
		ScoreDelta:    delta,
		RoundComplete: true,
	}, nil
}

func (r *higherLowerRunner) IsRoundComplete() bool {
	return r.roundAnswered
}
