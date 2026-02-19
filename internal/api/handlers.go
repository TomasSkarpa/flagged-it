package api

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/facts"
	"flagged-it/internal/games/guessing"
	"flagged-it/internal/games/round"
	_ "flagged-it/internal/games/round/adapters"
	"flagged-it/internal/utils"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type FlagGameHandler struct{}

type GameQuestion struct {
	FlagURL    string           `json:"flagUrl"`
	Options    []models.Country `json:"options"`
	QuestionID string           `json:"questionId"`
}

type GameAnswer struct {
	Correct     bool   `json:"correct"`
	CorrectCCA2 string `json:"correctCca2"`
	CorrectName string `json:"correctName"`
}

type GameSession struct {
	SessionID string          `json:"sessionId"`
	Score     int             `json:"score"`
	Total     int             `json:"total"`
	Region    string          `json:"region"`
	Locale   string          `json:"-"` // User's locale for translations
	Runner   round.RoundRunner `json:"-"`
}

var (
	sessions      = make(map[string]*GameSession)
	sessionsMutex sync.RWMutex
	countries     = []models.Country{}
	factsData     = make(map[string]models.CountryFacts)
)

func init() {
	countries = data.LoadCountries()
	factsData = data.LoadCountryFacts()
	rand.Seed(time.Now().UnixNano())
}

// getTranslatedCountry creates a copy of the country with translated name in Common field
func getTranslatedCountry(country models.Country, locale string) models.Country {
	// Get the translated name first - use pointer to ensure method works correctly
	translatedName := (&country).GetTranslatedName(locale)

	// Log warning only if translation actually failed (translation doesn't exist in map)
	if locale != "en" && country.Name.Translations != nil {
		if _, exists := country.Name.Translations[locale]; !exists && translatedName == country.Name.Common {
			log.Printf("WARNING: Translation missing for locale '%s', country '%s' (CCA2: %s)",
				locale, country.Name.Common, country.CCA2)
		}
	}

	// Create a new country struct with translated name
	translated := country
	// Set the common field to the translated name
	translated.Name.Common = translatedName
	// Clear the translations map to avoid confusion - the common field is already translated
	// This ensures the frontend uses the translated common field
	translated.Name.Translations = nil
	return translated
}

// getTranslatedCountries translates an array of countries
func getTranslatedCountries(countries []models.Country, locale string) []models.Country {
	result := make([]models.Country, len(countries))
	for i, country := range countries {
		result[i] = getTranslatedCountry(country, locale)
	}
	return result
}

func (h *FlagGameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Region string `json:"region"`
		Locale string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	runner, err := round.NewRunner(round.GameTypeFlag, round.RoundOptions{Region: req.Region, Locale: locale})
	if err != nil {
		http.Error(w, "Failed to start game", http.StatusInternalServerError)
		return
	}
	if err := runner.StartRound(round.RoundOptions{Region: req.Region, Locale: locale}); err != nil {
		http.Error(w, "No countries available", http.StatusBadRequest)
		return
	}

	sessionID := generateSessionID()
	session := &GameSession{
		SessionID: sessionID,
		Score:     0,
		Total:     0,
		Region:    req.Region,
		Locale:   locale,
		Runner:   runner,
	}
	sessionsMutex.Lock()
	sessions[sessionID] = session
	sessionsMutex.Unlock()

	payload, _ := runner.Payload(locale)
	var question map[string]interface{}
	if payload != nil {
		_ = json.Unmarshal(payload.Data, &question)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId": sessionID,
		"question":  question,
	})
}

func (h *FlagGameHandler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	sessionsMutex.RLock()
	session, exists := sessions[sessionID]
	sessionsMutex.RUnlock()
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if locale != "" {
		session.Locale = locale
	} else if session.Locale == "" {
		session.Locale = "en"
	}
	locale = session.Locale

	payload, _ := session.Runner.Payload(locale)
	var question map[string]interface{}
	if payload != nil {
		_ = json.Unmarshal(payload.Data, &question)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)
}

func (h *FlagGameHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID  string `json:"sessionId"`
		QuestionID string `json:"questionId"`
		AnswerCCA2 string `json:"answerCca2"`
		Locale     string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	sessionsMutex.RLock()
	session, exists := sessions[req.SessionID]
	sessionsMutex.RUnlock()
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	inputData, _ := json.Marshal(map[string]string{"cca2": req.AnswerCCA2})
	result, err := session.Runner.Submit(&round.SubmitInput{Data: inputData})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Score += result.ScoreDelta
	if result.RoundComplete {
		session.Total++
		if session.Total < 10 {
			_ = session.Runner.StartRound(round.RoundOptions{Region: session.Region, Locale: session.Locale})
		}
	}

	correctCCA2, correctName := "", ""
	if m, ok := result.RevealedAnswer.(map[string]string); ok {
		correctCCA2 = m["correctCca2"]
		correctName = m["correctName"]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"correct":     result.Correct,
		"correctCca2": correctCCA2,
		"correctName": correctName,
		"score":       session.Score,
		"total":       session.Total,
		"finished":    session.Total >= 10,
	})
}

