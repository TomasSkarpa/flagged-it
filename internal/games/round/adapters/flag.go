package adapters

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/flag"
	"flagged-it/internal/games/round"
	"math/rand"
	"time"
)

func init() {
	countries := data.LoadCountries()
	round.Register(round.GameTypeFlag, func(opts round.RoundOptions) (round.RoundRunner, error) {
		return newFlagRunner(countries, opts)
	})
}

type flagRunner struct {
	logic         *flag.Logic
	opts          round.RoundOptions
	roundAnswered bool
	questionID    string
}

func newFlagRunner(countries []models.Country, opts round.RoundOptions) (*flagRunner, error) {
	logic := flag.NewLogic(countries)
	if opts.Region != "" && opts.Region != "World" {
		logic.SetRegion(opts.Region)
	}
	return &flagRunner{logic: logic, opts: opts}, nil
}

func (r *flagRunner) StartRound(opts round.RoundOptions) error {
	if opts.Region != "" && opts.Region != "World" {
		r.logic.SetRegion(opts.Region)
	}
	r.roundAnswered = false
	r.questionID = generateQuestionID()
	return r.logic.NewRound()
}

func (r *flagRunner) Payload(locale string) (*round.RoundPayload, error) {
	st := r.logic.GetState()
	if st.CurrentCountry == nil {
		return nil, nil
	}
	options := translateCountries(st.Options, locale)
	flagURL := "/assets/twemoji_flags_cca2/" + st.CurrentCountry.CCA2 + ".svg"
	data, err := json.Marshal(map[string]interface{}{
		"flagUrl":    flagURL,
		"options":    options,
		"questionId": r.questionID,
	})
	if err != nil {
		return nil, err
	}
	return &round.RoundPayload{GameType: round.GameTypeFlag, Data: data}, nil
}

func (r *flagRunner) Submit(input *round.SubmitInput) (*round.RoundResult, error) {
	if input == nil || len(input.Data) == 0 {
		return &round.RoundResult{Correct: false, RoundComplete: true}, nil
	}
	var req struct {
		Cca2 string `json:"cca2"`
	}
	if err := json.Unmarshal(input.Data, &req); err != nil {
		return &round.RoundResult{Correct: false, RoundComplete: true}, nil
	}
	st := r.logic.GetState()
	var guessed *models.Country
	for i := range st.Options {
		if st.Options[i].CCA2 == req.Cca2 {
			guessed = &st.Options[i]
			break
		}
	}
	if guessed == nil {
		return &round.RoundResult{Correct: false, RoundComplete: true}, nil
	}
	res, err := r.logic.MakeGuess(guessed)
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
			"correctCca2": res.CorrectCountry.CCA2,
			"correctName": res.CorrectCountry.Name.Common,
		}
	}
	return &round.RoundResult{
		Correct:       res.IsCorrect,
		ScoreDelta:    delta,
		RoundComplete: true,
		RevealedAnswer: revealed,
	}, nil
}

func (r *flagRunner) IsRoundComplete() bool {
	return r.roundAnswered
}

func generateQuestionID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 12)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func translateCountries(cs []models.Country, locale string) []models.Country {
	out := make([]models.Country, len(cs))
	for i := range cs {
		out[i] = cs[i]
		out[i].Name.Common = cs[i].GetTranslatedName(locale)
		out[i].Name.Translations = nil
	}
	return out
}
