package api

import (
	"log"
	"net/http"
	"sync"
	"time"
)

// Solo game sessions live in plain maps. Every request for one game runs under
// that game's lock, so map access and session updates never race. The sweeper
// takes the same lock to drop sessions older than sessionMaxAge.

const (
	sessionMaxAge     = 2 * time.Hour
	sessionSweepEvery = 5 * time.Minute
)

var (
	flagGameLock        sync.Mutex
	shapeGameLock       sync.Mutex
	capitalGameLock     sync.Mutex
	higherLowerGameLock sync.Mutex
	worldleGameLock     sync.Mutex
	factsGameLock       sync.Mutex
	flagColorGameLock   sync.Mutex
)

// serialize runs h while holding mu.
func serialize(mu *sync.Mutex, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		h(w, r)
	}
}

// expireOld deletes entries created before cutoff and returns how many it removed.
func expireOld[T any](m map[string]T, createdAt func(T) time.Time, cutoff time.Time) int {
	removed := 0
	for id, s := range m {
		if createdAt(s).Before(cutoff) {
			delete(m, id)
			removed++
		}
	}
	return removed
}

func withLock(mu *sync.Mutex, fn func() int) int {
	mu.Lock()
	defer mu.Unlock()
	return fn()
}

// sweepSessions removes solo sessions older than sessionMaxAge.
func sweepSessions(now time.Time) int {
	cutoff := now.Add(-sessionMaxAge)
	n := 0
	n += withLock(&flagGameLock, func() int {
		return expireOld(sessions, func(s *GameSession) time.Time { return s.CreatedAt }, cutoff)
	})
	n += withLock(&shapeGameLock, func() int {
		return expireOld(shapeSessions, func(s *ShapeGameSession) time.Time { return s.CreatedAt }, cutoff)
	})
	n += withLock(&capitalGameLock, func() int {
		return expireOld(capitalSessions, func(s *CapitalGameSession) time.Time { return s.CreatedAt }, cutoff)
	})
	n += withLock(&higherLowerGameLock, func() int {
		return expireOld(higherLowerSessions, func(s *HigherLowerSession) time.Time { return s.CreatedAt }, cutoff)
	})
	n += withLock(&worldleGameLock, func() int {
		return expireOld(worldleSessions, func(s *WorldleSession) time.Time { return s.CreatedAt }, cutoff)
	})
	n += withLock(&factsGameLock, func() int {
		return expireOld(factsGameSessions, func(s *FactsGameSession) time.Time { return s.CreatedAt }, cutoff)
	})
	n += withLock(&flagColorGameLock, func() int {
		return expireOld(flagColorSessions, func(s *flagColorSession) time.Time { return s.CreatedAt }, cutoff)
	})
	return n
}

var startSweeperOnce sync.Once

func startSessionSweeper() {
	startSweeperOnce.Do(func() {
		go func() {
			t := time.NewTicker(sessionSweepEvery)
			defer t.Stop()
			for now := range t.C {
				if n := sweepSessions(now); n > 0 {
					log.Printf("session sweeper: removed %d expired sessions", n)
				}
			}
		}()
	})
}
