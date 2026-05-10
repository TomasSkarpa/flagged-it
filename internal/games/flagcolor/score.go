package flagcolor

import (
	"math"
	"strings"
)

// PointsMaxPerRound is the highest score achievable on one round before difficulty scaling.
const PointsMaxPerRound = 100

// Scoring curve thresholds (ΔE76 in Lab space).
// Anything at or below deltaPerfect awards full points; anything at or beyond
// deltaZero awards none. Between them, the falloff uses a concave power curve
// so small/medium colour mistakes are forgiven more than a strict linear ramp.
//
// deltaPerfect of 5 covers the "perceptible only on side-by-side inspection"
// range, so visually-identical matches always score a clean 10.00.
//
// deltaZero is wider than a strict perceptual cutoff so large lightness/chroma
// shifts in the same hue family (e.g. navy vs sky blue) can still earn partial
// credit. PointsFromGuessHex applies an extra ab-plane alignment factor before
// mapping to points; the JSON API still reports raw ΔE76 for transparency.
const (
	deltaPerfect = 5.0
	deltaZero    = 72.0
	// decayExponent > 1 makes the score curve concave: small errors lose
	// fewer points, while large errors still fall off toward zero.
	decayExponent = 1.35

	minChromaHueGate = 5.0
	// hueRelaxMax scales raw ΔE down when (a*,b*) vectors align — same broad
	// colour direction with different L*/chroma — without helping opposite hues.
	hueRelaxMax = 0.32
)

// scoringDelta applies a hue-family relaxation to raw ΔE76 for point mapping only.
func scoringDelta(hexA, hexB string, rawDelta float64) float64 {
	if rawDelta <= deltaPerfect || rawDelta == 0 {
		return rawDelta
	}
	_, aa, ba := hexToLab(hexA)
	_, ab, bb := hexToLab(hexB)
	ca := math.Hypot(aa, ba)
	cb := math.Hypot(ab, bb)
	if math.Min(ca, cb) < minChromaHueGate {
		return rawDelta
	}
	denom := ca*cb + 1e-9
	cosTheta := (aa*ab + ba*bb) / denom
	if cosTheta < -1 {
		cosTheta = -1
	}
	if cosTheta > 1 {
		cosTheta = 1
	}
	// 0 = opposite direction in ab plane, 1 = same direction (hue ballpark).
	align := (cosTheta + 1) / 2
	chromaRatio := math.Min(ca, cb) / math.Max(ca, cb)
	weight := align * chromaRatio
	factor := 1 - hueRelaxMax*weight
	if factor < 0.55 {
		factor = 0.55
	}
	return rawDelta * factor
}

// EffectiveScoringDelta is the relaxed ΔE fed to PointsFromDeltaE after hue-family
// adjustment. The API reports raw DeltaE76 only; this stays internal to scoring.
func EffectiveScoringDelta(correctHex, guessHex string) float64 {
	raw := DeltaE76(correctHex, guessHex)
	return scoringDelta(correctHex, guessHex, raw)
}

// PointsFromGuessHex maps a guessed colour to a score using raw ΔE76 plus a
// small relaxation when both colours share a similar direction in CIELAB a*b*.
func PointsFromGuessHex(correctHex, guessHex string) int {
	return PointsFromDeltaE(EffectiveScoringDelta(correctHex, guessHex))
}

// DeltaE76 computes CIE76 ΔE* in Lab space (sRGB hex inputs, D65).
func DeltaE76(hexA, hexB string) float64 {
	la, aa, ba := hexToLab(hexA)
	lb, ab, bb := hexToLab(hexB)
	dL := la - lb
	da := aa - ab
	db := ba - bb
	return math.Sqrt(dL*dL + da*da + db*db)
}

// PointsFromDeltaE maps ΔE to 0..PointsMaxPerRound using a concave decay
// between deltaPerfect and deltaZero. Players are rewarded generously for
// near-misses; the score still trends toward zero as the colour drifts far.
func PointsFromDeltaE(deltaE float64) int {
	if deltaE <= deltaPerfect {
		return PointsMaxPerRound
	}
	if deltaE >= deltaZero {
		return 0
	}
	x := (deltaE - deltaPerfect) / (deltaZero - deltaPerfect)
	t := 1 - math.Pow(x, decayExponent)
	return int(math.Round(float64(PointsMaxPerRound) * t))
}

func hexToLab(hex string) (L, a, b float64) {
	h := strings.TrimPrefix(strings.TrimSpace(strings.ToUpper(hex)), "#")
	if len(h) != 6 {
		return 0, 0, 0
	}
	r := srgbToLinear(float64(hexByte(h[0:2])) / 255)
	g := srgbToLinear(float64(hexByte(h[2:4])) / 255)
	bl := srgbToLinear(float64(hexByte(h[4:6])) / 255)
	x, y, z := linearRGBToXYZ(r, g, bl)
	return xyzToLab(x, y, z)
}

func hexByte(two string) int {
	return hexNibble(two[0])<<4 | hexNibble(two[1])
}

func hexNibble(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return 0
	}
}

func srgbToLinear(u float64) float64 {
	if u <= 0.04045 {
		return u / 12.92
	}
	return math.Pow((u+0.055)/1.055, 2.4)
}

func linearRGBToXYZ(r, g, b float64) (x, y, z float64) {
	x = r*0.4124564 + g*0.3575761 + b*0.1804375
	y = r*0.2126729 + g*0.7151522 + b*0.0721750
	z = r*0.0193339 + g*0.1191920 + b*0.9503041
	return x, y, z
}

func xyzToLab(x, y, z float64) (L, a, b float64) {
	xn, yn, zn := 0.95047, 1.00000, 1.08883
	x /= xn
	y /= yn
	z /= zn

	f := func(t float64) float64 {
		if t > 0.008856 {
			return math.Cbrt(t)
		}
		return (7.787*t + 16.0/116.0)
	}

	fx, fy, fz := f(x), f(y), f(z)
	L = 116*fy - 16
	a = 500 * (fx - fy)
	b = 200 * (fy - fz)
	return L, a, b
}
