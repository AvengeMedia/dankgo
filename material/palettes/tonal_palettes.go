// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package palettes

import (
	"math"

	"github.com/AvengeMedia/dankgo/material/color"
)

// TonalPalette is a convenience type for retrieving colors that are constant in
// hue and chroma, but vary in tone.
//
// Each TonalPalette is initialized with a hue and chroma, and provides a cache
// for efficient tone retrieval.
type TonalPalette struct {
	cache    map[float64]color.ARGB
	Hue      float64
	Chroma   float64
	KeyColor color.Hct
}

// NewFromARGB creates a TonalPalette from an ARGB color.
//
// The palette’s hue and chroma will match those of the provided color,
// allowing tone values to vary while keeping the same hue and chroma.
func NewFromARGB(c color.ARGB) *TonalPalette {
	hct := c.ToHct()
	return NewFromHct(hct)
}

// NewFromHct creates a TonalPalette from a given HCT color.
//
// The resulting palette will have the same hue and chroma as the provided HCT.
func NewFromHct(hct color.Hct) *TonalPalette {
	return &TonalPalette{
		Hue:      hct.Hue,
		Chroma:   hct.Chroma,
		KeyColor: hct,
	}
}

// FromHueAndChroma creates a TonalPalette from explicit hue and chroma values.
//
// It internally generates a "key color" that best represents the given hue and
// chroma, and constructs a tonal palette around it.
func FromHueAndChroma(hue, chroma float64) *TonalPalette {
	return &TonalPalette{
		Hue:      hue,
		Chroma:   chroma,
		KeyColor: NewKeyColor(hue, chroma).Create(),
	}
}

// Tone returns the ARGB representation of a color at a given tone (0–100).
//
// The tone defines the perceived lightness of the color, where 0 is black and
// 100 is white. Results are cached for subsequent retrievals.
func (tp *TonalPalette) Tone(tone float64) color.ARGB {
	if tp.cache == nil {
		tp.cache = map[float64]color.ARGB{}
	}

	argb, ok := tp.cache[tone]
	if ok {
		return argb
	}
	if tone == 99 && color.IsYellow(tp.Hue) {
		argb = averageARGB(tp.Tone(98), tp.Tone(100))
	} else {
		argb = color.NewHct(tp.Hue, tp.Chroma, tone).ToARGB()
	}
	tp.cache[tone] = argb
	return argb
}

func averageARGB(a, b color.ARGB) color.ARGB {
	mean := func(x, y uint8) uint8 {
		return uint8(math.Round((float64(x) + float64(y)) / 2))
	}
	return color.ARGBFromRGB(mean(a.Red(), b.Red()), mean(a.Green(), b.Green()), mean(a.Blue(), b.Blue()))
}

// GetHct returns the HCT representation of a color with the specified tone.
//
// This provides access to hue, chroma, and tone values directly in HCT space.
func (tp *TonalPalette) GetHct(tone float64) color.Hct {
	return tp.Tone(tone).ToHct()
}

// IsBlue determines if a hue is in the blue range.
func (tp *TonalPalette) IsBlue() bool {
	return tp.Hue >= 250 && tp.Hue < 270
}

// IsYellow determines if a hue is in the yellow range.
func (tp *TonalPalette) IsYellow() bool {
	return tp.Hue >= 105 && tp.Hue < 125
}

// IsCyan determines if a hue is in the cyan range.
func (tp *TonalPalette) IsCyan() bool {
	return tp.Hue >= 170 && tp.Hue < 207
}
