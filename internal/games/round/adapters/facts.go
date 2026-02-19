package adapters

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/facts"
	"flagged-it/internal/games/round"
)

func init() {
	countries := data.LoadCountries()
	factsData := data.LoadCountryFacts()
	round.Register(round.GameTypeFacts, func(opts round.RoundOptions) (round.RoundRunner, error) {
		return newFactsRunner(countries, factsData, opts)
	})
}

type factsRunner struct {
	logic         *facts.Logic
	opts          round.RoundOptions
	roundComplete bool
}

func newFactsRunner(countries []models.Country, factsData map[string]models.CountryFacts, opts round.RoundOptions) (*factsRunner, error) {
	logic := facts.NewLogic(countries, factsData)
	return &factsRunner{logic: logic, opts: opts}, nil
}

func (r *factsRunner) StartRound(opts round.RoundOptions) error {
	r.roundComplete = false
	if err := r.logic.NewRound(); err != nil {
		return err
	}
	_, _ = r.logic.GetCurrentFact()
	return nil
}

func (r *factsRunner) Payload(locale string) (*round.RoundPayload, error) {
	st := r.logic.GetState()
	if st.CurrentCountry == nil {
		return nil, nil
	}
	fact := st.CurrentFactText
	if fact == "" {
		var err error
		fact, err = r.logic.GetCurrentFact()
		if err != nil {
			return nil, err
		}
	}
	data, err := json.Marshal(map[string]interface{}{
		"fact":         fact,
		"triesLeft":    st.TriesLeft,
		"guessHistory": st.GuessHistory,
	})
	if err != nil {
		return nil, err
	}
	return &round.RoundPayload{GameType: round.GameTypeFacts, Data: data}, nil
}

func (r *factsRunner) Submit(input *round.SubmitInput) (*round.RoundResult, error) {
	if input == nil || len(input.Data) == 0 {
		return &round.RoundResult{}, nil
	}
	var req struct {
		Action string `json:"action"`
		Guess  string `json:"guess"`
	}
	_ = json.Unmarshal(input.Data, &req)
	if req.Action == "skip" {
		res, err := r.logic.Skip()
		if err != nil {
			return nil, err
		}
		r.roundComplete = true
		return r.factsResult(res, true)
	}
	res, err := r.logic.MakeGuess(req.Guess)
	if err != nil {
		return nil, err
	}
	roundComplete := res.IsComplete || res.TriesLeft == 0
	if roundComplete {
		r.roundComplete = true
	}
	var nextPayload *round.RoundPayload
	if !roundComplete && !res.IsCorrect {
		_, _ = r.logic.GetCurrentFact()
		payload, _ := r.Payload(r.opts.Locale)
		nextPayload = payload
	}
	return r.factsResult(res, roundComplete, nextPayload)
}

func (r *factsRunner) factsResult(res *facts.GuessResult, roundComplete bool, nextPayload ...*round.RoundPayload) (*round.RoundResult, error) {
	delta := 0
	if res.IsCorrect {
		delta = 1
	}
	var np *round.RoundPayload
	if len(nextPayload) > 0 {
		np = nextPayload[0]
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
		RoundComplete: roundComplete,
		NextPayload:   np,
		RevealedAnswer: revealed,
	}, nil
}

func (r *factsRunner) IsRoundComplete() bool {
	return r.roundComplete
}
