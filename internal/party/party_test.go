package party

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"flagged-it/internal/data/models"
)

// fakeSource returns flag rounds: answer is always "AA", options AA..DD.
type fakeSource struct{ kind Kind }

func (f fakeSource) Generate(game, region string, n int) ([]*Question, error) {
	qs := make([]*Question, n)
	for i := range qs {
		k := f.kind
		if k == "" {
			k = KindFlag
		}
		qs[i] = &Question{Kind: k, Subject: "AA", Answer: "AA", Options: []string{"BB", "AA", "CC", "DD"}}
		if k == KindColor {
			qs[i] = &Question{Kind: k, Subject: "AA", GuessableID: "g1", TargetHex: "#ff0000"}
		}
	}
	return qs, nil
}

type fakeLookup struct{}

func (fakeLookup) Country(cca2 string) (models.Country, bool) {
	return models.Country{CCA2: cca2, Name: models.CountryName{Common: "Land " + cca2}, Capital: []string{"City " + cca2}}, true
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time      { return c.t }
func (c *clock) add(d time.Duration) { c.t = c.t.Add(d) }
func newTestHub(k Kind) (*Hub, *clock) {
	c := &clock{t: time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)}
	h := NewHub(fakeSource{kind: k}, fakeLookup{})
	h.now = c.now
	return h, c
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestLiveGameHappyPath(t *testing.T) {
	h, c := newTestHub(KindFlag)
	r, host, err := h.Create(ModeLive, "flag", "", "Ana")
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Code) != 4 {
		t.Fatalf("code %q", r.Code)
	}
	r.Lock()
	defer r.Unlock()
	guest, err := r.Join("  Ben  ", c.now())
	if err != nil {
		t.Fatal(err)
	}
	if guest.Name != "Ben" {
		t.Fatalf("name %q", guest.Name)
	}
	if err := r.Start(guest, c.now()); err != ErrNotHost {
		t.Fatalf("guest start: %v", err)
	}
	if err := r.Start(host, c.now()); err != nil {
		t.Fatal(err)
	}

	// The question view must not reveal the answer.
	v := r.View(guest, "en", fakeLookup{}, c.now())
	if v.Phase != PhaseQuestion || v.Round != 1 || v.Reveal != nil {
		t.Fatalf("bad question view: %+v", v)
	}
	if strings.Contains(mustJSON(t, v), `"answer"`) {
		t.Fatal("question view leaks the answer")
	}

	// Host answers fast and right, guest slow and wrong.
	c.add(time.Second)
	a, err := r.Answer(host, 0, "AA", nil, c.now())
	if err != nil || !a.Correct || a.Points <= 900 {
		t.Fatalf("host answer: %+v %v", a, err)
	}
	if _, err := r.Answer(host, 0, "AA", nil, c.now()); err != ErrAlreadyAnswered {
		t.Fatalf("double answer: %v", err)
	}
	if r.Phase != PhaseQuestion {
		t.Fatal("should wait for the guest")
	}
	c.add(5 * time.Second)
	if _, err := r.Answer(guest, 0, "BB", nil, c.now()); err != nil {
		t.Fatal(err)
	}
	if r.Phase != PhaseReveal {
		t.Fatal("everyone answered: should reveal")
	}
	v = r.View(guest, "en", fakeLookup{}, c.now())
	if v.Reveal == nil || v.Reveal.AnswerID != "AA" || v.Players[0].ID != host.ID {
		t.Fatalf("bad reveal: %+v", v.Reveal)
	}

	// Play through the rest on timers.
	for i := 0; i < 600 && r.Phase != PhaseFinished; i++ {
		c.add(time.Second)
		r.Seen(host, c.now())
		r.Seen(guest, c.now())
		r.Tick(c.now())
	}
	if r.Phase != PhaseFinished {
		t.Fatalf("game did not finish, phase %s round %d", r.Phase, r.Round)
	}
	v = r.View(host, "en", fakeLookup{}, c.now())
	if v.Round != RoundsPerGame || v.Players[0].Name != "Ana" {
		t.Fatalf("bad final view: %+v", v)
	}

	// Play again keeps players, resets scores.
	qs, _ := h.NewQuestions(r)
	if err := r.Again(host, qs, c.now()); err != nil {
		t.Fatal(err)
	}
	if r.Phase != PhaseLobby || host.Score != 0 || len(r.Players) != 2 {
		t.Fatal("again did not reset")
	}
}

