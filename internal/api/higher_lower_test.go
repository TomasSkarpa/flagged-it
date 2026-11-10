package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The right-hand value is the answer, so it must not reach the client before
// the player picks higher or lower. The answer response still reveals it.
func TestHigherLowerHidesRightValue(t *testing.T) {
	h := &HigherLowerHandler{}
	rec := httptest.NewRecorder()
	h.StartGame(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"category":"population"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("start: %d %s", rec.Code, rec.Body.String())
	}
	var start struct {
		SessionID  string         `json:"sessionId"`
		Comparison map[string]any `json:"comparison"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &start); err != nil {
		t.Fatal(err)
	}
	if _, ok := start.Comparison["right"].(map[string]any)["value"]; ok {
		t.Fatal("start response leaks right.value")
	}
	if _, ok := start.Comparison["left"].(map[string]any)["value"]; !ok {
		t.Fatal("start response is missing left.value")
	}

	rec = httptest.NewRecorder()
	h.SubmitAnswer(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"sessionId":"`+start.SessionID+`","answer":"higher"}`)))
	var ans map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &ans); err != nil {
		t.Fatal(err)
	}
	if _, ok := ans["rightValue"]; !ok {
		t.Fatal("answer response is missing rightValue")
	}
	if next, ok := ans["nextComparison"].(map[string]any); ok {
		if _, leaked := next["right"].(map[string]any)["value"]; leaked {
			t.Fatal("nextComparison leaks right.value")
		}
	}
}
