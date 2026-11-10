package party

import (
	"errors"
	"math/rand"
	"strings"

	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/flagcolor"
)

// Kind is one type of round.
type Kind string

const (
	KindFlag    Kind = "flag"    // which country has this flag?
	KindShape   Kind = "shape"   // which country has this shape?
	KindCapital Kind = "capital" // what is the capital of X?
	KindCompare Kind = "compare" // which has more people / more area?
	KindColor   Kind = "color"   // pick the hidden colour of X's flag
	KindFact    Kind = "fact"    // which country is this fact about?
)

// GameMix mixes every kind into one game.
const GameMix = "mix"

var allKinds = []Kind{KindFlag, KindShape, KindCapital, KindCompare, KindColor, KindFact}

// ValidGame reports whether game is "mix" or a single kind.
func ValidGame(game string) bool {
	if game == GameMix {
		return true
	}
	for _, k := range allKinds {
		if string(k) == game {
			return true
		}
	}
	return false
}

// Question is one round. Everything the server needs to check an answer lives
// here; the client only ever sees a View built from it.
type Question struct {
	Kind        Kind               `json:"kind"`
	Subject     string             `json:"subject,omitempty"` // CCA2 the round is about
	Options     []string           `json:"options,omitempty"` // CCA2 per option
	Answer      string             `json:"answer,omitempty"`  // CCA2 of the right option
	Metric      string             `json:"metric,omitempty"`  // compare: population | area
	Values      map[string]float64 `json:"values,omitempty"`  // compare: value per option
	Fact        string             `json:"fact,omitempty"`
	GuessableID string             `json:"guessableId,omitempty"` // color: hidden region id
	TargetHex   string             `json:"targetHex,omitempty"`   // color: hidden colour
}

// Source builds questions. The real one reads the embedded country data;
// tests use a fake.
type Source interface {
	Generate(game, region string, n int) ([]*Question, error)
}

// DataSource builds questions from the embedded country data.
type DataSource struct {
	countries []models.Country
	byCCA2    map[string]models.Country
	facts     map[string]models.CountryFacts
}

func NewDataSource() *DataSource {
	cs := data.LoadCountries()
	by := make(map[string]models.Country, len(cs))
	for _, c := range cs {
		by[c.CCA2] = c
	}
	return &DataSource{countries: cs, byCCA2: by, facts: data.LoadCountryFacts()}
}

// Country returns the country for a CCA2 code.
func (s *DataSource) Country(cca2 string) (models.Country, bool) {
	c, ok := s.byCCA2[cca2]
	return c, ok
}

var errNoCandidate = errors.New("no candidate country")

