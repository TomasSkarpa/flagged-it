package api

import (
	"encoding/json"
	"fmt"
	"flagged-it/internal/data/models"
	"flagged-it/internal/games/flagcolor"
	"net/http"
	"strings"
	"sync"
)

type FlagColorGameHandler struct{}

type flagColorSession struct {
	SessionID          string
	Score              int
	Total              int
	RoundLimit         int
	Region             string
	Locale             string
	Difficulty         string
	UsedCountries      map[string]bool
	CurrentQuestionID  string
	CurrentCCA2        string
	CurrentGuessableID string
	TargetHex          string
	CurrentCountryName string
}

var (
	flagColorSessions      = make(map[string]*flagColorSession)
	flagColorSessionsMutex sync.RWMutex
)

func buildFlagColorPool(region string) ([]models.Country, error) {
	list, err := flagcolor.AllowlistCCA2()
	if err != nil {
		return nil, err
	}
	allow := make(map[string]bool)
	for _, c := range list {
		allow[strings.ToUpper(strings.TrimSpace(c))] = true
	}

	var pool []models.Country
	for _, country := range countries {
		if !allow[country.CCA2] {
			continue
		}
		if region != "" && region != "World" && country.Region != region {
			continue
		}
		pool = append(pool, country)
	}
	return pool, nil
}

