// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package dynamic

import (
	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/AvengeMedia/dankgo/material/dislike"
	"github.com/AvengeMedia/dankgo/material/num"
	"github.com/AvengeMedia/dankgo/material/palettes"
	"github.com/AvengeMedia/dankgo/material/temperature"
)

// SchemePalettesDelegate is an interface for the palettes of a DynamicScheme
type SchemePalettesDelegate interface {
	GetPrimaryPalette(
		variant Variant,
		sourceColorHct color.Hct,
		dark bool,
		platform Platform,
		contrastLevel float64,
	) *palettes.TonalPalette
	GetSecondaryPalette(
		variant Variant,
		sourceColorHct color.Hct,
		dark bool,
		platform Platform,
		contrastLevel float64,
	) *palettes.TonalPalette
	GetTertiaryPalette(
		variant Variant,
		sourceColorHct color.Hct,
		dark bool,
		platform Platform,
		contrastLevel float64,
	) *palettes.TonalPalette
	GetNeutralPalette(
		variant Variant,
		sourceColorHct color.Hct,
		dark bool,
		platform Platform,
		contrastLevel float64,
	) *palettes.TonalPalette
	GetNeutralVariantPalette(
		variant Variant,
		sourceColorHct color.Hct,
		dark bool,
		platform Platform,
		contrastLevel float64,
	) *palettes.TonalPalette
	GetErrorPalette(
		variant Variant,
		sourceColorHct color.Hct,
		dark bool,
		platform Platform,
		contrastLevel float64,
	) *palettes.TonalPalette
}

type (
	// schemePalettesDelegateImpl2021 implements the palettes delegate for the
	// 2021
	// spec.
	schemePalettesDelegateImpl2021 struct{}
	// schemePalettesDelegateImpl2025 extends the 2021 implementation for the
	// 2025
	// spec.
	schemePalettesDelegateImpl2025 struct{ schemePalettesDelegateImpl2021 }
)

var (
	_ SchemePalettesDelegate = (*schemePalettesDelegateImpl2021)(nil)
	_ SchemePalettesDelegate = (*schemePalettesDelegateImpl2025)(nil)
)

// GetPrimaryPalette returns the primary palette for a given variant and color
func (d *schemePalettesDelegateImpl2021) GetPrimaryPalette(
	variant Variant, sourceColorHct color.Hct, _ bool, _ Platform, _ float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantContent, VariantFidelity:
		return palettes.FromHueAndChroma(
			sourceColorHct.Hue,
			sourceColorHct.Chroma,
		)
	case VariantFruitSalad:
		return palettes.FromHueAndChroma(
			num.NormalizeDegree(sourceColorHct.Hue-50.0),
			48.0,
		)
	case VariantMonochrome:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 0.0)
	case VariantNeutral:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 12.0)
	case VariantRainbow:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 48.0)
	case VariantTonalSpot:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 36.0)
	case VariantExpressive:
		return palettes.FromHueAndChroma(
			num.NormalizeDegree(sourceColorHct.Hue+240),
			40,
		)
	case VariantVibrant:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 200.0)
	default:
		panic("Unsupported variant")
	}
}

// GetSecondaryPalette returns the secondary palette for a given variant and
// color
func (d *schemePalettesDelegateImpl2021) GetSecondaryPalette(
	variant Variant, sourceColorHct color.Hct, _ bool, _ Platform, _ float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantContent, VariantFidelity:
		return palettes.FromHueAndChroma(
			sourceColorHct.Hue,
			max(sourceColorHct.Chroma-32.0, sourceColorHct.Chroma*0.5))
	case VariantFruitSalad:
		return palettes.FromHueAndChroma(
			num.NormalizeDegree(sourceColorHct.Hue-50.0),
			36.0,
		)
	case VariantMonochrome:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 0.0)
	case VariantNeutral:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 8.0)
	case VariantRainbow:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 16.0)
	case VariantTonalSpot:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 16.0)
	case VariantExpressive:
		return palettes.FromHueAndChroma(
			GetRotatedHue(
				sourceColorHct,
				[]float64{0, 21, 51, 121, 151, 191, 271, 321, 360},
				[]float64{45, 95, 45, 20, 45, 90, 45, 45, 45}),
			24.0)
	case VariantVibrant:
		return palettes.FromHueAndChroma(
			GetRotatedHue(
				sourceColorHct,
				[]float64{0, 41, 61, 101, 131, 181, 251, 301, 360},
				[]float64{18, 15, 10, 12, 15, 18, 15, 12, 12}),
			24.0)
	default:
		panic("Unsupported variant")
	}
}

