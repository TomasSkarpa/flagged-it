package flagcolor

import (
	"fmt"
	"math/rand"

	"flagged-it/internal/data/models"
)

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
		eligible := FilterPartsByDifficulty(parts, difficulty)
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
	return nil, fmt.Errorf("no eligible flag color challenge")
}
