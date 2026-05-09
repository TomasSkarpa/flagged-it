package utils

import "testing"

func TestNormalizeAnswerForCompare(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Κόσοβο", "κοσοβο"},
		{"Κοσοβο", "κοσοβο"},
		{"Κoσοβο", "κοσοβο"}, // Latin "o" after Greek Κ (common mixed-script typo)
		{"κόσοβο", "κοσοβο"},
		{"  Αθήνα ", "αθηνα"},
		{"Αθήνα", "αθηνα"},
		{"United States", "υnιtεd stαtεs"}, // Latin vowels folded to Greek homoglyphs
		// Czech (cs) and other Latin diacritics: compare without accents, any case
		{"jizni korea", "jιznι kοrεα"},
		{"Jižní Korea", "jιznι kοrεα"},
		{"JIZNI KOREA", "jιznι kοrεα"},
		{"réunion", "rευnιοn"},
		{"RÉUNION", "rευnιοn"},
	}
	for _, tt := range tests {
		if got := NormalizeAnswerForCompare(tt.in); got != tt.want {
			t.Errorf("NormalizeAnswerForCompare(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