func (h *FlagGameHandler) GetScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	sessionsMutex.RLock()
	session, exists := sessions[sessionID]
	sessionsMutex.RUnlock()
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score": session.Score,
		"total": session.Total,
	})
}

func generateSessionID() string {
	return time.Now().Format("20060102150405") + "-" + string(rune(rand.Intn(1000)))
}

func generateQuestionID() string {
	return time.Now().Format("20060102150405.000000")
}

// ============================================
// Debug/Browse Handlers
// ============================================

type DebugHandler struct{}

// GetAllCountries returns all countries with their data
func (h *DebugHandler) GetAllCountries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get locale from query parameter or default to "en"
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	// Translate country names based on locale
	translatedCountries := getTranslatedCountries(countries, locale)

	// Set cache headers - countries data is static and only changes on rebuild
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=604800") // Cache for 1 week (data only changes on rebuild)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"countries": translatedCountries,
		"total":     len(translatedCountries),
	})
}

// GetCountryGeoJSON returns GeoJSON for a specific country
func (h *DebugHandler) GetCountryGeoJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cca3 := r.URL.Query().Get("cca3")
	if cca3 == "" {
		http.Error(w, "cca3 parameter required", http.StatusBadRequest)
		return
	}

	geoData, err := data.LoadGeoData(cca3)
	if err != nil {
		http.Error(w, "GeoJSON not found for country: "+cca3, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(geoData)
}

// GetAllGeoJSON returns list of all available GeoJSON country codes
func (h *DebugHandler) GetAllGeoJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return countries that have GeoJSON data
	countriesWithGeo := []models.Country{}
	for _, country := range countries {
		_, err := data.LoadGeoData(country.CCA3)
		if err == nil {
			countriesWithGeo = append(countriesWithGeo, country)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"countries": countriesWithGeo,
		"total":     len(countriesWithGeo),
	})
}

// GetWorldGeoJSON returns the world GeoJSON data
func (h *DebugHandler) GetWorldGeoJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	worldGeo, err := data.LoadWorldGeoData()
	if err != nil {
		http.Error(w, "Failed to load world GeoJSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(worldGeo)
}

// ============================================
// Shape Game Handler
// ============================================

type ShapeGameHandler struct{}

type ShapeGameSession struct {
	SessionID string            `json:"sessionId"`
	Score     int               `json:"score"`
	Total     int               `json:"total"`
	Region    string            `json:"region"`
	Locale    string            `json:"-"`
	Runner    round.RoundRunner `json:"-"`
}

type ShapeQuestion struct {
	GeoJSON    interface{}      `json:"geoJson"`
	Options    []models.Country `json:"options"`
	QuestionID string           `json:"questionId"`
}

var (
	shapeSessions      = make(map[string]*ShapeGameSession)
	shapeSessionsMutex sync.RWMutex
)

// countriesWithGeoCache caches the list of countries with geo data per region
var (
	countriesWithGeoCache      = make(map[string][]models.Country)
	countriesWithGeoCacheMutex sync.RWMutex
	geoAvailableSet            = make(map[string]bool) // Set of CCA3 codes with geo data
	geoAvailableOnce           sync.Once
)

// countriesWithGeo returns countries that have geo data available (cached)
func countriesWithGeo(region string) []models.Country {
	// Pre-populate list of countries with geo data (once, lazy load on first use)
	geoAvailableOnce.Do(func() {
		// Build set by trying to load geo for each country
		// This will populate the cache as we check
		for _, country := range countries {
			if country.CCA3 != "" {
				_, err := data.LoadGeoData(country.CCA3)
				if err == nil {
					geoAvailableSet[country.CCA3] = true
				}
			}
		}
	})

	// Check cache first
	cacheKey := region
	if cacheKey == "" {
		cacheKey = "World"
	}

	countriesWithGeoCacheMutex.RLock()
	if cached, exists := countriesWithGeoCache[cacheKey]; exists {
		countriesWithGeoCacheMutex.RUnlock()
		return cached
	}
	countriesWithGeoCacheMutex.RUnlock()

	// Build result
	var result []models.Country
	for _, country := range countries {
		if region != "" && region != "World" && country.Region != region {
			continue
		}
		if country.CCA3 != "" && geoAvailableSet[country.CCA3] {
			result = append(result, country)
		}
	}

	// Cache the result
	countriesWithGeoCacheMutex.Lock()
	countriesWithGeoCache[cacheKey] = result
	countriesWithGeoCacheMutex.Unlock()

	return result
}

func (h *ShapeGameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Region string `json:"region"`
		Locale string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	runner, err := round.NewRunner(round.GameTypeShape, round.RoundOptions{Region: req.Region, Locale: locale})
	if err != nil {
		http.Error(w, "Failed to start game", http.StatusInternalServerError)
		return
	}
	if err := runner.StartRound(round.RoundOptions{Region: req.Region, Locale: locale}); err != nil {
		http.Error(w, "No countries available", http.StatusBadRequest)
		return
	}

	sessionID := generateSessionID()
	session := &ShapeGameSession{
		SessionID: sessionID,
		Score:     0,
		Total:     0,
		Region:    req.Region,
		Locale:    locale,
		Runner:    runner,
	}
	shapeSessions[sessionID] = session

	payload, _ := runner.Payload(locale)
	var question map[string]interface{}
	if payload != nil {
		_ = json.Unmarshal(payload.Data, &question)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId": sessionID,
		"question":  question,
	})
}

