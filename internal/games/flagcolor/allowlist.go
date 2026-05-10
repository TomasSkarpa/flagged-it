package flagcolor

import (
	"encoding/json"
	"fmt"
)

// AllowlistCCA2 returns curated country codes that may appear in flag color mode.
func AllowlistCCA2() ([]string, error) {
	var out []string
	if err := json.Unmarshal(allowlistRaw, &out); err != nil {
		return nil, fmt.Errorf("flagcolor allowlist: %w", err)
	}
	return out, nil
}
