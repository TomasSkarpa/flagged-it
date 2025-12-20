package assetsembed

import (
	"embed"
	"fyne.io/fyne/v2"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed assets
var AssetsFS embed.FS

// LoadResource loads a resource from the embedded assets filesystem
func LoadResource(path string) (fyne.Resource, error) {
	// For web builds, use URL loading
	if runtime.GOOS == "js" {
		return fyne.LoadResourceFromURLString(path)
	}
	
	// Normalize path - ensure it starts with "assets/"
	cleanPath := path
	if !strings.HasPrefix(cleanPath, "assets/") {
		cleanPath = "assets/" + cleanPath
	}
	
	data, err := AssetsFS.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}
	
	return fyne.NewStaticResource(filepath.Base(path), data), nil
}

// LoadResourceFromPath is compatible with fyne.LoadResourceFromPath
func LoadResourceFromPath(path string) (fyne.Resource, error) {
	return LoadResource(path)
}

// ReadDir reads a directory from embedded assets
func ReadDir(path string) ([]fs.DirEntry, error) {
	cleanPath := path
	if !strings.HasPrefix(cleanPath, "assets/") {
		cleanPath = "assets/" + cleanPath
	}
	return AssetsFS.ReadDir(cleanPath)
}