func (h *ShapeGameHandler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	session, exists := shapeSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if locale != "" {
		session.Locale = locale
	} else if session.Locale == "" {
		session.Locale = "en"
	}
	locale = session.Locale

	payload, _ := session.Runner.Payload(locale)
	var question map[string]interface{}
	if payload != nil {
		_ = json.Unmarshal(payload.Data, &question)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)
}

func (h *ShapeGameHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID  string `json:"sessionId"`
		QuestionID string `json:"questionId"`
		AnswerCCA2 string `json:"answerCca2"`
		Locale     string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session, exists := shapeSessions[req.SessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	inputData, _ := json.Marshal(map[string]string{"cca2": req.AnswerCCA2})
	result, err := session.Runner.Submit(&round.SubmitInput{Data: inputData})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Score += result.ScoreDelta
	if result.RoundComplete {
		session.Total++
		if session.Total < 10 {
			_ = session.Runner.StartRound(round.RoundOptions{Region: session.Region, Locale: session.Locale})
		}
	}

	correctCCA2, correctName := "", ""
	if m, ok := result.RevealedAnswer.(map[string]string); ok {
		correctCCA2 = m["correctCca2"]
		correctName = m["correctName"]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"correct":     result.Correct,
		"correctCca2": correctCCA2,
		"correctName": correctName,
		"score":       session.Score,
		"total":       session.Total,
		"finished":    session.Total >= 10,
	})
}

func (h *ShapeGameHandler) GetScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	session, exists := shapeSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score": session.Score,
		"total": session.Total,
	})
}

// ============================================
// Capital Game Handler
// ============================================

type CapitalGameHandler struct{}

type CapitalGameSession struct {
	SessionID string            `json:"sessionId"`
	Score     int               `json:"score"`
	Total     int               `json:"total"`
	Region    string            `json:"region"`
	Locale    string            `json:"-"`
	Runner    round.RoundRunner `json:"-"`
}

type CapitalQuestion struct {
	CountryName string   `json:"countryName"`
	CountryCCA2 string   `json:"countryCca2"`
	Options     []string `json:"options"`
	QuestionID  string   `json:"questionId"`
}

var (
	capitalSessions      = make(map[string]*CapitalGameSession)
	capitalSessionsMutex sync.RWMutex
)

func (h *CapitalGameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Region string `json:"region"`
		Locale string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	runner, err := round.NewRunner(round.GameTypeCapital, round.RoundOptions{Region: req.Region, Locale: locale})
	if err != nil {
		http.Error(w, "Failed to start game", http.StatusInternalServerError)
		return
	}
	if err := runner.StartRound(round.RoundOptions{Region: req.Region, Locale: locale}); err != nil {
		http.Error(w, "Not enough countries with capitals available", http.StatusBadRequest)
		return
	}

	sessionID := generateSessionID()
	session := &CapitalGameSession{
		SessionID: sessionID,
		Score:     0,
		Total:     0,
		Region:    req.Region,
		Locale:    locale,
		Runner:    runner,
	}
	capitalSessions[sessionID] = session

	payload, _ := runner.Payload(locale)
	var question map[string]interface{}
	if payload != nil {
		_ = json.Unmarshal(payload.Data, &question)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId": sessionID,
		"question":  question,
	})
}

func (h *CapitalGameHandler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	session, exists := capitalSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if locale != "" {
		session.Locale = locale
	} else if session.Locale == "" {
		session.Locale = "en"
	}
	locale = session.Locale

	payload, _ := session.Runner.Payload(locale)
	var question map[string]interface{}
	if payload != nil {
		_ = json.Unmarshal(payload.Data, &question)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)
}

func (h *CapitalGameHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID  string `json:"sessionId"`
		QuestionID string `json:"questionId"`
		Answer     string `json:"answer"`
		Locale     string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session, exists := capitalSessions[req.SessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	inputData, _ := json.Marshal(map[string]string{"answer": req.Answer})
	result, err := session.Runner.Submit(&round.SubmitInput{Data: inputData})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Score += result.ScoreDelta
	if result.RoundComplete {
		session.Total++
		if session.Total < 10 {
			_ = session.Runner.StartRound(round.RoundOptions{Region: session.Region, Locale: session.Locale})
		}
	}

	correctCapital, correctCountry := "", ""
	if m, ok := result.RevealedAnswer.(map[string]string); ok {
		correctCapital = m["correctCapital"]
		correctCountry = m["correctCountry"]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"correct":        result.Correct,
		"correctCapital": correctCapital,
		"correctCountry": correctCountry,
		"score":          session.Score,
		"total":          session.Total,
	})
}

