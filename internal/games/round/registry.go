package round

import (
	"fmt"
	"sync"
)

// Factory creates a new RoundRunner for a game type with the given options.
type Factory func(opts RoundOptions) (RoundRunner, error)

var (
	registry   = make(map[GameType]Factory)
	registryMu sync.RWMutex
)

// Register registers a factory for the given game type.
// Typically called from adapter init() in round/adapters.
func Register(gt GameType, f Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[gt] = f
}

// NewRunner creates a new RoundRunner for the given game type and options.
// Returns an error if the game type is not registered.
func NewRunner(gt GameType, opts RoundOptions) (RoundRunner, error) {
	registryMu.RLock()
	f, ok := registry[gt]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("round: game type %q not registered", gt)
	}
	return f(opts)
}

// Registered returns all registered game types (e.g. for quiz preset options).
func Registered() []GameType {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]GameType, 0, len(registry))
	for gt := range registry {
		out = append(out, gt)
	}
	return out
}
