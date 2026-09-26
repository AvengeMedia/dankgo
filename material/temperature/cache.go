// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package temperature

import (
	"math"
	"slices"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/AvengeMedia/dankgo/material/num"
)

// HctMap is map containing Hct hash as key and temperature as value.
type HctMap map[color.Hash]float64

// Cache provides design utilities using color temperature theory.
//
// It handles analogous colors, complementary color, and uses cache to
// efficiently and lazily generate data for calculations when needed.
type Cache struct {
	input                         color.Hct
	hctsByTempCache               []color.Hct
	hctsByHueCache                []color.Hct
	tempsByHctCache               HctMap
	inputRelativeTemperatureCache float64
	complementCache               *color.Hct
}

// NewCache creates a new TemperatureCache with the given color.Hct color.
func NewCache(input color.Hct) Cache {
	return Cache{
		input:                         input,
		hctsByTempCache:               []color.Hct{},
		hctsByHueCache:                []color.Hct{},
		tempsByHctCache:               HctMap{},
		inputRelativeTemperatureCache: -1.0,
		complementCache:               nil,
	}
}

// HctsByTemp returns a slice of Hct colors sorted by temperature.
func (t *Cache) HctsByTemp() []color.Hct {
	if len(t.hctsByTempCache) > 0 {
		return t.hctsByTempCache
	}

	hcts := append(t.HctsByHue(), t.input)
	temperaturesByHct := t.TempsByHct()

	slices.SortStableFunc(hcts, func(a, b color.Hct) int {
		return num.SignCmp(temperaturesByHct[a.Hash()], temperaturesByHct[b.Hash()])
	})
	t.hctsByTempCache = hcts
	return hcts
}

// Warmest returns the warmest color in the cache.
func (t *Cache) Warmest() color.Hct {
	hcts := t.HctsByTemp()
	return hcts[len(hcts)-1]
}

// Coldest returns the coldest color in the cache.
func (t *Cache) Coldest() color.Hct {
	return t.HctsByTemp()[0]
}

// Analogous returns a set of colors with differing hues, equidistant in
// temperature.
//
// In art, this is usually described as a set of 5 colors on a color wheel
// divided into 12 sections. This method allows provision of either of those
// values.
//
// Behavior is undefined when count or divisions is 0. When divisions < count,
// colors repeat.
//
// Parameters:
//   - count: The number of colors to return, includes the input color. Default
//     is 5.
//   - divisions: The number of divisions on the color wheel. Default is 12.
func (t *Cache) Analogous(count, divisions int) []color.Hct {
	if count == 0 {
		count = 5
	}
	if divisions == 0 {
		divisions = 12
	}

	startHue := int(math.Round(t.input.Hue))
	startHct := t.HctsByHue()[startHue]
	allColors := []color.Hct{startHct}

	hctByHue := t.HctsByHue()

	lastTemp := t.RelativeTemperature(startHct)
	absoluteTotalTempDelta := 0.0

	for i := range 360 {
		hue := num.NormalizeDegreeInt(startHue + i)
		hct := hctByHue[hue]
		temp := t.RelativeTemperature(hct)
		tempDelta := math.Abs(temp - lastTemp)
		lastTemp = temp
		absoluteTotalTempDelta += tempDelta
	}

	hueAddend := 1
	tempStep := absoluteTotalTempDelta / float64(divisions)
	totalTempDelta := 0.0
	lastTemp = t.RelativeTemperature(startHct)

	for len(allColors) < divisions {
		hue := num.NormalizeDegreeInt(startHue + hueAddend)
		hct := t.HctsByHue()[hue]
		temp := t.RelativeTemperature(hct)
		tempDelta := math.Abs(temp - lastTemp)
		totalTempDelta += tempDelta

		desiredTotalTempDeltaForIndex := float64(len(allColors)) * tempStep
		indexSatisfied := totalTempDelta >= desiredTotalTempDeltaForIndex
		indexAddend := 1

		// Keep adding this hue to the answers until its temperature is
		// insufficient. This ensures consistent behavior when there aren't
		// [divisions] discrete steps between 0 and 360 in hue with [tempStep]
		// delta in temperature between them.
		//
		// For example, white and black have no analogues: there are no other
		// colors at T100/T0. Therefore, they should just be added to the array
		// as answers.
		for indexSatisfied && len(allColors) < divisions {
			allColors = append(allColors, hct)
			desiredTotalTempDeltaForIndex = float64(
				len(allColors)+indexAddend,
			) * tempStep
			indexSatisfied = totalTempDelta >= desiredTotalTempDeltaForIndex
			indexAddend++
		}

		lastTemp = temp
		hueAddend++
		if hueAddend > 360 {
			for len(allColors) < divisions {
				allColors = append(allColors, hct)
			}
			break
		}
	}

	answers := []color.Hct{t.input}

	// First, generate analogues from rotating counter-clockwise.
	increaseHueCount := int(math.Floor(float64(count-1) / 2.0))
	for i := 1; i < (increaseHueCount + 1); i++ {
		index := 0 - i
		for index < 0 {
			index = len(allColors) + index
		}
		if index >= len(allColors) {
			index = index % len(allColors)
		}
		answers = append([]color.Hct{allColors[index]}, answers...)
	}

	// Second, generate analogues from rotating clockwise.
	decreaseHueCount := count - increaseHueCount - 1
	for i := 1; i < (decreaseHueCount + 1); i++ {
		index := i
		for index < 0 {
			index = len(allColors) + index
		}
		if index >= len(allColors) {
			index = index % len(allColors)
		}
		answers = append(answers, allColors[index])
	}

	return answers
}

