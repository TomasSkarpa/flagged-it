package flagcolor

import (
	"fmt"
)

// GuessHexFromRGB builds a normalized uppercase hex color from sRGB bytes.
func GuessHexFromRGB(r, g, b int) (string, error) {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return "", fmt.Errorf("rgb out of range")
	}
	return fmt.Sprintf("#%02X%02X%02X", r, g, b), nil
}