// GetTertiaryPalette returns the tertiary palette for a given variant and color
func (d *schemePalettesDelegateImpl2021) GetTertiaryPalette(
	variant Variant, sourceColorHct color.Hct, _ bool, _ Platform, _ float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantContent:
		tempCache := temperature.NewCache(sourceColorHct)
		analogous := tempCache.Analogous(3, 6)
		return palettes.NewFromHct(dislike.FixIfDisliked(analogous[2]))
	case VariantFidelity:
		tempCache := temperature.NewCache(sourceColorHct)
		return palettes.NewFromHct(
			dislike.FixIfDisliked(tempCache.Complement()),
		)
	case VariantFruitSalad:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 36.0)
	case VariantMonochrome:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 0.0)
	case VariantNeutral:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 16.0)
	case VariantRainbow, VariantTonalSpot:
		return palettes.FromHueAndChroma(
			num.NormalizeDegree(sourceColorHct.Hue+60.0),
			24.0,
		)
	case VariantExpressive:
		return palettes.FromHueAndChroma(
			GetRotatedHue(
				sourceColorHct,
				[]float64{0, 21, 51, 121, 151, 191, 271, 321, 360},
				[]float64{120, 120, 20, 45, 20, 15, 20, 120, 120}),
			32.0)
	case VariantVibrant:
		return palettes.FromHueAndChroma(
			GetRotatedHue(
				sourceColorHct,
				[]float64{0, 41, 61, 101, 131, 181, 251, 301, 360},
				[]float64{35, 30, 20, 25, 30, 35, 30, 25, 25}),
			32.0)
	default:
		panic("Unsupported variant")
	}
}

// GetNeutralPalette returns the neutral palette for a given variant and color
func (d *schemePalettesDelegateImpl2021) GetNeutralPalette(
	variant Variant, sourceColorHct color.Hct, _ bool, _ Platform, _ float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantContent, VariantFidelity:
		return palettes.FromHueAndChroma(
			sourceColorHct.Hue,
			sourceColorHct.Chroma/8.0,
		)
	case VariantFruitSalad:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 10.0)
	case VariantMonochrome:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 0.0)
	case VariantNeutral:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 2.0)
	case VariantRainbow:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 0.0)
	case VariantTonalSpot:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 6.0)
	case VariantExpressive:
		return palettes.FromHueAndChroma(
			num.NormalizeDegree(sourceColorHct.Hue+15),
			8,
		)
	case VariantVibrant:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 10)
	default:
		panic("Unsupported variant")
	}
}

// GetNeutralVariantPalette returns the neutral variant palette for a given
// variant and color
func (d *schemePalettesDelegateImpl2021) GetNeutralVariantPalette(
	variant Variant, sourceColorHct color.Hct, _ bool, _ Platform, _ float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantContent:
		return palettes.FromHueAndChroma(
			sourceColorHct.Hue,
			(sourceColorHct.Chroma/8.0)+4.0,
		)
	case VariantFidelity:
		return palettes.FromHueAndChroma(
			sourceColorHct.Hue,
			(sourceColorHct.Chroma/8.0)+4.0,
		)
	case VariantFruitSalad:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 16.0)
	case VariantMonochrome:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 0.0)
	case VariantNeutral:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 2.0)
	case VariantRainbow:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 0.0)
	case VariantTonalSpot:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 8.0)
	case VariantExpressive:
		return palettes.FromHueAndChroma(
			num.NormalizeDegree(sourceColorHct.Hue+15),
			12,
		)
	case VariantVibrant:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 12)
	default:
		panic("Unsupported variant")
	}
}