func (h *FlagColorGameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Region     string `json:"region"`
		Locale     string `json:"locale"`
		RoundCount int    `json:"roundCount"`
		Difficulty string `json:"difficulty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	locale := req.Locale
	if locale == "" {
		locale = "en"
	}
	difficulty := strings.ToLower(strings.TrimSpace(req.Difficulty))
	if difficulty != "hard" {
		difficulty = "easy"
	}
	roundLimit := req.RoundCount
	if roundLimit <= 0 {
		roundLimit = 10
	}

	pool, err := buildFlagColorPool(req.Region)
	if err != nil || len(pool) == 0 {
		http.Error(w, "No countries available for flag color mode", http.StatusBadRequest)
		return
	}

	sessionID := generateSessionID()
	session := &flagColorSession{
		SessionID:     sessionID,
		Score:         0,
		Total:         0,
		RoundLimit:    roundLimit,
		Region:        req.Region,
		Locale:        locale,
		Difficulty:    difficulty,
		UsedCountries: make(map[string]bool),
	}

	q, err := generateFlagColorQuestion(pool, session, locale)
	if err != nil {
		http.Error(w, "Could not build question", http.StatusBadRequest)
		return
	}

	flagColorSessionsMutex.Lock()
	flagColorSessions[sessionID] = session
	flagColorSessionsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionId": sessionID,
		"question":  q,
	})
}

func generateFlagColorQuestion(pool []models.Country, session *flagColorSession, locale string) (map[string]interface{}, error) {
	if len(session.UsedCountries) >= len(pool) {
		session.UsedCountries = make(map[string]bool)
	}

	ch, err := flagcolor.PickChallenge(session.Difficulty, pool, session.UsedCountries)
	if err != nil {
		return nil, err
	}
	session.UsedCountries[ch.Country.CCA2] = true

	translated := getTranslatedCountry(ch.Country, locale)
	qid := generateQuestionID()
	session.CurrentQuestionID = qid
	session.CurrentCCA2 = ch.Country.CCA2
	session.CurrentGuessableID = ch.GuessableID
	session.TargetHex = ch.TargetHex
	session.CurrentCountryName = translated.Name.Common

	flagURL := "/assets/twemoji_flags_cca2/" + ch.Country.CCA2 + ".svg"
	return map[string]interface{}{
		"questionId":     qid,
		"flagUrl":        flagURL,
		"cca2":           ch.Country.CCA2,
		"guessableId":    ch.GuessableID,
		"countryName":    translated.Name.Common,
		"difficulty":     session.Difficulty,
		"maxPointsRound": flagcolor.PointsMaxPerRound,
	}, nil
}

func regenerateFlagColorQuestion(session *flagColorSession, locale string) (map[string]interface{}, error) {
	pool, err := buildFlagColorPool(session.Region)
	if err != nil || len(pool) == 0 {
		return nil, err
	}
	var country models.Country
	for _, c := range pool {
		if c.CCA2 == session.CurrentCCA2 {
			country = c
			break
		}
	}
	if country.CCA2 == "" {
		return nil, fmt.Errorf("country not in pool")
	}
	translated := getTranslatedCountry(country, locale)
	qid := session.CurrentQuestionID
	if qid == "" {
		qid = generateQuestionID()
		session.CurrentQuestionID = qid
	}
	flagURL := "/assets/twemoji_flags_cca2/" + country.CCA2 + ".svg"
	return map[string]interface{}{
		"questionId":     qid,
		"flagUrl":        flagURL,
		"cca2":           country.CCA2,
		"guessableId":    session.CurrentGuessableID,
		"countryName":    translated.Name.Common,
		"difficulty":     session.Difficulty,
		"maxPointsRound": flagcolor.PointsMaxPerRound,
	}, nil
}

func (h *FlagColorGameHandler) GetQuestion(w http.ResponseWriter, r *http.Request) {
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

	flagColorSessionsMutex.RLock()
	session, ok := flagColorSessions[sessionID]
	flagColorSessionsMutex.RUnlock()
	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	session.Locale = locale

	pool, err := buildFlagColorPool(session.Region)
	if err != nil || len(pool) == 0 {
		http.Error(w, "No countries available", http.StatusBadRequest)
		return
	}

	var q map[string]interface{}
	if session.CurrentQuestionID != "" && session.CurrentCCA2 != "" {
		q, err = regenerateFlagColorQuestion(session, locale)
	} else {
		q, err = generateFlagColorQuestion(pool, session, locale)
	}
	if err != nil {
		http.Error(w, "Could not build question", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)
}

func (h *FlagColorGameHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID  string `json:"sessionId"`
		QuestionID string `json:"questionId"`
		R          int    `json:"r"`
		G          int    `json:"g"`
		B          int    `json:"b"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	flagColorSessionsMutex.Lock()
	session, ok := flagColorSessions[req.SessionID]
	if !ok {
		flagColorSessionsMutex.Unlock()
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	if session.Total >= session.RoundLimit {
		flagColorSessionsMutex.Unlock()
		http.Error(w, "Game already finished", http.StatusBadRequest)
		return
	}

	if req.QuestionID == "" || req.QuestionID != session.CurrentQuestionID {
		flagColorSessionsMutex.Unlock()
		http.Error(w, "Invalid question", http.StatusBadRequest)
		return
	}

	guessHex, err := flagcolor.GuessHexFromRGB(req.R, req.G, req.B)
	if err != nil {
		flagColorSessionsMutex.Unlock()
		http.Error(w, "Invalid color", http.StatusBadRequest)
		return
	}

	correctHex := session.TargetHex
	delta := flagcolor.DeltaE76(correctHex, guessHex)
	points := flagcolor.PointsFromDeltaE(delta)
	if session.Difficulty == "hard" {
		points = int(float64(points) * 0.9)
		if points < 0 {
			points = 0
		}
	}

	session.Score += points
	session.Total++
	finished := session.Total >= session.RoundLimit

	session.CurrentQuestionID = ""
	session.CurrentCCA2 = ""
	session.CurrentGuessableID = ""
	session.TargetHex = ""

	pool, poolErr := buildFlagColorPool(session.Region)
	if poolErr == nil && !finished && len(pool) > 0 {
		_, _ = generateFlagColorQuestion(pool, session, session.Locale)
	}

	flagColorSessionsMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pointsEarned":      points,
		"deltaE":            delta,
		"correctHex":        correctHex,
		"guessHex":          guessHex,
		"score":             session.Score,
		"total":             session.Total,
		"finished":          finished,
		"maxPointsPerRound": flagcolor.PointsMaxPerRound,
	})
}

func (h *FlagColorGameHandler) GetScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}
	flagColorSessionsMutex.RLock()
	session, ok := flagColorSessions[sessionID]
	flagColorSessionsMutex.RUnlock()
	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"score": session.Score,
		"total": session.Total,
	})
}
