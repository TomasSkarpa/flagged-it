package multiplayer

import (
	"crypto/rand"
	"encoding/hex"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"fmt"
	mathrand "math/rand"
	"sync"
	"time"
)

// RoomManager manages all game rooms
type RoomManager struct {
	rooms      map[string]*Room
	roomsMutex sync.RWMutex
	countries  []models.Country
}

var (
	globalRoomManager *RoomManager
	roomManagerOnce   sync.Once
)

// GetRoomManager returns the global room manager instance
func GetRoomManager() *RoomManager {
	roomManagerOnce.Do(func() {
		globalRoomManager = &RoomManager{
			rooms:     make(map[string]*Room),
			countries: data.LoadCountries(),
		}
		// Start cleanup goroutine
		go globalRoomManager.cleanupEmptyRooms()
	})
	return globalRoomManager
}

// GameEngineFactory creates a game engine for a given mode and config
type GameEngineFactory func(GameMode, RoomConfig, []models.Country) (GameEngine, error)

var engineFactory GameEngineFactory

// SetEngineFactory sets the factory function for creating game engines
func SetEngineFactory(factory GameEngineFactory) {
	engineFactory = factory
}

// CreateRoom creates a new game room
func (rm *RoomManager) CreateRoom(hostID string, hostName string, config RoomConfig) (*Room, error) {
	// Validate config
	if config.NumQuestions < 1 || config.NumQuestions > 100 {
		return nil, fmt.Errorf("numQuestions must be between 1 and 100")
	}
	if config.MaxPlayers < 2 || config.MaxPlayers > 20 {
		return nil, fmt.Errorf("maxPlayers must be between 2 and 20")
	}
	if config.TimeLimit < 0 {
		return nil, fmt.Errorf("timeLimit must be >= 0")
	}

	// Generate unique room ID and code
	roomID := generateRoomID()
	roomCode := generateRoomCode()

	// Create game engine using factory
	if engineFactory == nil {
		return nil, fmt.Errorf("game engine factory not set")
	}
	gameEngine, err := engineFactory(config.GameMode, config, rm.countries)
	if err != nil {
		return nil, fmt.Errorf("failed to create game engine: %w", err)
	}

	// Create room
	room := &Room{
		ID:            roomID,
		Code:          roomCode,
		HostID:        hostID,
		Config:        config,
		Status:        RoomStatusWaiting,
		Players:       make(map[string]*Player),
		QuestionIndex: 0,
		Answers:       make(map[string]map[string]*AnswerSubmission),
		CreatedAt:     time.Now(),
		gameEngine:    gameEngine,
	}

	// Add host as first player
	room.Players[hostID] = &Player{
		ID:       hostID,
		Name:     hostName,
		IsHost:   true,
		IsReady:  false,
		Score:    0,
		Streak:   0,
		JoinedAt: time.Now(),
	}

	// Store room
	rm.roomsMutex.Lock()
	rm.rooms[roomID] = room
	rm.roomsMutex.Unlock()

	return room, nil
}

// GetRoom retrieves a room by ID
func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.roomsMutex.RLock()
	defer rm.roomsMutex.RUnlock()
	room, exists := rm.rooms[roomID]
	return room, exists
}

// GetRoomByCode retrieves a room by code
func (rm *RoomManager) GetRoomByCode(code string) (*Room, bool) {
	rm.roomsMutex.RLock()
	defer rm.roomsMutex.RUnlock()
	for _, room := range rm.rooms {
		if room.Code == code {
			return room, true
		}
	}
	return nil, false
}

// DeleteRoom deletes a room
func (rm *RoomManager) DeleteRoom(roomID string) bool {
	rm.roomsMutex.Lock()
	defer rm.roomsMutex.Unlock()
	if _, exists := rm.rooms[roomID]; exists {
		delete(rm.rooms, roomID)
		return true
	}
	return false
}

// GetPublicRooms returns a list of public rooms
func (rm *RoomManager) GetPublicRooms() []PublicRoomInfo {
	rm.roomsMutex.RLock()
	defer rm.roomsMutex.RUnlock()

	var publicRooms []PublicRoomInfo
	for _, room := range rm.rooms {
		if room.Config.IsPublic && room.Status == RoomStatusWaiting {
			publicRooms = append(publicRooms, PublicRoomInfo{
				ID:           room.ID,
				Code:         room.Code,
				GameMode:     room.Config.GameMode,
				NumQuestions: room.Config.NumQuestions,
				Difficulty:   room.Config.Difficulty,
				PlayerCount:  len(room.Players),
				MaxPlayers:   room.Config.MaxPlayers,
				Status:       room.Status,
				CreatedAt:    room.CreatedAt,
			})
		}
	}

	return publicRooms
}

// AddPlayer adds a player to a room
func (rm *RoomManager) AddPlayer(roomID string, playerID string, playerName string, password string) error {
	rm.roomsMutex.Lock()
	defer rm.roomsMutex.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found")
	}

	if room.Status != RoomStatusWaiting {
		return fmt.Errorf("room is not accepting new players")
	}

	if len(room.Players) >= room.Config.MaxPlayers {
		return fmt.Errorf("room is full")
	}

	// Check password if set
	if room.Config.Password != "" && room.Config.Password != password {
		return fmt.Errorf("invalid password")
	}

	// Check if player already in room
	if _, exists := room.Players[playerID]; exists {
		return fmt.Errorf("player already in room")
	}

	// Add player
	room.Players[playerID] = &Player{
		ID:       playerID,
		Name:     playerName,
		IsHost:   false,
		IsReady:  false,
		Score:    0,
		Streak:   0,
		JoinedAt: time.Now(),
	}

	return nil
}

// RemovePlayer removes a player from a room
func (rm *RoomManager) RemovePlayer(roomID string, playerID string) {
	rm.roomsMutex.Lock()
	defer rm.roomsMutex.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return
	}

	delete(room.Players, playerID)

	// If host left and room not started, assign new host or delete room
	if room.HostID == playerID && room.Status == RoomStatusWaiting {
		if len(room.Players) == 0 {
			delete(rm.rooms, roomID)
		} else {
			// Assign new host (first remaining player)
			for id := range room.Players {
				room.HostID = id
				room.Players[id].IsHost = true
				break
			}
		}
	}
}

// cleanupEmptyRooms periodically removes empty rooms older than 1 hour
func (rm *RoomManager) cleanupEmptyRooms() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rm.roomsMutex.Lock()
		now := time.Now()
		for id, room := range rm.rooms {
			// Delete empty rooms older than 1 hour
			if len(room.Players) == 0 && now.Sub(room.CreatedAt) > time.Hour {
				delete(rm.rooms, id)
			}
			// Delete finished rooms older than 30 minutes
			if room.Status == RoomStatusFinished && room.FinishedAt != nil && now.Sub(*room.FinishedAt) > 30*time.Minute {
				delete(rm.rooms, id)
			}
		}
		rm.roomsMutex.Unlock()
	}
}

// generateRoomID generates a unique room ID
func generateRoomID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateRoomCode generates a short, human-readable room code
func generateRoomCode() string {
	// Generate 6-character alphanumeric code
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude confusing chars
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[mathrand.Intn(len(charset))]
	}
	return string(b)
}
