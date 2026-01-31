package multiplayer

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now (can be restricted later)
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a WebSocket client connection
type Client struct {
	conn     *websocket.Conn
	roomID   string
	playerID string
	send     chan []byte
	room     *Room
}

// Hub manages all active clients and rooms
type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool // roomID -> clients
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	rm         *RoomManager
	mu         sync.RWMutex
}

// NewHub creates a new hub
func NewHub(rm *RoomManager) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rm:         rm,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if h.rooms[client.roomID] == nil {
				h.rooms[client.roomID] = make(map[*Client]bool)
			}
			h.rooms[client.roomID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if roomClients, ok := h.rooms[client.roomID]; ok {
					delete(roomClients, client)
					if len(roomClients) == 0 {
						delete(h.rooms, client.roomID)
					}
				}
				close(client.send)
			}
			h.mu.Unlock()

			// Only remove player from room if they're not the host
			// Host should persist in room even if disconnected (they own it)
			// Regular players are removed on disconnect but can rejoin
			if client.playerID != "" && client.roomID != "" {
				room, exists := h.rm.GetRoom(client.roomID)
				if exists && room != nil && room.HostID != client.playerID {
					// Only remove non-host players
					h.rm.RemovePlayer(client.roomID, client.playerID)
					
					// Get updated room after removal
					room, exists = h.rm.GetRoom(client.roomID)
					if exists && room != nil {
						// Create sanitized room copy for broadcast
						roomCopy := *room
						roomCopy.gameEngine = nil // Don't send game engine
						
						// Sanitize current question if present
						if roomCopy.CurrentQuestion != nil {
							qCopy := *roomCopy.CurrentQuestion
							qCopy.CorrectAnswer = ""
							roomCopy.CurrentQuestion = &qCopy
						}
						
						// Broadcast updated room state to remaining players
						stateMsg := RoomStateMessage{
							Type:      "ROOM_STATE",
							Room:      &roomCopy,
							Players:   room.Players,
							Timestamp: time.Now(),
						}
						msgBytes, _ := json.Marshal(stateMsg)
						h.BroadcastToRoom(client.roomID, msgBytes)
					}
				}
				// Host stays in room even if disconnected
			}
		}
	}
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomClients, ok := h.rooms[roomID]; ok {
		for client := range roomClients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(roomClients, client)
			}
		}
	}
}

