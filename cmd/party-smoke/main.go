// party-smoke plays a full live game (2 players) and a full challenge
// (2 players) against a running API, and fails loudly if anything is off.
//
//	go run ./cmd/party-smoke -api https://api.flaggedit.app
//	go run ./cmd/party-smoke -api http://localhost:8080 -game flag
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	api    = flag.String("api", "https://api.flaggedit.app", "API base URL (without /api)")
	game   = flag.String("game", "mix", "game: mix, flag, shape, capital, compare, color, fact")
	region = flag.String("region", "", "region, empty for the whole world")
	client = &http.Client{Timeout: 20 * time.Second}
)

type view struct {
	Code     string `json:"code"`
	Mode     string `json:"mode"`
	Phase    string `json:"phase"`
	Round    int    `json:"round"`
	Rounds   int    `json:"rounds"`
	Question *struct {
		Kind    string `json:"kind"`
		Options []struct {
			ID string `json:"id"`
		} `json:"options"`
	} `json:"question"`
	Reveal *struct {
		AnswerID string `json:"answerId"`
	} `json:"reveal"`
	You struct {
		Score int `json:"score"`
	} `json:"you"`
	Players []struct {
		Name     string `json:"name"`
		Score    int    `json:"score"`
		Finished bool   `json:"finished"`
	} `json:"players"`
}

type seat struct {
	Token string `json:"token"`
	View  view   `json:"view"`
}

func fail(format string, args ...any) {
	fmt.Printf("FAIL: "+format+"\n", args...)
	os.Exit(1)
}

func ok(format string, args ...any) { fmt.Printf("ok   "+format+"\n", args...) }

func base(path string) string { return strings.TrimRight(*api, "/") + "/api/party" + path }

func q(code, token string) string {
	return "?code=" + code + "&t=" + url.QueryEscape(token) + "&locale=en"
}

func do(method, u string, body any, out any) int {
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, u, r)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := client.Do(req)
	if err != nil {
		fail("%s %s: %v", method, u, err)
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if out != nil && res.StatusCode == 200 {
		if err := json.Unmarshal(data, out); err != nil {
			fail("%s %s: bad JSON: %v", method, u, err)
		}
	}
	return res.StatusCode
}

func create(mode, name string) seat {
	var s seat
	if c := do("POST", base("/create"), map[string]string{"mode": mode, "game": *game, "region": *region, "name": name}, &s); c != 200 {
		fail("create %s: HTTP %d", mode, c)
	}
	return s
}

func join(code, name string) seat {
	var s seat
	if c := do("POST", base("/join"), map[string]string{"code": code, "name": name}, &s); c != 200 {
		fail("join %s: HTTP %d", code, c)
	}
	return s
}

func state(code, token string) view {
	var v view
	if c := do("GET", base("/state"+q(code, token)), nil, &v); c != 200 {
		fail("state: HTTP %d", c)
	}
	return v
}

func answer(code, token string, v view) {
	body := map[string]any{"round": v.Round}
	if v.Question.Kind == "color" {
		body["r"], body["g"], body["b"] = 200, 30, 40
	} else {
		body["optionId"] = v.Question.Options[0].ID
	}
	if c := do("POST", base("/answer"+q(code, token)), body, nil); c != 200 {
		fail("answer round %d (%s): HTTP %d", v.Round, v.Question.Kind, c)
	}
}

// checkMedia fetches the round picture for flag and shape rounds.
func checkMedia(code, token string, v view) {
	kind := v.Question.Kind
	if kind != "flag" && kind != "shape" {
		return
	}
	res, err := client.Get(base("/" + kind + q(code, token) + fmt.Sprintf("&round=%d", v.Round)))
	if err != nil {
		fail("%s image: %v", kind, err)
	}
	b, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 || len(b) < 50 {
		fail("%s image round %d: HTTP %d (%d bytes)", kind, v.Round, res.StatusCode, len(b))
	}
	if kind == "shape" && bytes.Contains(b, []byte(`"name"`)) {
		fail("shape round %d leaks a country name", v.Round)
	}
}

// listen reads the event stream and reports each state's phase:round.
func listen(code, token string) (<-chan string, func()) {
	out := make(chan string, 256)
	req, _ := http.NewRequest("GET", base("/events"+q(code, token)), nil)
	streamClient := &http.Client{} // no timeout: the stream stays open
	res, err := streamClient.Do(req)
	if err != nil {
		fail("events: %v", err)
	}
	if res.StatusCode != 200 || !strings.HasPrefix(res.Header.Get("Content-Type"), "text/event-stream") {
		fail("events: HTTP %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	go func() {
		sc := bufio.NewScanner(res.Body)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "data: ") {
				var v view
				if json.Unmarshal([]byte(line[6:]), &v) == nil && v.Code != "" {
					out <- fmt.Sprintf("%s:%d", v.Phase, v.Round)
				}
			}
		}
		close(out)
	}()
	return out, func() { res.Body.Close() }
}

