// annotate-twemoji-flags adds data-fi-guess (and matching id) to every solid hex-filled
// path, circle, or rect in assets/twemoji_flags_cca2/*.svg that is not yet annotated.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	tagRe   = regexp.MustCompile(`<(path|circle|rect)([\s\S]*?)/>`)
	fillRe  = regexp.MustCompile(`\bfill="#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})"`)
	guessRe = regexp.MustCompile(`\bdata-fi-guess="([0-9]+)"`)
)

func main() {
	root := filepath.Join("assets", "twemoji_flags_cca2")
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	var nfiles, ntags int
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".svg" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		next, changed := annotate(string(raw))
		if changed {
			if err := os.WriteFile(path, []byte(next), 0o644); err != nil {
				return err
			}
			nfiles++
		}
		ntags += strings.Count(next, "data-fi-guess=")
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("done: wrote %d files under %s (total data-fi-guess attrs in tree scan: %d)\n", nfiles, root, ntags)
}

func annotate(s string) (string, bool) {
	maxID := 0
	for _, m := range guessRe.FindAllStringSubmatch(s, -1) {
		if v, err := strconv.Atoi(m[1]); err == nil && v > maxID {
			maxID = v
		}
	}

	next := tagRe.ReplaceAllStringFunc(s, func(full string) string {
		sub := tagRe.FindStringSubmatch(full)
		if len(sub) != 3 {
			return full
		}
		name, attrs := sub[1], sub[2]
		if guessRe.MatchString(attrs) {
			return full
		}
		if !fillRe.MatchString(attrs) {
			return full
		}
		maxID++
		id := strconv.Itoa(maxID)
		insert := fmt.Sprintf(` id="fi-guess-%s" data-fi-guess="%s"`, id, id)
		return "<" + name + insert + attrs + "/>"
	})

	changed := next != s
	return next, changed
}
