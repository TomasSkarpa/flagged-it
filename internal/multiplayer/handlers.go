package multiplayer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// RoomHandler handles HTTP requests for room management
type RoomHandler struct {
	rm  *RoomManager
	hub *Hub
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(rm *RoomManager, hub *Hub) *RoomHandler {
	return &RoomHandler{
		rm:  rm,
		hub: hub,
	}
}

// CreateRoom handles POST /api/rooms
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		HostName string     `json:"hostName"`
		Config   RoomConfig `json:"config"`
	}

	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	log.Printf("Create room request body: %s", string(bodyBytes))

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("Error decoding create room request: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Create room request parsed: hostName=%s, config=%+v", req.HostName, req.Config)

	// Validate host name
	if req.HostName == "" {
		req.HostName = "Player"
	}

	// Use default config if not provided
	if req.Config.GameMode == "" {
		req.Config = DefaultRoomConfig()
	}

	// Generate player ID for host
	hostID := uuid.New().String()

	// Create room
	room, err := h.rm.CreateRoom(hostID, req.HostName, req.Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create response
	response := map[string]interface{}{
		"roomId":   room.ID,
		"roomCode": room.Code,
		"hostId":   hostID,
		"shareUrl": fmt.Sprintf("/multiplayer/%s", room.Code),
		"room":     sanitizeRoom(room),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRoom handles GET /api/rooms/:roomId
func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract room ID from path
	// Expected: /api/rooms/:roomId
	pathParts := strings.Split(r.URL.Path, "/")
	var roomID string
	for i, part := range pathParts {
		if part == "rooms" && i+1 < len(pathParts) {
			roomID = pathParts[i+1]
			break
		}
	}

	if roomID == "" {
		http.Error(w, "Room ID required", http.StatusBadRequest)
		return
	}

	room, exists := h.rm.GetRoom(roomID)
	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sanitizeRoom(room))
}

// GetRoomByCode handles GET /api/rooms/code/:code
func (h *RoomHandler) GetRoomByCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract code from path
	pathParts := strings.Split(r.URL.Path, "/")
	var code string
	for i, part := range pathParts {
		if part == "code" && i+1 < len(pathParts) {
			code = pathParts[i+1]
			break
		}
	}

	if code == "" {
		http.Error(w, "Room code required", http.StatusBadRequest)
		return
	}

	room, exists := h.rm.GetRoomByCode(code)
	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sanitizeRoom(room))
}

// GetPublicRooms handles GET /api/rooms/public
func (h *RoomHandler) GetPublicRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	publicRooms := h.rm.GetPublicRooms()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms": publicRooms,
	})
}

// DeleteRoom handles DELETE /api/rooms/:roomId
func (h *RoomHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract room ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	var roomID string
	for i, part := range pathParts {
		if part == "rooms" && i+1 < len(pathParts) {
			roomID = pathParts[i+1]
			break
		}
	}

	if roomID == "" {
		http.Error(w, "Room ID required", http.StatusBadRequest)
		return
	}

	deleted := h.rm.DeleteRoom(roomID)
	if !deleted {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"roomId":  roomID,
	})
}

// sanitizeRoom creates a sanitized copy of room for JSON response
func sanitizeRoom(room *Room) *Room {
	if room == nil {
		return nil
	}

	// Create copy
	roomCopy := *room

	// Remove internal fields
	roomCopy.gameEngine = nil

	// Sanitize current question (remove correct answer)
	if roomCopy.CurrentQuestion != nil {
		qCopy := *roomCopy.CurrentQuestion
		qCopy.CorrectAnswer = ""
		roomCopy.CurrentQuestion = &qCopy
	}

	// Answers are already sanitized (they don't contain correct answers)
	// Just copy the map structure
	if roomCopy.Answers != nil {
		sanitizedAnswers := make(map[string]map[string]*AnswerSubmission)
		for qID, playerAnswers := range roomCopy.Answers {
			sanitizedAnswers[qID] = make(map[string]*AnswerSubmission)
			for pID, answer := range playerAnswers {
				answerCopy := *answer
				sanitizedAnswers[qID][pID] = &answerCopy
			}
		}
		roomCopy.Answers = sanitizedAnswers
	}

	return &roomCopy
}

// SetupMultiplayerRoutes sets up multiplayer HTTP routes
// corsMiddleware and loggingMiddleware should be provided from the api package
func SetupMultiplayerRoutes(rm *RoomManager, hub *Hub, corsMiddleware func(http.HandlerFunc) http.HandlerFunc, loggingMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	handler := NewRoomHandler(rm, hub)

	// Apply middleware to routes
	applyMiddleware := func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return loggingMiddleware(corsMiddleware(handlerFunc))
	}

	// Room management routes
	http.HandleFunc("/api/rooms", applyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateRoom(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/rooms/public", applyMiddleware(handler.GetPublicRooms))
	http.HandleFunc("/api/rooms/code/", applyMiddleware(handler.GetRoomByCode))

	// Dynamic route handler for /api/rooms/:roomId
	http.HandleFunc("/api/rooms/", applyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/public") {
			handler.GetPublicRooms(w, r)
			return
		}
		if strings.Contains(path, "/code/") {
			handler.GetRoomByCode(w, r)
			return
		}

		// Extract room ID
		parts := strings.Split(path, "/")
		if len(parts) >= 4 && parts[2] == "rooms" {
			roomID := parts[3]
			if roomID != "" && roomID != "public" && !strings.HasPrefix(roomID, "code") {
				if r.Method == http.MethodGet {
					// Temporarily modify path for GetRoom
					r.URL.Path = "/api/rooms/" + roomID
					handler.GetRoom(w, r)
				} else if r.Method == http.MethodDelete {
					r.URL.Path = "/api/rooms/" + roomID
					handler.DeleteRoom(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}
		}

		http.Error(w, "Invalid route", http.StatusNotFound)
	}))

	// WebSocket route (CORS handled in WebSocket upgrader)
	// Use /api/ws/rooms/ for consistency with other API routes
	http.HandleFunc("/api/ws/rooms/", func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocket(hub, w, r)
	})

	log.Println("Multiplayer routes configured")
}
