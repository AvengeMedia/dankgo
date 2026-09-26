// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package palettes

import (
	"math"

	"github.com/AvengeMedia/dankgo/material/color"
)

// KeyColor is a color that represents the hue and chroma of a tonal palette
type KeyColor struct {
	// hue is the hue of the key color
	hue float64
	// requestedChroma is the chroma of the key color
	requestedChroma float64
	// chromaCache maps tone to max chroma to avoid duplicated HCT calculation
	chromaCache map[float64]float64
	// maxChromaValue is the maximum possible chroma value
	maxChromaValue float64
}

// NewKeyColor creates a new KeyColor with the given hue and chroma
func NewKeyColor(hue, requestedChroma float64) *KeyColor {
	return &KeyColor{
		hue:             hue,
		requestedChroma: requestedChroma,
		chromaCache:     map[float64]float64{},
		maxChromaValue:  200.0,
	}
}

// maxChroma calculates the maximum chroma for a given tone
// This is a placeholder for the actual implementation
func (k *KeyColor) maxChroma(tone float64) float64 {
	if chroma, exists := k.chromaCache[tone]; exists {
		return chroma
	}

	chroma := color.NewHct(k.hue, k.maxChromaValue, tone).Chroma
	k.chromaCache[tone] = chroma
	return chroma
}

// Create creates a key color from the hue and chroma
// The key color is the first tone, starting from T50, matching the given hue
// and chroma.
//
// Returns an Hct color value
func (k *KeyColor) Create() color.Hct {
	// Pivot around T50 because T50 has the most chroma available, on
	// average. Thus it is most likely to have a direct answer.
	pivotTone := 50.0
	toneStepSize := 1.0
	// Epsilon to accept values slightly higher than the requested chroma.
	epsilon := 0.01

	lowerTone := 0.0
	upperTone := 100.0

	for lowerTone < upperTone {
		midTone := math.Floor((lowerTone + upperTone) / 2)
		isAscending := k.maxChroma(midTone) < k.maxChroma(midTone+toneStepSize)
		sufficientChroma := k.maxChroma(midTone) >= k.requestedChroma-epsilon

		if sufficientChroma {
			if math.Abs(lowerTone-pivotTone) < math.Abs(upperTone-pivotTone) {
				upperTone = midTone
			} else {
				if lowerTone == midTone {
					return color.NewHct(k.hue, k.requestedChroma, lowerTone)
				}
				lowerTone = midTone
			}
		} else {
			// As there is no sufficient chroma in the midTone, follow the
			// direction
			// to the chroma peak.
			if isAscending {
				lowerTone = midTone + toneStepSize
			} else {
				// Keep midTone for potential chroma peak.
				upperTone = midTone
			}
		}
	}

	return color.NewHct(k.hue, k.requestedChroma, lowerTone)
}
