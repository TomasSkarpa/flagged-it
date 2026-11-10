package party

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"flagged-it/internal/games/flagcolor"
)

// Mode says how a room is played.
type Mode string

const (
	ModeLive      Mode = "live"      // everyone answers the same question at the same time
	ModeChallenge Mode = "challenge" // same questions, everyone plays at their own pace
)

// Phase of a live room, or of one player in a challenge.
type Phase string

const (
	PhaseLobby    Phase = "lobby"    // live: waiting for the host to start
	PhaseReady    Phase = "ready"    // challenge: joined, not started yet
	PhaseQuestion Phase = "question" // answering
	PhaseReveal   Phase = "reveal"   // showing the right answer
	PhaseFinished Phase = "finished"
)

const (
	RoundsPerGame       = 10
	MaxLivePlayers      = 10
	MaxChallengePlayers = 50
	maxNameLen          = 16

	revealDuration = 4 * time.Second
	answerGrace    = 1500 * time.Millisecond // network slack after the timer hits 0
	hostAwayAfter  = 10 * time.Second
	connectedFor   = 6 * time.Second // a player polling within this window counts as connected

	pointsMax = 1000
)

var (
	ErrNotFound        = errors.New("room not found")
	ErrNotPlayer       = errors.New("not a player in this room")
	ErrFull            = errors.New("room is full")
	ErrNotHost         = errors.New("only the host can do that")
	ErrWrongPhase      = errors.New("not possible right now")
	ErrWrongRound      = errors.New("that round is over")
	ErrAlreadyAnswered = errors.New("already answered")
	ErrBadAnswer       = errors.New("invalid answer")
)

// RoundDuration is how long players get to answer a round of kind k.
func RoundDuration(k Kind) time.Duration {
	switch k {
	case KindColor:
		return 30 * time.Second
	case KindFact:
		return 20 * time.Second
	}
	return 15 * time.Second
}

// Answer is one player's answer to one round.
type Answer struct {
	OptionID string `json:"optionId,omitempty"`
	Hex      string `json:"hex,omitempty"`
	Correct  bool   `json:"correct"`
	Points   int    `json:"points"`
	TimedOut bool   `json:"timedOut,omitempty"`
}

// Player in a room. Token is the player's secret; ID is safe to show others.
type Player struct {
	ID       string          `json:"id"`
	Token    string          `json:"token"`
	Name     string          `json:"name"`
	JoinedAt time.Time       `json:"joinedAt"`
	Score    int             `json:"score"`
	Answers  map[int]*Answer `json:"answers"` // by round index

	// Challenge mode: each player moves through the rounds on their own.
	Phase      Phase     `json:"phase,omitempty"`
	Round      int       `json:"round"`
	StartedAt  time.Time `json:"startedAt"` // when the current question was shown
	FinishedAt time.Time `json:"finishedAt"`

	conns    int       // open event streams
	lastSeen time.Time // last request from this player
}

// Room is one game for a group of players. All fields are guarded by mu.
type Room struct {
	mu sync.Mutex

	Code      string      `json:"code"`
	Mode      Mode        `json:"mode"`
	Game      string      `json:"game"`
	Region    string      `json:"region"`
	CreatedAt time.Time   `json:"createdAt"`
	Questions []*Question `json:"questions"`
	Players   []*Player   `json:"players"`
	HostID    string      `json:"hostId"`

	// Live mode only.
	Phase       Phase     `json:"phase"`
	Round       int       `json:"round"` // index into Questions
	PhaseStart  time.Time `json:"phaseStart"`
	PhaseEnd    time.Time `json:"phaseEnd"`
	LastChanged time.Time `json:"lastChanged"`

	version uint64
	changed chan struct{}
}

