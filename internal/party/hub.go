package party

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ" // no I or O: easy to read out loud
	codeLen      = 4
	maxRooms     = 2000

	liveIdleTTL    = 15 * time.Minute // live room with nobody around
	liveMaxAge     = 6 * time.Hour
	challengeTTL   = 7 * 24 * time.Hour
	tickEvery      = 200 * time.Millisecond
	sweepEvery     = time.Minute
	persistEvery   = 10 * time.Second
	persistFile    = "party-challenges.json"
	persistDirEnv  = "FLAGGED_IT_DATA_DIR"
	defaultDataDir = "/opt/flagged-it/data"
)

var ErrTooManyRooms = errors.New("too many rooms, try again later")

// Hub holds all rooms.
type Hub struct {
	mu    sync.Mutex
	rooms map[string]*Room

	src    Source
	lookup Lookup
	now    func() time.Time

	persistPath string
	dirty       bool
}

// NewHub creates a hub. lookup turns CCA2 codes into country names.
func NewHub(src Source, lookup Lookup) *Hub {
	return &Hub{rooms: map[string]*Room{}, src: src, lookup: lookup, now: time.Now}
}

// Lookup returns the country lookup used for views.
func (h *Hub) Lookup() Lookup { return h.lookup }

// Now returns the hub's clock.
func (h *Hub) Now() time.Time { return h.now() }

func newCode() string {
	var b strings.Builder
	for i := 0; i < codeLen; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		b.WriteByte(codeAlphabet[n.Int64()])
	}
	return b.String()
}

// NormalizeCode upper-cases and trims a code typed by a player.
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// Create makes a room and adds its creator as the host.
func (h *Hub) Create(mode Mode, game, region, name string) (*Room, *Player, error) {
	if mode != ModeLive && mode != ModeChallenge {
		mode = ModeLive
	}
	if !ValidGame(game) {
		game = GameMix
	}
	qs, err := h.src.Generate(game, region, RoundsPerGame)
	if err != nil {
		return nil, nil, err
	}
	now := h.now()

	h.mu.Lock()
	if len(h.rooms) >= maxRooms {
		h.mu.Unlock()
		return nil, nil, ErrTooManyRooms
	}
	code := newCode()
	for h.rooms[code] != nil {
		code = newCode()
	}
	r := newRoom(code, mode, game, region, qs, now)
	h.rooms[code] = r
	if mode == ModeChallenge {
		h.dirty = true
	}
	h.mu.Unlock()

	r.Lock()
	p, err := r.Join(name, now)
	r.Unlock()
	return r, p, err
}

// Get returns the room with this code, or nil.
func (h *Hub) Get(code string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.rooms[NormalizeCode(code)]
}

// NewQuestions makes a fresh set of questions for a room's game.
func (h *Hub) NewQuestions(r *Room) ([]*Question, error) {
	return h.src.Generate(r.Game, r.Region, RoundsPerGame)
}

// MarkDirty asks for challenge rooms to be saved soon.
func (h *Hub) MarkDirty() {
	h.mu.Lock()
	h.dirty = true
	h.mu.Unlock()
}

func (h *Hub) snapshot() []*Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	rs := make([]*Room, 0, len(h.rooms))
	for _, r := range h.rooms {
		rs = append(rs, r)
	}
	return rs
}

// Tick advances timers in every room.
func (h *Hub) Tick() {
	now := h.now()
	for _, r := range h.snapshot() {
		r.Lock()
		if r.Tick(now) && r.Mode == ModeChallenge {
			h.MarkDirty()
		}
		r.Unlock()
	}
}

// Sweep removes rooms nobody will come back to.
func (h *Hub) Sweep() int {
	now := h.now()
	var drop []string
	for _, r := range h.snapshot() {
		r.Lock()
		expired := false
		if r.Mode == ModeChallenge {
			expired = now.Sub(r.CreatedAt) > challengeTTL
		} else {
			expired = now.Sub(r.CreatedAt) > liveMaxAge ||
				(!r.anyoneConnected(now) && now.Sub(r.LastChanged) > liveIdleTTL)
		}
		if expired {
			drop = append(drop, r.Code)
		}
		r.Unlock()
	}
	if len(drop) > 0 {
		h.mu.Lock()
		for _, c := range drop {
			if r := h.rooms[c]; r != nil && r.Mode == ModeChallenge {
				h.dirty = true
			}
			delete(h.rooms, c)
		}
		h.mu.Unlock()
	}
	return len(drop)
}

// Run ticks rooms, sweeps old ones and saves challenges until stop is closed.
func (h *Hub) Run(stop <-chan struct{}) {
	tick := time.NewTicker(tickEvery)
	sweep := time.NewTicker(sweepEvery)
	save := time.NewTicker(persistEvery)
	defer tick.Stop()
	defer sweep.Stop()
	defer save.Stop()
	for {
		select {
		case <-stop:
			h.Save()
			return
		case <-tick.C:
			h.Tick()
		case <-sweep.C:
			if n := h.Sweep(); n > 0 {
				log.Printf("party: removed %d old rooms", n)
			}
		case <-save.C:
			h.Save()
		}
	}
}

// UsePersistence picks a writable folder for challenge rooms and loads any
// saved ones. Without a writable folder challenges live in memory only.
func (h *Hub) UsePersistence() {
	dirs := []string{os.Getenv(persistDirEnv), defaultDataDir}
	if d, err := os.UserCacheDir(); err == nil {
		dirs = append(dirs, filepath.Join(d, "flagged-it"))
	}
	for _, d := range dirs {
		if d == "" {
			continue
		}
		if err := os.MkdirAll(d, 0o755); err != nil {
			continue
		}
		probe := filepath.Join(d, ".write-test")
		if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
			continue
		}
		_ = os.Remove(probe)
		h.persistPath = filepath.Join(d, persistFile)
		log.Printf("party: saving challenges to %s", h.persistPath)
		h.load()
		return
	}
	log.Printf("party: no writable data folder, challenges are kept in memory only")
}

func (h *Hub) load() {
	b, err := os.ReadFile(h.persistPath)
	if err != nil {
		return
	}
	var rooms []*Room
	if err := json.Unmarshal(b, &rooms); err != nil {
		log.Printf("party: could not read saved challenges: %v", err)
		return
	}
	now := h.now()
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, r := range rooms {
		if r.Mode != ModeChallenge || now.Sub(r.CreatedAt) > challengeTTL || len(r.Questions) == 0 {
			continue
		}
		r.changed = make(chan struct{})
		for _, p := range r.Players {
			if p.Answers == nil {
				p.Answers = map[int]*Answer{}
			}
		}
		h.rooms[r.Code] = r
	}
	log.Printf("party: loaded %d saved challenges", len(h.rooms))
}

// Save writes challenge rooms to disk if anything changed.
func (h *Hub) Save() {
	h.mu.Lock()
	if h.persistPath == "" || !h.dirty {
		h.mu.Unlock()
		return
	}
	h.dirty = false
	var rooms []*Room
	for _, r := range h.rooms {
		if r.Mode == ModeChallenge {
			rooms = append(rooms, r)
		}
	}
	path := h.persistPath
	h.mu.Unlock()

	// Marshal each room under its own lock.
	parts := make([]json.RawMessage, 0, len(rooms))
	for _, r := range rooms {
		r.Lock()
		b, err := json.Marshal(r)
		r.Unlock()
		if err == nil {
			parts = append(parts, b)
		}
	}
	b, err := json.Marshal(parts)
	if err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		log.Printf("party: save failed: %v", err)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		log.Printf("party: save failed: %v", err)
	}
}
