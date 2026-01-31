package adapters

import (
	"flagged-it/internal/data/models"
	"flagged-it/internal/multiplayer"
)

// Stub adapters for games not yet implemented

func NewCapitalGameEngine(config multiplayer.RoomConfig, countries []models.Country) (multiplayer.GameEngine, error) {
	return nil, &multiplayer.GameError{Message: "capital game mode not yet implemented"}
}

func NewHigherLowerGameEngine(config multiplayer.RoomConfig, countries []models.Country) (multiplayer.GameEngine, error) {
	return nil, &multiplayer.GameError{Message: "higherlower game mode not yet implemented"}
}

func NewFactsGameEngine(config multiplayer.RoomConfig, countries []models.Country) (multiplayer.GameEngine, error) {
	return nil, &multiplayer.GameError{Message: "facts game mode not yet implemented"}
}

func NewWorldleGameEngine(config multiplayer.RoomConfig, countries []models.Country) (multiplayer.GameEngine, error) {
	return nil, &multiplayer.GameError{Message: "worldle game mode not yet implemented"}
}
