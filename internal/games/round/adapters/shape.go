package adapters

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/round"
	"flagged-it/internal/games/shape"
	"math/rand"
	"sync"
)

var (
	shapeCountries     []models.Country
	shapeCountriesOnce sync.Once
)

func init() {
	round.Register(round.GameTypeShape, func(opts round.RoundOptions) (round.RoundRunner, error) {
		return newShapeRunner(opts)
	})
}

func countriesWithGeo(region string) []models.Country {
	shapeCountriesOnce.Do(func() {
		all := data.LoadCountries()
		for _, c := range all {
			if c.CCA3 != "" {
				if _, err := data.LoadGeoData(c.CCA3); err == nil {
					shapeCountries = append(shapeCountries, c)
				}
			}
		}
	})
	if region == "" || region == "World" {
		return shapeCountries
	}
	var out []models.Country
	for _, c := range shapeCountries {
		if c.Region == region {
			out = append(out, c)
		}
	}
	return out
}

type shapeRunner struct {
	logic         *shape.Logic
	opts          round.RoundOptions
	roundAnswered bool
	questionID    string
}

func newShapeRunner(opts round.RoundOptions) (*shapeRunner, error) {
	countries := countriesWithGeo(opts.Region)
	logic := shape.NewLogic(countries)
	if opts.Region != "" && opts.Region != "World" {
		logic.SetRegion(opts.Region)
	}
	return &shapeRunner{logic: logic, opts: opts}, nil
}

func (r *shapeRunner) StartRound(opts round.RoundOptions) error {
	if opts.Region != "" && opts.Region != "World" {
		r.logic.SetRegion(opts.Region)
	}
	r.roundAnswered = false
	r.questionID = generateQuestionID()
	return r.logic.NewRound()
}

func (r *shapeRunner) Payload(locale string) (*round.RoundPayload, error) {
	st := r.logic.GetState()
	if st.CurrentCountry == nil {
		return nil, nil
	}
	geoData, err := data.LoadGeoData(st.CurrentCountry.CCA3)
	if err != nil {
		return nil, err
	}
	countries := countriesWithGeo(r.opts.Region)
	options := []models.Country{*st.CurrentCountry}
	used := map[string]bool{st.CurrentCountry.CCA2: true}
	for len(options) < 4 {
		c := countries[rand.Intn(len(countries))]
		if !used[c.CCA2] {
			options = append(options, c)
			used[c.CCA2] = true
		}
	}
	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	options = translateCountries(options, locale)
	data, err := json.Marshal(map[string]interface{}{
		"geoJson":    geoData,
		"options":    options,
		"questionId": r.questionID,
	})
	if err != nil {
		return nil, err
	}
	return &round.RoundPayload{GameType: round.GameTypeShape, Data: data}, nil
}

func (r *shapeRunner) Submit(input *round.SubmitInput) (*round.RoundResult, error) {
	if input == nil || len(input.Data) == 0 {
		return &round.RoundResult{RoundComplete: true}, nil
	}
	var req struct {
		Guess string `json:"guess"`
		Cca2  string `json:"cca2"`
	}
	_ = json.Unmarshal(input.Data, &req)
	guess := req.Guess
	if guess == "" && req.Cca2 != "" {
		st := r.logic.GetState()
		for i := range st.Countries {
			if st.Countries[i].CCA2 == req.Cca2 {
				guess = st.Countries[i].Name.Common
				break
			}
		}
	}
	if guess == "" {
		return &round.RoundResult{RoundComplete: true}, nil
	}
	res, err := r.logic.MakeGuess(guess)
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

func (r *shapeRunner) IsRoundComplete() bool {
	return r.roundAnswered
}
