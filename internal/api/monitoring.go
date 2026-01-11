package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks server performance metrics
type Metrics struct {
	TotalRequests    uint64            // Total number of requests processed
	ActiveConnections int32            // Current active connections
	ResponseTimeSum   uint64            // Sum of all response times in milliseconds
	SlowRequests      uint64            // Count of requests taking > 1 second
	RequestsByStatus  map[int]uint64    // Requests grouped by HTTP status code
	RequestsByMethod  map[string]uint64 // Requests grouped by HTTP method
	RequestsByPath    map[string]uint64 // Requests grouped by path
	ActiveIPs         map[string]int64  // Active connections by IP address (last seen timestamp)
	mu                sync.RWMutex
}

var globalMetrics = &Metrics{
	RequestsByStatus: make(map[int]uint64),
	RequestsByMethod: make(map[string]uint64),
	RequestsByPath:   make(map[string]uint64),
	ActiveIPs:        make(map[string]int64),
}

// GetMetrics returns a copy of current metrics
func GetMetrics() Metrics {
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()

	// Copy maps
	requestsByStatus := make(map[int]uint64)
	for k, v := range globalMetrics.RequestsByStatus {
		requestsByStatus[k] = v
	}

	requestsByMethod := make(map[string]uint64)
	for k, v := range globalMetrics.RequestsByMethod {
		requestsByMethod[k] = v
	}

	requestsByPath := make(map[string]uint64)
	for k, v := range globalMetrics.RequestsByPath {
		requestsByPath[k] = v
	}

	// Copy active IPs with recent connections only (within last 5 minutes)
	activeIPs := make(map[string]int64)
	now := time.Now().Unix()
	for ip, lastSeen := range globalMetrics.ActiveIPs {
		if now-lastSeen < 300 { // 5 minutes
			activeIPs[ip] = lastSeen
		}
	}

	return Metrics{
		TotalRequests:     atomic.LoadUint64(&globalMetrics.TotalRequests),
		ActiveConnections: atomic.LoadInt32(&globalMetrics.ActiveConnections),
		ResponseTimeSum:   atomic.LoadUint64(&globalMetrics.ResponseTimeSum),
		SlowRequests:      atomic.LoadUint64(&globalMetrics.SlowRequests),
		RequestsByStatus:  requestsByStatus,
		RequestsByMethod:  requestsByMethod,
		RequestsByPath:    requestsByPath,
		ActiveIPs:         activeIPs,
	}
}

// incrementCounter safely increments a counter in a map
func incrementCounter(m map[string]uint64, key string) {
	globalMetrics.mu.Lock()
	m[key]++
	globalMetrics.mu.Unlock()
}

// incrementStatusCounter safely increments a status code counter
func incrementStatusCounter(code int) {
	globalMetrics.mu.Lock()
	globalMetrics.RequestsByStatus[code]++
	globalMetrics.mu.Unlock()
}

// updateActiveIP updates the last seen time for an IP
func updateActiveIP(ip string) {
	globalMetrics.mu.Lock()
	globalMetrics.ActiveIPs[ip] = time.Now().Unix()
	globalMetrics.mu.Unlock()
}

// cleanOldIPs removes IPs that haven't been seen in the last 30 minutes
func cleanOldIPs() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		globalMetrics.mu.Lock()
		now := time.Now().Unix()
		for ip, lastSeen := range globalMetrics.ActiveIPs {
			if now-lastSeen > 1800 { // 30 minutes
				delete(globalMetrics.ActiveIPs, ip)
			}
		}
		globalMetrics.mu.Unlock()
	}
}

// LoggingMiddleware logs requests with response time and tracks metrics
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment active connections
		atomic.AddInt32(&globalMetrics.ActiveConnections, 1)
		defer atomic.AddInt32(&globalMetrics.ActiveConnections, -1)

		// Track request
		atomic.AddUint64(&globalMetrics.TotalRequests, 1)
		incrementCounter(globalMetrics.RequestsByMethod, r.Method)
		incrementCounter(globalMetrics.RequestsByPath, r.URL.Path)

		// Get IP and update active IPs
		ip := getIP(r)
		updateActiveIP(ip)

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next(wrapper, r)

		// Calculate response time
		duration := time.Since(start)
		responseTimeMs := uint64(duration.Milliseconds())

		// Track response time
		atomic.AddUint64(&globalMetrics.ResponseTimeSum, responseTimeMs)

		// Track slow requests (> 1 second)
		if duration > time.Second {
			atomic.AddUint64(&globalMetrics.SlowRequests, 1)
			log.Printf("[SLOW] %s %s %s - %d - %v (IP: %s)", 
				r.Method, r.URL.Path, r.Proto, wrapper.statusCode, duration, ip)
		}

		// Track status code
		incrementStatusCounter(wrapper.statusCode)

		// Log request (only log API requests, not static assets to reduce noise)
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			log.Printf("[REQUEST] %s %s - %d - %v (IP: %s, Active: %d)", 
				r.Method, r.URL.Path, wrapper.statusCode, duration, ip, atomic.LoadInt32(&globalMetrics.ActiveConnections))
		}
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetStatsHandler returns current server statistics
func GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := GetMetrics()

	// Calculate average response time
	var avgResponseTime float64
	if metrics.TotalRequests > 0 {
		avgResponseTime = float64(metrics.ResponseTimeSum) / float64(metrics.TotalRequests)
	}

	// Count unique active IPs
	uniqueIPs := len(metrics.ActiveIPs)

	// Build response
	stats := map[string]interface{}{
		"totalRequests":    metrics.TotalRequests,
		"activeConnections": metrics.ActiveConnections,
		"uniqueActiveIPs":   uniqueIPs,
		"avgResponseTimeMs": avgResponseTime,
		"slowRequests":      metrics.SlowRequests,
		"requestsByStatus":  metrics.RequestsByStatus,
		"requestsByMethod":  metrics.RequestsByMethod,
		"requestsByPath":    metrics.RequestsByPath,
		"timestamp":         time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	
	// Simple JSON encoding (could use json package for prettier output)
	log.Printf("[STATS] Stats requested - Active: %d, Total: %d, AvgTime: %.2fms, IPs: %d",
		metrics.ActiveConnections, metrics.TotalRequests, avgResponseTime, uniqueIPs)

	json.NewEncoder(w).Encode(stats)
}

// Initialize monitoring cleanup goroutine
func init() {
	go cleanOldIPs()
}