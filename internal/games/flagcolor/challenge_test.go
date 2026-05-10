package flagcolor

import "testing"

func TestDedupeGuessableParts(t *testing.T) {
	parts := []GuessablePart{
		{GuessID: "4", Fill: "#BF9300"},
		{GuessID: "4", Fill: "#BF9300"},
		{GuessID: "5", Fill: "#FFFFFF"},
	}
	got := dedupeGuessableParts(parts)
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d %#v", len(got), got)
	}

	conflict := []GuessablePart{
		{GuessID: "1", Fill: "#FF0000"},
		{GuessID: "1", Fill: "#00FF00"},
	}
	got2 := dedupeGuessableParts(conflict)
	if len(got2) != 2 {
		t.Fatalf("same GuessID different fills should both remain, got %d", len(got2))
	}
}