// Generate returns n questions for game ("mix" or one kind) in region ("" = world).
func (s *DataSource) Generate(game, region string, n int) ([]*Question, error) {
	kinds := kindSequence(game, n)
	used := map[string]bool{}
	out := make([]*Question, 0, n)
	for _, k := range kinds {
		q, err := s.build(k, region, used)
		if err != nil {
			// Fall back to a flag round so a thin region never breaks a game.
			q, err = s.build(KindFlag, region, used)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, nil
}

// kindSequence spreads kinds evenly for mix, without the same kind twice in a row.
func kindSequence(game string, n int) []Kind {
	if game != GameMix {
		seq := make([]Kind, n)
		for i := range seq {
			seq[i] = Kind(game)
		}
		return seq
	}
	var seq []Kind
	for len(seq) < n {
		block := append([]Kind(nil), allKinds...)
		rand.Shuffle(len(block), func(i, j int) { block[i], block[j] = block[j], block[i] })
		if len(seq) > 0 && block[0] == seq[len(seq)-1] {
			block[0], block[len(block)-1] = block[len(block)-1], block[0]
		}
		seq = append(seq, block...)
	}
	return seq[:n]
}

func (s *DataSource) pool(region string) []models.Country {
	if region == "" || region == "World" {
		return s.countries
	}
	var p []models.Country
	for _, c := range s.countries {
		if c.Region == region {
			p = append(p, c)
		}
	}
	if len(p) < 6 {
		return s.countries
	}
	return p
}

// pick returns a random unused country from pool that passes ok.
func pick(pool []models.Country, used map[string]bool, ok func(models.Country) bool) (models.Country, bool) {
	for _, i := range rand.Perm(len(pool)) {
		c := pool[i]
		if !used[c.CCA2] && ok(c) {
			return c, true
		}
	}
	return models.Country{}, false
}

// options returns answer plus up to 3 other countries passing ok, shuffled.
func options(pool []models.Country, answer models.Country, ok func(models.Country) bool) []string {
	opts := []string{answer.CCA2}
	seen := map[string]bool{answer.CCA2: true}
	for _, i := range rand.Perm(len(pool)) {
		if len(opts) == 4 {
			break
		}
		c := pool[i]
		if !seen[c.CCA2] && ok(c) {
			seen[c.CCA2] = true
			opts = append(opts, c.CCA2)
		}
	}
	rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	return opts
}

func anyCountry(models.Country) bool { return true }

func (s *DataSource) build(k Kind, region string, used map[string]bool) (*Question, error) {
	q, err := s.buildIn(k, s.pool(region), used)
	if errors.Is(err, errNoCandidate) {
		// Region ran out for this kind: use the whole world.
		q, err = s.buildIn(k, s.countries, used)
	}
	return q, err
}

func (s *DataSource) buildIn(k Kind, pool []models.Country, used map[string]bool) (*Question, error) {
	switch k {
	case KindFlag:
		c, ok := pick(pool, used, anyCountry)
		if !ok {
			return nil, errNoCandidate
		}
		used[c.CCA2] = true
		return &Question{Kind: k, Subject: c.CCA2, Answer: c.CCA2, Options: options(pool, c, anyCountry)}, nil

	case KindShape:
		hasGeo := func(c models.Country) bool {
			g, err := data.LoadGeoData(c.CCA3)
			return err == nil && len(g.Features) > 0
		}
		c, ok := pick(pool, used, hasGeo)
		if !ok {
			return nil, errNoCandidate
		}
		used[c.CCA2] = true
		return &Question{Kind: k, Subject: c.CCA2, Answer: c.CCA2, Options: options(pool, c, anyCountry)}, nil

	case KindCapital:
		hasCapital := func(c models.Country) bool { return len(c.Capital) > 0 && c.Capital[0] != "" }
		c, ok := pick(pool, used, hasCapital)
		if !ok {
			return nil, errNoCandidate
		}
		used[c.CCA2] = true
		capitals := map[string]bool{c.Capital[0]: true}
		distinctCapital := func(o models.Country) bool {
			if !hasCapital(o) || capitals[o.Capital[0]] {
				return false
			}
			capitals[o.Capital[0]] = true
			return true
		}
		return &Question{Kind: k, Subject: c.CCA2, Answer: c.CCA2, Options: options(pool, c, distinctCapital)}, nil

	case KindCompare:
		metric := "population"
		value := func(c models.Country) float64 { return float64(c.Population) }
		if rand.Intn(2) == 0 {
			metric = "area"
			value = func(c models.Country) float64 { return c.Area }
		}
		positive := func(c models.Country) bool { return value(c) > 0 }
		a, ok := pick(pool, used, positive)
		if !ok {
			return nil, errNoCandidate
		}
		// The two values must differ clearly, or the round is a coin flip.
		b, ok := pick(pool, used, func(c models.Country) bool {
			if c.CCA2 == a.CCA2 || !positive(c) {
				return false
			}
			hi, lo := value(a), value(c)
			if lo > hi {
				hi, lo = lo, hi
			}
			return hi >= lo*1.1
		})
		if !ok {
			return nil, errNoCandidate
		}
		used[a.CCA2], used[b.CCA2] = true, true
		answer := a.CCA2
		if value(b) > value(a) {
			answer = b.CCA2
		}
		opts := []string{a.CCA2, b.CCA2}
		rand.Shuffle(2, func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
		return &Question{
			Kind: k, Answer: answer, Options: opts, Metric: metric,
			Values: map[string]float64{a.CCA2: value(a), b.CCA2: value(b)},
		}, nil

	case KindColor:
		ch, err := flagcolor.PickChallenge("easy", pool, used)
		if err != nil {
			return nil, errNoCandidate
		}
		used[ch.Country.CCA2] = true
		return &Question{Kind: k, Subject: ch.Country.CCA2, GuessableID: ch.GuessableID, TargetHex: ch.TargetHex}, nil

	case KindFact:
		var fact string
		c, ok := pick(pool, used, func(c models.Country) bool {
			fact = s.pickFact(c)
			return fact != ""
		})
		if !ok {
			return nil, errNoCandidate
		}
		used[c.CCA2] = true
		return &Question{Kind: k, Subject: c.CCA2, Answer: c.CCA2, Fact: fact, Options: options(pool, c, anyCountry)}, nil
	}
	return nil, errors.New("unknown kind")
}

// pickFact returns a random fact about c that doesn't name c, or "".
func (s *DataSource) pickFact(c models.Country) string {
	f, ok := s.facts[c.CCA2]
	if !ok {
		return ""
	}
	names := []string{strings.ToLower(c.Name.Common), strings.ToLower(c.Name.Official)}
	for _, i := range rand.Perm(len(f.Facts)) {
		text := f.Facts[i]
		lower := strings.ToLower(text)
		named := false
		for _, n := range names {
			if n != "" && strings.Contains(lower, n) {
				named = true
				break
			}
		}
		if !named {
			return text
		}
	}
	return ""
}
