package api

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func SetupRoutes() {
	flagHandler := &FlagGameHandler{}
	shapeHandler := &ShapeGameHandler{}
	capitalHandler := &CapitalGameHandler{}
	higherLowerHandler := &HigherLowerHandler{}
	debugHandler := &DebugHandler{}

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

	// Flag game routes
	http.HandleFunc("/api/game/flag/start", corsMiddleware(flagHandler.StartGame))
	http.HandleFunc("/api/game/flag/question", corsMiddleware(flagHandler.GetQuestion))
	http.HandleFunc("/api/game/flag/answer", corsMiddleware(flagHandler.SubmitAnswer))
	http.HandleFunc("/api/game/flag/score", corsMiddleware(flagHandler.GetScore))

	// Shape game routes
	http.HandleFunc("/api/game/shape/start", corsMiddleware(shapeHandler.StartGame))
	http.HandleFunc("/api/game/shape/question", corsMiddleware(shapeHandler.GetQuestion))
	http.HandleFunc("/api/game/shape/answer", corsMiddleware(shapeHandler.SubmitAnswer))
	http.HandleFunc("/api/game/shape/score", corsMiddleware(shapeHandler.GetScore))

	// Capital game routes
	http.HandleFunc("/api/game/capital/start", corsMiddleware(capitalHandler.StartGame))
	http.HandleFunc("/api/game/capital/question", corsMiddleware(capitalHandler.GetQuestion))
	http.HandleFunc("/api/game/capital/answer", corsMiddleware(capitalHandler.SubmitAnswer))
	http.HandleFunc("/api/game/capital/score", corsMiddleware(capitalHandler.GetScore))

	// Higher/Lower game routes
	http.HandleFunc("/api/game/higherlower/start", corsMiddleware(higherLowerHandler.StartGame))
	http.HandleFunc("/api/game/higherlower/answer", corsMiddleware(higherLowerHandler.SubmitAnswer))
	http.HandleFunc("/api/game/higherlower/score", corsMiddleware(higherLowerHandler.GetScore))

	// Debug/Browse routes
	http.HandleFunc("/api/debug/countries", corsMiddleware(debugHandler.GetAllCountries))
	http.HandleFunc("/api/debug/geojson", corsMiddleware(debugHandler.GetCountryGeoJSON))
	http.HandleFunc("/api/debug/geojson/all", corsMiddleware(debugHandler.GetAllGeoJSON))

	// Health check
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
