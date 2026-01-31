package api

import (
	"flagged-it/internal/multiplayer"
	"flagged-it/internal/multiplayer/adapters"
	"flagged-it/internal/data/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func SetupRoutes() {
	flagHandler := &FlagGameHandler{}
	shapeHandler := &ShapeGameHandler{}
	capitalHandler := &CapitalGameHandler{}
	higherLowerHandler := &HigherLowerHandler{}
	worldleHandler := &WorldleGameHandler{}
	factsHandler := &FactsGameHandler{}
	debugHandler := &DebugHandler{}

	// Rate limiter: 100 requests per minute, burst of 20
	rateLimiter := NewRateLimiter(600*time.Millisecond, 20) // ~100 requests per minute
	rateLimitMiddleware := RateLimitMiddleware(rateLimiter)

	// CORS middleware - handles preflight and actual requests
	corsMiddleware := CORSMiddleware()

	// Combine logging, rate limiting and CORS middleware
	// IMPORTANT: CORS must run first to handle OPTIONS preflight requests
	// Rate limiting should skip OPTIONS requests
	// Logging tracks all requests
	combinedMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return LoggingMiddleware(corsMiddleware(rateLimitMiddleware(next)))
	}

	// Flag game routes
	http.HandleFunc("/api/game/flag/start", combinedMiddleware(flagHandler.StartGame))
	http.HandleFunc("/api/game/flag/question", combinedMiddleware(flagHandler.GetQuestion))
	http.HandleFunc("/api/game/flag/answer", combinedMiddleware(flagHandler.SubmitAnswer))
	http.HandleFunc("/api/game/flag/score", combinedMiddleware(flagHandler.GetScore))

	// Shape game routes
	http.HandleFunc("/api/game/shape/start", combinedMiddleware(shapeHandler.StartGame))
	http.HandleFunc("/api/game/shape/question", combinedMiddleware(shapeHandler.GetQuestion))
	http.HandleFunc("/api/game/shape/answer", combinedMiddleware(shapeHandler.SubmitAnswer))
	http.HandleFunc("/api/game/shape/score", combinedMiddleware(shapeHandler.GetScore))

	// Capital game routes
	http.HandleFunc("/api/game/capital/start", combinedMiddleware(capitalHandler.StartGame))
	http.HandleFunc("/api/game/capital/question", combinedMiddleware(capitalHandler.GetQuestion))
	http.HandleFunc("/api/game/capital/answer", combinedMiddleware(capitalHandler.SubmitAnswer))
	http.HandleFunc("/api/game/capital/score", combinedMiddleware(capitalHandler.GetScore))

	// Higher/Lower game routes
	http.HandleFunc("/api/game/higherlower/start", combinedMiddleware(higherLowerHandler.StartGame))
	http.HandleFunc("/api/game/higherlower/answer", combinedMiddleware(higherLowerHandler.SubmitAnswer))
	http.HandleFunc("/api/game/higherlower/score", combinedMiddleware(higherLowerHandler.GetScore))

	// Worldle game routes
	http.HandleFunc("/api/game/worldle/start", combinedMiddleware(worldleHandler.StartGame))
	http.HandleFunc("/api/game/worldle/guess", combinedMiddleware(worldleHandler.SubmitGuess))
	http.HandleFunc("/api/game/worldle/state", combinedMiddleware(worldleHandler.GetState))

	// Facts game routes
	http.HandleFunc("/api/game/facts/start", combinedMiddleware(factsHandler.StartGame))
	http.HandleFunc("/api/game/facts/guess", combinedMiddleware(factsHandler.SubmitGuess))
	http.HandleFunc("/api/game/facts/skip", combinedMiddleware(factsHandler.Skip))
	http.HandleFunc("/api/game/facts/next", combinedMiddleware(factsHandler.NextRound))

	// Debug/Browse routes
	http.HandleFunc("/api/debug/countries", combinedMiddleware(debugHandler.GetAllCountries))
	http.HandleFunc("/api/debug/geojson", combinedMiddleware(debugHandler.GetCountryGeoJSON))
	http.HandleFunc("/api/debug/geojson/all", combinedMiddleware(debugHandler.GetAllGeoJSON))
	http.HandleFunc("/api/debug/geojson/world", combinedMiddleware(debugHandler.GetWorldGeoJSON))

	// Health check (no rate limiting for monitoring, but include CORS and logging)
	http.HandleFunc("/api/health", LoggingMiddleware(corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})))

	// Stats/metrics endpoint (no rate limiting for monitoring, but include CORS and logging)
	http.HandleFunc("/api/stats", LoggingMiddleware(corsMiddleware(GetStatsHandler)))

	// Multiplayer routes
	rm := multiplayer.GetRoomManager()
	
	// Set up game engine factory to avoid import cycles
	multiplayer.SetEngineFactory(func(mode multiplayer.GameMode, config multiplayer.RoomConfig, countries []models.Country) (multiplayer.GameEngine, error) {
		switch mode {
		case multiplayer.GameModeFlag:
			return adapters.NewFlagGameEngine(config, countries)
		case multiplayer.GameModeShape:
			return adapters.NewShapeGameEngine(config, countries)
		case multiplayer.GameModeCapital:
			return adapters.NewCapitalGameEngine(config, countries)
		case multiplayer.GameModeHigherLower:
			return adapters.NewHigherLowerGameEngine(config, countries)
		case multiplayer.GameModeFacts:
			return adapters.NewFactsGameEngine(config, countries)
		case multiplayer.GameModeWorldle:
			return adapters.NewWorldleGameEngine(config, countries)
		default:
			return nil, fmt.Errorf("unsupported game mode: %s", mode)
		}
	})
	
	hub := multiplayer.NewHub(rm)
	go hub.Run()
	multiplayer.SetupMultiplayerRoutes(rm, hub, corsMiddleware, LoggingMiddleware)

	log.Println("API routes configured")
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