// GetErrorPalette returns the error palette for a given variant and color
func (d *schemePalettesDelegateImpl2021) GetErrorPalette(
	Variant, color.Hct, bool, Platform, float64,
) *palettes.TonalPalette {
	return nil
}

// GetPrimaryPalette overrides the 2021 implementation for the 2025 spec
func (d *schemePalettesDelegateImpl2025) GetPrimaryPalette(
	variant Variant,
	sourceColorHct color.Hct,
	dark bool,
	platform Platform,
	contrastLevel float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantNeutral:
		var chroma float64
		if platform == PlatformPhone {
			if sourceColorHct.IsBlue() {
				chroma = 12.0
			} else {
				chroma = 8.0
			}
		} else {
			if sourceColorHct.IsBlue() {
				chroma = 16.0
			} else {
				chroma = 12.0
			}
		}
		return palettes.FromHueAndChroma(sourceColorHct.Hue, chroma)
	case VariantTonalSpot:
		chroma := 32.0
		if platform == PlatformPhone && dark {
			chroma = 26.0
		}
		return palettes.FromHueAndChroma(sourceColorHct.Hue, chroma)
	case VariantExpressive:
		chroma := 40.0
		if platform == PlatformPhone {
			if dark {
				chroma = 36.0
			} else {
				chroma = 48.0
			}
		}
		return palettes.FromHueAndChroma(sourceColorHct.Hue, chroma)
	case VariantVibrant:
		chroma := 56.0
		if platform == PlatformPhone {
			chroma = 74.0
		}
		return palettes.FromHueAndChroma(sourceColorHct.Hue, chroma)
	default:
		return d.schemePalettesDelegateImpl2021.GetPrimaryPalette(
			variant,
			sourceColorHct,
			dark,
			platform,
			contrastLevel,
		)
	}
}

// GetSecondaryPalette overrides the 2021 implementation for the 2025 spec
func (d *schemePalettesDelegateImpl2025) GetSecondaryPalette(
	variant Variant,
	sourceColorHct color.Hct,
	dark bool,
	platform Platform,
	contrastLevel float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantNeutral:
		var chroma float64
		if platform == PlatformPhone {
			if sourceColorHct.IsBlue() {
				chroma = 6.0
			} else {
				chroma = 4.0
			}
		} else {
			if sourceColorHct.IsBlue() {
				chroma = 10.0
			} else {
				chroma = 6.0
			}
		}
		return palettes.FromHueAndChroma(sourceColorHct.Hue, chroma)
	case VariantTonalSpot:
		return palettes.FromHueAndChroma(sourceColorHct.Hue, 16.0)
	case VariantExpressive:
		chroma := 24.0
		if platform == PlatformPhone && dark {
			chroma = 16.0
		}
		hueKeys := []float64{0, 105, 140, 204, 253, 278, 300, 333, 360}
		rotations := []float64{-160, 155, -100, 96, -96, -156, -165, -160}
		return palettes.FromHueAndChroma(
			GetRotatedHue(sourceColorHct, hueKeys, rotations),
			chroma)
	case VariantVibrant:
		chroma := 36.0
		if platform == PlatformPhone {
			chroma = 56.0
		}
		hueKeys := []float64{0, 38, 105, 140, 333, 360}
		rotations := []float64{-14, 10, -14, 10, -14}
		return palettes.FromHueAndChroma(
			GetRotatedHue(sourceColorHct, hueKeys, rotations),
			chroma)
	default:
		return d.schemePalettesDelegateImpl2021.GetSecondaryPalette(
			variant,
			sourceColorHct,
			dark,
			platform,
			contrastLevel,
		)
	}
}