// Complement returns a color that complements the input color aesthetically.
//
// In art, this is usually described as being across the color wheel.
// History of this shows intent as a color that is just as cool-warm as the
// input color is warm-cool.
func (t *Cache) Complement() color.Hct {
	if t.complementCache != nil {
		return *t.complementCache
	}

	coldestHue := t.Coldest().Hue
	coldestTemp := t.TempsByHct()[t.Coldest().Hash()]

	warmestHue := t.Warmest().Hue
	warmestTemp := t.TempsByHct()[t.Warmest().Hash()]

	r := warmestTemp - coldestTemp
	startHueIsColdestToWarmest := isBetween(t.input.Hue, coldestHue, warmestHue)

	var startHue, endHue float64
	if startHueIsColdestToWarmest {
		startHue = warmestHue
		endHue = coldestHue
	} else {
		startHue = coldestHue
		endHue = warmestHue
	}

	directionOfRotation := 1.0
	smallestError := 1000.0
	answer := t.HctsByHue()[int(math.Round(t.input.Hue))]

	complementRelativeTemp := 1.0 - t.InputRelativeTemperature()

	// Find the color in the other section, closest to the inverse percentile
	// of the input color. This is the complement.
	for hueAddend := 0.0; hueAddend <= 360.0; hueAddend += 1.0 {
		hue := num.NormalizeDegree(startHue + directionOfRotation*hueAddend)
		if !isBetween(hue, startHue, endHue) {
			continue
		}

		possibleAnswer := t.HctsByHue()[int(math.Round(hue))]
		relativeTemp := (t.TempsByHct()[possibleAnswer.Hash()] - coldestTemp) / r
		err := math.Abs(complementRelativeTemp - relativeTemp)

		if err < smallestError {
			smallestError = err
			answer = possibleAnswer
		}
	}

	t.complementCache = &answer
	return answer
}

// RelativeTemperature returns temperature relative to all colors with the same
// chroma and tone.
// Value on a scale from 0 to 1.
func (t *Cache) RelativeTemperature(hct color.Hct) float64 {
	r := t.TempsByHct()[t.Warmest().Hash()] - t.TempsByHct()[t.Coldest().Hash()]
	differenceFromColdest := t.TempsByHct()[hct.Hash()] - t.TempsByHct()[t.Coldest().Hash()]

	// Handle when there's no difference in temperature between warmest and
	// coldest: for example, at T100, only one color is available, white.
	if r == 0.0 {
		return 0.5
	}

	return differenceFromColdest / r
}

// InputRelativeTemperature returns the relative temperature of the input color.
func (t *Cache) InputRelativeTemperature() float64 {
	if t.inputRelativeTemperatureCache >= 0.0 {
		return t.inputRelativeTemperatureCache
	}

	t.inputRelativeTemperatureCache = t.RelativeTemperature(t.input)
	return t.inputRelativeTemperatureCache
}

// TempsByHct returns a map with keys of HCTs and values of raw temperature.
func (t *Cache) TempsByHct() HctMap {
	if len(t.tempsByHctCache) > 0 {
		return t.tempsByHctCache
	}

	allHcts := append(t.HctsByHue(), t.input)
	temperaturesByHct := HctMap{}

	for _, e := range allHcts {
		temperaturesByHct[e.Hash()] = RawTemperature(e)
	}

	t.tempsByHctCache = temperaturesByHct
	return t.tempsByHctCache
}

// HctsByHue returns HCTs for all hues, with the same chroma/tone as the input.
// Sorted ascending, hue 0 to 360.
func (t *Cache) HctsByHue() []color.Hct {
	if len(t.hctsByHueCache) > 0 {
		return t.hctsByHueCache
	}

	hcts := make([]color.Hct, 361)
	for hue := 0.0; hue <= 360.0; hue += 1.0 {
		colorAtHue := color.NewHct(hue, t.input.Chroma, t.input.Tone)
		hcts[int(hue)] = colorAtHue
	}

	t.hctsByHueCache = hcts
	return t.hctsByHueCache
}

// isBetween determines if an angle is between two other angles, rotating
// clockwise.
func isBetween(angle, a, b float64) bool {
	if a < b {
		return a <= angle && angle <= b
	}
	return a <= angle || angle <= b
}

// RawTemperature calculates the raw temperature value of a color.
//
// Value representing cool-warm factor of a color. Values below 0 are considered
// cool, above, warm.
//
// Color science has researched emotion and harmony, which art uses to select
// colors. Warm-cool is the foundation of analogous and complementary colors.
// See:
// - Li-Chen Ou's Chapter 19 in Handbook of Color Psychology (2015).
// - Josef Albers' Interaction of Color chapters 19 and 21.
//
// Implementation of Ou, Woodcock and Wright's algorithm, which uses
// L*a*b* / LCH color space.
// Return value has these properties:
//   - Values below 0 are cool, above 0 are warm.
//   - Lower bound: -0.52 - (chroma ^ 1.07 / 20). L*a*b* chroma is infinite.
//     Assuming max of 130 chroma, -9.66.
//   - Upper bound: -0.52 + (chroma ^ 1.07 / 20). L*a*b* chroma is infinite.
//     Assuming max of 130 chroma, 8.61.
func RawTemperature(c color.Hct) float64 {
	lab := c.ToARGB().ToLab()

	hue := num.NormalizeDegree(math.Atan2(lab.B, lab.A) * 180.0 / math.Pi)
	chroma := math.Sqrt((lab.A * lab.A) + (lab.B * lab.B))

	temperature := -0.5 + 0.02*math.Pow(chroma, 1.07)*
		math.Cos(num.NormalizeDegree(hue-50.0)*math.Pi/180.0)
	return temperature
}
