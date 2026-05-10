package flagcolor

import (
	"math"
	"strings"
)

// PointsMaxPerRound is the highest score achievable on one round before difficulty scaling.
const PointsMaxPerRound = 100

const deltaPerfect = 2.0
const deltaZero = 40.0

// DeltaE76 computes CIE76 ΔE* in Lab space (sRGB hex inputs, D65).
func DeltaE76(hexA, hexB string) float64 {
	la, aa, ba := hexToLab(hexA)
	lb, ab, bb := hexToLab(hexB)
	dL := la - lb
	da := aa - ab
	db := ba - bb
	return math.Sqrt(dL*dL + da*da + db*db)
}

// PointsFromDeltaE maps ΔE to 0..PointsMaxPerRound (linear decay between thresholds).
func PointsFromDeltaE(deltaE float64) int {
	if deltaE <= deltaPerfect {
		return PointsMaxPerRound
	}
	if deltaE >= deltaZero {
		return 0
	}
	t := 1 - (deltaE-deltaPerfect)/(deltaZero-deltaPerfect)
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