func (h *CapitalGameHandler) GetScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	session, exists := capitalSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score": session.Score,
		"total": session.Total,
	})
}

// ============================================
// Higher Lower Game Handler
// ============================================

type HigherLowerHandler struct{}

type HigherLowerCategory string

const (
	CategoryPopulation HigherLowerCategory = "population"
	CategoryArea       HigherLowerCategory = "area"
	CategoryContinents HigherLowerCategory = "continents"
)

type HigherLowerItem struct {
	Name     string  `json:"name"`
	Value    float64 `json:"value"`
	CCA2     string  `json:"cca2,omitempty"`
	ImageURL string  `json:"imageUrl,omitempty"`
}

type HigherLowerComparison struct {
	Left       HigherLowerItem     `json:"left"`
	Right      HigherLowerItem     `json:"right"`
	Category   HigherLowerCategory `json:"category"`
	ValueLabel string              `json:"valueLabel"`
}

type HigherLowerSession struct {
	SessionID    string              `json:"sessionId"`
	Score        int                 `json:"score"`
	HighScore    int                 `json:"highScore"`
	Category     HigherLowerCategory `json:"category"`
	Locale       string              `json:"-"`
	Runner       round.RoundRunner   `json:"-"`
	GameOver     bool                `json:"gameOver"`
	LastLeftVal  float64             `json:"-"`
	LastRightVal float64             `json:"-"`
}

var (
	higherLowerSessions      = make(map[string]*HigherLowerSession)
	higherLowerSessionsMutex sync.RWMutex
)

func (h *HigherLowerHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Category string `json:"category"`
		Locale   string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Category = "population"
	}

	category := HigherLowerCategory(req.Category)
	if category != CategoryPopulation && category != CategoryArea && category != CategoryContinents {
		category = CategoryPopulation
	}

	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	compType := "population"
	if category == CategoryArea {
		compType = "area"
	}

	runner, err := round.NewRunner(round.GameTypeHigherLower, round.RoundOptions{Locale: locale, ComparisonType: compType})
	if err != nil {
		http.Error(w, "Failed to start game", http.StatusInternalServerError)
		return
	}
	if err := runner.StartRound(round.RoundOptions{Locale: locale, ComparisonType: compType}); err != nil {
		http.Error(w, "Failed to start game", http.StatusInternalServerError)
		return
	}

	sessionID := generateSessionID()
	session := &HigherLowerSession{
		SessionID: sessionID,
		Score:     0,
		HighScore: 0,
		Category:  category,
		Locale:    locale,
		Runner:    runner,
		GameOver:  false,
	}
	higherLowerSessions[sessionID] = session

	comparison := higherLowerPayloadToComparison(session)
	if comparison == nil {
		http.Error(w, "Failed to get comparison", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId":  sessionID,
		"comparison": comparison,
		"score":      session.Score,
		"highScore":  session.HighScore,
	})
}

func (h *HigherLowerHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
		Answer    string `json:"answer"`
		Locale    string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session, exists := higherLowerSessions[req.SessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	if session.GameOver {
		http.Error(w, "Game is over", http.StatusBadRequest)
		return
	}

	inputData, _ := json.Marshal(map[string]string{"direction": req.Answer})
	result, err := session.Runner.Submit(&round.SubmitInput{Data: inputData})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Score += result.ScoreDelta
	if session.Score > session.HighScore {
		session.HighScore = session.Score
	}
	if !result.Correct {
		session.GameOver = true
	} else if result.RoundComplete {
		compType := "population"
		if session.Category == CategoryArea {
			compType = "area"
		}
		_ = session.Runner.StartRound(round.RoundOptions{Locale: session.Locale, ComparisonType: compType})
	}

	var nextComparison *HigherLowerComparison
	if result.Correct && !session.GameOver {
		nextComparison = higherLowerPayloadToComparison(session)
	}

	response := map[string]interface{}{
		"correct":    result.Correct,
		"leftValue":  session.LastLeftVal,
		"rightValue": session.LastRightVal,
		"score":      session.Score,
		"highScore":  session.HighScore,
		"gameOver":   session.GameOver,
	}
	if nextComparison != nil {
		response["nextComparison"] = nextComparison
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *HigherLowerHandler) GetScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	session, exists := higherLowerSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score":     session.Score,
		"highScore": session.HighScore,
		"gameOver":  session.GameOver,
	})
}