// GetTertiaryPalette overrides the 2021 implementation for the 2025 spec
func (d *schemePalettesDelegateImpl2025) GetTertiaryPalette(
	variant Variant,
	sourceColorHct color.Hct,
	dark bool,
	platform Platform,
	contrastLevel float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantNeutral:
		chroma := 36.0
		if platform == PlatformPhone {
			chroma = 20.0
		}
		hueKeys := []float64{0, 38, 105, 161, 204, 278, 333, 360}
		rotations := []float64{-32, 26, 10, -39, 24, -15, -32}
		return palettes.FromHueAndChroma(
			GetRotatedHue(sourceColorHct, hueKeys, rotations),
			chroma)
	case VariantTonalSpot:
		chroma := 32.0
		if platform == PlatformPhone {
			chroma = 28.0
		}
		hueKeys := []float64{0, 20, 71, 161, 333, 360}
		rotations := []float64{-40, 48, -32, 40, -32}
		return palettes.FromHueAndChroma(
			GetRotatedHue(sourceColorHct, hueKeys, rotations),
			chroma)
	case VariantExpressive:
		hueKeys := []float64{0, 105, 140, 204, 253, 278, 300, 333, 360}
		rotations := []float64{-165, 160, -105, 101, -101, -160, -170, -165}
		return palettes.FromHueAndChroma(
			GetRotatedHue(sourceColorHct, hueKeys, rotations),
			48.0)
	case VariantVibrant:
		hueKeys := []float64{0, 38, 71, 105, 140, 161, 253, 333, 360}
		rotations := []float64{-72, 35, 24, -24, 62, 50, 62, -72}
		return palettes.FromHueAndChroma(
			GetRotatedHue(sourceColorHct, hueKeys, rotations),
			56.0)
	default:
		return d.schemePalettesDelegateImpl2021.GetTertiaryPalette(
			variant,
			sourceColorHct,
			dark,
			platform,
			contrastLevel,
		)
	}
}

// getExpressiveNeutralHue is a helper for getting the neutral hue in expressive
// variant
func getExpressiveNeutralHue(sourceColorHct color.Hct) float64 {
	hueKeys := []float64{0, 71, 124, 253, 278, 300, 360}
	rotations := []float64{10, 0, 10, 0, 10, 0}
	return GetRotatedHue(sourceColorHct, hueKeys, rotations)
}

// getExpressiveNeutralChroma is a helper for getting the neutral chroma in
// expressive variant
func getExpressiveNeutralChroma(
	sourceColorHct color.Hct,
	dark bool,
	platform Platform,
) float64 {
	neutralHue := getExpressiveNeutralHue(sourceColorHct)
	if platform == PlatformPhone {
		if dark {
			if color.IsYellow(neutralHue) {
				return 6.0
			}
			return 14.0
		}
		return 18.0
	}
	return 12.0
}

// getVibrantNeutralHue is a helper for getting the neutral hue in vibrant
// variant
func getVibrantNeutralHue(sourceColorHct color.Hct) float64 {
	hueKeys := []float64{0, 38, 105, 140, 333, 360}
	rotations := []float64{-14, 10, -14, 10, -14}
	return GetRotatedHue(sourceColorHct, hueKeys, rotations)
}

// getVibrantNeutralChroma is a helper for getting the neutral chroma in vibrant
// variant
func getVibrantNeutralChroma(
	sourceColorHct color.Hct,
	platform Platform,
) float64 {
	neutralHue := getVibrantNeutralHue(sourceColorHct)
	if platform == PlatformPhone {
		return 28.0
	}
	if color.IsBlue(neutralHue) {
		return 28.0
	}
	return 20.0
}

// GetNeutralPalette overrides the 2021 implementation for the 2025 spec
func (d *schemePalettesDelegateImpl2025) GetNeutralPalette(
	variant Variant,
	sourceColorHct color.Hct,
	dark bool,
	platform Platform,
	contrastLevel float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantNeutral:
		chroma := 6.0
		if platform == PlatformPhone {
			chroma = 1.4
		}
		return palettes.FromHueAndChroma(sourceColorHct.Hue, chroma)
	case VariantTonalSpot:
		chroma := 10.0
		if platform == PlatformPhone {
			chroma = 5.0
		}
		return palettes.FromHueAndChroma(sourceColorHct.Hue, chroma)
	case VariantExpressive:
		return palettes.FromHueAndChroma(
			getExpressiveNeutralHue(sourceColorHct),
			getExpressiveNeutralChroma(sourceColorHct, dark, platform))
	case VariantVibrant:
		return palettes.FromHueAndChroma(
			getVibrantNeutralHue(sourceColorHct),
			getVibrantNeutralChroma(sourceColorHct, platform))
	default:
		return d.schemePalettesDelegateImpl2021.GetNeutralPalette(
			variant,
			sourceColorHct,
			dark,
			platform,
			contrastLevel,
		)
	}
}

