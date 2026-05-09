package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func isCombiningMark(r rune) bool {
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Me, r)
}

// stripUnicodeMarks applies NFD and removes combining marks so typed answers without diacritics
// match canonical names (Czech, Polish, Vietnamese tones, Greek monotonic, etc.).
func stripUnicodeMarks(s string) string {
	nfd := norm.NFD.String(s)
	var b strings.Builder
	b.Grow(len(nfd))
	for _, r := range nfd {
		if isCombiningMark(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// latinGreekHomoglyph maps ASCII vowels to Greek lowercase letters that look the same, so mixed-script
// typing (e.g. Greek letters with a Latin "o" by mistake) still matches fully Greek canonical names.
var latinGreekHomoglyph = map[rune]rune{
	'a': '\u03b1', 'e': '\u03b5', 'i': '\u03b9', 'o': '\u03bf', 'u': '\u03c5',
}

func foldLatinGreekHomoglyphs(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if mapped, ok := latinGreekHomoglyph[r]; ok {
			b.WriteRune(mapped)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// NormalizeAnswerForCompare trims whitespace, lowercases using Unicode rules, strips combining marks
// after NFD (accents / Greek tonos / Czech háček, etc.), and folds Latin vowel lookalikes into Greek
// letters so mixed-script Greek guesses still match.
func NormalizeAnswerForCompare(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = stripUnicodeMarks(s)
	return foldLatinGreekHomoglyphs(s)
}
