package adapters

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/capital"
	"flagged-it/internal/games/round"
)

func init() {
	countries := data.LoadCountries()
	round.Register(round.GameTypeCapital, func(opts round.RoundOptions) (round.RoundRunner, error) {
		return newCapitalRunner(countries, opts)
	})
}

type capitalRunner struct {
	logic         *capital.Logic
	opts          round.RoundOptions
	roundAnswered bool
	questionID    string
}

func newCapitalRunner(countries []models.Country, opts round.RoundOptions) (*capitalRunner, error) {
	logic := capital.NewLogic(countries)
	if opts.Region != "" && opts.Region != "World" {
		logic.SetRegion(opts.Region)
	}
	return &capitalRunner{logic: logic, opts: opts}, nil
}

func (r *capitalRunner) StartRound(opts round.RoundOptions) error {
	if opts.Region != "" && opts.Region != "World" {
		r.logic.SetRegion(opts.Region)
	}
	r.roundAnswered = false
	r.questionID = generateQuestionID()
	return r.logic.NewRound()
}

func (r *capitalRunner) Payload(locale string) (*round.RoundPayload, error) {
	st := r.logic.GetState()
	if st.CurrentCountry == nil {
		return nil, nil
	}
	countryName := st.CurrentCountry.GetTranslatedName(locale)
	data, err := json.Marshal(map[string]interface{}{
		"countryName": countryName,
		"countryCca2": st.CurrentCountry.CCA2,
		"options":     st.Options,
		"questionId":  r.questionID,
	})
	if err != nil {
		return nil, err
	}
	return &round.RoundPayload{GameType: round.GameTypeCapital, Data: data}, nil
}

func (r *capitalRunner) Submit(input *round.SubmitInput) (*round.RoundResult, error) {
	if input == nil || len(input.Data) == 0 {
		return &round.RoundResult{RoundComplete: true}, nil
	}
	var req struct {
		Answer string `json:"answer"`
	}
	if err := json.Unmarshal(input.Data, &req); err != nil {
		return &round.RoundResult{RoundComplete: true}, nil
	}
	res, err := r.logic.MakeGuess(req.Answer)
	if err != nil {
		return nil, err
	}
	r.roundAnswered = true
	delta := 0
	if res.IsCorrect {
		delta = 1
	}
	var revealed interface{}
	if res.CorrectCountry != nil {
		revealed = map[string]string{
			"correctCapital": res.CorrectCapital,
			"correctCountry": res.CorrectCountry.GetTranslatedName(r.opts.Locale),
		}
	}
	return &round.RoundResult{
		Correct:       res.IsCorrect,
		ScoreDelta:    delta,
		RoundComplete: true,
		RevealedAnswer: revealed,
	}, nil
}

func (r *capitalRunner) IsRoundComplete() bool {
	return r.roundAnswered
}