// GetNeutralVariantPalette overrides the 2021 implementation for the 2025 spec
func (d *schemePalettesDelegateImpl2025) GetNeutralVariantPalette(
	variant Variant,
	sourceColorHct color.Hct,
	dark bool,
	platform Platform,
	contrastLevel float64,
) *palettes.TonalPalette {
	switch variant {
	case VariantNeutral:
		baseChroma := 6.0
		if platform == PlatformPhone {
			baseChroma = 1.4
		}
		return palettes.FromHueAndChroma(sourceColorHct.Hue, baseChroma*2.2)
	case VariantTonalSpot:
		baseChroma := 10.0
		if platform == PlatformPhone {
			baseChroma = 5.0
		}
		return palettes.FromHueAndChroma(sourceColorHct.Hue, baseChroma*1.7)
	case VariantExpressive:
		expressiveNeutralHue := getExpressiveNeutralHue(sourceColorHct)
		expressiveNeutralChroma := getExpressiveNeutralChroma(
			sourceColorHct,
			dark,
			platform,
		)
		multiplier := 2.3
		if expressiveNeutralHue >= 105 && expressiveNeutralHue < 125 {
			multiplier = 1.6
		}
		return palettes.FromHueAndChroma(
			expressiveNeutralHue,
			expressiveNeutralChroma*multiplier)
	case VariantVibrant:
		vibrantNeutralHue := getVibrantNeutralHue(sourceColorHct)
		vibrantNeutralChroma := getVibrantNeutralChroma(
			sourceColorHct,
			platform,
		)
		return palettes.FromHueAndChroma(
			vibrantNeutralHue,
			vibrantNeutralChroma*1.29)
	default:
		return d.schemePalettesDelegateImpl2021.GetNeutralVariantPalette(
			variant,
			sourceColorHct,
			dark,
			platform,
			contrastLevel,
		)
	}
}

// GetErrorPalette overrides the 2021 implementation for the 2025 spec
func (d *schemePalettesDelegateImpl2025) GetErrorPalette(
	variant Variant,
	sourceColorHct color.Hct,
	dark bool,
	platform Platform,
	contrastLevel float64,
) *palettes.TonalPalette {
	errorHue := GetPiecewiseHue(
		sourceColorHct,
		[]float64{0, 3, 13, 23, 33, 43, 153, 273, 360},
		[]float64{12, 22, 32, 12, 22, 32, 22, 12})

	var palette *palettes.TonalPalette

	switch variant {
	case VariantNeutral:
		chroma := 40.0
		if platform == PlatformPhone {
			chroma = 50.0
		}
		palette = palettes.FromHueAndChroma(errorHue, chroma)
	case VariantTonalSpot:
		chroma := 48.0
		if platform == PlatformPhone {
			chroma = 60.0
		}
		palette = palettes.FromHueAndChroma(errorHue, chroma)
	case VariantExpressive:
		chroma := 48.0
		if platform == PlatformPhone {
			chroma = 64.0
		}
		palette = palettes.FromHueAndChroma(errorHue, chroma)
	case VariantVibrant:
		chroma := 60.0
		if platform == PlatformPhone {
			chroma = 80.0
		}
		palette = palettes.FromHueAndChroma(errorHue, chroma)
	default:
		return d.schemePalettesDelegateImpl2021.GetErrorPalette(
			variant,
			sourceColorHct,
			dark,
			platform,
			contrastLevel,
		)
	}

	return palette
}
