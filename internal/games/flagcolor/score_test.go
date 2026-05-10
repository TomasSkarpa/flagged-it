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

// TestPointsCurveIsBenevolent guards the concave decay: small and medium
// ΔE values should award noticeably more than a pure linear ramp would.
func TestPointsFromGuessHex_hueBallparkBonus(t *testing.T) {
	// Navy vs lighter sky blue: large ΔE76 (lightness + chroma), same broad hue direction.
	correct, guess := "#002495", "#1D89BF"
	raw := DeltaE76(correct, guess)
	if raw < 50 {
		t.Fatalf("expected large ΔE for fixture, got %.2f", raw)
	}
	fromRaw := PointsFromDeltaE(raw)
	fromGuess := PointsFromGuessHex(correct, guess)
	if fromGuess <= fromRaw {
		t.Fatalf("hue relax should score higher than raw ΔE alone: raw=%d guess=%d (ΔE=%.2f)", fromRaw, fromGuess, raw)
	}
	if fromGuess <= 0 {
		t.Fatalf("same-family shade miss should not floor at zero, got %d", fromGuess)
	}
}

func TestPointsFromGuessHex_identicalFullScore(t *testing.T) {
	p := PointsFromGuessHex("#112233", "#112233")
	if p != PointsMaxPerRound {
		t.Fatalf("got %d", p)
	}
}

func TestEffectiveScoringDelta_matchesRaw_whenIdentical(t *testing.T) {
	h := "#AABBCC"
	d := DeltaE76(h, h)
	es := EffectiveScoringDelta(h, h)
	if es != d {
		t.Fatalf("effective=%v raw=%v", es, d)
	}
}

func TestPointsCurveIsBenevolent(t *testing.T) {
	cases := []struct {
		deltaE float64
		min    int
	}{
		{5, PointsMaxPerRound}, // visually identical matches must be perfect
		{8, 95},                // barely-perceptible miss still feels perfect-ish
		{12, 90},               // mid-range mistakes stay in the "excellent" tier
		{20, 75},               // visible miss still scores well into "great"
		{30, 55},               // clearly off but not punished into the floor
	}
	for _, c := range cases {
		got := PointsFromDeltaE(c.deltaE)
		if got < c.min {
			t.Fatalf("ΔE %.1f: got %d, want >= %d", c.deltaE, got, c.min)
		}
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