func higherLowerPayloadToComparison(session *HigherLowerSession) *HigherLowerComparison {
	payload, err := session.Runner.Payload(session.Locale)
	if err != nil || payload == nil {
		return nil
	}
	var data struct {
		CurrentCountry struct {
			Name       struct { Common string `json:"common"` }
			CCA2       string  `json:"cca2"`
			Population int     `json:"population"`
			Area       float64 `json:"area"`
		} `json:"currentCountry"`
		NextCountry struct {
			Name       struct { Common string `json:"common"` }
			CCA2       string  `json:"cca2"`
			Population int     `json:"population"`
			Area       float64 `json:"area"`
		} `json:"nextCountry"`
		ComparisonType string `json:"comparisonType"`
	}
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		return nil
	}
	var leftVal, rightVal float64
	if data.ComparisonType == "area" {
		leftVal = data.CurrentCountry.Area
		rightVal = data.NextCountry.Area
	} else {
		leftVal = float64(data.CurrentCountry.Population)
		rightVal = float64(data.NextCountry.Population)
	}
	session.LastLeftVal = leftVal
	session.LastRightVal = rightVal

	valueLabel := "population"
	if data.ComparisonType == "area" {
		valueLabel = "km² area"
	}

	return &HigherLowerComparison{
		Left: HigherLowerItem{
			Name:     data.CurrentCountry.Name.Common,
			Value:    leftVal,
			CCA2:     data.CurrentCountry.CCA2,
			ImageURL: "/assets/twemoji_flags_cca2/" + data.CurrentCountry.CCA2 + ".svg",
		},
		Right: HigherLowerItem{
			Name:     data.NextCountry.Name.Common,
			Value:    rightVal,
			CCA2:     data.NextCountry.CCA2,
			ImageURL: "/assets/twemoji_flags_cca2/" + data.NextCountry.CCA2 + ".svg",
		},
		Category:   session.Category,
		ValueLabel: valueLabel,
	}
}

// Worldle game handler
type WorldleGameHandler struct{}

type WorldleSession struct {
	SessionID string
	Locale    string // User's locale for translations
	Logic     *guessing.Logic
}

var (
	worldleSessions      = make(map[string]*WorldleSession)
	worldleSessionsMutex sync.RWMutex
)

func (h *WorldleGameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Locale string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body, use default locale
		req.Locale = "en"
	}

	// Default to English if locale not provided
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	// Filter countries (all countries available for worldle)
	filteredCountries := countries

	if len(filteredCountries) == 0 {
		http.Error(w, "No countries available", http.StatusBadRequest)
		return
	}

	// Create game logic instance
	gameLogic := guessing.NewLogic(filteredCountries)
	if err := gameLogic.NewGame(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID := generateSessionID()
	session := &WorldleSession{
		SessionID: sessionID,
		Locale:    locale,
		Logic:     gameLogic,
	}
	worldleSessions[sessionID] = session

	state := gameLogic.GetState()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId":  sessionID,
		"guessCount": len(state.Guesses),
		"isComplete": state.IsComplete,
	})
}

func (h *WorldleGameHandler) SubmitGuess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID   string `json:"sessionId"`
		CountryName string `json:"countryName"`
		Locale      string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session, exists := worldleSessions[req.SessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update locale if provided
	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	// Make guess using game logic
	result, err := session.Logic.MakeGuess(req.CountryName, locale)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !result.IsValidGuess {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isValidGuess": false,
			"error":        result.Error,
		})
		return
	}

	// Build response with guess entry details
	response := map[string]interface{}{
		"isValidGuess": result.IsValidGuess,
		"isCorrect":    result.IsCorrect,
		"guessCount":   result.GuessCount,
		"isComplete":   result.IsCorrect,
		"guessEntry": map[string]interface{}{
			"country": map[string]interface{}{
				"cca2":       result.GuessEntry.Country.CCA2,
				"name":       result.GuessEntry.Country.GetTranslatedName(locale),
				"flagUrl":    "/assets/twemoji_flags_cca2/" + result.GuessEntry.Country.CCA2 + ".svg",
				"continent":  result.GuessEntry.Continent,
				"population": result.GuessEntry.Country.Population,
				"area":       result.GuessEntry.Country.Area,
			},
			"isCorrect":        result.GuessEntry.IsCorrect,
			"continent":        result.GuessEntry.Continent,
			"continentCorrect": result.GuessEntry.ContinentCorrect,
			"population": map[string]interface{}{
				"value":     result.GuessEntry.Population.Value,
				"direction": result.GuessEntry.Population.Direction,
				"proximity": result.GuessEntry.Population.Proximity,
			},
			"area": map[string]interface{}{
				"value":     result.GuessEntry.Area.Value,
				"direction": result.GuessEntry.Area.Direction,
				"proximity": result.GuessEntry.Area.Proximity,
			},
		},
	}

	// Only reveal correct country if guess is correct
	if result.IsCorrect && result.CorrectCountry != nil {
		response["correctCountry"] = map[string]interface{}{
			"cca2":    result.CorrectCountry.CCA2,
			"name":    result.CorrectCountry.GetTranslatedName(locale),
			"flagUrl": "/assets/twemoji_flags_cca2/" + result.CorrectCountry.CCA2 + ".svg",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *WorldleGameHandler) GetState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	session, exists := worldleSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get locale from query parameter
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	state := session.Logic.GetState()

	// Build guesses array for response
	guesses := []map[string]interface{}{}
	for _, guessEntry := range state.Guesses {
		guesses = append(guesses, map[string]interface{}{
			"country": map[string]interface{}{
				"cca2":       guessEntry.Country.CCA2,
				"name":       guessEntry.Country.GetTranslatedName(locale),
				"flagUrl":    "/assets/twemoji_flags_cca2/" + guessEntry.Country.CCA2 + ".svg",
				"continent":  guessEntry.Continent,
				"population": guessEntry.Country.Population,
				"area":       guessEntry.Country.Area,
			},
			"isCorrect":        guessEntry.IsCorrect,
			"continent":        guessEntry.Continent,
			"continentCorrect": guessEntry.ContinentCorrect,
			"population": map[string]interface{}{
				"value":     guessEntry.Population.Value,
				"direction": guessEntry.Population.Direction,
				"proximity": guessEntry.Population.Proximity,
			},
			"area": map[string]interface{}{
				"value":     guessEntry.Area.Value,
				"direction": guessEntry.Area.Direction,
				"proximity": guessEntry.Area.Proximity,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId":  sessionID,
		"guesses":    guesses,
		"guessCount": len(state.Guesses),
		"isComplete": state.IsComplete,
	})
}

// Facts game handler
type FactsGameHandler struct{}

type FactsGameSession struct {
	SessionID string
	Locale    string // User's locale for translations
	Logic     *facts.Logic
}

var (
	factsGameSessions      = make(map[string]*FactsGameSession)
	factsGameSessionsMutex sync.RWMutex
)

func (h *FactsGameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Locale string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body, use default locale
		req.Locale = "en"
	}

	// Default to English if locale not provided
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	// Create game logic instance
	gameLogic := facts.NewLogic(countries, factsData)
	if err := gameLogic.NewRound(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get first fact
	currentFact, err := gameLogic.GetCurrentFact()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID := generateSessionID()
	session := &FactsGameSession{
		SessionID: sessionID,
		Locale:    locale,
		Logic:     gameLogic,
	}
	factsGameSessions[sessionID] = session

	state := gameLogic.GetState()

	// Format fact with number prefix
	formattedFact := fmt.Sprintf("Fact %d: %s", state.CurrentFact, currentFact)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId":   sessionID,
		"currentFact": formattedFact,
		"factNumber":  state.CurrentFact,
		"triesLeft":   state.TriesLeft,
		"score":       state.Score,
		"total":       state.Total,
		"isComplete":  state.IsComplete,
	})
}

