package flagcolor

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// GuessablePart describes one annotated region in a flag SVG.
type GuessablePart struct {
	GuessID string
	Tier    string
	Fill    string
}

// ParseGuessableParts returns elements tagged with data-fi-guess.
func ParseGuessableParts(svgXML []byte) ([]GuessablePart, error) {
	dec := xml.NewDecoder(bytes.NewReader(svgXML))
	var out []GuessablePart
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		name := se.Name.Local
		if name != "path" && name != "circle" && name != "rect" {
			continue
		}
		guessID := attrVal(se.Attr, "data-fi-guess")
		if guessID == "" {
			continue
		}
		tier := attrVal(se.Attr, "data-fi-tier")
		if tier == "" {
			tier = "both"
		}
		fill := attrVal(se.Attr, "fill")
		norm, err := normalizeHexFill(fill)
		if err != nil {
			continue
		}
		out = append(out, GuessablePart{
			GuessID: guessID,
			Tier:    strings.ToLower(tier),
			Fill:    norm,
		})
	}
	return out, nil
}

func attrVal(attrs []xml.Attr, key string) string {
	for _, a := range attrs {
		if a.Name.Local == key {
			return strings.TrimSpace(a.Value)
		}
	}
	return ""
}

func normalizeHexFill(fill string) (string, error) {
	fill = strings.TrimSpace(strings.ToUpper(fill))
	if !strings.HasPrefix(fill, "#") {
		return "", fmt.Errorf("not hex fill")
	}
	h := fill[1:]
	switch len(h) {
	case 3:
		return fmt.Sprintf("#%c%c%c%c%c%c", h[0], h[0], h[1], h[1], h[2], h[2]), nil
	case 6:
		if !isHex(h) {
			return "", fmt.Errorf("invalid hex")
		}
		return "#" + h, nil
	default:
		return "", fmt.Errorf("unsupported hex length")
	}
}

// NormalizeHexFill canonicalizes solid SVG fills (#RGB / #RRGGBB) for tooling.
func NormalizeHexFill(fill string) (string, error) {
	return normalizeHexFill(fill)
}

func isHex(s string) bool {
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

// TierMatchesDifficulty reports whether a part may be used for the given mode.
func TierMatchesDifficulty(tier, difficulty string) bool {
	tier = strings.ToLower(tier)
	difficulty = strings.ToLower(difficulty)
	if tier == "both" {
		return true
	}
	return tier == difficulty
}

// FilterPartsByDifficulty keeps parts eligible for easy or hard rounds.
func FilterPartsByDifficulty(parts []GuessablePart, difficulty string) []GuessablePart {
	var out []GuessablePart
	for _, p := range parts {
		if TierMatchesDifficulty(p.Tier, difficulty) {
			out = append(out, p)
		}
	}
	return out
}
