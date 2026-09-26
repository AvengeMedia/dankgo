// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package contrast

import (
	"math"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/AvengeMedia/dankgo/material/num"
)

// RatioOfTones calculates the contrast ratio between two tones. toneA and toneB
// should be in the range [0, 100]. Values outside this range are clamped. The
// return value is the contrast ratio in the range [1, 21].
func RatioOfTones(toneA, toneB float64) float64 {
	toneA = num.Clamp(0, 100, toneA)
	toneB = num.Clamp(0, 100, toneB)
	return RatioOfYs(color.YFromLstar(toneA), color.YFromLstar(toneB))
}

// RatioOfYs calculates the contrast ratio between two relative luminance
// values. The ratio is computed as (lighter + 5) / (darker + 5).
func RatioOfYs(y1, y2 float64) float64 {
	lighter := max(y1, y2)
	darker := y2
	if lighter == y2 {
		darker = y1
	}
	return (lighter + 5.0) / (darker + 5.0)
}

// Lighter returns a tone lighter than the given tone and that meets the
// specified contrast ratio. The tone must be in the range [0, 100] and ratio
// must be in the range [1, 21]. It returns -1 if the ratio cannot be met.
func Lighter(tone float64, ratio float64) float64 {
	if tone < 0 || tone > 100 {
		return -1
	}

	darkY := color.YFromLstar(tone)
	lightY := ratio*(darkY+5.0) - 5.0
	realContrast := RatioOfYs(lightY, darkY)
	delta := math.Abs(realContrast - ratio)
	if realContrast < ratio && delta > 0.04 {
		return -1
	}

	// Ensure gamut mapping, which requires a 'range' on tone, will still result
	// in the correct ratio by brightening slightly.
	returnValue := color.LstarFromY(lightY) + 0.4
	if returnValue < 0 || returnValue > 100 {
		return -1
	}
	return returnValue
}

// Darker returns a tone darker than the given tone and that meets the specified
// contrast ratio. The tone must be in the range [0, 100] and ratio must be in
// the range [1, 21]. It returns -1 if the ratio cannot be met.
func Darker(tone, ratio float64) float64 {
	if tone < 0.0 || tone > 100.0 {
		return -1.0
	}

	lightY := color.YFromLstar(tone)
	darkY := ((lightY + 5.0) / ratio) - 5.0
	realContrast := RatioOfYs(lightY, darkY)

	delta := math.Abs(realContrast - ratio)
	if realContrast < ratio && delta > 0.04 {
		return -1
	}

	// Ensure gamut mapping, which requires a 'range' on tone, will still result
	// in the correct ratio by darkening slightly.
	returnValue := color.LstarFromY(darkY) - 0.4
	if returnValue < 0 || returnValue > 100 {
		return -1
	}
	return returnValue
}

// LighterUnsafe is like Lighter but always returns a valid tone even if the
// ratio cannot be met. If the ratio cannot be met, it returns 100.
func LighterUnsafe(tone, ratio float64) float64 {
	lighterSafe := Lighter(tone, ratio)
	if lighterSafe < 0 {
		return 100
	}
	return lighterSafe
}

// DarkerUnsafe is like Darker but always returns a valid tone even if the ratio
// cannot be met. If the ratio cannot be met, it returns 0.
func DarkerUnsafe(tone, ratio float64) float64 {
	darkerSafe := Darker(tone, ratio)
	if darkerSafe < 0 {
		return 0
	}
	return darkerSafe
}