func (h *FactsGameHandler) SubmitGuess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID   string `json:"sessionId"`
		CountryName string `json:"countryName"`
		Locale      string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session, exists := factsGameSessions[req.SessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update locale if provided
	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	// Make guess using game logic
	result, err := session.Logic.MakeGuess(req.CountryName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If guess is invalid (country not found), return error immediately
	if !result.IsValidGuess {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isValidGuess": false,
			"error":        result.Error,
		})
		return
	}

	// Build guess history with formatted facts
	// Each entry corresponds to a guess, and each guess reveals a new fact
	// Entry 0 = guess 1 (fact 1), entry 1 = guess 2 (fact 2), etc.
	countries := data.LoadCountries()
	guessHistory := []map[string]interface{}{}
	for i, entry := range result.GuessHistory {
		// Format fact with number prefix
		var factText string
		if entry.Fact != "" {
			factNum := i + 1
			factText = fmt.Sprintf("Fact %d: %s", factNum, entry.Fact)
		}

		// Determine if this guess was correct
		// The last entry is correct if result.IsCorrect is true
		// Previous entries are all wrong (they didn't match)
		isEntryCorrect := result.IsCorrect && i == len(result.GuessHistory)-1

		// Find country from guess string (skip if guess is "Skip")
		var countryInfo map[string]interface{}
		if strings.ToLower(strings.TrimSpace(entry.Guess)) != "skip" {
			for _, country := range countries {
				if utils.MatchCountry(entry.Guess, country, utils.MatchAll) {
					countryInfo = map[string]interface{}{
						"cca2":    country.CCA2,
						"name":    country.GetTranslatedName(locale),
						"flagUrl": "/assets/twemoji_flags_cca2/" + country.CCA2 + ".svg",
					}
					break
				}
			}
		}

		entryMap := map[string]interface{}{
			"guess":     entry.Guess,
			"fact":      factText,
			"isCorrect": isEntryCorrect,
		}
		if countryInfo != nil {
			entryMap["country"] = countryInfo
		}

		guessHistory = append(guessHistory, entryMap)
	}

	// Get next fact if wrong guess and tries left
	var nextFact string
	var nextFactNumber int
	if !result.IsCorrect && result.TriesLeft > 0 {
		fact, err := session.Logic.GetCurrentFact()
		if err == nil {
			state := session.Logic.GetState()
			nextFactNumber = state.CurrentFact
			// Format fact with number prefix
			nextFact = fmt.Sprintf("Fact %d: %s", nextFactNumber, fact)
		}
	}

	// Build response
	response := map[string]interface{}{
		"isCorrect":    result.IsCorrect,
		"triesLeft":    result.TriesLeft,
		"score":        result.Score,
		"total":        result.Total,
		"isComplete":   result.IsComplete,
		"guessHistory": guessHistory,
		"nextFact":     nextFact,
		"factNumber":   nextFactNumber,
	}

	// Only reveal correct country if correct or game over
	if result.IsCorrect && result.CorrectCountry != nil {
		response["correctCountry"] = map[string]interface{}{
			"cca2":    result.CorrectCountry.CCA2,
			"name":    result.CorrectCountry.GetTranslatedName(locale),
			"flagUrl": "/assets/twemoji_flags_cca2/" + result.CorrectCountry.CCA2 + ".svg",
		}
	} else if result.TriesLeft == 0 && result.CorrectCountry != nil {
		// Game over - reveal answer
		response["correctCountry"] = map[string]interface{}{
			"cca2":    result.CorrectCountry.CCA2,
			"name":    result.CorrectCountry.GetTranslatedName(locale),
			"flagUrl": "/assets/twemoji_flags_cca2/" + result.CorrectCountry.CCA2 + ".svg",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *FactsGameHandler) NextRound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session, exists := factsGameSessions[req.SessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Start new round
	if err := session.Logic.NewRound(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get first fact of new round
	currentFact, err := session.Logic.GetCurrentFact()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state := session.Logic.GetState()

	// Format fact with number prefix
	formattedFact := fmt.Sprintf("Fact %d: %s", state.CurrentFact, currentFact)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId":   req.SessionID,
		"currentFact": formattedFact,
		"factNumber":  state.CurrentFact,
		"triesLeft":   state.TriesLeft,
		"score":       state.Score,
		"total":       state.Total,
		"isComplete":  state.IsComplete,
	})
}

