package utils

type ScoreManager struct {
	scores     map[string]int
	totals     map[string]int
	listRegion string
}

func NewScoreManager() *ScoreManager {
	return &ScoreManager{
		scores: make(map[string]int),
		totals: make(map[string]int),
	}
}

func (sm *ScoreManager) UpdateScore(game string, score int) {
	if score > sm.scores[game] {
		sm.scores[game] = score
	}
}

func (sm *ScoreManager) SetTotal(game string, total int) {
	sm.totals[game] = total
}

func (sm *ScoreManager) SetListRegion(region string) {
	sm.listRegion = region
}

func (sm *ScoreManager) GetListRegion() string {
	return sm.listRegion
}

func (sm *ScoreManager) GetScore(game string) int {
	return sm.scores[game]
}

func (sm *ScoreManager) GetTotal(game string) int {
	return sm.totals[game]
}

func (sm *ScoreManager) GetAllScores() map[string]int {
	return sm.scores
}

func (sm *ScoreManager) GetAllTotals() map[string]int {
	return sm.totals
}
