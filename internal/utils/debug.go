package utils

import (
	"os"
	"time"
)

type DebugManager struct {
	debugEnabled bool
	clickCount   int
	lastClick    time.Time
}

func NewDebugManager() *DebugManager {
	return &DebugManager{
		debugEnabled: false,
		clickCount:   0,
	}
}

func (dm *DebugManager) IsDebugEnabled() bool {
	if os.Getenv("FLAGGED_IT_DEBUG") == "true" {
		return true
	}

	if _, err := os.Stat(".debug_mode"); err == nil {
		return true
	}

	for _, arg := range os.Args {
		if arg == "-v" {
			return true
		}
	}

	return dm.debugEnabled
}

func (dm *DebugManager) HandleSecretClick() bool {
	now := time.Now()

	if now.Sub(dm.lastClick) > 2*time.Second {
		dm.clickCount = 0
	}

	dm.clickCount++
	dm.lastClick = now

	if dm.clickCount >= 7 {
		dm.debugEnabled = true
		dm.clickCount = 0
		return true
	}

	return false
}

func (dm *DebugManager) EnableDebugMode() {
	dm.debugEnabled = true
}

func (dm *DebugManager) DisableDebugMode() {
	dm.debugEnabled = false
	os.Remove(".debug_mode")
}

func (dm *DebugManager) CreateDebugFile() error {
	file, err := os.Create(".debug_mode")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("debug_enabled")
	return err
}

func (dm *DebugManager) GetDebugInfo() map[string]interface{} {
	info := make(map[string]interface{})

	info["debug_enabled"] = dm.IsDebugEnabled()
	info["click_count"] = dm.clickCount
	info["env_debug"] = os.Getenv("FLAGGED_IT_DEBUG")

	if _, err := os.Stat(".debug_mode"); err == nil {
		info["debug_file_exists"] = true
	} else {
		info["debug_file_exists"] = false
	}

	return info
}
