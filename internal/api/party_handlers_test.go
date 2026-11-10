package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

var partySetupOnce sync.Once

func partyServer(t *testing.T) *httptest.Server {
	t.Helper()
	t.Setenv("FLAGGED_IT_DATA_DIR", t.TempDir())
	partySetupOnce.Do(func() {
		setupPartyRoutes(func(h http.HandlerFunc) http.HandlerFunc { return h })
	})
	srv := httptest.NewServer(http.DefaultServeMux)
	t.Cleanup(srv.Close)
	return srv
}

type testView struct {
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
	Players []struct {
		Name  string `json:"name"`
		Score int    `json:"score"`
	} `json:"players"`
}

func post(t *testing.T, u string, body any, out any) int {
	t.Helper()
	b, _ := json.Marshal(body)
	res, err := http.Post(u, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if out != nil && res.StatusCode == http.StatusOK {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			t.Fatal(err)
		}
	}
	return res.StatusCode
}

func TestPartyHTTPLiveFlow(t *testing.T) {
	srv := partyServer(t)
	api := srv.URL + "/api/party"

	var host, guest struct {
		Token string   `json:"token"`
		View  testView `json:"view"`
	}
	if c := post(t, api+"/create", map[string]string{"mode": "live", "game": "flag", "name": "Ana"}, &host); c != 200 {
		t.Fatalf("create: %d", c)
	}
	code := host.View.Code
	if c := post(t, api+"/join", map[string]string{"code": strings.ToLower(code), "name": "Ben"}, &guest); c != 200 {
		t.Fatalf("join: %d", c)
	}
	q := func(token string) string { return "?code=" + code + "&t=" + url.QueryEscape(token) + "&locale=en" }

	// Guest listens to the event stream.
	res, err := http.Get(api + "/events" + q(guest.Token))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if ct := res.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type %q", ct)
	}
	events := make(chan testView, 32)
	go func() {
		sc := bufio.NewScanner(res.Body)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "data: ") {
				var v testView
				if json.Unmarshal([]byte(line[6:]), &v) == nil {
					events <- v
				}
			}
		}
	}()
	waitFor := func(pred func(testView) bool) testView {
		t.Helper()
		deadline := time.After(10 * time.Second)
		for {
			select {
			case v := <-events:
				if pred(v) {
					return v
				}
			case <-deadline:
				t.Fatal("timed out waiting for event")
			}
		}
	}
	waitFor(func(v testView) bool { return v.Phase == "lobby" && len(v.Players) == 2 })

	// Only the host can start.
	if c := post(t, api+"/start"+q(guest.Token), nil, nil); c != http.StatusForbidden {
		t.Fatalf("guest start: %d", c)
	}
	var started testView
	if c := post(t, api+"/start"+q(host.Token), nil, &started); c != 200 || started.Phase != "question" {
		t.Fatalf("start: %d %+v", c, started)
	}
	v := waitFor(func(v testView) bool { return v.Phase == "question" && v.Round == 1 })
	if v.Question == nil || len(v.Question.Options) != 4 || v.Reveal != nil {
		t.Fatalf("question view: %+v", v)
	}

	// Flag image comes through the room; shape endpoint refuses a flag round.
	img, err := http.Get(api + "/flag" + q(guest.Token) + "&round=1")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(img.Body)
	img.Body.Close()
	if img.StatusCode != 200 || !bytes.Contains(body, []byte("<svg")) {
		t.Fatalf("flag image: %d", img.StatusCode)
	}
	if res, _ := http.Get(api + "/flag" + q(guest.Token) + "&round=2"); res.StatusCode != http.StatusNotFound {
		t.Fatalf("future round image should be hidden, got %d", res.StatusCode)
	}

	// Both answer: the stream moves to the reveal.
	opt := v.Question.Options[0].ID
	if c := post(t, api+"/answer"+q(host.Token), map[string]any{"round": 1, "optionId": opt}, nil); c != 200 {
		t.Fatalf("answer: %d", c)
	}
	if c := post(t, api+"/answer"+q(host.Token), map[string]any{"round": 1, "optionId": opt}, nil); c != http.StatusConflict {
		t.Fatalf("second answer: %d", c)
	}
	if c := post(t, api+"/answer"+q(guest.Token), map[string]any{"round": 1, "optionId": opt}, nil); c != 200 {
		t.Fatalf("guest answer: %d", c)
	}
	rv := waitFor(func(v testView) bool { return v.Phase == "reveal" })
	if rv.Reveal == nil || rv.Reveal.AnswerID == "" {
		t.Fatal("reveal has no answer")
	}

	// Unknown token is refused.
	if c := post(t, api+"/start?code="+code+"&t=nope", nil, nil); c != http.StatusForbidden {
		t.Fatalf("bad token: %d", c)
	}
}

func TestPartyHTTPChallengeAndShape(t *testing.T) {
	srv := partyServer(t)
	api := srv.URL + "/api/party"
	var me struct {
		Token string   `json:"token"`
		View  testView `json:"view"`
	}
	if c := post(t, api+"/create", map[string]string{"mode": "challenge", "game": "shape", "region": "Europe", "name": "Ana"}, &me); c != 200 {
		t.Fatalf("create: %d", c)
	}
	if me.View.Phase != "ready" {
		t.Fatalf("challenge should start ready, got %s", me.View.Phase)
	}
	q := "?code=" + me.View.Code + "&t=" + me.Token
	var v testView
	if post(t, api+"/start"+q, nil, &v); v.Phase != "question" || v.Question.Kind != "shape" {
		t.Fatalf("start: %+v", v)
	}

	res, err := http.Get(api + "/shape" + q + "&round=1")
	if err != nil {
		t.Fatal(err)
	}
	var geo struct {
		Features []struct {
			ID         string         `json:"id"`
			Properties map[string]any `json:"properties"`
			Geometry   map[string]any `json:"geometry"`
		} `json:"features"`
	}
	_ = json.NewDecoder(res.Body).Decode(&geo)
	res.Body.Close()
	if res.StatusCode != 200 || len(geo.Features) == 0 {
		t.Fatalf("shape: %d", res.StatusCode)
	}
	for _, f := range geo.Features {
		if f.ID != "" || len(f.Properties) != 0 || f.Geometry == nil {
			t.Fatalf("shape leaks the country: %+v", f)
		}
	}

	for round := 1; round <= v.Rounds; round++ {
		var after testView
		if c := post(t, api+"/answer"+q, map[string]any{"round": round, "optionId": v.Question.Options[0].ID}, &after); c != 200 || after.Phase != "reveal" {
			t.Fatalf("answer %d: %d %+v", round, c, after)
		}
		if c := post(t, api+"/next"+q, nil, &v); c != 200 {
			t.Fatalf("next: %d", c)
		}
	}
	if v.Phase != "finished" || len(v.Players) != 1 {
		t.Fatalf("finish: %+v", v)
	}
}
