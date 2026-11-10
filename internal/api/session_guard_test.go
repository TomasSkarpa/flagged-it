package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Parallel game starts used to write the session maps without a lock.
// Run with -race to check the per-game locks.
func TestParallelStartsDoNotRace(t *testing.T) {
	starts := []http.HandlerFunc{
		serialize(&flagGameLock, (&FlagGameHandler{}).StartGame),
		serialize(&shapeGameLock, (&ShapeGameHandler{}).StartGame),
		serialize(&capitalGameLock, (&CapitalGameHandler{}).StartGame),
		serialize(&higherLowerGameLock, (&HigherLowerHandler{}).StartGame),
		serialize(&worldleGameLock, (&WorldleGameHandler{}).StartGame),
		serialize(&factsGameLock, (&FactsGameHandler{}).StartGame),
	}
	var wg sync.WaitGroup
	for _, h := range starts {
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func(h http.HandlerFunc) {
				defer wg.Done()
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"region":"Europe","locale":"en"}`))
				rec := httptest.NewRecorder()
				h(rec, req)
				if rec.Code != http.StatusOK {
					t.Errorf("start returned %d: %s", rec.Code, rec.Body.String())
				}
			}(h)
		}
	}
	// Sweep at the same time to cover sweeper vs handler access.
	wg.Add(1)
	go func() {
		defer wg.Done()
		sweepSessions(time.Now())
	}()
	wg.Wait()
}

func TestSweepRemovesOnlyOldSessions(t *testing.T) {
	flagGameLock.Lock()
	sessions["old"] = &GameSession{SessionID: "old", CreatedAt: time.Now().Add(-3 * time.Hour)}
	sessions["new"] = &GameSession{SessionID: "new", CreatedAt: time.Now()}
	flagGameLock.Unlock()

	sweepSessions(time.Now())

	flagGameLock.Lock()
	defer flagGameLock.Unlock()
	if _, ok := sessions["old"]; ok {
		t.Error("old session was not removed")
	}
	if _, ok := sessions["new"]; !ok {
		t.Error("new session was removed")
	}
	delete(sessions, "new")
}
