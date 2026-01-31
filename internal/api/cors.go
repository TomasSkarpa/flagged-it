package api

import (
	"net/http"
	"strings"
)

// CORSMiddleware returns a CORS middleware function
func CORSMiddleware() func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
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
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Max-Age", "3600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Set CORS headers for actual requests
			if shouldAllow && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "3600")

			next(w, r)
		}
	}
}
