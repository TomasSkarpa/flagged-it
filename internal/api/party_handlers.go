package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"flagged-it/internal/data"
	"flagged-it/internal/games/flagcolor"
	"flagged-it/internal/party"
)

// Multiplayer ("party") rooms: live rooms and async challenges.
// Players authenticate with the secret token they got on create/join, sent as
// the `t` query parameter (EventSource can't send headers).

var (
	partyData *party.DataSource
	partyHub  *party.Hub
	// Keyed by player token, so a class on one Wi-Fi doesn't share a limit.
	partyLimiter = NewRateLimiter(100*time.Millisecond, 40)
)

const partyHeartbeat = 15 * time.Second

func setupPartyRoutes(cors func(http.HandlerFunc) http.HandlerFunc) {
	partyData = party.NewDataSource()
	partyHub = party.NewHub(partyData, partyData)
	partyHub.UsePersistence()
	go partyHub.Run(nil)

	mw := func(h http.HandlerFunc) http.HandlerFunc {
		return LoggingMiddleware(cors(partyRateLimit(h)))
	}
	http.HandleFunc("/api/party/create", mw(partyCreate))
	http.HandleFunc("/api/party/join", mw(partyJoin))
	http.HandleFunc("/api/party/state", mw(partyState))
	http.HandleFunc("/api/party/events", mw(partyEvents))
	http.HandleFunc("/api/party/start", mw(partyAction(func(r *party.Room, p *party.Player, _ *http.Request, now time.Time) error {
		return r.Start(p, now)
	})))
	http.HandleFunc("/api/party/answer", mw(partyAnswer))
	http.HandleFunc("/api/party/next", mw(partyAction(func(r *party.Room, p *party.Player, _ *http.Request, now time.Time) error {
		return r.Next(p, now)
	})))
	http.HandleFunc("/api/party/again", mw(partyAgain))
	http.HandleFunc("/api/party/leave", mw(partyLeave))
	http.HandleFunc("/api/party/flag", mw(partyFlag))
	http.HandleFunc("/api/party/shape", mw(partyShape))
}

func partyRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		key := r.URL.Query().Get("t")
		if key == "" {
			key = "ip:" + getIP(r)
		}
		if !partyLimiter.Allow(key) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func partyLocale(r *http.Request) string {
	l := r.URL.Query().Get("locale")
	if l == "" || len(l) > 5 {
		return "en"
	}
	return l
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// partyError maps engine errors to HTTP codes the frontend can show.
func partyError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, party.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, party.ErrNotPlayer), errors.Is(err, party.ErrNotHost):
		status = http.StatusForbidden
	case errors.Is(err, party.ErrFull), errors.Is(err, party.ErrWrongPhase),
		errors.Is(err, party.ErrWrongRound), errors.Is(err, party.ErrAlreadyAnswered):
		status = http.StatusConflict
	case errors.Is(err, party.ErrTooManyRooms):
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

type partyJoinResponse struct {
	Token string     `json:"token"`
	View  party.View `json:"view"`
}

func partyCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Mode   string `json:"mode"`
		Game   string `json:"game"`
		Region string `json:"region"`
		Name   string `json:"name"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		partyError(w, party.ErrBadAnswer)
		return
	}
	room, p, err := partyHub.Create(party.Mode(req.Mode), req.Game, req.Region, req.Name)
	if err != nil {
		log.Printf("party: create failed: %v", err)
		partyError(w, err)
		return
	}
	room.Lock()
	v := room.View(p, partyLocale(r), partyHub.Lookup(), partyHub.Now())
	room.Unlock()
	writeJSON(w, http.StatusOK, partyJoinResponse{Token: p.Token, View: v})
}

func partyJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		partyError(w, party.ErrBadAnswer)
		return
	}
	room := partyHub.Get(req.Code)
	if room == nil {
		partyError(w, party.ErrNotFound)
		return
	}
	now := partyHub.Now()
	room.Lock()
	p, err := room.Join(req.Name, now)
	var v party.View
	if err == nil {
		v = room.View(p, partyLocale(r), partyHub.Lookup(), now)
	}
	mode := room.Mode
	room.Unlock()
	if err != nil {
		partyError(w, err)
		return
	}
	if mode == party.ModeChallenge {
		partyHub.MarkDirty()
	}
	writeJSON(w, http.StatusOK, partyJoinResponse{Token: p.Token, View: v})
}

// withPlayer finds the room and player from ?code=&t= and runs fn with the
// room locked. It returns false (and has written the error) on failure.
func withPlayer(w http.ResponseWriter, r *http.Request, fn func(room *party.Room, p *party.Player, now time.Time)) bool {
	q := r.URL.Query()
	room := partyHub.Get(q.Get("code"))
	if room == nil {
		partyError(w, party.ErrNotFound)
		return false
	}
	room.Lock()
	defer room.Unlock()
	p := room.PlayerByToken(q.Get("t"))
	if p == nil {
		partyError(w, party.ErrNotPlayer)
		return false
	}
	now := partyHub.Now()
	room.Seen(p, now)
	fn(room, p, now)
	return true
}

func partyState(w http.ResponseWriter, r *http.Request) {
	var v party.View
	if withPlayer(w, r, func(room *party.Room, p *party.Player, now time.Time) {
		v = room.View(p, partyLocale(r), partyHub.Lookup(), now)
	}) {
		writeJSON(w, http.StatusOK, v)
	}
}

// partyAction wraps a POST action that returns the player's fresh view.
func partyAction(act func(*party.Room, *party.Player, *http.Request, time.Time) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var (
			v         party.View
			err       error
			challenge bool
		)
		if !withPlayer(w, r, func(room *party.Room, p *party.Player, now time.Time) {
			if err = act(room, p, r, now); err == nil {
				v = room.View(p, partyLocale(r), partyHub.Lookup(), now)
			}
			challenge = room.Mode == party.ModeChallenge
		}) {
			return
		}
		if err != nil {
			partyError(w, err)
			return
		}
		if challenge {
			partyHub.MarkDirty()
		}
		writeJSON(w, http.StatusOK, v)
	}
}

func partyAnswer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Round    int    `json:"round"` // 1-based, as shown
		OptionID string `json:"optionId"`
		R        *int   `json:"r"`
		G        *int   `json:"g"`
		B        *int   `json:"b"`
	}
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
			partyError(w, party.ErrBadAnswer)
			return
		}
	}
	partyAction(func(room *party.Room, p *party.Player, _ *http.Request, now time.Time) error {
		var color *party.ColorGuess
		if req.R != nil && req.G != nil && req.B != nil {
			color = &party.ColorGuess{R: *req.R, G: *req.G, B: *req.B}
		}
		_, err := room.Answer(p, req.Round-1, req.OptionID, color, now)
		return err
	})(w, r)
}

func partyAgain(w http.ResponseWriter, r *http.Request) {
	// Build the new questions before taking the room lock.
	room := partyHub.Get(r.URL.Query().Get("code"))
	if room == nil {
		partyError(w, party.ErrNotFound)
		return
	}
	qs, err := partyHub.NewQuestions(room)
	if err != nil {
		partyError(w, err)
		return
	}
	partyAction(func(room *party.Room, p *party.Player, _ *http.Request, now time.Time) error {
		return room.Again(p, qs, now)
	})(w, r)
}

func partyLeave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if withPlayer(w, r, func(room *party.Room, p *party.Player, now time.Time) {
		room.Leave(p, now)
	}) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// partyEvents streams the player's view as server-sent events: one message
// per change, plus a comment every 15s so proxies keep the line open.
func partyEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	token, locale := q.Get("t"), partyLocale(r)
	room := partyHub.Get(q.Get("code"))
	if room == nil {
		partyError(w, party.ErrNotFound)
		return
	}
	room.Lock()
	p := room.PlayerByToken(token)
	if p != nil {
		room.Connect(p, partyHub.Now())
	}
	room.Unlock()
	if p == nil {
		partyError(w, party.ErrNotPlayer)
		return
	}
	defer func() {
		room.Lock()
		room.Disconnect(p, partyHub.Now())
		room.Unlock()
	}()

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache, no-transform")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no") // nginx: don't buffer this response
	w.WriteHeader(http.StatusOK)
	rc := http.NewResponseController(w)

	heartbeat := time.NewTicker(partyHeartbeat)
	defer heartbeat.Stop()
	ctx := r.Context()

	for {
		room.Lock()
		member := room.PlayerByToken(token) != nil
		var v party.View
		if member {
			now := partyHub.Now()
			room.Seen(p, now)
			v = room.View(p, locale, partyHub.Lookup(), now)
		}
		changed := room.Changed()
		room.Unlock()

		if !member {
			fmt.Fprint(w, "event: gone\ndata: {}\n\n")
			_ = rc.Flush()
			return
		}
		b, _ := json.Marshal(v)
		if _, err := fmt.Fprintf(w, "data: %s\n\n", b); err != nil {
			return
		}
		if err := rc.Flush(); err != nil {
			return
		}

	wait:
		for {
			select {
			case <-ctx.Done():
				return
			case <-changed:
				break wait
			case <-heartbeat.C:
				// A real event (not a comment) so the client can tell the stream is alive.
				if _, err := fmt.Fprint(w, "event: ping\ndata: {}\n\n"); err != nil {
					return
				}
				if err := rc.Flush(); err != nil {
					return
				}
				room.Lock()
				room.Seen(p, partyHub.Now())
				room.Unlock()
			}
		}
	}
}

// roundQuestion returns the question for ?round= (1-based) if the player has reached it.
func roundQuestion(w http.ResponseWriter, r *http.Request, kind party.Kind) (*party.Question, bool) {
	round, err := strconv.Atoi(r.URL.Query().Get("round"))
	if err != nil {
		http.Error(w, "round required", http.StatusBadRequest)
		return nil, false
	}
	var (
		q  *party.Question
		ok bool
	)
	if !withPlayer(w, r, func(room *party.Room, p *party.Player, _ time.Time) {
		q, ok = room.VisibleQuestion(p, round-1)
	}) {
		return nil, false
	}
	if !ok || q.Kind != kind {
		http.Error(w, "not available", http.StatusNotFound)
		return nil, false
	}
	return q, true
}

// partyFlag serves the flag for a flag round without its country code in the URL.
func partyFlag(w http.ResponseWriter, r *http.Request) {
	q, ok := roundQuestion(w, r, party.KindFlag)
	if !ok {
		return
	}
	svg, err := flagcolor.ReadSVGBytes(q.Subject)
	if err != nil {
		http.Error(w, "flag not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	_, _ = w.Write(svg)
}

// partyShape serves the outline for a shape round with names and ids removed.
func partyShape(w http.ResponseWriter, r *http.Request) {
	q, ok := roundQuestion(w, r, party.KindShape)
	if !ok {
		return
	}
	c, found := partyData.Country(q.Subject)
	if !found {
		http.Error(w, "shape not found", http.StatusNotFound)
		return
	}
	geo, err := data.LoadGeoData(c.CCA3)
	if err != nil {
		http.Error(w, "shape not found", http.StatusNotFound)
		return
	}
	features := make([]map[string]any, 0, len(geo.Features))
	for _, f := range geo.Features {
		features = append(features, map[string]any{
			"type":       f.Type,
			"properties": map[string]any{},
			"geometry":   f.Geometry,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"type": strings.TrimSpace(geo.Type), "features": features})
}