func TestLiveAnswerAfterTimeIsRejected(t *testing.T) {
	h, c := newTestHub(KindFlag)
	r, host, _ := h.Create(ModeLive, "flag", "", "Ana")
	r.Lock()
	defer r.Unlock()
	_ = r.Start(host, c.now())
	c.add(RoundDuration(KindFlag) + answerGrace + time.Second)
	if _, err := r.Answer(host, 0, "AA", nil, c.now()); err != ErrWrongRound {
		t.Fatalf("late answer: %v", err)
	}
}

func TestHostHandsOverWhenAway(t *testing.T) {
	h, c := newTestHub(KindFlag)
	r, host, _ := h.Create(ModeLive, "flag", "", "Ana")
	r.Lock()
	defer r.Unlock()
	guest, _ := r.Join("Ben", c.now())
	r.Connect(guest, c.now())
	c.add(hostAwayAfter + time.Second)
	r.Tick(c.now())
	if r.HostID != guest.ID {
		t.Fatal("host should move to the connected guest")
	}
	_ = host
}

func TestChallengeHappyPath(t *testing.T) {
	h, c := newTestHub(KindFlag)
	r, a, _ := h.Create(ModeChallenge, "flag", "", "Ana")
	r.Lock()
	defer r.Unlock()
	if a.Phase != PhaseReady {
		t.Fatal("challenge players start ready")
	}
	b, _ := r.Join("Ben", c.now())

	play := func(p *Player, option string) {
		if err := r.Start(p, c.now()); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < RoundsPerGame; i++ {
			v := r.View(p, "en", fakeLookup{}, c.now())
			if v.Phase != PhaseQuestion || v.Round != i+1 || v.Reveal != nil {
				t.Fatalf("round %d view: %+v", i, v)
			}
			c.add(2 * time.Second)
			if _, err := r.Answer(p, i, option, nil, c.now()); err != nil {
				t.Fatal(err)
			}
			if v := r.View(p, "en", fakeLookup{}, c.now()); v.Reveal == nil {
				t.Fatal("reveal missing after answer")
			}
			if err := r.Next(p, c.now()); err != nil {
				t.Fatal(err)
			}
		}
		if p.Phase != PhaseFinished {
			t.Fatal("should be finished")
		}
	}
	play(a, "AA")
	c.add(time.Hour) // Ben plays later
	play(b, "BB")

	v := r.View(b, "en", fakeLookup{}, c.now())
	if v.Players[0].Name != "Ana" || !v.Players[1].Finished || v.Players[1].Score != 0 {
		t.Fatalf("bad leaderboard: %+v", v.Players)
	}
}

func TestChallengeTimeout(t *testing.T) {
	h, c := newTestHub(KindFlag)
	r, a, _ := h.Create(ModeChallenge, "flag", "", "Ana")
	r.Lock()
	defer r.Unlock()
	_ = r.Start(a, c.now())
	c.add(RoundDuration(KindFlag) + answerGrace + time.Second)
	r.Tick(c.now())
	if a.Phase != PhaseReveal || !a.Answers[0].TimedOut {
		t.Fatal("timed-out round should reveal with no points")
	}
}

func TestColorRoundScoresCloseness(t *testing.T) {
	h, c := newTestHub(KindColor)
	r, a, _ := h.Create(ModeLive, "color", "", "Ana")
	r.Lock()
	defer r.Unlock()
	b, _ := r.Join("Ben", c.now())
	_ = r.Start(a, c.now())
	exact, err := r.Answer(a, 0, "", &ColorGuess{255, 0, 0}, c.now())
	if err != nil {
		t.Fatal(err)
	}
	far, _ := r.Answer(b, 0, "", &ColorGuess{0, 0, 255}, c.now())
	if exact.Points != pointsMax || far.Points >= exact.Points {
		t.Fatalf("exact %d far %d", exact.Points, far.Points)
	}
	if _, err := r.Answer(a, 0, "AA", nil, c.now()); err == nil {
		t.Fatal("option answer in a color round should fail")
	}
}

