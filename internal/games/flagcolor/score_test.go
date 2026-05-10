package flagcolor

import (
	"testing"
)

func TestDeltaE76_identical(t *testing.T) {
	d := DeltaE76("#AABBCC", "#AABBCC")
	if d > 0.01 {
		t.Fatalf("expected ~0, got %v", d)
	}
}

func TestPointsFromDeltaE(t *testing.T) {
	if PointsFromDeltaE(0) != PointsMaxPerRound {
		t.Fatal()
	}
	if PointsFromDeltaE(100) != 0 {
		t.Fatal()
	}
	mid := PointsFromDeltaE(21)
	if mid <= 0 || mid >= PointsMaxPerRound {
		t.Fatalf("mid %d", mid)
	}
}

func TestDeltaE76_ordering(t *testing.T) {
	close := DeltaE76("#FF0000", "#FE0101")
	far := DeltaE76("#FF0000", "#000000")
	if !(close < far) {
		t.Fatalf("close=%v far=%v", close, far)
	}
}

func TestPointsDecreasesWithDelta(t *testing.T) {
	pNear := PointsFromDeltaE(8)
	pFar := PointsFromDeltaE(30)
	if pNear <= pFar {
		t.Fatalf("near=%d far=%d", pNear, pFar)
	}
}

func TestGuessHexFromRGB(t *testing.T) {
	h, err := GuessHexFromRGB(10, 20, 30)
	if err != nil {
		t.Fatal(err)
	}
	if h != "#0A141E" {
		t.Fatal(h)
	}
	_, err = GuessHexFromRGB(300, 0, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeHexFill_shortForm(t *testing.T) {
	n, err := normalizeHexFill("#abc")
	if err != nil {
		t.Fatal(err)
	}
	if n != "#AABBCC" {
		t.Fatal(n)
	}
}
