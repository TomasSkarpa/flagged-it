//go:build ignore

// audit-flag-guess-colors checks annotated flag SVGs for merge/tier consistency.
//
// Run from repo root: go run ./internal/utils/audit-flag-guess-colors.go [path/to/twemoji_flags_cca2]
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"flagged-it/internal/games/flagcolor"
)

var (
	tagRe      = regexp.MustCompile(`<(path|circle|rect)([^>]*)/>`)
	fillRe     = regexp.MustCompile(`\bfill="#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})"`)
	guessRe    = regexp.MustCompile(`\bdata-fi-guess="([^"]*)"`)
	tierRe     = regexp.MustCompile(`\bdata-fi-tier="(easy|hard|both)"`)
	mergeGrpRe = regexp.MustCompile(`\bdata-fi-merge-group="([^"]+)"`)
)

type row struct {
	fill  string
	guess string
	tier  string
	merge string
}

func main() {
	root := filepath.Join("assets", "twemoji_flags_cca2")
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		root = os.Args[1]
	}

	var invSlot, invFill, invMG, parseErr int
	mergeGroupUsers := make(map[string]bool)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".svg" {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			return nil
		}
		cc := filepath.Base(path[:len(path)-4])
		rows := scanRows(string(raw))
		if len(rows) == 0 {
			fmt.Printf("WARN %s: no guessable solid hex paths\n", cc)
			return nil
		}

		for _, r := range rows {
			if r.merge != "" {
				mergeGroupUsers[cc] = true
			}
		}

		slotGuess := make(map[string]string)
		for _, r := range rows {
			slot := r.fill + "|" + r.merge
			if g, ok := slotGuess[slot]; ok {
				if g != r.guess {
					fmt.Printf("ERROR %s: slot %q has guesses %s vs %s\n", cc, slot, g, r.guess)
					invSlot++
				}
			} else {
				slotGuess[slot] = r.guess
			}
		}

		guessFill := make(map[string]string)
		for _, r := range rows {
			if f, ok := guessFill[r.guess]; ok {
				if f != r.fill {
					fmt.Printf("ERROR %s: guess %s uses fills %s vs %s\n", cc, r.guess, f, r.fill)
					invFill++
				}
			} else {
				guessFill[r.guess] = r.fill
			}
		}

		guessMG := make(map[string]string)
		guessHasEmptyMG := make(map[string]bool)
		for _, r := range rows {
			if r.merge == "" {
				guessHasEmptyMG[r.guess] = true
				continue
			}
			if prev, ok := guessMG[r.guess]; ok {
				if prev != r.merge {
					fmt.Printf("ERROR %s: guess %s merge-groups %q vs %q\n", cc, r.guess, prev, r.merge)
					invMG++
				}
			} else {
				guessMG[r.guess] = r.merge
			}
		}
		for g, mg := range guessMG {
			if guessHasEmptyMG[g] {
				fmt.Printf("ERROR %s: guess %s mixes empty merge-group with %q\n", cc, g, mg)
				invMG++
			}
		}

		guessTiers := make(map[string]map[string]bool)
		for _, r := range rows {
			if guessTiers[r.guess] == nil {
				guessTiers[r.guess] = make(map[string]bool)
			}
			guessTiers[r.guess][r.tier] = true
		}
		for g, ts := range guessTiers {
			if len(ts) > 1 {
				fmt.Printf("WARN %s: guess %s mixed tiers %v\n", cc, g, tierKeys(ts))
			}
		}

		if len(rows) >= 14 {
			fmt.Printf("WARN %s: %d shards — spot-check if same hex needs merge-groups\n", cc, len(rows))
		}

		nw := nearWhiteHexes(rows)
		if len(nw) >= 2 {
			sort.Strings(nw)
			fmt.Printf("WARN %s: multiple near-white fills %v\n", cc, nw)
		}

		parts, perr := flagcolor.ParseGuessableParts(raw)
		if perr != nil {
			fmt.Printf("ERROR %s: ParseGuessableParts: %v\n", cc, perr)
			parseErr++
		} else if len(flagcolor.DedupeGuessableParts(parts)) < 2 && len(rows) >= 4 {
			fmt.Printf("WARN %s: deduped to <2 challenges but many paths — review\n", cc)
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	var mg []string
	for c := range mergeGroupUsers {
		mg = append(mg, c)
	}
	sort.Strings(mg)
	fmt.Printf("\nmerge-group users (%d): %s\n", len(mg), strings.Join(mg, ", "))
	fmt.Printf("errors: slot=%d guessFill=%d mergeGroup=%d parse=%d\n", invSlot, invFill, invMG, parseErr)

	if invSlot > 0 || invFill > 0 || invMG > 0 || parseErr > 0 {
		os.Exit(1)
	}
}

func scanRows(s string) []row {
	var out []row
	for _, m := range tagRe.FindAllStringSubmatch(s, -1) {
		if len(m) != 3 {
			continue
		}
		attrs := m[2]
		if !guessRe.MatchString(attrs) || !fillRe.MatchString(attrs) {
			continue
		}
		fillMatch := fillRe.FindStringSubmatch(attrs)
		norm, err := flagcolor.NormalizeHexFill("#" + fillMatch[1])
		if err != nil {
			continue
		}
		guess := guessRe.FindStringSubmatch(attrs)[1]
		tier := "both"
		if tm := tierRe.FindStringSubmatch(attrs); len(tm) > 1 {
			tier = strings.ToLower(tm[1])
		}
		merge := ""
		if mg := mergeGrpRe.FindStringSubmatch(attrs); len(mg) > 1 {
			merge = mg[1]
		}
		out = append(out, row{fill: norm, guess: guess, tier: tier, merge: merge})
	}
	return out
}

func tierKeys(m map[string]bool) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func nearWhiteHexes(rows []row) []string {
	seen := make(map[string]bool)
	for _, r := range rows {
		h := strings.TrimPrefix(strings.ToUpper(r.fill), "#")
		if len(h) != 6 {
			continue
		}
		R, _ := strconv.ParseInt(h[0:2], 16, 64)
		G, _ := strconv.ParseInt(h[2:4], 16, 64)
		B, _ := strconv.ParseInt(h[4:6], 16, 64)
		if R >= 230 && G >= 230 && B >= 230 && !(R == 255 && G == 255 && B == 255) {
			seen[r.fill] = true
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}
