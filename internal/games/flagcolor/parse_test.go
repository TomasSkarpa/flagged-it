package flagcolor

import (
	"os"
	"testing"
)

func TestParseGuessableParts(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.svg")
	if err != nil {
		t.Fatal(err)
	}
	parts, err := ParseGuessableParts(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 2 {
		t.Fatalf("want 2 parts, got %d", len(parts))
	}
	if parts[0].GuessID != "1" || parts[0].Fill != "#FF0000" || parts[0].Tier != "easy" {
		t.Fatalf("part0 %+v", parts[0])
	}
	if parts[1].GuessID != "2" || parts[1].Tier != "hard" {
		t.Fatalf("part1 %+v", parts[1])
	}
}

func TestFilterPartsByDifficulty(t *testing.T) {
	parts := []GuessablePart{
		{GuessID: "1", Tier: "easy", Fill: "#FF0000"},
		{GuessID: "2", Tier: "hard", Fill: "#00FF00"},
		{GuessID: "3", Tier: "both", Fill: "#0000FF"},
	}
	easy := FilterPartsByDifficulty(parts, "easy")
	if len(easy) != 2 {
		t.Fatalf("easy count %d", len(easy))
	}
	hard := FilterPartsByDifficulty(parts, "hard")
	if len(hard) != 2 {
		t.Fatalf("hard count %d", len(hard))
	}
}