func (h *FactsGameHandler) Skip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
		Locale    string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session, exists := factsGameSessions[req.SessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update locale if provided
	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	// Skip current round
	result, err := session.Logic.Skip()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build guess history with formatted facts
	countries := data.LoadCountries()
	guessHistory := []map[string]interface{}{}
	for i, entry := range result.GuessHistory {
		var factText string
		if entry.Fact != "" {
			factNum := i + 1
			factText = fmt.Sprintf("Fact %d: %s", factNum, entry.Fact)
		}

		// Skip entries are not correct
		isEntryCorrect := false

		// Find country from guess string (skip if guess is "Skip")
		var countryInfo map[string]interface{}
		if strings.ToLower(strings.TrimSpace(entry.Guess)) != "skip" {
			for _, country := range countries {
				if utils.MatchCountry(entry.Guess, country, utils.MatchAll) {
					countryInfo = map[string]interface{}{
						"cca2":    country.CCA2,
						"name":    country.GetTranslatedName(locale),
						"flagUrl": "/assets/twemoji_flags_cca2/" + country.CCA2 + ".svg",
					}
					break
				}
			}
		}

		entryMap := map[string]interface{}{
			"guess":     entry.Guess,
			"fact":      factText,
			"isCorrect": isEntryCorrect,
		}
		if countryInfo != nil {
			entryMap["country"] = countryInfo
		}

		guessHistory = append(guessHistory, entryMap)
	}

	// Build response - same format as SubmitGuess when triesLeft = 0
	response := map[string]interface{}{
		"isCorrect":    false,
		"triesLeft":    0,
		"score":        result.Score,
		"total":        result.Total,
		"isComplete":   result.IsComplete,
		"guessHistory": guessHistory,
	}

	// Reveal correct country
	if result.CorrectCountry != nil {
		response["correctCountry"] = map[string]interface{}{
			"cca2":    result.CorrectCountry.CCA2,
			"name":    result.CorrectCountry.GetTranslatedName(locale),
			"flagUrl": "/assets/twemoji_flags_cca2/" + result.CorrectCountry.CCA2 + ".svg",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================
// Quiz Handler (combined rounds from multiple game types)
// ============================================

type QuizHandler struct{}

type QuizSession struct {
	SessionID    string
	RoundTypes   []round.GameType
	CurrentIndex int
	Score        int
	Runner       round.RoundRunner
	Locale       string
}

var (
	quizSessions      = make(map[string]*QuizSession)
	quizSessionsMutex sync.RWMutex
)

func (h *QuizHandler) Start(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoundTypes []string `json:"roundTypes"`
		Locale     string   `json:"locale"`
		Region     string   `json:"region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.RoundTypes) == 0 {
		http.Error(w, "roundTypes required", http.StatusBadRequest)
		return
	}

	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	types := make([]round.GameType, len(req.RoundTypes))
	for i, s := range req.RoundTypes {
		types[i] = round.GameType(s)
	}

	runner, err := round.NewRunner(types[0], round.RoundOptions{Region: req.Region, Locale: locale})
	if err != nil {
		http.Error(w, "Failed to create round: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := runner.StartRound(round.RoundOptions{Region: req.Region, Locale: locale}); err != nil {
		http.Error(w, "Failed to start round", http.StatusInternalServerError)
		return
	}

	sessionID := generateSessionID()
	session := &QuizSession{
		SessionID:    sessionID,
		RoundTypes:   types,
		CurrentIndex: 0,
		Score:        0,
		Runner:       runner,
		Locale:       locale,
	}
	quizSessionsMutex.Lock()
	quizSessions[sessionID] = session
	quizSessionsMutex.Unlock()

	payload, _ := runner.Payload(locale)
	resp := map[string]interface{}{
		"quizSessionId": sessionID,
		"score":         session.Score,
		"totalRounds":   len(types),
		"currentRound":  1,
	}
	if payload != nil {
		resp["gameType"] = string(payload.GameType)
		resp["data"] = payload.Data
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *QuizHandler) GetRound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("quizSessionId")
	if sessionID == "" {
		http.Error(w, "quizSessionId required", http.StatusBadRequest)
		return
	}

	quizSessionsMutex.RLock()
	session, exists := quizSessions[sessionID]
	quizSessionsMutex.RUnlock()
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.CurrentIndex >= len(session.RoundTypes) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"complete":     true,
			"score":        session.Score,
			"totalRounds":  len(session.RoundTypes),
		})
		return
	}

	payload, _ := session.Runner.Payload(session.Locale)
	resp := map[string]interface{}{
		"score":        session.Score,
		"totalRounds":  len(session.RoundTypes),
		"currentRound": session.CurrentIndex + 1,
	}
	if payload != nil {
		resp["gameType"] = string(payload.GameType)
		resp["data"] = payload.Data
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *QuizHandler) SubmitRound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		QuizSessionID string          `json:"quizSessionId"`
		Data          json.RawMessage `json:"data"`
		Locale        string          `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	quizSessionsMutex.Lock()
	session, exists := quizSessions[req.QuizSessionID]
	quizSessionsMutex.Unlock()
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if req.Locale != "" {
		session.Locale = req.Locale
	}

	result, err := session.Runner.Submit(&round.SubmitInput{Data: req.Data})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Score += result.ScoreDelta

	resp := map[string]interface{}{
		"correct":      result.Correct,
		"scoreDelta":  result.ScoreDelta,
		"score":       session.Score,
		"totalRounds": len(session.RoundTypes),
		"currentRound": session.CurrentIndex + 1,
	}
	if result.RevealedAnswer != nil {
		resp["revealedAnswer"] = result.RevealedAnswer
	}

	if result.RoundComplete {
		session.CurrentIndex++
		if session.CurrentIndex >= len(session.RoundTypes) {
			resp["complete"] = true
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		nextType := session.RoundTypes[session.CurrentIndex]
		runner, err := round.NewRunner(nextType, round.RoundOptions{Region: "", Locale: session.Locale})
		if err != nil {
			resp["complete"] = true
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		if err := runner.StartRound(round.RoundOptions{Locale: session.Locale}); err != nil {
			resp["complete"] = true
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		session.Runner = runner
		payload, _ := runner.Payload(session.Locale)
		if payload != nil {
			resp["nextGameType"] = string(payload.GameType)
			resp["nextData"] = payload.Data
		}
	} else if result.NextPayload != nil {
		resp["nextGameType"] = string(result.NextPayload.GameType)
		resp["nextData"] = result.NextPayload.Data
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ============================================
// Random one-question (stateless-style consumer)
// ============================================

var (
	randomRoundSessions      = make(map[string]round.RoundRunner)
	randomRoundSessionsMutex sync.RWMutex
)

func (h *QuizHandler) RandomRound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Locale string `json:"locale"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	loc := req.Locale
	if loc == "" {
		loc = "en"
	}

	registered := round.Registered()
	if len(registered) == 0 {
		http.Error(w, "No games registered", http.StatusInternalServerError)
		return
	}
	gt := registered[rand.Intn(len(registered))]

	runner, err := round.NewRunner(gt, round.RoundOptions{Locale: loc})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := runner.StartRound(round.RoundOptions{Locale: loc}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := generateSessionID()
	randomRoundSessionsMutex.Lock()
	randomRoundSessions[token] = runner
	randomRoundSessionsMutex.Unlock()

	payload, _ := runner.Payload(loc)
	resp := map[string]interface{}{
		"token":    token,
		"gameType": string(gt),
	}
	if payload != nil {
		resp["data"] = payload.Data
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *QuizHandler) RandomRoundSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token string          `json:"token"`
		Data  json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	randomRoundSessionsMutex.Lock()
	runner, exists := randomRoundSessions[req.Token]
	if exists {
		delete(randomRoundSessions, req.Token)
	}
	randomRoundSessionsMutex.Unlock()

	if !exists {
		http.Error(w, "Session not found or expired", http.StatusNotFound)
		return
	}

	result, err := runner.Submit(&round.SubmitInput{Data: req.Data})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"correct":       result.Correct,
		"scoreDelta":   result.ScoreDelta,
		"roundComplete": result.RoundComplete,
	}
	if result.RevealedAnswer != nil {
		resp["revealedAnswer"] = result.RevealedAnswer
	}
	if result.NextPayload != nil {
		resp["nextGameType"] = string(result.NextPayload.GameType)
		resp["nextData"] = result.NextPayload.Data
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
