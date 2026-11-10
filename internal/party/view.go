package party

import (
	"strconv"
	"time"

	"flagged-it/internal/data/models"
)

// Lookup finds countries by CCA2 code (for labels in the player's language).
type Lookup interface {
	Country(cca2 string) (models.Country, bool)
}

// View is what one player sees. It never contains the answer to a question
// that player hasn't finished.
type View struct {
	Code        string        `json:"code"`
	Mode        Mode          `json:"mode"`
	Game        string        `json:"game"`
	Region      string        `json:"region"`
	Rounds      int           `json:"rounds"`
	Phase       Phase         `json:"phase"`
	Round       int           `json:"round"` // 1-based; 0 before the first round
	RemainingMs int64         `json:"remainingMs"`
	DurationMs  int64         `json:"durationMs"`
	You         YouView       `json:"you"`
	Players     []PlayerView  `json:"players"`
	Question    *QuestionView `json:"question,omitempty"`
	Reveal      *RevealView   `json:"reveal,omitempty"`
	Version     uint64        `json:"version"`
}

type YouView struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	IsHost bool        `json:"isHost"`
	Score  int         `json:"score"`
	Answer *AnswerView `json:"answer,omitempty"` // your answer to the current round
}

type PlayerView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Score     int    `json:"score"`
	IsHost    bool   `json:"isHost"`
	IsYou     bool   `json:"isYou"`
	Connected bool   `json:"connected"`
	Answered  bool   `json:"answered"`         // live: answered the current round
	Points    *int   `json:"points,omitempty"` // live reveal: points this round
	Correct   *bool  `json:"correct,omitempty"`
	Done      int    `json:"done"` // challenge: rounds finished
	Finished  bool   `json:"finished"`
}

type QuestionView struct {
	Kind        Kind         `json:"kind"`
	Key         string       `json:"key"` // stable id for this round, e.g. ABCD-3
	CountryName string       `json:"countryName,omitempty"`
	FlagURL     string       `json:"flagUrl,omitempty"` // only when the country is named anyway
	Fact        string       `json:"fact,omitempty"`
	Metric      string       `json:"metric,omitempty"`
	GuessableID string       `json:"guessableId,omitempty"`
	Options     []OptionView `json:"options,omitempty"`
}

type OptionView struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	FlagURL string `json:"flagUrl,omitempty"` // compare rounds only
}

type RevealView struct {
	AnswerID    string             `json:"answerId,omitempty"`
	AnswerLabel string             `json:"answerLabel,omitempty"`
	CountryName string             `json:"countryName,omitempty"`
	CCA2        string             `json:"cca2,omitempty"`
	TargetHex   string             `json:"targetHex,omitempty"`
	Values      map[string]float64 `json:"values,omitempty"`
}

type AnswerView struct {
	OptionID string `json:"optionId,omitempty"`
	Hex      string `json:"hex,omitempty"`
	Correct  bool   `json:"correct"`
	Points   int    `json:"points"`
	TimedOut bool   `json:"timedOut,omitempty"`
}

func flagURL(cca2 string) string { return "/assets/twemoji_flags_cca2/" + cca2 + ".svg" }

func countryName(l Lookup, cca2, locale string) string {
	c, ok := l.Country(cca2)
	if !ok {
		return cca2
	}
	return c.GetTranslatedName(locale)
}

// View builds what player p sees right now. Hold the lock.
func (r *Room) View(p *Player, locale string, l Lookup, now time.Time) View {
	v := View{
		Code: r.Code, Mode: r.Mode, Game: r.Game, Region: r.Region,
		Rounds: len(r.Questions), Version: r.version,
		You: YouView{ID: p.ID, Name: p.Name, IsHost: p.ID == r.HostID, Score: p.Score},
	}

	// Which round and phase this player is in.
	phase, round := r.Phase, r.Round
	var start, end time.Time
	if r.Mode == ModeChallenge {
		phase, round = p.Phase, p.Round
		if phase == PhaseQuestion {
			start = p.StartedAt
			end = start.Add(RoundDuration(r.Questions[round].Kind))
		}
	} else if phase == PhaseQuestion || phase == PhaseReveal {
		start, end = r.PhaseStart, r.PhaseEnd
	}
	v.Phase = phase
	if phase == PhaseQuestion || phase == PhaseReveal {
		v.Round = round + 1
		if !end.IsZero() {
			v.DurationMs = end.Sub(start).Milliseconds()
			if rem := end.Sub(now).Milliseconds(); rem > 0 {
				v.RemainingMs = rem
			}
		}
	}
	if phase == PhaseFinished {
		v.Round = len(r.Questions)
	}

	if phase == PhaseQuestion || phase == PhaseReveal {
		q := r.Questions[round]
		v.Question = questionView(q, r.Code, round, locale, l)
		if a, ok := p.Answers[round]; ok {
			v.You.Answer = &AnswerView{OptionID: a.OptionID, Hex: a.Hex, Correct: a.Correct, Points: a.Points, TimedOut: a.TimedOut}
		}
		if phase == PhaseReveal {
			v.Reveal = revealView(q, locale, l)
		}
	}

	for _, pl := range r.ranked() {
		pv := PlayerView{
			ID: pl.ID, Name: pl.Name, Score: pl.Score, IsHost: pl.ID == r.HostID,
			IsYou: pl == p, Connected: pl.connected(now), Done: len(pl.Answers),
			Finished: pl.Phase == PhaseFinished,
		}
		if r.Mode == ModeLive && (phase == PhaseQuestion || phase == PhaseReveal) {
			a, ok := pl.Answers[round]
			pv.Answered = ok
			if ok && phase == PhaseReveal {
				pts, correct := a.Points, a.Correct
				pv.Points, pv.Correct = &pts, &correct
			}
		}
		v.Players = append(v.Players, pv)
	}
	return v
}

func questionView(q *Question, code string, round int, locale string, l Lookup) *QuestionView {
	qv := &QuestionView{Kind: q.Kind, Key: code + "-" + strconv.Itoa(round+1)}
	switch q.Kind {
	case KindCapital:
		qv.CountryName = countryName(l, q.Subject, locale)
		qv.FlagURL = flagURL(q.Subject)
	case KindColor:
		qv.CountryName = countryName(l, q.Subject, locale)
		qv.FlagURL = flagURL(q.Subject)
		qv.GuessableID = q.GuessableID
	case KindFact:
		qv.Fact = q.Fact
	case KindCompare:
		qv.Metric = q.Metric
	}
	for _, id := range q.Options {
		o := OptionView{ID: id, Label: countryName(l, id, locale)}
		switch q.Kind {
		case KindCapital:
			if c, ok := l.Country(id); ok && len(c.Capital) > 0 {
				o.Label = c.Capital[0]
			}
		case KindCompare:
			o.FlagURL = flagURL(id)
		}
		qv.Options = append(qv.Options, o)
	}
	return qv
}

func revealView(q *Question, locale string, l Lookup) *RevealView {
	rv := &RevealView{AnswerID: q.Answer, Values: q.Values, TargetHex: q.TargetHex}
	subject := q.Subject
	if subject == "" {
		subject = q.Answer
	}
	rv.CCA2 = subject
	rv.CountryName = countryName(l, subject, locale)
	if q.Answer != "" {
		rv.AnswerLabel = countryName(l, q.Answer, locale)
		if q.Kind == KindCapital {
			if c, ok := l.Country(q.Answer); ok && len(c.Capital) > 0 {
				rv.AnswerLabel = c.Capital[0]
			}
		}
	}
	return rv
}