func waitEvent(events <-chan string, want string, within time.Duration) {
	deadline := time.After(within)
	for {
		select {
		case e, open := <-events:
			if !open {
				fail("event stream closed while waiting for %s", want)
			}
			if e == want {
				return
			}
		case <-deadline:
			fail("no %s event within %s (stream buffered by a proxy?)", want, within)
		}
	}
}

func liveGame() {
	start := time.Now()
	host := create("live", "Smoke Host")
	code := host.View.Code
	ok("live: created room %s (%s)", code, *game)
	guest := join(strings.ToLower(code), "Smoke Guest")
	ok("live: guest joined with a lower-case code")

	events, stop := listen(code, guest.Token)
	defer stop()
	waitEvent(events, "lobby:0", 5*time.Second)
	ok("live: event stream delivers the lobby")

	if c := do("POST", base("/start"+q(code, guest.Token)), nil, nil); c != http.StatusForbidden {
		fail("live: guest could start the game (HTTP %d)", c)
	}
	var v view
	if c := do("POST", base("/start"+q(code, host.Token)), nil, &v); c != 200 || v.Phase != "question" {
		fail("live: start: HTTP %d phase %s", c, v.Phase)
	}
	kinds := []string{}
	for round := 1; round <= v.Rounds; round++ {
		waitEvent(events, fmt.Sprintf("question:%d", round), 25*time.Second)
		hv := state(code, host.Token)
		if hv.Reveal != nil {
			fail("live: round %d shows the answer before anyone answered", round)
		}
		kinds = append(kinds, hv.Question.Kind)
		checkMedia(code, guest.Token, hv)
		answer(code, host.Token, hv)
		answer(code, guest.Token, hv)
		waitEvent(events, fmt.Sprintf("reveal:%d", round), 5*time.Second)
	}
	waitEvent(events, fmt.Sprintf("finished:%d", v.Rounds), 10*time.Second)
	final := state(code, host.Token)
	if final.Phase != "finished" || len(final.Players) != 2 {
		fail("live: final state %+v", final)
	}
	ok("live: %d rounds %v, final %s %d / %s %d (%s)", v.Rounds, kinds,
		final.Players[0].Name, final.Players[0].Score, final.Players[1].Name, final.Players[1].Score,
		time.Since(start).Round(time.Second))

	do("POST", base("/leave"+q(code, guest.Token)), nil, nil)
	do("POST", base("/leave"+q(code, host.Token)), nil, nil)
}

func playChallenge(code, token string) view {
	var v view
	if c := do("POST", base("/start"+q(code, token)), nil, &v); c != 200 {
		fail("challenge: start: HTTP %d", c)
	}
	for v.Phase != "finished" {
		if v.Phase != "question" {
			fail("challenge: expected a question, got %s", v.Phase)
		}
		checkMedia(code, token, v)
		answer(code, token, v)
		after := state(code, token)
		if after.Phase != "reveal" || after.Reveal == nil {
			fail("challenge: no reveal after answering round %d", v.Round)
		}
		if c := do("POST", base("/next"+q(code, token)), nil, &v); c != 200 {
			fail("challenge: next: HTTP %d", c)
		}
	}
	return v
}

func challenge() {
	me := create("challenge", "Smoke A")
	code := me.View.Code
	if me.View.Phase != "ready" {
		fail("challenge: new challenge is %s, want ready", me.View.Phase)
	}
	a := playChallenge(code, me.Token)
	ok("challenge %s: first player finished with %d", code, a.You.Score)
	friend := join(code, "Smoke B")
	b := playChallenge(code, friend.Token)
	done := 0
	for _, p := range b.Players {
		if p.Finished {
			done++
		}
	}
	if done != 2 {
		fail("challenge: leaderboard shows %d finished players, want 2", done)
	}
	ok("challenge %s: second player finished with %d, leaderboard has both", code, b.You.Score)
}

func main() {
	flag.Parse()
	fmt.Printf("party smoke against %s\n", *api)
	if c := do("GET", strings.TrimRight(*api, "/")+"/api/health", nil, nil); c != 200 {
		fail("health: HTTP %d", c)
	}
	ok("health")
	if c := do("GET", base("/state?code=ZZZZ&t=x"), nil, nil); c != http.StatusNotFound {
		fail("unknown room: HTTP %d, want 404", c)
	}
	ok("unknown room is 404")
	liveGame()
	challenge()
	fmt.Println("PASS")
}
