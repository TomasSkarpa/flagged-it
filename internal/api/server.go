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
	flagColorHandler := &FlagColorGameHandler{}
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

				// Allow all Vercel preview deployments (any subdomain of vercel.app)
				if strings.HasSuffix(origin, ".vercel.app") && strings.HasPrefix(origin, "https://") {
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

	// Combine logging, rate limiting and CORS middleware
	// IMPORTANT: CORS must run first to handle OPTIONS preflight requests
	// Rate limiting should skip OPTIONS requests
	// Logging tracks all requests
	combinedMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return LoggingMiddleware(corsMiddleware(rateLimitMiddleware(next)))
	}

	// Flag game routes
	http.HandleFunc("/api/game/flag/start", combinedMiddleware(serialize(&flagGameLock, flagHandler.StartGame)))
	http.HandleFunc("/api/game/flag/question", combinedMiddleware(serialize(&flagGameLock, flagHandler.GetQuestion)))
	http.HandleFunc("/api/game/flag/answer", combinedMiddleware(serialize(&flagGameLock, flagHandler.SubmitAnswer)))
	http.HandleFunc("/api/game/flag/score", combinedMiddleware(serialize(&flagGameLock, flagHandler.GetScore)))

	// Shape game routes
	http.HandleFunc("/api/game/shape/start", combinedMiddleware(serialize(&shapeGameLock, shapeHandler.StartGame)))
	http.HandleFunc("/api/game/shape/question", combinedMiddleware(serialize(&shapeGameLock, shapeHandler.GetQuestion)))
	http.HandleFunc("/api/game/shape/answer", combinedMiddleware(serialize(&shapeGameLock, shapeHandler.SubmitAnswer)))
	http.HandleFunc("/api/game/shape/score", combinedMiddleware(serialize(&shapeGameLock, shapeHandler.GetScore)))

	// Capital game routes
	http.HandleFunc("/api/game/capital/start", combinedMiddleware(serialize(&capitalGameLock, capitalHandler.StartGame)))
	http.HandleFunc("/api/game/capital/question", combinedMiddleware(serialize(&capitalGameLock, capitalHandler.GetQuestion)))
	http.HandleFunc("/api/game/capital/answer", combinedMiddleware(serialize(&capitalGameLock, capitalHandler.SubmitAnswer)))
	http.HandleFunc("/api/game/capital/score", combinedMiddleware(serialize(&capitalGameLock, capitalHandler.GetScore)))

	// Higher/Lower game routes
	http.HandleFunc("/api/game/higherlower/start", combinedMiddleware(serialize(&higherLowerGameLock, higherLowerHandler.StartGame)))
	http.HandleFunc("/api/game/higherlower/answer", combinedMiddleware(serialize(&higherLowerGameLock, higherLowerHandler.SubmitAnswer)))
	http.HandleFunc("/api/game/higherlower/score", combinedMiddleware(serialize(&higherLowerGameLock, higherLowerHandler.GetScore)))

	// Worldle game routes
	http.HandleFunc("/api/game/worldle/start", combinedMiddleware(serialize(&worldleGameLock, worldleHandler.StartGame)))
	http.HandleFunc("/api/game/worldle/guess", combinedMiddleware(serialize(&worldleGameLock, worldleHandler.SubmitGuess)))
	http.HandleFunc("/api/game/worldle/state", combinedMiddleware(serialize(&worldleGameLock, worldleHandler.GetState)))

	// Facts game routes
	http.HandleFunc("/api/game/facts/start", combinedMiddleware(serialize(&factsGameLock, factsHandler.StartGame)))
	http.HandleFunc("/api/game/facts/guess", combinedMiddleware(serialize(&factsGameLock, factsHandler.SubmitGuess)))
	http.HandleFunc("/api/game/facts/skip", combinedMiddleware(serialize(&factsGameLock, factsHandler.Skip)))
	http.HandleFunc("/api/game/facts/next", combinedMiddleware(serialize(&factsGameLock, factsHandler.NextRound)))

	// Flag color quiz
	http.HandleFunc("/api/game/flag-color/start", combinedMiddleware(serialize(&flagColorGameLock, flagColorHandler.StartGame)))
	http.HandleFunc("/api/game/flag-color/question", combinedMiddleware(serialize(&flagColorGameLock, flagColorHandler.GetQuestion)))
	http.HandleFunc("/api/game/flag-color/answer", combinedMiddleware(serialize(&flagColorGameLock, flagColorHandler.SubmitAnswer)))
	http.HandleFunc("/api/game/flag-color/score", combinedMiddleware(serialize(&flagColorGameLock, flagColorHandler.GetScore)))

	// Countries and GeoJSON (used by library, games, map)
	http.HandleFunc("/api/countries", combinedMiddleware(debugHandler.GetAllCountries))
	http.HandleFunc("/api/geojson", combinedMiddleware(debugHandler.GetCountryGeoJSON))
	http.HandleFunc("/api/geojson/all", combinedMiddleware(debugHandler.GetAllGeoJSON))
	http.HandleFunc("/api/geojson/world", combinedMiddleware(debugHandler.GetWorldGeoJSON))

	// Health check (no rate limiting for monitoring, but include CORS and logging)
	http.HandleFunc("/api/health", LoggingMiddleware(corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})))

	// Stats/metrics endpoint (no rate limiting for monitoring, but include CORS and logging)
	http.HandleFunc("/api/stats", LoggingMiddleware(corsMiddleware(GetStatsHandler)))

	startSessionSweeper()

	log.Println("API routes configured")
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
