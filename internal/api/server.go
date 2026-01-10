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

	// CORS middleware
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is from localhost or local network
			isLocal := false
			if origin != "" {
				// Allow localhost variants
				if strings.HasPrefix(origin, "http://localhost:") ||
					strings.HasPrefix(origin, "http://127.0.0.1:") {
					isLocal = true
				}

				// Allow local network IPs (192.168.x.x, 10.x.x.x, 172.16-31.x.x)
				if strings.HasPrefix(origin, "http://192.168.") ||
					strings.HasPrefix(origin, "http://10.") ||
					strings.HasPrefix(origin, "http://172.16.") ||
					strings.HasPrefix(origin, "http://172.17.") ||
					strings.HasPrefix(origin, "http://172.18.") ||
					strings.HasPrefix(origin, "http://172.19.") ||
					strings.HasPrefix(origin, "http://172.2") ||
					strings.HasPrefix(origin, "http://172.3") {
					isLocal = true
				}
			}

			// Allow localhost for dev, local network IPs, and vercel for production
			allowedOrigins := []string{
				"http://localhost:5173",
				"http://localhost:3000",
				"https://flaggedit.vercel.app",
				"http://flaggedit.vercel.app", // Fallback for HTTP
			}

			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			// Also allow if it's a local network request
			if isLocal {
				allowed = true
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "3600")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next(w, r)
		}
	}

	// Combine rate limiting and CORS middleware
	combinedMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return rateLimitMiddleware(corsMiddleware(next))
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
	http.HandleFunc("/api/game/facts/next", combinedMiddleware(factsHandler.NextRound))

	// Debug/Browse routes
	http.HandleFunc("/api/debug/countries", combinedMiddleware(debugHandler.GetAllCountries))
	http.HandleFunc("/api/debug/geojson", combinedMiddleware(debugHandler.GetCountryGeoJSON))
	http.HandleFunc("/api/debug/geojson/all", combinedMiddleware(debugHandler.GetAllGeoJSON))

	// Health check (no rate limiting for monitoring)
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("API routes configured")
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
