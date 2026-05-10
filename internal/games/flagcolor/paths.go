package flagcolor

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	resolvedDir     string
	resolvedDirOnce sync.Once
)

// ResolveSVGDir finds the folder containing twemoji_flags_cca2 SVGs.
// Order: FLAG_SVG_DIR env, ASSETS_PATH/twemoji_flags_cca2, repo web/static path, repo assets path.
func ResolveSVGDir() string {
	resolvedDirOnce.Do(func() {
		if d := os.Getenv("FLAG_SVG_DIR"); d != "" {
			if isDir(d) {
				resolvedDir = d
				return
			}
		}
		if base := os.Getenv("ASSETS_PATH"); base != "" {
			candidate := filepath.Join(base, "twemoji_flags_cca2")
			if isDir(candidate) {
				resolvedDir = candidate
				return
			}
		}
		if root, ok := findRepoRoot(); ok {
			for _, rel := range []string{
				filepath.Join("web", "static", "assets", "twemoji_flags_cca2"),
				filepath.Join("assets", "twemoji_flags_cca2"),
			} {
				candidate := filepath.Join(root, rel)
				if isDir(candidate) {
					resolvedDir = candidate
					return
				}
			}
		}
	})
	return resolvedDir
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func findRepoRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for i := 0; i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

// ReadSVGBytes loads raw SVG XML for a country code (uppercase CCA2).
func ReadSVGBytes(cca2 string) ([]byte, error) {
	dir := ResolveSVGDir()
	if dir == "" {
		return nil, os.ErrNotExist
	}
	code := strings.ToUpper(strings.TrimSpace(cca2))
	p := filepath.Join(dir, code+".svg")
	return os.ReadFile(p)
}
