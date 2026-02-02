package api

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/facts"
	"flagged-it/internal/games/guessing"
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
	SessionID              string          `json:"sessionId"`
	Score                  int             `json:"score"`
	Total                  int             `json:"total"`
	Region                 string          `json:"region"`
	Locale                 string          `json:"-"` // User's locale for translations
	UsedCountries          map[string]bool `json:"-"`
	CurrentCorrectCCA2     string          `json:"-"`
	CurrentCorrectName     string          `json:"-"`
	CurrentQuestionID      string          `json:"-"` // Track current question ID to prevent regeneration
	CurrentQuestionCountry string          `json:"-"` // Track current question's country CCA2
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

	// Default to English if locale not provided
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	// Filter countries by region if specified
	filteredCountries := countries
	if req.Region != "" && req.Region != "World" {
		filteredCountries = []models.Country{}
		for _, country := range countries {
			if country.Region == req.Region {
				filteredCountries = append(filteredCountries, country)
			}
		}
	}

	if len(filteredCountries) == 0 {
		http.Error(w, "No countries available", http.StatusBadRequest)
		return
	}

	// Create session
	sessionID := generateSessionID()
	session := &GameSession{
		SessionID:     sessionID,
		Score:         0,
		Total:         0,
		Region:        req.Region,
		Locale:        locale,
		UsedCountries: make(map[string]bool),
	}
	sessionsMutex.Lock()
	sessions[sessionID] = session
	sessionsMutex.Unlock()

	// Generate first question
	question := generateQuestion(filteredCountries, session, locale)

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

	// Get locale from query parameter, fallback to session locale or English
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

	// Update session locale if provided
	if locale != "" {
		session.Locale = locale
	} else if session.Locale == "" {
		session.Locale = "en"
	}

	locale = session.Locale

	// Get filtered countries
	filteredCountries := countries
	if session.Region != "" && session.Region != "World" {
		filteredCountries = []models.Country{}
		for _, country := range countries {
			if country.Region == session.Region {
				filteredCountries = append(filteredCountries, country)
			}
		}
	}

	// Only generate a new question if one hasn't been generated yet for this round
	var question GameQuestion
	if session.CurrentQuestionID != "" && session.CurrentQuestionCountry != "" {
		// Return the same question (regenerate with same country to ensure translations are up to date)
		question = regenerateQuestionWithCountry(filteredCountries, session, locale, session.CurrentQuestionCountry)
	} else {
		// Generate a new question
		question = generateQuestion(filteredCountries, session, locale)
		session.CurrentQuestionID = question.QuestionID
		session.CurrentQuestionCountry = session.CurrentCorrectCCA2
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

	// Update locale if provided, otherwise use session locale
	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	// Find correct answer from question (we need to store it in session)
	// For now, we'll look it up from the question ID
	// In a real implementation, you'd store the correct answer in the session

	// This is a simplified version - in production, store question->answer mapping
	correct := false
	correctCCA2 := ""
	correctName := ""

	// For now, we'll need to store the current question's answer in the session
	// Let's add a field to track the current correct answer
	if session.CurrentCorrectCCA2 != "" {
		correct = req.AnswerCCA2 == session.CurrentCorrectCCA2
		correctCCA2 = session.CurrentCorrectCCA2
		// Get translated name
		for _, country := range countries {
			if country.CCA2 == correctCCA2 {
				correctName = country.GetTranslatedName(locale)
				break
			}
		}
		if correctName == "" {
			correctName = session.CurrentCorrectName
		}
	}

	// Check if game is already finished before incrementing
	alreadyFinished := session.Total >= 10
	if !alreadyFinished {
		session.Total++
		if correct {
			session.Score++
		}
		// Clear current question so next GetQuestion call will generate a new one
		session.CurrentQuestionID = ""
		session.CurrentQuestionCountry = ""
	}

	response := map[string]interface{}{
		"correct":     correct,
		"correctCca2": correctCCA2,
		"correctName": correctName,
		"score":       session.Score,
		"total":       session.Total,
		"finished":    session.Total >= 10,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

func generateQuestion(availableCountries []models.Country, session *GameSession, locale string) GameQuestion {
	// Find a country that hasn't been used
	var country models.Country
	maxAttempts := len(availableCountries) * 2 // Safety limit
	attempts := 0
	for {
		country = availableCountries[rand.Intn(len(availableCountries))]
		if !session.UsedCountries[country.CCA2] {
			break
		}
		attempts++
		// If all countries used, reset
		if len(session.UsedCountries) >= len(availableCountries) {
			session.UsedCountries = make(map[string]bool)
			break
		}
		// Safety: prevent infinite loop
		if attempts >= maxAttempts {
			session.UsedCountries = make(map[string]bool)
			break
		}
	}
	session.UsedCountries[country.CCA2] = true

	// Store correct answer in session (store English name as fallback)
	session.CurrentCorrectCCA2 = country.CCA2
	session.CurrentCorrectName = country.Name.Common

	// Generate options with translated names
	options := []models.Country{country}
	usedOptions := make(map[string]bool)
	usedOptions[country.CCA2] = true

	for len(options) < 4 {
		option := availableCountries[rand.Intn(len(availableCountries))]
		if !usedOptions[option.CCA2] {
			options = append(options, option)
			usedOptions[option.CCA2] = true
		}
	}

	// Shuffle options
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	flagURL := "/assets/twemoji_flags_cca2/" + country.CCA2 + ".svg"
	questionID := generateQuestionID()

	// Translate country names in options
	translatedOptions := getTranslatedCountries(options, locale)

	return GameQuestion{
		FlagURL:    flagURL,
		Options:    translatedOptions,
		QuestionID: questionID,
	}
}

// regenerateQuestionWithCountry regenerates a question with the same country (for locale updates)
func regenerateQuestionWithCountry(availableCountries []models.Country, session *GameSession, locale string, countryCCA2 string) GameQuestion {
	// Find the country
	var country models.Country
	for _, c := range availableCountries {
		if c.CCA2 == countryCCA2 {
			country = c
			break
		}
	}

	// If country not found, generate a new question instead
	if country.CCA2 == "" {
		return generateQuestion(availableCountries, session, locale)
	}

	// Store correct answer in session
	session.CurrentCorrectCCA2 = country.CCA2
	session.CurrentCorrectName = country.Name.Common

	// Generate options with translated names (same country, but options may vary)
	options := []models.Country{country}
	usedOptions := make(map[string]bool)
	usedOptions[country.CCA2] = true

	for len(options) < 4 {
		option := availableCountries[rand.Intn(len(availableCountries))]
		if !usedOptions[option.CCA2] {
			options = append(options, option)
			usedOptions[option.CCA2] = true
		}
	}

	// Shuffle options
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	flagURL := "/assets/twemoji_flags_cca2/" + country.CCA2 + ".svg"
	// Keep the same question ID to maintain consistency
	questionID := session.CurrentQuestionID
	if questionID == "" {
		// If no question ID exists, generate one (shouldn't happen, but safety check)
		questionID = generateQuestionID()
		session.CurrentQuestionID = questionID
	}

	// Translate country names in options
	translatedOptions := getTranslatedCountries(options, locale)

	return GameQuestion{
		FlagURL:    flagURL,
		Options:    translatedOptions,
		QuestionID: questionID,
	}
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
	SessionID          string          `json:"sessionId"`
	Score              int             `json:"score"`
	Total              int             `json:"total"`
	Region             string          `json:"region"`
	Locale             string          `json:"-"` // User's locale for translations
	UsedCountries      map[string]bool `json:"-"`
	CurrentCorrectCCA2 string          `json:"-"`
	CurrentCorrectCCA3 string          `json:"-"`
	CurrentCorrectName string          `json:"-"`
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

	// Default to English if locale not provided
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	// Filter countries with geo data
	filteredCountries := countriesWithGeo(req.Region)

	if len(filteredCountries) == 0 {
		http.Error(w, "No countries available", http.StatusBadRequest)
		return
	}

	// Create session
	sessionID := generateSessionID()
	session := &ShapeGameSession{
		SessionID:     sessionID,
		Score:         0,
		Total:         0,
		Region:        req.Region,
		Locale:        locale,
		UsedCountries: make(map[string]bool),
	}
	shapeSessions[sessionID] = session

	// Generate first question
	question := generateShapeQuestion(filteredCountries, session, locale)

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

	// Get locale from query parameter
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	session, exists := shapeSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update session locale if provided
	if locale != "" {
		session.Locale = locale
	} else if session.Locale == "" {
		session.Locale = "en"
	}
	locale = session.Locale

	filteredCountries := countriesWithGeo(session.Region)
	question := generateShapeQuestion(filteredCountries, session, locale)

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

	// Update locale if provided
	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	correct := false
	correctCCA2 := ""
	correctName := ""

	if session.CurrentCorrectCCA2 != "" {
		correct = req.AnswerCCA2 == session.CurrentCorrectCCA2
		correctCCA2 = session.CurrentCorrectCCA2
		// Get translated name
		for _, country := range countries {
			if country.CCA2 == correctCCA2 {
				correctName = country.GetTranslatedName(locale)
				break
			}
		}
		if correctName == "" {
			correctName = session.CurrentCorrectName
		}
	}

	// Check if game is already finished before incrementing
	alreadyFinished := session.Total >= 10
	if !alreadyFinished {
		session.Total++
		if correct {
			session.Score++
		}
	}

	response := map[string]interface{}{
		"correct":     correct,
		"correctCca2": correctCCA2,
		"correctName": correctName,
		"score":       session.Score,
		"total":       session.Total,
		"finished":    session.Total >= 10,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

func generateShapeQuestion(availableCountries []models.Country, session *ShapeGameSession, locale string) ShapeQuestion {
	// Find a country that hasn't been used and has geo data
	var country models.Country
	var geoData models.GeoJSON
	var err error

	maxAttempts := len(availableCountries) * 2 // Safety limit
	attempts := 0
	for {
		country = availableCountries[rand.Intn(len(availableCountries))]
		if !session.UsedCountries[country.CCA2] {
			geoData, err = data.LoadGeoData(country.CCA3)
			if err == nil {
				break
			}
		}
		attempts++
		// If all countries used, reset
		if len(session.UsedCountries) >= len(availableCountries) {
			session.UsedCountries = make(map[string]bool)
			break
		}
		// Safety: prevent infinite loop
		if attempts >= maxAttempts {
			session.UsedCountries = make(map[string]bool)
			// Get any country that has geo data
			for _, c := range availableCountries {
				if geoData, err = data.LoadGeoData(c.CCA3); err == nil {
					country = c
					break
				}
			}
			break
		}
	}
	session.UsedCountries[country.CCA2] = true

	// Store correct answer in session
	session.CurrentCorrectCCA2 = country.CCA2
	session.CurrentCorrectCCA3 = country.CCA3
	session.CurrentCorrectName = country.Name.Common

	// Generate options
	options := []models.Country{country}
	usedOptions := make(map[string]bool)
	usedOptions[country.CCA2] = true

	for len(options) < 4 {
		option := availableCountries[rand.Intn(len(availableCountries))]
		if !usedOptions[option.CCA2] {
			options = append(options, option)
			usedOptions[option.CCA2] = true
		}
	}

	// Shuffle options
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	// Translate country names in options
	translatedOptions := getTranslatedCountries(options, locale)

	questionID := generateQuestionID()

	return ShapeQuestion{
		GeoJSON:    geoData,
		Options:    translatedOptions,
		QuestionID: questionID,
	}
}

// ============================================
// Capital Game Handler
// ============================================

type CapitalGameHandler struct{}

type CapitalGameSession struct {
	SessionID             string          `json:"sessionId"`
	Score                 int             `json:"score"`
	Total                 int             `json:"total"`
	Region                string          `json:"region"`
	Locale                string          `json:"-"` // User's locale for translations
	UsedCountries         map[string]bool `json:"-"`
	CurrentCorrectCCA2    string          `json:"-"`
	CurrentCorrectName    string          `json:"-"`
	CurrentCorrectCapital string          `json:"-"`
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

	// Default to English if locale not provided
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	// Filter countries by region and ensure they have capitals
	filteredCountries := []models.Country{}
	for _, country := range countries {
		if len(country.Capital) > 0 && country.Capital[0] != "" {
			if req.Region == "" || req.Region == "World" || country.Region == req.Region {
				filteredCountries = append(filteredCountries, country)
			}
		}
	}

	if len(filteredCountries) < 4 {
		http.Error(w, "Not enough countries with capitals available", http.StatusBadRequest)
		return
	}

	sessionID := generateSessionID()
	session := &CapitalGameSession{
		SessionID:     sessionID,
		Score:         0,
		Total:         0,
		Region:        req.Region,
		Locale:        locale,
		UsedCountries: make(map[string]bool),
	}
	capitalSessions[sessionID] = session

	question := generateCapitalQuestion(filteredCountries, session, locale)

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

	// Get locale from query parameter
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	session, exists := capitalSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Update session locale if provided
	if locale != "" {
		session.Locale = locale
	} else if session.Locale == "" {
		session.Locale = "en"
	}
	locale = session.Locale

	// Filter countries
	filteredCountries := []models.Country{}
	for _, country := range countries {
		if len(country.Capital) > 0 && country.Capital[0] != "" {
			if session.Region == "" || session.Region == "World" || country.Region == session.Region {
				filteredCountries = append(filteredCountries, country)
			}
		}
	}

	question := generateCapitalQuestion(filteredCountries, session, locale)

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

	// Update locale if provided
	locale := req.Locale
	if locale == "" {
		locale = session.Locale
	}
	if locale == "" {
		locale = "en"
	}
	session.Locale = locale

	// Get translated country name
	correctCountryName := session.CurrentCorrectName
	if session.CurrentCorrectCCA2 != "" {
		for _, country := range countries {
			if country.CCA2 == session.CurrentCorrectCCA2 {
				correctCountryName = country.GetTranslatedName(locale)
				break
			}
		}
	}

	// Check if game is already finished before incrementing
	alreadyFinished := session.Total >= 10
	var correct bool
	if !alreadyFinished {
		session.Total++
		correct = req.Answer == session.CurrentCorrectCapital
		if correct {
			session.Score++
		}
	} else {
		correct = false
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"correct":        correct,
		"correctCapital": session.CurrentCorrectCapital,
		"correctCountry": correctCountryName,
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

func generateCapitalQuestion(availableCountries []models.Country, session *CapitalGameSession, locale string) CapitalQuestion {
	// Find a country that hasn't been used
	var country models.Country
	for {
		country = availableCountries[rand.Intn(len(availableCountries))]
		if !session.UsedCountries[country.CCA2] && len(country.Capital) > 0 {
			break
		}
		if len(session.UsedCountries) >= len(availableCountries) {
			session.UsedCountries = make(map[string]bool)
		}
	}
	session.UsedCountries[country.CCA2] = true

	correctCapital := country.Capital[0]
	session.CurrentCorrectCCA2 = country.CCA2
	session.CurrentCorrectName = country.Name.Common
	session.CurrentCorrectCapital = correctCapital

	// Generate options (capitals from other countries)
	options := []string{correctCapital}
	usedCapitals := make(map[string]bool)
	usedCapitals[correctCapital] = true

	for len(options) < 4 {
		option := availableCountries[rand.Intn(len(availableCountries))]
		if len(option.Capital) > 0 && !usedCapitals[option.Capital[0]] {
			options = append(options, option.Capital[0])
			usedCapitals[option.Capital[0]] = true
		}
	}

	// Shuffle options
	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	// Get translated country name
	translatedCountryName := country.GetTranslatedName(locale)

	return CapitalQuestion{
		CountryName: translatedCountryName,
		CountryCCA2: country.CCA2,
		Options:     options,
		QuestionID:  generateQuestionID(),
	}
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
	Locale       string              `json:"-"` // User's locale for translations
	CurrentLeft  HigherLowerItem     `json:"-"`
	CurrentRight HigherLowerItem     `json:"-"`
	UsedPairs    map[string]bool     `json:"-"`
	GameOver     bool                `json:"gameOver"`
}

var (
	higherLowerSessions      = make(map[string]*HigherLowerSession)
	higherLowerSessionsMutex sync.RWMutex
)

// ContinentData holds continent comparison data
type ContinentData struct {
	Name         string
	CountryCount int
}

func getContinentData() []ContinentData {
	continentCounts := make(map[string]int)
	for _, country := range countries {
		if country.Region != "" {
			continentCounts[country.Region]++
		}
	}

	result := []ContinentData{}
	for name, count := range continentCounts {
		result = append(result, ContinentData{Name: name, CountryCount: count})
	}
	return result
}

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
		req.Category = "population" // default
	}

	category := HigherLowerCategory(req.Category)
	if category != CategoryPopulation && category != CategoryArea && category != CategoryContinents {
		category = CategoryPopulation
	}

	// Default to English if locale not provided
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	sessionID := generateSessionID()
	session := &HigherLowerSession{
		SessionID: sessionID,
		Score:     0,
		HighScore: 0,
		Category:  category,
		Locale:    locale,
		UsedPairs: make(map[string]bool),
		GameOver:  false,
	}
	higherLowerSessions[sessionID] = session

	comparison := generateHigherLowerComparison(session, locale)

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
		Answer    string `json:"answer"` // "higher" or "lower"
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

	// Update locale if provided
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

	// Determine if answer is correct
	leftValue := session.CurrentLeft.Value
	rightValue := session.CurrentRight.Value

	var correct bool
	if req.Answer == "higher" {
		correct = rightValue >= leftValue
	} else {
		correct = rightValue <= leftValue
	}

	var nextComparison *HigherLowerComparison
	if correct {
		session.Score++
		if session.Score > session.HighScore {
			session.HighScore = session.Score
		}
		// Move right to left and generate new right
		session.CurrentLeft = session.CurrentRight
		nextComparison = generateHigherLowerComparison(session, locale)
	} else {
		session.GameOver = true
	}

	response := map[string]interface{}{
		"correct":    correct,
		"leftValue":  leftValue,
		"rightValue": rightValue,
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

func generateHigherLowerComparison(session *HigherLowerSession, locale string) *HigherLowerComparison {
	var left, right HigherLowerItem
	var valueLabel string

	switch session.Category {
	case CategoryPopulation:
		valueLabel = "population"
		left, right = generateCountryPair(session, locale, func(c models.Country) float64 {
			return float64(c.Population)
		})
	case CategoryArea:
		valueLabel = "km² area"
		left, right = generateCountryPair(session, locale, func(c models.Country) float64 {
			return c.Area
		})
	case CategoryContinents:
		valueLabel = "countries"
		left, right = generateContinentPair(session)
	default:
		valueLabel = "population"
		left, right = generateCountryPair(session, locale, func(c models.Country) float64 {
			return float64(c.Population)
		})
	}

	// If this is first comparison, store both
	if session.CurrentLeft.Name == "" {
		session.CurrentLeft = left
	} else {
		// Use existing left
		left = session.CurrentLeft
	}
	session.CurrentRight = right

	return &HigherLowerComparison{
		Left:       left,
		Right:      right,
		Category:   session.Category,
		ValueLabel: valueLabel,
	}
}

func generateCountryPair(session *HigherLowerSession, locale string, getValue func(models.Country) float64) (HigherLowerItem, HigherLowerItem) {
	// Filter countries with valid data
	validCountries := []models.Country{}
	for _, c := range countries {
		if getValue(c) > 0 {
			validCountries = append(validCountries, c)
		}
	}

	var left, right models.Country

	// Get left (if not already set)
	if session.CurrentLeft.Name == "" {
		for {
			left = validCountries[rand.Intn(len(validCountries))]
			if getValue(left) > 0 {
				break
			}
		}
	}

	// Get right (different from left)
	for {
		right = validCountries[rand.Intn(len(validCountries))]
		pairKey := session.CurrentLeft.Name + "-" + right.Name.Common
		if right.CCA2 != session.CurrentLeft.CCA2 && !session.UsedPairs[pairKey] {
			session.UsedPairs[pairKey] = true
			break
		}
	}

	leftItem := HigherLowerItem{
		Name:     left.GetTranslatedName(locale),
		Value:    getValue(left),
		CCA2:     left.CCA2,
		ImageURL: "/assets/twemoji_flags_cca2/" + left.CCA2 + ".svg",
	}

	rightItem := HigherLowerItem{
		Name:     right.GetTranslatedName(locale),
		Value:    getValue(right),
		CCA2:     right.CCA2,
		ImageURL: "/assets/twemoji_flags_cca2/" + right.CCA2 + ".svg",
	}

	// If left was already set, use it (but update name if locale changed)
	if session.CurrentLeft.Name != "" {
		leftItem = session.CurrentLeft
		// Update name if locale changed
		leftItem.Name = left.GetTranslatedName(locale)
	}

	return leftItem, rightItem
}

func generateContinentPair(session *HigherLowerSession) (HigherLowerItem, HigherLowerItem) {
	continents := getContinentData()

	var left, right ContinentData

	// Get left (if not already set)
	if session.CurrentLeft.Name == "" {
		left = continents[rand.Intn(len(continents))]
	}

	// Get right (different from left)
	for {
		right = continents[rand.Intn(len(continents))]
		leftName := session.CurrentLeft.Name
		if leftName == "" {
			leftName = left.Name
		}
		pairKey := leftName + "-" + right.Name
		if right.Name != leftName && !session.UsedPairs[pairKey] {
			session.UsedPairs[pairKey] = true
			break
		}
	}

	// Continent emoji mapping
	continentEmojis := map[string]string{
		"Africa":   "🌍",
		"Americas": "🌎",
		"Asia":     "🌏",
		"Europe":   "🏰",
		"Oceania":  "🏝️",
	}

	leftItem := HigherLowerItem{
		Name:     left.Name,
		Value:    float64(left.CountryCount),
		ImageURL: continentEmojis[left.Name],
	}

	rightItem := HigherLowerItem{
		Name:     right.Name,
		Value:    float64(right.CountryCount),
		ImageURL: continentEmojis[right.Name],
	}

	// If left was already set, use it
	if session.CurrentLeft.Name != "" {
		leftItem = session.CurrentLeft
	}

	return leftItem, rightItem
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
// Speed Challenge Game Handler
// ============================================

type SpeedChallengeHandler struct{}

type SpeedChallengeGameType string

const (
	GameTypeFlag        SpeedChallengeGameType = "flag"
	GameTypeShape       SpeedChallengeGameType = "shape"
	GameTypeCapital     SpeedChallengeGameType = "capital"
	GameTypeFacts       SpeedChallengeGameType = "facts"
	GameTypeHigherLower SpeedChallengeGameType = "higher_lower"
	GameTypeWorldle     SpeedChallengeGameType = "worldle"
)

type SpeedChallengeSession struct {
	SessionID       string                   `json:"sessionId"`
	Score           int                      `json:"score"`
	Total           int                      `json:"total"`
	Locale          string                   `json:"-"`
	CurrentRound    int                      `json:"currentRound"`
	CurrentGameType SpeedChallengeGameType   `json:"currentGameType"`
	TimeLimit       int                      `json:"timeLimit"` // seconds per question
	StartTime       time.Time                `json:"-"`
	GameTypeOrder   []SpeedChallengeGameType `json:"-"`
	SubSessions     map[string]interface{}   `json:"-"` // Store sub-game sessions
	RoundHistory    []RoundResult            `json:"-"`
	MaxRounds       int                      `json:"maxRounds"`
	IsComplete      bool                     `json:"isComplete"`
}

type RoundResult struct {
	RoundNumber int    `json:"roundNumber"`
	GameType    string `json:"gameType"`
	Correct     bool   `json:"correct"`
	TimeTaken   int    `json:"timeTaken"` // milliseconds
	Points      int    `json:"points"`
}

var (
	speedChallengeSessions      = make(map[string]*SpeedChallengeSession)
	speedChallengeSessionsMutex sync.RWMutex
)

// Available game types (excluding hangman and worldle for now)
var availableGameTypes = []SpeedChallengeGameType{
	GameTypeFlag,
	GameTypeShape,
	GameTypeCapital,
	GameTypeFacts,
	GameTypeHigherLower,
	// GameTypeWorldle, // Skip worldle for now as it requires more complex logic
}

func (h *SpeedChallengeHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TimeLimit int    `json:"timeLimit"` // seconds per question, default 30
		MaxRounds int    `json:"maxRounds"` // total rounds, default 10
		Locale    string `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use defaults if body is empty
		req.TimeLimit = 30
		req.MaxRounds = 10
		req.Locale = "en"
	}

	// Default values
	timeLimit := req.TimeLimit
	if timeLimit <= 0 {
		timeLimit = 30
	}
	maxRounds := req.MaxRounds
	if maxRounds <= 0 {
		maxRounds = 10
	}

	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	// Create game type order (shuffle for variety)
	gameTypeOrder := make([]SpeedChallengeGameType, len(availableGameTypes))
	copy(gameTypeOrder, availableGameTypes)
	rand.Shuffle(len(gameTypeOrder), func(i, j int) {
		gameTypeOrder[i], gameTypeOrder[j] = gameTypeOrder[j], gameTypeOrder[i]
	})

	// Create session
	sessionID := generateSessionID()
	session := &SpeedChallengeSession{
		SessionID:     sessionID,
		Score:         0,
		Total:         0,
		Locale:        locale,
		CurrentRound:  0,
		TimeLimit:     timeLimit,
		StartTime:     time.Now(),
		GameTypeOrder: gameTypeOrder,
		SubSessions:   make(map[string]interface{}),
		RoundHistory:  []RoundResult{},
		MaxRounds:     maxRounds,
		IsComplete:    false,
	}

	speedChallengeSessionsMutex.Lock()
	speedChallengeSessions[sessionID] = session
	speedChallengeSessionsMutex.Unlock()

	// Generate first question
	question, err := h.generateQuestion(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId":       sessionID,
		"question":        question,
		"currentRound":    session.CurrentRound,
		"currentGameType": string(session.CurrentGameType),
		"timeLimit":       session.TimeLimit,
		"maxRounds":       session.MaxRounds,
		"score":           session.Score,
		"total":           session.Total,
	})
}

func (h *SpeedChallengeHandler) GetQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	speedChallengeSessionsMutex.RLock()
	session, exists := speedChallengeSessions[sessionID]
	speedChallengeSessionsMutex.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	question, err := h.generateQuestion(session)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"question":        question,
		"currentRound":    session.CurrentRound,
		"currentGameType": string(session.CurrentGameType),
		"timeLimit":       session.TimeLimit,
		"score":           session.Score,
		"total":           session.Total,
	})
}

func (h *SpeedChallengeHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID  string      `json:"sessionId"`
		QuestionID string      `json:"questionId"`
		Answer     interface{} `json:"answer"`    // Can be string, CCA2, or "higher"/"lower"
		TimeTaken  int         `json:"timeTaken"` // milliseconds
		Locale     string      `json:"locale"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	speedChallengeSessionsMutex.Lock()
	session, exists := speedChallengeSessions[req.SessionID]
	speedChallengeSessionsMutex.Unlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.IsComplete {
		http.Error(w, "Game is complete", http.StatusBadRequest)
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

	// Submit answer to appropriate game handler
	correct, _, err := h.submitAnswerToGame(session, req.Answer, req.QuestionID, locale)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate points based on speed
	timeTaken := req.TimeTaken
	if timeTaken <= 0 {
		timeTaken = session.TimeLimit * 1000 // Default to full time if not provided
	}
	calculatedPoints := calculateSpeedPoints(correct, timeTaken, session.TimeLimit*1000)

	// Update session
	session.Total++
	if correct {
		session.Score += calculatedPoints
	}

	// Record round result
	roundResult := RoundResult{
		RoundNumber: session.CurrentRound,
		GameType:    string(session.CurrentGameType),
		Correct:     correct,
		TimeTaken:   timeTaken,
		Points:      calculatedPoints,
	}
	session.RoundHistory = append(session.RoundHistory, roundResult)

	// Extract correct answer from session metadata
	var correctAnswerText string
	if session.SubSessions != nil {
		if name, ok := session.SubSessions["currentCorrectName"].(string); ok && name != "" {
			correctAnswerText = name
		} else if capital, ok := session.SubSessions["currentCorrectCapital"].(string); ok && capital != "" {
			correctAnswerText = capital
		} else if cca2, ok := session.SubSessions["currentCorrectCCA2"].(string); ok && cca2 != "" {
			// Try to get country name from CCA2
			for _, c := range countries {
				if c.CCA2 == cca2 {
					correctAnswerText = c.GetTranslatedName(locale)
					break
				}
			}
		}
	}

	// Check if game is complete
	isComplete := session.Total >= session.MaxRounds
	session.IsComplete = isComplete

	// Generate next question if not complete
	var nextQuestion interface{}
	var nextGameType string
	if !isComplete {
		session.CurrentRound++
		question, err := h.generateQuestion(session)
		if err == nil {
			nextQuestion = question
			nextGameType = string(session.CurrentGameType)
		}
	}

	response := map[string]interface{}{
		"correct":       correct,
		"points":        calculatedPoints,
		"score":         session.Score,
		"total":         session.Total,
		"timeTaken":     timeTaken,
		"finished":      isComplete,
		"currentRound":  session.CurrentRound,
		"correctAnswer": correctAnswerText,
	}

	if nextQuestion != nil {
		response["nextQuestion"] = nextQuestion
		response["nextGameType"] = nextGameType
		response["timeLimit"] = session.TimeLimit
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *SpeedChallengeHandler) GetScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	speedChallengeSessionsMutex.RLock()
	session, exists := speedChallengeSessions[sessionID]
	speedChallengeSessionsMutex.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score":        session.Score,
		"total":        session.Total,
		"maxRounds":    session.MaxRounds,
		"isComplete":   session.IsComplete,
		"roundHistory": session.RoundHistory,
	})
}

// Helper methods

func (h *SpeedChallengeHandler) generateQuestion(session *SpeedChallengeSession) (interface{}, error) {
	// Determine current game type based on round
	gameTypeIndex := session.CurrentRound % len(session.GameTypeOrder)
	session.CurrentGameType = session.GameTypeOrder[gameTypeIndex]

	// Delegate to appropriate game handler
	switch session.CurrentGameType {
	case GameTypeFlag:
		return h.generateFlagQuestion(session)
	case GameTypeShape:
		return h.generateShapeQuestion(session)
	case GameTypeCapital:
		return h.generateCapitalQuestion(session)
	case GameTypeFacts:
		return h.generateFactsQuestion(session)
	case GameTypeHigherLower:
		return h.generateHigherLowerQuestion(session)
	case GameTypeWorldle:
		return h.generateWorldleQuestion(session)
	default:
		return nil, fmt.Errorf("unknown game type: %s", session.CurrentGameType)
	}
}

func (h *SpeedChallengeHandler) generateFlagQuestion(session *SpeedChallengeSession) (interface{}, error) {
	// Use existing flag game handler logic
	filteredCountries := countries
	usedCountries := make(map[string]bool)

	// Get a random country
	var country models.Country
	maxAttempts := len(filteredCountries) * 2
	attempts := 0
	for {
		country = filteredCountries[rand.Intn(len(filteredCountries))]
		if !usedCountries[country.CCA2] {
			break
		}
		attempts++
		if attempts >= maxAttempts {
			usedCountries = make(map[string]bool)
			break
		}
	}

	// Generate options
	options := []models.Country{country}
	usedOptions := make(map[string]bool)
	usedOptions[country.CCA2] = true

	for len(options) < 4 {
		option := filteredCountries[rand.Intn(len(filteredCountries))]
		if !usedOptions[option.CCA2] {
			options = append(options, option)
			usedOptions[option.CCA2] = true
		}
	}

	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	translatedOptions := getTranslatedCountries(options, session.Locale)
	flagURL := "/assets/twemoji_flags_cca2/" + country.CCA2 + ".svg"
	questionID := generateQuestionID()

	// Store correct answer in session metadata
	if session.SubSessions == nil {
		session.SubSessions = make(map[string]interface{})
	}
	session.SubSessions["currentCorrectCCA2"] = country.CCA2
	session.SubSessions["currentCorrectName"] = country.Name.Common

	return map[string]interface{}{
		"gameType":   "flag",
		"flagUrl":    flagURL,
		"options":    translatedOptions,
		"questionId": questionID,
	}, nil
}

func (h *SpeedChallengeHandler) generateShapeQuestion(session *SpeedChallengeSession) (interface{}, error) {
	// Use countries with geo data
	filteredCountries := countriesWithGeo("")
	if len(filteredCountries) == 0 {
		return nil, fmt.Errorf("no countries with geo data available")
	}

	var country models.Country
	var geoData models.GeoJSON
	var err error

	maxAttempts := len(filteredCountries) * 2
	attempts := 0
	for {
		country = filteredCountries[rand.Intn(len(filteredCountries))]
		geoData, err = data.LoadGeoData(country.CCA3)
		if err == nil {
			break
		}
		attempts++
		if attempts >= maxAttempts {
			return nil, fmt.Errorf("failed to load geo data")
		}
	}

	// Generate options
	options := []models.Country{country}
	usedOptions := make(map[string]bool)
	usedOptions[country.CCA2] = true

	for len(options) < 4 {
		option := filteredCountries[rand.Intn(len(filteredCountries))]
		if !usedOptions[option.CCA2] {
			options = append(options, option)
			usedOptions[option.CCA2] = true
		}
	}

	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	translatedOptions := getTranslatedCountries(options, session.Locale)
	questionID := generateQuestionID()

	// Store correct answer
	if session.SubSessions == nil {
		session.SubSessions = make(map[string]interface{})
	}
	session.SubSessions["currentCorrectCCA2"] = country.CCA2
	session.SubSessions["currentCorrectName"] = country.Name.Common

	return map[string]interface{}{
		"gameType":   "shape",
		"geoJson":    geoData,
		"options":    translatedOptions,
		"questionId": questionID,
	}, nil
}

func (h *SpeedChallengeHandler) generateCapitalQuestion(session *SpeedChallengeSession) (interface{}, error) {
	// Filter countries with capitals
	filteredCountries := []models.Country{}
	for _, country := range countries {
		if len(country.Capital) > 0 && country.Capital[0] != "" {
			filteredCountries = append(filteredCountries, country)
		}
	}

	if len(filteredCountries) < 4 {
		return nil, fmt.Errorf("not enough countries with capitals")
	}

	var country models.Country
	for {
		country = filteredCountries[rand.Intn(len(filteredCountries))]
		if len(country.Capital) > 0 {
			break
		}
	}

	correctCapital := country.Capital[0]
	translatedCountryName := country.GetTranslatedName(session.Locale)

	// Generate options
	options := []string{correctCapital}
	usedCapitals := make(map[string]bool)
	usedCapitals[correctCapital] = true

	for len(options) < 4 {
		option := filteredCountries[rand.Intn(len(filteredCountries))]
		if len(option.Capital) > 0 && !usedCapitals[option.Capital[0]] {
			options = append(options, option.Capital[0])
			usedCapitals[option.Capital[0]] = true
		}
	}

	rand.Shuffle(len(options), func(i, j int) {
		options[i], options[j] = options[j], options[i]
	})

	questionID := generateQuestionID()

	// Store correct answer
	if session.SubSessions == nil {
		session.SubSessions = make(map[string]interface{})
	}
	session.SubSessions["currentCorrectCapital"] = correctCapital
	session.SubSessions["currentCorrectCCA2"] = country.CCA2
	session.SubSessions["currentCorrectName"] = translatedCountryName

	return map[string]interface{}{
		"gameType":    "capital",
		"countryName": translatedCountryName,
		"countryCca2": country.CCA2,
		"options":     options,
		"questionId":  questionID,
	}, nil
}

func (h *SpeedChallengeHandler) generateFactsQuestion(session *SpeedChallengeSession) (interface{}, error) {
	// Find countries with facts
	var availableCountries []models.Country
	for _, country := range countries {
		if _, hasFacts := factsData[country.CCA2]; hasFacts {
			availableCountries = append(availableCountries, country)
		}
	}

	if len(availableCountries) == 0 {
		return nil, fmt.Errorf("no countries with facts available")
	}

	country := availableCountries[rand.Intn(len(availableCountries))]
	facts := factsData[country.CCA2].Facts

	if len(facts) == 0 {
		return nil, fmt.Errorf("no facts available for country")
	}

	// Get a random fact
	factIndex := rand.Intn(len(facts))
	fact := facts[factIndex]

	questionID := generateQuestionID()

	// Store correct answer
	if session.SubSessions == nil {
		session.SubSessions = make(map[string]interface{})
	}
	session.SubSessions["currentCorrectCCA2"] = country.CCA2
	session.SubSessions["currentCorrectName"] = country.Name.Common
	session.SubSessions["currentFact"] = fact

	return map[string]interface{}{
		"gameType":   "facts",
		"fact":       fact,
		"questionId": questionID,
	}, nil
}

func (h *SpeedChallengeHandler) generateHigherLowerQuestion(session *SpeedChallengeSession) (interface{}, error) {
	// Use population comparison
	validCountries := []models.Country{}
	for _, c := range countries {
		if c.Population > 0 {
			validCountries = append(validCountries, c)
		}
	}

	if len(validCountries) < 2 {
		return nil, fmt.Errorf("not enough countries available")
	}

	left := validCountries[rand.Intn(len(validCountries))]
	var right models.Country
	for {
		right = validCountries[rand.Intn(len(validCountries))]
		if right.CCA2 != left.CCA2 {
			break
		}
	}

	leftItem := map[string]interface{}{
		"name":     left.GetTranslatedName(session.Locale),
		"value":    float64(left.Population),
		"cca2":     left.CCA2,
		"imageUrl": "/assets/twemoji_flags_cca2/" + left.CCA2 + ".svg",
	}

	rightItem := map[string]interface{}{
		"name":     right.GetTranslatedName(session.Locale),
		"value":    float64(right.Population),
		"cca2":     right.CCA2,
		"imageUrl": "/assets/twemoji_flags_cca2/" + right.CCA2 + ".svg",
	}

	questionID := generateQuestionID()

	// Store correct answer (right should be higher or lower than left)
	if session.SubSessions == nil {
		session.SubSessions = make(map[string]interface{})
	}
	session.SubSessions["leftValue"] = float64(left.Population)
	session.SubSessions["rightValue"] = float64(right.Population)
	session.SubSessions["leftCountry"] = left.CCA2
	session.SubSessions["rightCountry"] = right.CCA2

	return map[string]interface{}{
		"gameType":   "higher_lower",
		"left":       leftItem,
		"right":      rightItem,
		"category":   "population",
		"valueLabel": "population",
		"questionId": questionID,
	}, nil
}

func (h *SpeedChallengeHandler) generateWorldleQuestion(session *SpeedChallengeSession) (interface{}, error) {
	// For speed challenge, skip worldle for now as it's complex
	// Instead, generate another game type
	// This is a placeholder - can be implemented later
	return nil, fmt.Errorf("worldle not implemented for speed challenge")
}

func (h *SpeedChallengeHandler) submitAnswerToGame(session *SpeedChallengeSession, answer interface{}, questionID string, locale string) (bool, int, error) {
	switch session.CurrentGameType {
	case GameTypeFlag:
		return h.submitFlagAnswer(session, answer, locale)
	case GameTypeShape:
		return h.submitShapeAnswer(session, answer, locale)
	case GameTypeCapital:
		return h.submitCapitalAnswer(session, answer, locale)
	case GameTypeFacts:
		return h.submitFactsAnswer(session, answer, locale)
	case GameTypeHigherLower:
		return h.submitHigherLowerAnswer(session, answer, locale)
	case GameTypeWorldle:
		return h.submitWorldleAnswer(session, answer, locale)
	default:
		return false, 0, fmt.Errorf("unknown game type: %s", session.CurrentGameType)
	}
}

func (h *SpeedChallengeHandler) submitFlagAnswer(session *SpeedChallengeSession, answer interface{}, locale string) (bool, int, error) {
	answerCCA2, ok := answer.(string)
	if !ok {
		return false, 0, fmt.Errorf("invalid answer format")
	}

	correctCCA2, ok := session.SubSessions["currentCorrectCCA2"].(string)
	if !ok {
		return false, 0, fmt.Errorf("correct answer not found")
	}

	return answerCCA2 == correctCCA2, 0, nil
}

func (h *SpeedChallengeHandler) submitShapeAnswer(session *SpeedChallengeSession, answer interface{}, locale string) (bool, int, error) {
	answerCCA2, ok := answer.(string)
	if !ok {
		return false, 0, fmt.Errorf("invalid answer format")
	}

	correctCCA2, ok := session.SubSessions["currentCorrectCCA2"].(string)
	if !ok {
		return false, 0, fmt.Errorf("correct answer not found")
	}

	return answerCCA2 == correctCCA2, 0, nil
}

func (h *SpeedChallengeHandler) submitCapitalAnswer(session *SpeedChallengeSession, answer interface{}, locale string) (bool, int, error) {
	answerStr, ok := answer.(string)
	if !ok {
		return false, 0, fmt.Errorf("invalid answer format")
	}

	correctCapital, ok := session.SubSessions["currentCorrectCapital"].(string)
	if !ok {
		return false, 0, fmt.Errorf("correct answer not found")
	}

	return answerStr == correctCapital, 0, nil
}

func (h *SpeedChallengeHandler) submitFactsAnswer(session *SpeedChallengeSession, answer interface{}, locale string) (bool, int, error) {
	answerStr, ok := answer.(string)
	if !ok {
		return false, 0, fmt.Errorf("invalid answer format")
	}

	correctCCA2, ok := session.SubSessions["currentCorrectCCA2"].(string)
	if !ok {
		return false, 0, fmt.Errorf("correct answer not found")
	}

	// Match country name
	for _, country := range countries {
		if country.CCA2 == correctCCA2 {
			if utils.MatchCountry(answerStr, country, utils.MatchAll) {
				return true, 0, nil
			}
			break
		}
	}

	return false, 0, nil
}

func (h *SpeedChallengeHandler) submitHigherLowerAnswer(session *SpeedChallengeSession, answer interface{}, locale string) (bool, int, error) {
	answerStr, ok := answer.(string)
	if !ok {
		return false, 0, fmt.Errorf("invalid answer format")
	}

	leftValue, ok1 := session.SubSessions["leftValue"].(float64)
	rightValue, ok2 := session.SubSessions["rightValue"].(float64)

	if !ok1 || !ok2 {
		return false, 0, fmt.Errorf("comparison values not found")
	}

	var correct bool
	if answerStr == "higher" {
		correct = rightValue >= leftValue
	} else if answerStr == "lower" {
		correct = rightValue <= leftValue
	} else {
		return false, 0, fmt.Errorf("invalid answer, must be 'higher' or 'lower'")
	}

	return correct, 0, nil
}

func (h *SpeedChallengeHandler) submitWorldleAnswer(session *SpeedChallengeSession, answer interface{}, locale string) (bool, int, error) {
	// Simplified: for speed challenge, worldle can be skipped or simplified
	// For now, return false as it requires more complex logic
	return false, 0, fmt.Errorf("worldle not fully implemented for speed challenge")
}

// Calculate points based on speed and correctness
func calculateSpeedPoints(correct bool, timeTaken int, timeLimit int) int {
	if !correct {
		return 0
	}

	basePoints := 10
	timeRatio := float64(timeTaken) / float64(timeLimit)
	speedBonus := int((1.0 - timeRatio) * 10) // 0-10 bonus based on speed

	if speedBonus < 0 {
		speedBonus = 0
	}
	if speedBonus > 10 {
		speedBonus = 10
	}

	return basePoints + speedBonus
}
