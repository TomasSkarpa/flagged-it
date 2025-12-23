//go:build !js || !wasm
// +build !js !wasm

package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ScoreEntry represents a single score entry
type ScoreEntry struct {
	GameMode string    `json:"game_mode"`
	Score    int       `json:"score"`
	Total    int       `json:"total"`
	Percent  float64   `json:"percent"`
	Date     time.Time `json:"date"`
	Duration int       `json:"duration"`
	Region   string    `json:"region,omitempty"`
}

// getScoreboardPath returns the path to the scoreboard file
func getScoreboardPath() (string, error) {
	// Use user's config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, "flagged-it")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, "scoreboard.json"), nil
}

// GetScoreboard retrieves all scores from file
func GetScoreboard() []ScoreEntry {
	path, err := getScoreboardPath()
	if err != nil {
		return []ScoreEntry{}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return []ScoreEntry{}
	}

	var scores []ScoreEntry
	if err := json.Unmarshal(data, &scores); err != nil {
		return []ScoreEntry{}
	}

	return scores
}

// SaveScore adds a new score entry to the scoreboard
func SaveScore(entry ScoreEntry) error {
	scores := GetScoreboard()

	// Add new entry
	entry.Date = time.Now()
	scores = append(scores, entry)

	// Keep only top 100 entries per game mode
	scoresByGame := make(map[string][]ScoreEntry)
	for _, s := range scores {
		scoresByGame[s.GameMode] = append(scoresByGame[s.GameMode], s)
	}

	// Sort and trim each game mode
	var trimmedScores []ScoreEntry
	for _, gameScores := range scoresByGame {
		sort.Slice(gameScores, func(i, j int) bool {
			if gameScores[i].Percent != gameScores[j].Percent {
				return gameScores[i].Percent > gameScores[j].Percent
			}
			return gameScores[i].Date.After(gameScores[j].Date)
		})

		// Keep top 100 per game
		if len(gameScores) > 100 {
			gameScores = gameScores[:100]
		}
		trimmedScores = append(trimmedScores, gameScores...)
	}

	// Save to file
	path, err := getScoreboardPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(trimmedScores, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetTopScores returns top N scores for a specific game mode
func GetTopScores(gameMode string, limit int) []ScoreEntry {
	allScores := GetScoreboard()

	// Filter by game mode
	var filtered []ScoreEntry
	for _, s := range allScores {
		if s.GameMode == gameMode {
			filtered = append(filtered, s)
		}
	}

	// Sort by percentage, then by date
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Percent != filtered[j].Percent {
			return filtered[i].Percent > filtered[j].Percent
		}
		return filtered[i].Date.After(filtered[j].Date)
	})

	// Return top N
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

// GetPersonalBest returns the best score for a specific game mode
func GetPersonalBest(gameMode string) *ScoreEntry {
	scores := GetTopScores(gameMode, 1)
	if len(scores) > 0 {
		return &scores[0]
	}
	return nil
}

// ClearScoreboard removes all scores
func ClearScoreboard() {
	path, err := getScoreboardPath()
	if err != nil {
		return
	}
	os.Remove(path)
}

