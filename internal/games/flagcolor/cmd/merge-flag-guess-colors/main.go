// merge-flag-guess-colors rewrites annotated flag SVGs so every solid fill shares one
// data-fi-guess id per distinct colour (stable order: first occurrence in file).
// Shards of the same colour get unique id="fi-guess-{n}{suffix}" for DOM validity.
// data-fi-tier is merged per colour (easy+hard -> both; unanimous tier kept).
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"flagged-it/internal/games/flagcolor"
)

var (
	tagRe    = regexp.MustCompile(`<(path|circle|rect)([^>]*)/>`)
	fillRe   = regexp.MustCompile(`\bfill="#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})"`)
	guessRe  = regexp.MustCompile(`\bdata-fi-guess="([^"]*)"`)
	tierRe   = regexp.MustCompile(`\bdata-fi-tier="(easy|hard|both)"`)
	mergeGrpRe = regexp.MustCompile(`\bdata-fi-merge-group="([^"]+)"`)
	idAttrRe = regexp.MustCompile(`\s*id="[^"]*"`)
	guessAttrRe = regexp.MustCompile(`\s*data-fi-guess="[^"]*"`)
	tierAttrRe = regexp.MustCompile(`\s*data-fi-tier="[^"]*"`)
)

type shard struct {
	full    string
	tagName string
	attrs   string
	fillKey string // normalized hex + merge-group or "" if invalid
	tier    string // easy | hard | both (both when omitted)
}

func main() {
	root := filepath.Join("assets", "twemoji_flags_cca2")
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	var nfiles int
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
		next, changed := merge(string(raw))
		if changed {
			if err := os.WriteFile(path, []byte(next), 0o644); err != nil {
				return err
			}
			nfiles++
			fmt.Println(path)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("merge-flag-guess-colors: updated %d files under %s\n", nfiles, root)
}

func merge(s string) (string, bool) {
	var shards []shard
	for _, m := range tagRe.FindAllStringSubmatch(s, -1) {
		if len(m) != 3 {
			continue
		}
		full, tagName, attrs := m[0], m[1], m[2]
		if !guessRe.MatchString(attrs) || !fillRe.MatchString(attrs) {
			continue
		}
		fillMatch := fillRe.FindStringSubmatch(attrs)
		norm, err := flagcolor.NormalizeHexFill("#" + fillMatch[1])
		if err != nil {
			continue
		}
		mergeGrp := ""
		if mg := mergeGrpRe.FindStringSubmatch(attrs); len(mg) > 1 {
			mergeGrp = mg[1]
		}
		fillSlot := norm + "|" + mergeGrp
		tier := "both"
		if tm := tierRe.FindStringSubmatch(attrs); len(tm) > 1 {
			tier = strings.ToLower(tm[1])
		}
		shards = append(shards, shard{full: full, tagName: tagName, attrs: attrs, fillKey: fillSlot, tier: tier})
	}
	if len(shards) == 0 {
		return s, false
	}

	fillToGuess := make(map[string]string)
	fillTierVotes := make(map[string][]string)
	var fillOrder []string
	for _, sh := range shards {
		if _, ok := fillToGuess[sh.fillKey]; !ok {
			fillOrder = append(fillOrder, sh.fillKey)
			fillToGuess[sh.fillKey] = strconv.Itoa(len(fillOrder))
		}
		fillTierVotes[sh.fillKey] = append(fillTierVotes[sh.fillKey], sh.tier)
	}

	fillSuffixIndex := make(map[string]int)
	next := s
	changed := false
	for _, sh := range shards {
		g := fillToGuess[sh.fillKey]
		i := fillSuffixIndex[sh.fillKey]
		fillSuffixIndex[sh.fillKey] = i + 1
		localID := fmt.Sprintf("fi-guess-%s%s", g, shardSuffix(i))
		tierOut := mergeTier(fillTierVotes[sh.fillKey])
		neu := rebuildTag(sh.tagName, sh.attrs, localID, g, tierOut)
		idx := strings.Index(next, sh.full)
		if idx < 0 {
			continue
		}
		next = next[:idx] + neu + next[idx+len(sh.full):]
		changed = true
	}
	return next, changed
}

func shardSuffix(i int) string {
	if i == 0 {
		return ""
	}
	// Match existing convention: fi-guess-4, fi-guess-4b, fi-guess-4c, …
	const letters = "bcdefghijklmnopqrstuvwxyz"
	if i-1 < len(letters) {
		return string(letters[i-1])
	}
	return fmt.Sprintf("-%d", i)
}

func mergeTier(votes []string) string {
	easy, hard, both := false, false, false
	for _, v := range votes {
		switch v {
		case "easy":
			easy = true
		case "hard":
			hard = true
		default:
			both = true
		}
	}
	if both || (easy && hard) {
		return ""
	}
	if easy {
		return "easy"
	}
	if hard {
		return "hard"
	}
	return ""
}

func rebuildTag(tagName, attrs, localID, guessNum, tier string) string {
	inner := strings.TrimSpace(attrs)
	inner = idAttrRe.ReplaceAllString(inner, "")
	inner = guessAttrRe.ReplaceAllString(inner, "")
	inner = tierAttrRe.ReplaceAllString(inner, "")
	inner = mergeGrpRe.ReplaceAllString(inner, "")
	inner = strings.TrimSpace(inner)

	var b strings.Builder
	fmt.Fprintf(&b, `<%s id="%s" data-fi-guess="%s"`, tagName, localID, guessNum)
	if tier != "" {
		fmt.Fprintf(&b, ` data-fi-tier="%s"`, tier)
	}
	if inner != "" {
		b.WriteByte(' ')
		b.WriteString(inner)
	}
	b.WriteString("/>")
	return b.String()
}