// Message types
type WSMessage struct {
	Type      string          `json:"type"`
	PlayerID  string          `json:"playerId,omitempty"`
	RoomID    string          `json:"roomId,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// Message data types
type JoinMessage struct {
	RoomID   string `json:"roomId"`
	RoomCode string `json:"roomCode,omitempty"`
	PlayerID string `json:"playerId"`
	PlayerName string `json:"playerName"`
	Password string `json:"password,omitempty"`
}

type ReadyMessage struct {
	PlayerID string `json:"playerId"`
	Ready    bool   `json:"ready"`
}

type StartMessage struct {
	PlayerID string `json:"playerId"`
}

type AnswerMessage struct {
	PlayerID   string `json:"playerId"`
	QuestionID string `json:"questionId"`
	Answer     string `json:"answer"`
	TimeTaken  int    `json:"timeTaken"` // milliseconds
}

type ConfigUpdateMessage struct {
	PlayerID string     `json:"playerId"`
	Config   RoomConfig `json:"config"`
}

// Response message types
type RoomStateMessage struct {
	Type      string                 `json:"type"`
	Room      *Room                  `json:"room"`
	Players   map[string]*Player     `json:"players"`
	Timestamp time.Time              `json:"timestamp"`
}

type QuestionMessage struct {
	Type     string                 `json:"type"`
	Question *Question              `json:"question"`
	Index    int                    `json:"index"`
	Total    int                    `json:"total"`
}

type AnswerResultMessage struct {
	Type       string                 `json:"type"`
	PlayerID   string                 `json:"playerId"`
	QuestionID string                 `json:"questionId"`
	IsCorrect  bool                   `json:"isCorrect"`
	Points     int                    `json:"points"`
	CorrectAnswer string              `json:"correctAnswer"`
	Leaderboard map[string]*Player    `json:"leaderboard"`
}

type GameFinishedMessage struct {
	Type       string                 `json:"type"`
	Leaderboard map[string]*Player    `json:"leaderboard"`
	FinalScores map[string]int        `json:"finalScores"`
}

type ErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"` // Optional fallback, frontend should use Code for translation
	Code    string `json:"code"`              // Translation key (e.g., "multiplayer.error.room_not_found")
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Extract room ID from URL path
	// Expected path: /api/ws/rooms/:roomId
	roomID := r.URL.Path[len("/api/ws/rooms/"):]
	if roomID == "" {
		conn.Close()
		return
	}

	client := &Client{
		conn:   conn,
		roomID: roomID,
		send:   make(chan []byte, 256),
	}

	hub.register <- client

	// Start goroutines
	go client.writePump()
	go client.readPump(hub)
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Handle message
		c.handleMessage(hub, &msg)
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage routes messages to appropriate handlers
func (c *Client) handleMessage(hub *Hub, msg *WSMessage) {
	rm := hub.rm

	switch msg.Type {
	case "JOIN":
		var joinMsg JoinMessage
		if err := json.Unmarshal(msg.Data, &joinMsg); err != nil {
			c.sendError("Invalid JOIN message", "multiplayer.error.invalid_message")
			return
		}

		// Get room by ID or code
		var room *Room
		var exists bool
		if joinMsg.RoomID != "" {
			room, exists = rm.GetRoom(joinMsg.RoomID)
		} else if joinMsg.RoomCode != "" {
			room, exists = rm.GetRoomByCode(joinMsg.RoomCode)
		} else {
			c.sendError("Room ID or code required", "multiplayer.error.missing_room")
			return
		}

		if !exists {
			c.sendError("Room not found", "multiplayer.error.room_not_found")
			return
		}

		// Set client room and player info
		c.roomID = room.ID
		c.playerID = joinMsg.PlayerID
		
		// Check if player already exists in room
		rm.roomsMutex.RLock()
		existingPlayer, playerExists := room.Players[joinMsg.PlayerID]
		rm.roomsMutex.RUnlock()
		
		if !playerExists {
			// New player - add them to the room
			if err := rm.AddPlayer(room.ID, joinMsg.PlayerID, joinMsg.PlayerName, joinMsg.Password); err != nil {
				// Map specific errors to translation keys
				errorCode := "multiplayer.error.join_failed"
				if err.Error() == "room is full" {
					errorCode = "multiplayer.error.room_full"
				} else if err.Error() == "invalid password" {
					errorCode = "multiplayer.error.invalid_password"
				} else if err.Error() == "room is not accepting new players" {
					errorCode = "multiplayer.error.room_not_accepting"
				}
				c.sendError(err.Error(), errorCode)
				return
			}
		} else {
			// Player already exists - this is a reconnection (host or regular player reconnecting)
			// Update player name if provided (in case they changed it)
			if joinMsg.PlayerName != "" && existingPlayer.Name != joinMsg.PlayerName {
				rm.roomsMutex.Lock()
				if room, ok := rm.rooms[room.ID]; ok {
					if player, exists := room.Players[joinMsg.PlayerID]; exists {
						player.Name = joinMsg.PlayerName
					}
				}
				rm.roomsMutex.Unlock()
			}
			// Player reconnected - just send them the current room state
		}

		// Update client room reference
		c.room = room

		// Send room state to all players
		c.broadcastRoomState(hub, room)

	case "READY":
		var readyMsg ReadyMessage
		if err := json.Unmarshal(msg.Data, &readyMsg); err != nil {
			c.sendError("Invalid READY message", "multiplayer.error.invalid_message")
			return
		}

		room, exists := rm.GetRoom(c.roomID)
		if !exists {
			c.sendError("Room not found", "multiplayer.error.room_not_found")
			return
		}

		player, exists := room.Players[readyMsg.PlayerID]
		if !exists {
			c.sendError("Player not found", "multiplayer.error.player_not_found")
			return
		}

		player.IsReady = readyMsg.Ready
		c.broadcastRoomState(hub, room)

	case "START":
		var startMsg StartMessage
		if err := json.Unmarshal(msg.Data, &startMsg); err != nil {
			c.sendError("Invalid START message", "multiplayer.error.invalid_message")
			return
		}

		room, exists := rm.GetRoom(c.roomID)
		if !exists {
			c.sendError("Room not found", "multiplayer.error.room_not_found")
			return
		}

		// Check if player is host
		if room.HostID != startMsg.PlayerID {
			c.sendError("Only host can start the game", "multiplayer.error.not_host")
			return
		}

		// Check if there are at least 2 players
		if len(room.Players) < 2 {
			c.sendError("Need at least 2 players to start", "multiplayer.error.not_enough_players")
			return
		}

		// Check if all players are ready
		allReady := true
		for _, player := range room.Players {
			if !player.IsReady {
				allReady = false
				break
			}
		}

		if !allReady {
			c.sendError("Not all players are ready", "multiplayer.error.not_all_ready")
			return
		}

		// Start game
		if err := c.startGame(hub, room); err != nil {
			c.sendError(err.Error(), "multiplayer.error.start_failed")
			return
		}

	case "ANSWER":
		var answerMsg AnswerMessage
		if err := json.Unmarshal(msg.Data, &answerMsg); err != nil {
			log.Printf("Error unmarshaling ANSWER message: %v, Data: %s", err, string(msg.Data))
			c.sendError("Invalid ANSWER message", "multiplayer.error.invalid_message")
			return
		}

		log.Printf("Received ANSWER message: PlayerID=%s, QuestionID=%s, Answer=%s, TimeTaken=%d", 
			answerMsg.PlayerID, answerMsg.QuestionID, answerMsg.Answer, answerMsg.TimeTaken)

		room, exists := rm.GetRoom(c.roomID)
		if !exists {
			log.Printf("Room not found for ANSWER: roomID=%s", c.roomID)
			c.sendError("Room not found", "multiplayer.error.room_not_found")
			return
		}

		if room.Status != RoomStatusPlaying {
			log.Printf("Game not in progress for ANSWER: roomID=%s, status=%s", c.roomID, room.Status)
			c.sendError("Game is not in progress", "multiplayer.error.game_not_started")
			return
		}

		// Submit answer
		c.submitAnswer(hub, room, &answerMsg)

	case "CONFIG_UPDATE":
		var configMsg ConfigUpdateMessage
		if err := json.Unmarshal(msg.Data, &configMsg); err != nil {
			c.sendError("Invalid CONFIG_UPDATE message", "multiplayer.error.invalid_message")
			return
		}

		room, exists := rm.GetRoom(c.roomID)
		if !exists {
			c.sendError("Room not found", "multiplayer.error.room_not_found")
			return
		}

		// Check if player is host
		if room.HostID != configMsg.PlayerID {
			c.sendError("Only host can update config", "multiplayer.error.not_host")
			return
		}

		// Check if game hasn't started
		if room.Status != RoomStatusWaiting {
			c.sendError("Cannot update config after game started", "multiplayer.error.config_update_after_start")
			return
		}

		// Update config
		room.Config = configMsg.Config
		c.broadcastRoomState(hub, room)

	default:
		c.sendError("Unknown message type: "+msg.Type, "multiplayer.error.unknown_message_type")
	}
}

// sendError sends an error message to the client
// code should be a translation key (e.g., "multiplayer.error.room_not_found")
// message is optional fallback text for debugging/logging
func (c *Client) sendError(message string, code string) {
	errorMsg := ErrorMessage{
		Type:    "ERROR",
		Message: "", // Don't send hardcoded message - frontend will translate using Code
		Code:    code,
	}
	msgBytes, _ := json.Marshal(errorMsg)
	select {
	case c.send <- msgBytes:
	default:
	}
	// Log error for debugging
	if message != "" {
		log.Printf("WebSocket error [%s]: %s", code, message)
	}
}

// broadcastRoomState broadcasts room state to all clients in the room
func (c *Client) broadcastRoomState(hub *Hub, room *Room) {
	// Create sanitized room copy (remove internal fields)
	roomCopy := *room
	roomCopy.gameEngine = nil // Don't send game engine

	stateMsg := RoomStateMessage{
		Type:      "ROOM_STATE",
		Room:      &roomCopy,
		Players:   room.Players,
		Timestamp: time.Now(),
	}

	msgBytes, _ := json.Marshal(stateMsg)
	hub.BroadcastToRoom(room.ID, msgBytes)
}

// startGame starts the game
func (c *Client) startGame(hub *Hub, room *Room) error {
	room.Status = RoomStatusPlaying
	now := time.Now()
	room.StartedAt = &now
	room.QuestionIndex = 0

	// Initialize game engine
	if err := room.gameEngine.Initialize(room.Config, nil); err != nil {
		return err
	}

	// Get first question
	question, err := room.gameEngine.GetNextQuestion()
	if err != nil {
		return err
	}

	room.CurrentQuestion = question

	// Broadcast game started
	c.broadcastRoomState(hub, room)

	// Send first question
	c.sendQuestion(hub, room, question)

	// Start question timer if time limit set
	if room.Config.TimeLimit > 0 {
		go c.questionTimer(hub, room, question)
	}

	return nil
}

// sendQuestion sends a question to all players
func (c *Client) sendQuestion(hub *Hub, room *Room, question *Question) {
	// Create sanitized question (remove correct answer)
	questionCopy := *question
	questionCopy.CorrectAnswer = ""

	questionMsg := QuestionMessage{
		Type:     "QUESTION",
		Question: &questionCopy,
		Index:    room.QuestionIndex + 1,
		Total:    room.gameEngine.GetQuestionCount(),
	}

	msgBytes, _ := json.Marshal(questionMsg)
	hub.BroadcastToRoom(room.ID, msgBytes)
}

// submitAnswer handles answer submission
func (c *Client) submitAnswer(hub *Hub, room *Room, answerMsg *AnswerMessage) {
	log.Printf("submitAnswer called: PlayerID=%s, QuestionID=%s, Answer=%s", 
		answerMsg.PlayerID, answerMsg.QuestionID, answerMsg.Answer)
	
	// Check if already answered
	if room.Answers == nil {
		room.Answers = make(map[string]map[string]*AnswerSubmission)
	}
	if room.Answers[answerMsg.QuestionID] == nil {
		room.Answers[answerMsg.QuestionID] = make(map[string]*AnswerSubmission)
	}

	if _, exists := room.Answers[answerMsg.QuestionID][answerMsg.PlayerID]; exists {
		log.Printf("Player already answered: PlayerID=%s, QuestionID=%s", answerMsg.PlayerID, answerMsg.QuestionID)
		c.sendError("Already answered this question", "multiplayer.error.already_answered")
		return
	}

	// Check if game engine exists
	if room.gameEngine == nil {
		log.Printf("Game engine is nil for room: %s", room.ID)
		c.sendError("Game engine not initialized", "multiplayer.error.answer_failed")
		return
	}

	// Validate answer
	result, err := room.gameEngine.SubmitAnswer(answerMsg.QuestionID, answerMsg.Answer, answerMsg.TimeTaken)
	if err != nil {
		log.Printf("Error submitting answer: %v", err)
		c.sendError(err.Error(), "multiplayer.error.answer_failed")
		return
	}

	log.Printf("Answer validated: IsCorrect=%v, Points=%d", result.IsCorrect, result.Points)

	// Update player score
	player, exists := room.Players[answerMsg.PlayerID]
	if !exists {
		c.sendError("Player not found", "multiplayer.error.player_not_found")
		return
	}

	if result.IsCorrect {
		player.Score += result.Points
		player.Streak++
	} else {
		player.Streak = 0
	}

	// Store answer
	submission := &AnswerSubmission{
		PlayerID:    answerMsg.PlayerID,
		QuestionID:  answerMsg.QuestionID,
		Answer:      answerMsg.Answer,
		IsCorrect:   result.IsCorrect,
		Points:      result.Points,
		TimeTaken:   answerMsg.TimeTaken,
		SubmittedAt: time.Now(),
	}
	room.Answers[answerMsg.QuestionID][answerMsg.PlayerID] = submission
	player.LastAnswer = submission

	// Broadcast answer result
	c.broadcastAnswerResult(hub, room, answerMsg.PlayerID, answerMsg.QuestionID, result)

	// Check if all players answered or time expired
	c.checkQuestionComplete(hub, room)
}

// broadcastAnswerResult broadcasts answer result
func (c *Client) broadcastAnswerResult(hub *Hub, room *Room, playerID string, questionID string, result *AnswerResult) {
	// Build leaderboard
	leaderboard := make(map[string]*Player)
	for id, player := range room.Players {
		leaderboard[id] = player
	}

	resultMsg := AnswerResultMessage{
		Type:        "ANSWER_RESULT",
		PlayerID:    playerID,
		QuestionID:  questionID,
		IsCorrect:   result.IsCorrect,
		Points:      result.Points,
		CorrectAnswer: result.CorrectAnswer,
		Leaderboard: leaderboard,
	}

	msgBytes, _ := json.Marshal(resultMsg)
	hub.BroadcastToRoom(room.ID, msgBytes)
}

// checkQuestionComplete checks if question is complete and moves to next
func (c *Client) checkQuestionComplete(hub *Hub, room *Room) {
	// Check if all players answered
	allAnswered := true
	for playerID := range room.Players {
		if room.Answers[room.CurrentQuestion.ID] == nil {
			allAnswered = false
			break
		}
		if _, exists := room.Answers[room.CurrentQuestion.ID][playerID]; !exists {
			allAnswered = false
			break
		}
	}

	if allAnswered {
		c.nextQuestion(hub, room)
	}
}

// questionTimer handles question time limit
func (c *Client) questionTimer(hub *Hub, room *Room, question *Question) {
	time.Sleep(time.Duration(room.Config.TimeLimit) * time.Second)
	
	// Check if question is still current
	rm := hub.rm
	currentRoom, exists := rm.GetRoom(room.ID)
	if !exists || currentRoom.CurrentQuestion == nil || currentRoom.CurrentQuestion.ID != question.ID {
		return
	}

	// Move to next question
	c.nextQuestion(hub, currentRoom)
}

// nextQuestion moves to the next question
func (c *Client) nextQuestion(hub *Hub, room *Room) {
	// Check if game is complete
	if room.gameEngine.IsComplete() {
		c.finishGame(hub, room)
		return
	}

		// Get next question
		question, err := room.gameEngine.GetNextQuestion()
		if err != nil {
			c.sendError(err.Error(), "multiplayer.error.next_question_failed")
			return
		}

	room.QuestionIndex++
	room.CurrentQuestion = question

	// Broadcast new question
	c.broadcastRoomState(hub, room)
	c.sendQuestion(hub, room, question)

	// Start timer for new question
	if room.Config.TimeLimit > 0 {
		go c.questionTimer(hub, room, question)
	}
}

// finishGame finishes the game
func (c *Client) finishGame(hub *Hub, room *Room) {
	room.Status = RoomStatusFinished
	now := time.Now()
	room.FinishedAt = &now
	room.CurrentQuestion = nil

	// Build final leaderboard
	leaderboard := make(map[string]*Player)
	finalScores := make(map[string]int)
	for id, player := range room.Players {
		leaderboard[id] = player
		finalScores[id] = player.Score
	}

	finishedMsg := GameFinishedMessage{
		Type:        "GAME_FINISHED",
		Leaderboard: leaderboard,
		FinalScores: finalScores,
	}

	msgBytes, _ := json.Marshal(finishedMsg)
	hub.BroadcastToRoom(room.ID, msgBytes)

	// Also broadcast room state
	c.broadcastRoomState(hub, room)
}
