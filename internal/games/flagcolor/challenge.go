package flagcolor

import (
	"errors"
	"fmt"
	"math/rand"

	"flagged-it/internal/data/models"
)

// ErrNoEligibleChallenge means no country in the pool had an unused SVG region
// tagged with data-fi-guess for the current difficulty (after skipping invalid SVGs).
var ErrNoEligibleChallenge = errors.New("no eligible flag color challenge")

// Challenge is one round's hidden-color target.
type Challenge struct {
	Country     models.Country
	GuessableID string
	TargetHex   string
}

// PickChallenge chooses an unused country and one annotated region matching difficulty.
func PickChallenge(difficulty string, pool []models.Country, used map[string]bool) (*Challenge, error) {
	if len(pool) == 0 {
		return nil, fmt.Errorf("empty pool")
	}
	for _, i := range rand.Perm(len(pool)) {
		c := pool[i]
		if used[c.CCA2] {
			continue
		}
		raw, err := ReadSVGBytes(c.CCA2)
		if err != nil || len(raw) == 0 {
			continue
		}
		parts, err := ParseGuessableParts(raw)
		if err != nil || len(parts) == 0 {
			continue
		}
		eligible := dedupeGuessableParts(FilterPartsByDifficulty(parts, difficulty))
		if len(eligible) == 0 {
			continue
		}
		pick := eligible[rand.Intn(len(eligible))]
		return &Challenge{
			Country:     c,
			GuessableID: pick.GuessID,
			TargetHex:   pick.Fill,
		}, nil
	}
	return nil, fmt.Errorf("%w", ErrNoEligibleChallenge)
}

// dedupeGuessableParts keeps one entry per (GuessID, Fill) so multiple SVG paths
// that share the same guess id (e.g. emblem shards) do not skew random selection.
func dedupeGuessableParts(parts []GuessablePart) []GuessablePart {
	type key struct {
		id   string
		fill string
	}
	seen := make(map[key]bool)
	out := make([]GuessablePart, 0, len(parts))
	for _, p := range parts {
		k := key{p.GuessID, p.Fill}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, p)
	}
	return out
}
