package api

import (
	"log"
	"net/http"
	"os"
	"strings"
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
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin should be allowed
			shouldAllow := false

			if origin != "" {
				// Explicitly allow localhost:5173 (Svelte dev server default)
				if origin == "http://localhost:5173" {
					shouldAllow = true
				}
				// Allow localhost with any port
				if strings.HasPrefix(origin, "http://localhost:") {
					shouldAllow = true
				}
				// Allow 127.0.0.1 with any port
				if strings.HasPrefix(origin, "http://127.0.0.1:") {
					shouldAllow = true
				}
				// Allow bare localhost or 127.0.0.1
				if origin == "http://localhost" || origin == "http://127.0.0.1" {
					shouldAllow = true
				}

				// Allow local network IPs
				if strings.HasPrefix(origin, "http://192.168.") ||
					strings.HasPrefix(origin, "http://10.") ||
					strings.HasPrefix(origin, "http://172.16.") ||
					strings.HasPrefix(origin, "http://172.17.") ||
					strings.HasPrefix(origin, "http://172.18.") ||
					strings.HasPrefix(origin, "http://172.19.") ||
					strings.HasPrefix(origin, "http://172.2") ||
					strings.HasPrefix(origin, "http://172.3") {
					shouldAllow = true
				}

				// Allow production domains
				if origin == "https://flaggedit.vercel.app" ||
					origin == "http://flaggedit.vercel.app" ||
					origin == "https://flaggedit.app" ||
					origin == "http://flaggedit.app" {
					shouldAllow = true
				}
			}

			// Handle preflight OPTIONS request FIRST - before anything else
			if r.Method == http.MethodOptions {
				// Set CORS headers for preflight - browser requires these
				// Only set Allow-Origin if the origin is allowed
				if shouldAllow && origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				// Always set these headers for OPTIONS (even if origin is not allowed)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Max-Age", "3600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Set CORS headers for actual requests
			if shouldAllow && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "3600")

			next(w, r)
		}
	}

	// Combine rate limiting and CORS middleware
	// IMPORTANT: CORS must run first to handle OPTIONS preflight requests
	// Rate limiting should skip OPTIONS requests
	combinedMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return corsMiddleware(rateLimitMiddleware(next))
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

	// Health check (no rate limiting for monitoring, but include CORS)
	http.HandleFunc("/api/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))

	log.Println("API routes configured")
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