func TestNamesAreUniqueAndShort(t *testing.T) {
	h, c := newTestHub(KindFlag)
	r, _, _ := h.Create(ModeLive, "flag", "", "Ana")
	r.Lock()
	defer r.Unlock()
	p, _ := r.Join("ana", c.now())
	if p.Name != "ana 2" {
		t.Fatalf("got %q", p.Name)
	}
	p, _ = r.Join("a very very long player name", c.now())
	if len([]rune(p.Name)) > maxNameLen {
		t.Fatalf("got %q", p.Name)
	}
	for len(r.Players) < MaxLivePlayers {
		_, _ = r.Join("x", c.now())
	}
	if _, err := r.Join("late", c.now()); err != ErrFull {
		t.Fatalf("full room: %v", err)
	}
}

func TestSweepDropsIdleLiveRooms(t *testing.T) {
	h, c := newTestHub(KindFlag)
	r, _, _ := h.Create(ModeLive, "flag", "", "Ana")
	ch, _, _ := h.Create(ModeChallenge, "flag", "", "Ana")
	c.add(liveIdleTTL + time.Minute)
	h.Sweep()
	if h.Get(r.Code) != nil {
		t.Fatal("idle live room should be gone")
	}
	if h.Get(ch.Code) == nil {
		t.Fatal("challenge should stay")
	}
}

func TestChallengesSurviveRestart(t *testing.T) {
	t.Setenv(persistDirEnv, t.TempDir())
	h, c := newTestHub(KindFlag)
	h.UsePersistence()
	r, a, _ := h.Create(ModeChallenge, "flag", "", "Ana")
	r.Lock()
	_ = r.Start(a, c.now())
	_, _ = r.Answer(a, 0, "AA", nil, c.now())
	r.Unlock()
	h.MarkDirty()
	h.Save()

	h2, _ := newTestHub(KindFlag)
	h2.UsePersistence()
	r2 := h2.Get(strings.ToLower(r.Code))
	if r2 == nil {
		t.Fatal("challenge not restored")
	}
	r2.Lock()
	defer r2.Unlock()
	p := r2.PlayerByToken(a.Token)
	if p == nil || p.Score != a.Score || p.Phase != PhaseReveal {
		t.Fatalf("player not restored: %+v", p)
	}
}

// The real generator should make full games for every game type and region.
func TestDataSourceGeneratesEveryGame(t *testing.T) {
	src := NewDataSource()
	regions := []string{"", "Africa", "Americas", "Asia", "Europe", "Oceania"}
	games := append([]string{GameMix}, "flag", "shape", "capital", "compare", "color", "fact")
	for _, region := range regions {
		for _, game := range games {
			qs, err := src.Generate(game, region, RoundsPerGame)
			if err != nil {
				t.Fatalf("%s/%s: %v", game, region, err)
			}
			if len(qs) != RoundsPerGame {
				t.Fatalf("%s/%s: %d questions", game, region, len(qs))
			}
			seen := map[string]bool{}
			for _, q := range qs {
				switch q.Kind {
				case KindColor:
					if q.TargetHex == "" || q.GuessableID == "" {
						t.Fatalf("%s/%s: empty color round", game, region)
					}
				case KindCompare:
					if len(q.Options) != 2 || q.Values[q.Options[0]] == q.Values[q.Options[1]] {
						t.Fatalf("%s/%s: bad compare %+v", game, region, q)
					}
				default:
					if len(q.Options) != 4 {
						t.Fatalf("%s/%s: %s has %d options", game, region, q.Kind, len(q.Options))
					}
					if !contains(q.Options, q.Answer) {
						t.Fatalf("%s/%s: answer not in options", game, region)
					}
				}
				if q.Subject != "" {
					if seen[q.Subject] {
						t.Fatalf("%s/%s: %s used twice", game, region, q.Subject)
					}
					seen[q.Subject] = true
				}
			}
			if game == GameMix {
				for i := 1; i < len(qs); i++ {
					if qs[i].Kind == qs[i-1].Kind {
						t.Fatalf("mix repeats %s twice in a row", qs[i].Kind)
					}
				}
			}
		}
	}
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
