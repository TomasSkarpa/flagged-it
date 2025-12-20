package assets

import (
	"flagged-it/internal/assetsembed"
	"fyne.io/fyne/v2"
)

// LoadResourceFromPath loads a resource from embedded assets
func LoadResourceFromPath(path string) (fyne.Resource, error) {
	return assetsembed.LoadResourceFromPath(path)
}
