// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package dynamic

import (
	"math"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/AvengeMedia/dankgo/material/contrast"
	"github.com/AvengeMedia/dankgo/material/palettes"
)

// TonalPaletteFunc returns a TonalPalette based on the provided Scheme.
// A TonalPalette is defined by a hue and chroma, allowing chroma to be
// preserved when contrast adjustments are made, instead of directly specifying
// hue/chroma.
type TonalPaletteFunc func(s *Scheme) palettes.TonalPalette

// ToneFunc returns the tone (lightness) of the color for a given Scheme. If not
// explicitly provided, it defaults to the tone of the background, or 50 if
// there is no background.
type ToneFunc func(s *Scheme) float64

// ChromaMultiplier returns a multiplier for the chroma value based on the
// Scheme. This is used to scale the chroma of the color and defaults to 1 when
// unspecified.
type ChromaMultiplier func(s *Scheme) float64

// ColorFunc returns a pointer to a Color based on the Scheme. Typically used to
// reference background or related colors dynamically.
type ColorFunc func(s *Scheme) *Color

// ToneDeltaPairFunc returns a pointer to a ToneDeltaPair for the given Scheme.
// A ToneDeltaPair enforces a tone difference constraint between two colors,
// where one must be the color being constructed.
type ToneDeltaPairFunc func(s *Scheme) *ToneDeltaPair

// ContrastCurveFunc returns a pointer to a ContrastCurve based on the Scheme. A
// ContrastCurve defines how a color's contrast behaves at various contrast
// levels. This is typically used in conjunction with a background color.
type ContrastCurveFunc func(s *Scheme) *ContrastCurve

// Color represents a color in a dynamic color scheme
type Color struct {
	Name             string
	Palette          TonalPaletteFunc
	Tone             ToneFunc
	ChromaMultiplier ChromaMultiplier
	IsBackground     bool
	Background       ColorFunc
	SecondBackground ColorFunc
	ToneDeltaPair    ToneDeltaPairFunc
	ContrastCurve    ContrastCurveFunc
}

// ForegroundTone calculates a foreground tone that has sufficient contrast with
// a background tone
func ForegroundTone(bgTone, ratio float64) float64 {
	lighterTone := contrast.LighterUnsafe(bgTone, ratio)
	darkerTone := contrast.DarkerUnsafe(bgTone, ratio)
	lighterRatio := contrast.RatioOfTones(lighterTone, bgTone)
	darkerRatio := contrast.RatioOfTones(darkerTone, bgTone)
	preferLighter := TonePrefersLightForeground(bgTone)

	if preferLighter {
		negligibleDifference := (math.Abs(lighterRatio-darkerRatio) < 0.1 &&
			lighterRatio < ratio && darkerRatio < ratio)
		if lighterRatio >= ratio || lighterRatio >= darkerRatio ||
			negligibleDifference {
			return lighterTone
		}
		return darkerTone
	}
	if darkerRatio >= ratio || darkerRatio >= lighterRatio {
		return darkerTone
	}
	return lighterTone
}

// GetInitialToneFromBackground returns initial tone from given background
// ColorFunc. Returns 50 if background ColorFunc is nil.
func GetInitialToneFromBackground(background ColorFunc) ToneFunc {
	if background == nil {
		return func(*Scheme) float64 {
			return 50
		}
	}
	return func(s *Scheme) float64 { return background(s).GetTone(s) }
}

// EnableLightForeground adjusts a tone to enable light foreground if needed.
func EnableLightForeground(tone float64) float64 {
	if TonePrefersLightForeground(tone) && !ToneAllowsLightForeground(tone) {
		return 49.0
	}
	return tone
}

// TonePrefersLightForeground determines if a tone prefers light foreground.
func TonePrefersLightForeground(tone float64) bool {
	return math.Round(tone) < 60
}

// ToneAllowsLightForeground determines if a tone allows light foreground.
func ToneAllowsLightForeground(tone float64) bool {
	return math.Round(tone) <= 49
}

// FromPalette creates a DynamicColor from a palette and tone function.
func FromPalette(name string, palette TonalPaletteFunc, tone ToneFunc) *Color {
	return &Color{Name: name, Palette: palette, Tone: tone}
}

// GetArgb returns the ARGB value for the DynamicColor in the given scheme.
func (dc *Color) GetArgb(scheme *Scheme) color.ARGB {
	return dc.GetHct(scheme).ToARGB()
}

// GetHct returns the HCT color for the DynamicColor in the given scheme.
func (dc *Color) GetHct(scheme *Scheme) color.Hct {
	if scheme.Version == Version2025 {
		return ColorCalculation2025.GetHct(scheme, dc)
	}
	return ColorCalculation2021.GetHct(scheme, dc)
}

// GetTone retuns Tone for the dynamic color using given scheme.
func (dc *Color) GetTone(scheme *Scheme) float64 {
	if scheme.Version == Version2025 {
		return ColorCalculation2025.GetTone(scheme, dc)
	}
	return ColorCalculation2021.GetTone(scheme, dc)
}