func newRoom(code string, mode Mode, game, region string, qs []*Question, now time.Time) *Room {
	r := &Room{
		Code: code, Mode: mode, Game: game, Region: region,
		CreatedAt: now, Questions: qs, LastChanged: now,
		changed: make(chan struct{}),
	}
	if mode == ModeLive {
		r.Phase = PhaseLobby
	}
	return r
}

// Lock and Unlock guard the room; callers of the methods below must hold it.
func (r *Room) Lock()   { r.mu.Lock() }
func (r *Room) Unlock() { r.mu.Unlock() }

// Changed returns a channel that is closed on the next change. Hold the lock.
func (r *Room) Changed() <-chan struct{} {
	if r.changed == nil {
		r.changed = make(chan struct{})
	}
	return r.changed
}

func (r *Room) notify(now time.Time) {
	r.version++
	r.LastChanged = now
	if r.changed != nil {
		close(r.changed)
	}
	r.changed = make(chan struct{})
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// cleanName trims a display name and makes it unique in the room.
func (r *Room) cleanName(name string) string {
	name = strings.Join(strings.Fields(name), " ")
	for utf8.RuneCountInString(name) > maxNameLen {
		_, size := utf8.DecodeLastRuneInString(name)
		name = name[:len(name)-size]
	}
	if name == "" {
		name = "Player " + strconv.Itoa(len(r.Players)+1)
	}
	base, n := name, 2
	for r.nameTaken(name) {
		name = base + " " + strconv.Itoa(n)
		n++
	}
	return name
}

func (r *Room) nameTaken(name string) bool {
	for _, p := range r.Players {
		if strings.EqualFold(p.Name, name) {
			return true
		}
	}
	return false
}

// PlayerByToken finds the player with this secret token.
func (r *Room) PlayerByToken(token string) *Player {
	if token == "" {
		return nil
	}
	for _, p := range r.Players {
		if p.Token == token {
			return p
		}
	}
	return nil
}

// Join adds a player. The first player becomes the host.
func (r *Room) Join(name string, now time.Time) (*Player, error) {
	limit := MaxLivePlayers
	if r.Mode == ModeChallenge {
		limit = MaxChallengePlayers
	}
	if len(r.Players) >= limit {
		return nil, ErrFull
	}
	p := &Player{
		ID: randomHex(4), Token: randomHex(16), Name: r.cleanName(name),
		JoinedAt: now, Answers: map[int]*Answer{}, lastSeen: now,
	}
	if r.Mode == ModeChallenge {
		p.Phase = PhaseReady
	}
	r.Players = append(r.Players, p)
	if r.HostID == "" {
		r.HostID = p.ID
	}
	r.notify(now)
	return p, nil
}

// Leave removes a player and hands the host role on if needed.
func (r *Room) Leave(p *Player, now time.Time) {
	for i, q := range r.Players {
		if q == p {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			break
		}
	}
	if r.HostID == p.ID {
		r.HostID = ""
		if len(r.Players) > 0 {
			r.HostID = r.Players[0].ID
		}
	}
	r.notify(now)
}

// Seen records activity from a player (polling or an action).
func (r *Room) Seen(p *Player, now time.Time) { p.lastSeen = now }

// Connect and Disconnect track open event streams.
func (r *Room) Connect(p *Player, now time.Time) {
	p.conns++
	p.lastSeen = now
	if p.conns == 1 {
		r.notify(now)
	}
}

func (r *Room) Disconnect(p *Player, now time.Time) {
	if p.conns > 0 {
		p.conns--
	}
	p.lastSeen = now
	if p.conns == 0 {
		r.notify(now)
	}
}

func (p *Player) connected(now time.Time) bool {
	return p.conns > 0 || now.Sub(p.lastSeen) < connectedFor
}

// Start begins the game: the host starts a live room; in a challenge each
// player starts their own run.
func (r *Room) Start(p *Player, now time.Time) error {
	if r.Mode == ModeChallenge {
		if p.Phase != PhaseReady {
			return ErrWrongPhase
		}
		r.startPlayerRound(p, 0, now)
		r.notify(now)
		return nil
	}
	if p.ID != r.HostID {
		return ErrNotHost
	}
	if r.Phase != PhaseLobby {
		return ErrWrongPhase
	}
	r.startRound(0, now)
	r.notify(now)
	return nil
}

func (r *Room) startRound(i int, now time.Time) {
	r.Round = i
	r.Phase = PhaseQuestion
	r.PhaseStart = now
	r.PhaseEnd = now.Add(RoundDuration(r.Questions[i].Kind))
}

func (r *Room) startPlayerRound(p *Player, i int, now time.Time) {
	p.Round = i
	p.Phase = PhaseQuestion
	p.StartedAt = now
}

// ColorGuess is an RGB guess for a color round.
type ColorGuess struct{ R, G, B int }

// Answer records p's answer to round (index). For color rounds pass color.
func (r *Room) Answer(p *Player, round int, optionID string, color *ColorGuess, now time.Time) (*Answer, error) {
	var start, end time.Time
	if r.Mode == ModeChallenge {
		if p.Phase != PhaseQuestion {
			return nil, ErrWrongPhase
		}
		if round != p.Round {
			return nil, ErrWrongRound
		}
		start = p.StartedAt
		end = start.Add(RoundDuration(r.Questions[round].Kind))
	} else {
		if r.Phase != PhaseQuestion {
			return nil, ErrWrongPhase
		}
		if round != r.Round {
			return nil, ErrWrongRound
		}
		start, end = r.PhaseStart, r.PhaseEnd
	}
	if _, done := p.Answers[round]; done {
		return nil, ErrAlreadyAnswered
	}
	if now.After(end.Add(answerGrace)) {
		return nil, ErrWrongRound
	}

	a, err := score(r.Questions[round], optionID, color, timeLeft(start, end, now))
	if err != nil {
		return nil, err
	}
	p.Answers[round] = a
	p.Score += a.Points
	p.lastSeen = now

	if r.Mode == ModeChallenge {
		p.Phase = PhaseReveal
	} else if r.allAnswered(now) {
		r.toReveal(now)
	}
	r.notify(now)
	return a, nil
}

// timeLeft is the share of the round still left, 0..1.
func timeLeft(start, end, now time.Time) float64 {
	total := end.Sub(start)
	if total <= 0 {
		return 0
	}
	f := float64(end.Sub(now)) / float64(total)
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// score: a right choice earns 500 plus up to 500 for speed. A colour round
// earns up to 1000 for closeness (speed doesn't count there).
func score(q *Question, optionID string, color *ColorGuess, left float64) (*Answer, error) {
	if q.Kind == KindColor {
		if color == nil {
			return nil, ErrBadAnswer
		}
		hexGuess, err := flagcolor.GuessHexFromRGB(color.R, color.G, color.B)
		if err != nil {
			return nil, ErrBadAnswer
		}
		pts := flagcolor.PointsFromGuessHex(q.TargetHex, hexGuess)
		points := pts * pointsMax / flagcolor.PointsMaxPerRound
		return &Answer{Hex: hexGuess, Points: points, Correct: points >= pointsMax*7/10}, nil
	}
	valid := false
	for _, o := range q.Options {
		if o == optionID {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrBadAnswer
	}
	a := &Answer{OptionID: optionID, Correct: optionID == q.Answer}
	if a.Correct {
		a.Points = pointsMax/2 + int(float64(pointsMax/2)*left+0.5)
	}
	return a, nil
}

func (r *Room) allAnswered(now time.Time) bool {
	active := 0
	for _, p := range r.Players {
		if !p.connected(now) {
			continue
		}
		active++
		if _, ok := p.Answers[r.Round]; !ok {
			return false
		}
	}
	return active > 0
}

func (r *Room) toReveal(now time.Time) {
	r.Phase = PhaseReveal
	r.PhaseStart = now
	r.PhaseEnd = now.Add(revealDuration)
}

// Next moves a challenge player on from the reveal.
func (r *Room) Next(p *Player, now time.Time) error {
	if r.Mode != ModeChallenge || p.Phase != PhaseReveal {
		return ErrWrongPhase
	}
	if p.Round+1 >= len(r.Questions) {
		p.Phase = PhaseFinished
		p.FinishedAt = now
	} else {
		r.startPlayerRound(p, p.Round+1, now)
	}
	r.notify(now)
	return nil
}

// Again resets a finished live room to the lobby with new questions.
func (r *Room) Again(p *Player, qs []*Question, now time.Time) error {
	if r.Mode != ModeLive {
		return ErrWrongPhase
	}
	if p.ID != r.HostID {
		return ErrNotHost
	}
	if r.Phase != PhaseFinished {
		return ErrWrongPhase
	}
	r.Questions = qs
	r.Phase = PhaseLobby
	r.Round = 0
	for _, pl := range r.Players {
		pl.Score = 0
		pl.Answers = map[int]*Answer{}
	}
	r.notify(now)
	return nil
}

// Tick moves timed phases along. It returns true if anything changed.
func (r *Room) Tick(now time.Time) bool {
	changed := false
	if r.Mode == ModeChallenge {
		for _, p := range r.Players {
			if p.Phase != PhaseQuestion {
				continue
			}
			end := p.StartedAt.Add(RoundDuration(r.Questions[p.Round].Kind)).Add(answerGrace)
			if now.After(end) {
				if _, ok := p.Answers[p.Round]; !ok {
					p.Answers[p.Round] = &Answer{TimedOut: true}
				}
				p.Phase = PhaseReveal
				changed = true
			}
		}
	} else {
		switch r.Phase {
		case PhaseQuestion:
			if now.After(r.PhaseEnd) || r.allAnswered(now) {
				r.toReveal(now)
				changed = true
			}
		case PhaseReveal:
			if now.After(r.PhaseEnd) {
				if r.Round+1 >= len(r.Questions) {
					r.Phase = PhaseFinished
				} else {
					r.startRound(r.Round+1, now)
				}
				changed = true
			}
		}
		if r.handOverHost(now) {
			changed = true
		}
	}
	if changed {
		r.notify(now)
	}
	return changed
}

// handOverHost gives the host role to the longest-present connected player
// when the host has been away for a while.
func (r *Room) handOverHost(now time.Time) bool {
	var host *Player
	for _, p := range r.Players {
		if p.ID == r.HostID {
			host = p
		}
	}
	if host != nil && (host.conns > 0 || now.Sub(host.lastSeen) < hostAwayAfter) {
		return false
	}
	for _, p := range r.Players {
		if p != host && p.connected(now) {
			r.HostID = p.ID
			return true
		}
	}
	return false
}

// anyoneConnected reports whether any player is connected.
func (r *Room) anyoneConnected(now time.Time) bool {
	for _, p := range r.Players {
		if p.connected(now) {
			return true
		}
	}
	return false
}

// ranked returns players by score, best first (ties: who joined first).
func (r *Room) ranked() []*Player {
	ps := append([]*Player(nil), r.Players...)
	sort.SliceStable(ps, func(i, j int) bool { return ps[i].Score > ps[j].Score })
	return ps
}

// VisibleQuestion returns round i (0-based) if player p has already reached it.
// Used to serve round media (flag image, shape) without exposing future rounds.
func (r *Room) VisibleQuestion(p *Player, i int) (*Question, bool) {
	if i < 0 || i >= len(r.Questions) {
		return nil, false
	}
	if r.Mode == ModeChallenge {
		if p.Phase == PhaseReady || i > p.Round {
			return nil, false
		}
	} else if r.Phase == PhaseLobby || i > r.Round {
		return nil, false
	}
	return r.Questions[i], true
}
