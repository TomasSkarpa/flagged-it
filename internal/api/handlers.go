package api

import (
	"encoding/json"
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"math/rand"
	"net/http"
	"time"
)

type FlagGameHandler struct{}

type GameQuestion struct {
	FlagURL    string            `json:"flagUrl"`
	Options    []models.Country  `json:"options"`
	QuestionID string            `json:"questionId"`
}

type GameAnswer struct {
	Correct    bool   `json:"correct"`
	CorrectCCA2 string `json:"correctCca2"`
	CorrectName string `json:"correctName"`
}

type GameSession struct {
	SessionID         string            `json:"sessionId"`
	Score             int               `json:"score"`
	Total             int               `json:"total"`
	Region            string            `json:"region"`
	UsedCountries     map[string]bool   `json:"-"`
	CurrentCorrectCCA2 string           `json:"-"`
	CurrentCorrectName string           `json:"-"`
}

var sessions = make(map[string]*GameSession)
var countries = []models.Country{}

func init() {
	countries = data.LoadCountries()
	rand.Seed(time.Now().UnixNano())
}

func (h *FlagGameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Region string `json:"region"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
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
		UsedCountries: make(map[string]bool),
	}
	sessions[sessionID] = session

	// Generate first question
	question := generateQuestion(filteredCountries, session)

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

	session, exists := sessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

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

	question := generateQuestion(filteredCountries, session)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(question)
}

func (h *FlagGameHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
		QuestionID string `json:"questionId"`
		AnswerCCA2 string `json:"answerCca2"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	session, exists := sessions[req.SessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

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
		correctName = session.CurrentCorrectName
	}

	session.Total++
	if correct {
		session.Score++
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

	session, exists := sessions[sessionID]
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

func generateQuestion(availableCountries []models.Country, session *GameSession) GameQuestion {
	// Find a country that hasn't been used
	var country models.Country
	for {
		country = availableCountries[rand.Intn(len(availableCountries))]
		if !session.UsedCountries[country.CCA2] {
			break
		}
		// If all countries used, reset
		if len(session.UsedCountries) >= len(availableCountries) {
			session.UsedCountries = make(map[string]bool)
		}
	}
	session.UsedCountries[country.CCA2] = true

	// Store correct answer in session
	session.CurrentCorrectCCA2 = country.CCA2
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

	flagURL := "/assets/twemoji_flags_cca2/" + country.CCA2 + ".svg"
	questionID := generateQuestionID()

	return GameQuestion{
		FlagURL:    flagURL,
		Options:    options,
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"countries": countries,
		"total":     len(countries),
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

// ============================================
// Shape Game Handler
// ============================================

type ShapeGameHandler struct{}

type ShapeGameSession struct {
	SessionID          string          `json:"sessionId"`
	Score              int             `json:"score"`
	Total              int             `json:"total"`
	Region             string          `json:"region"`
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

var shapeSessions = make(map[string]*ShapeGameSession)

// countriesWithGeo returns countries that have geo data available
func countriesWithGeo(region string) []models.Country {
	var result []models.Country
	for _, country := range countries {
		if region != "" && region != "World" && country.Region != region {
			continue
		}
		// Check if geo data exists
		if country.CCA3 != "" {
			_, err := data.LoadGeoData(country.CCA3)
			if err == nil {
				result = append(result, country)
			}
		}
	}
	return result
}

func (h *ShapeGameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Region string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
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
		UsedCountries: make(map[string]bool),
	}
	shapeSessions[sessionID] = session

	// Generate first question
	question := generateShapeQuestion(filteredCountries, session)

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

	session, exists := shapeSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	filteredCountries := countriesWithGeo(session.Region)
	question := generateShapeQuestion(filteredCountries, session)

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

	correct := false
	correctCCA2 := ""
	correctName := ""

	if session.CurrentCorrectCCA2 != "" {
		correct = req.AnswerCCA2 == session.CurrentCorrectCCA2
		correctCCA2 = session.CurrentCorrectCCA2
		correctName = session.CurrentCorrectName
	}

	session.Total++
	if correct {
		session.Score++
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

func generateShapeQuestion(availableCountries []models.Country, session *ShapeGameSession) ShapeQuestion {
	// Find a country that hasn't been used and has geo data
	var country models.Country
	var geoData models.GeoJSON
	var err error

	for {
		country = availableCountries[rand.Intn(len(availableCountries))]
		if !session.UsedCountries[country.CCA2] {
			geoData, err = data.LoadGeoData(country.CCA3)
			if err == nil {
				break
			}
		}
		// If all countries used, reset
		if len(session.UsedCountries) >= len(availableCountries) {
			session.UsedCountries = make(map[string]bool)
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

	questionID := generateQuestionID()

	return ShapeQuestion{
		GeoJSON:    geoData,
		Options:    options,
		QuestionID: questionID,
	}
}

// ============================================
// Capital Game Handler
// ============================================

type CapitalGameHandler struct{}

type CapitalGameSession struct {
	SessionID           string          `json:"sessionId"`
	Score               int             `json:"score"`
	Total               int             `json:"total"`
	Region              string          `json:"region"`
	UsedCountries       map[string]bool `json:"-"`
	CurrentCorrectCCA2  string          `json:"-"`
	CurrentCorrectName  string          `json:"-"`
	CurrentCorrectCapital string        `json:"-"`
}

type CapitalQuestion struct {
	CountryName string   `json:"countryName"`
	CountryCCA2 string   `json:"countryCca2"`
	Options     []string `json:"options"`
	QuestionID  string   `json:"questionId"`
}

var capitalSessions = make(map[string]*CapitalGameSession)

func (h *CapitalGameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Region string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
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
		UsedCountries: make(map[string]bool),
	}
	capitalSessions[sessionID] = session

	question := generateCapitalQuestion(filteredCountries, session)

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

	session, exists := capitalSessions[sessionID]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Filter countries
	filteredCountries := []models.Country{}
	for _, country := range countries {
		if len(country.Capital) > 0 && country.Capital[0] != "" {
			if session.Region == "" || session.Region == "World" || country.Region == session.Region {
				filteredCountries = append(filteredCountries, country)
			}
		}
	}

	question := generateCapitalQuestion(filteredCountries, session)

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

	correct := req.Answer == session.CurrentCorrectCapital
	if correct {
		session.Score++
	}
	session.Total++

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"correct":        correct,
		"correctCapital": session.CurrentCorrectCapital,
		"correctCountry": session.CurrentCorrectName,
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

func generateCapitalQuestion(availableCountries []models.Country, session *CapitalGameSession) CapitalQuestion {
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

	return CapitalQuestion{
		CountryName: country.Name.Common,
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
	SessionID      string              `json:"sessionId"`
	Score          int                 `json:"score"`
	HighScore      int                 `json:"highScore"`
	Category       HigherLowerCategory `json:"category"`
	CurrentLeft    HigherLowerItem     `json:"-"`
	CurrentRight   HigherLowerItem     `json:"-"`
	UsedPairs      map[string]bool     `json:"-"`
	GameOver       bool                `json:"gameOver"`
}

var higherLowerSessions = make(map[string]*HigherLowerSession)

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
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Category = "population" // default
	}

	category := HigherLowerCategory(req.Category)
	if category != CategoryPopulation && category != CategoryArea && category != CategoryContinents {
		category = CategoryPopulation
	}

	sessionID := generateSessionID()
	session := &HigherLowerSession{
		SessionID: sessionID,
		Score:     0,
		HighScore: 0,
		Category:  category,
		UsedPairs: make(map[string]bool),
		GameOver:  false,
	}
	higherLowerSessions[sessionID] = session

	comparison := generateHigherLowerComparison(session)

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
		nextComparison = generateHigherLowerComparison(session)
	} else {
		session.GameOver = true
	}

	response := map[string]interface{}{
		"correct":      correct,
		"leftValue":    leftValue,
		"rightValue":   rightValue,
		"score":        session.Score,
		"highScore":    session.HighScore,
		"gameOver":     session.GameOver,
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

func generateHigherLowerComparison(session *HigherLowerSession) *HigherLowerComparison {
	var left, right HigherLowerItem
	var valueLabel string

	switch session.Category {
	case CategoryPopulation:
		valueLabel = "population"
		left, right = generateCountryPair(session, func(c models.Country) float64 {
			return float64(c.Population)
		})
	case CategoryArea:
		valueLabel = "km² area"
		left, right = generateCountryPair(session, func(c models.Country) float64 {
			return c.Area
		})
	case CategoryContinents:
		valueLabel = "countries"
		left, right = generateContinentPair(session)
	default:
		valueLabel = "population"
		left, right = generateCountryPair(session, func(c models.Country) float64 {
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

func generateCountryPair(session *HigherLowerSession, getValue func(models.Country) float64) (HigherLowerItem, HigherLowerItem) {
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
		Name:     left.Name.Common,
		Value:    getValue(left),
		CCA2:     left.CCA2,
		ImageURL: "/assets/twemoji_flags_cca2/" + left.CCA2 + ".svg",
	}

	rightItem := HigherLowerItem{
		Name:     right.Name.Common,
		Value:    getValue(right),
		CCA2:     right.CCA2,
		ImageURL: "/assets/twemoji_flags_cca2/" + right.CCA2 + ".svg",
	}

	// If left was already set, use it
	if session.CurrentLeft.Name != "" {
		leftItem = session.CurrentLeft
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
