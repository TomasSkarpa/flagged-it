package flag

import "flagged-it/internal/data/models"

// GameConfig holds configuration for the flag game
type GameConfig struct {
	RoundCount int
	Region    string
}

// DefaultGameConfig returns the default configuration
func DefaultGameConfig() GameConfig {
	return GameConfig{
		RoundCount: 10,
		Region:    "",
	}
}

// RegionInfo provides information about available regions
type RegionInfo struct {
	Name      string
	Countries []models.Country
}
