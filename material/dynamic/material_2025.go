// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package dynamic

import (
	"fmt"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/AvengeMedia/dankgo/material/num"
	"github.com/AvengeMedia/dankgo/material/palettes"
)

// tMaxC
// Paramters:
//   - lowerBound: 0
//   - upperBound: 100
//   - chromaMultiplier: 1
func tMaxC(palette palettes.TonalPalette, params ...float64) float64 {
	lowerBound := 0.0
	upperBound := 100.0
	chromaMultiplier := 1.0
	if len(params) > 2 {
		lowerBound = params[0]
		upperBound = params[1]
		chromaMultiplier = params[2]
	} else if len(params) > 1 {
		lowerBound = params[0]
		upperBound = params[1]
	} else if len(params) > 0 {
		lowerBound = params[0]
	}
	answer := findBestToneForChroma(palette.Hue, palette.Chroma*chromaMultiplier, 100, true)
	return num.Clamp(lowerBound, upperBound, answer)
}

// tMinC
// Paramters:
//   - lowerBound: 0
//   - upperBound: 100
func tMinC(palette palettes.TonalPalette, bounds ...float64) float64 {
	lowerBound := 0.0
	upperBound := 100.0
	if len(bounds) > 1 {
		lowerBound = bounds[0]
		upperBound = bounds[1]
	} else if len(bounds) > 0 {
		lowerBound = bounds[0]
	}

	answer := findBestToneForChroma(palette.Hue, palette.Chroma, 0, false)
	return num.Clamp(lowerBound, upperBound, answer)
}

// Unlike FindDesiredChromaByTone, keeps searching past the first chroma dip.
func findBestToneForChroma(hue, chroma, tone float64, byDecreasingTone bool) float64 {
	answer := tone
	best := color.NewHct(hue, chroma, answer)
	for best.Chroma < chroma {
		if tone < 0 || tone > 100 {
			break
		}
		if byDecreasingTone {
			tone--
		} else {
			tone++
		}
		candidate := color.NewHct(hue, chroma, tone)
		if best.Chroma < candidate.Chroma {
			best = candidate
			answer = tone
		}
	}
	return answer
}

// GetCurve returns the contrast curve for a given default contrast.
func GetCurve(defaultContrast float64) *ContrastCurve {
	switch defaultContrast {
	case 1.5:
		return NewContrastCurve(1.5, 1.5, 3, 5.5)
	case 3:
		return NewContrastCurve(3, 3, 4.5, 7)
	case 4.5:
		return NewContrastCurve(4.5, 4.5, 7, 11)
	case 6:
		return NewContrastCurve(6, 6, 7, 11)
	case 7:
		return NewContrastCurve(7, 7, 11, 21)
	case 9:
		return NewContrastCurve(9, 9, 11, 21)
	case 11:
		return NewContrastCurve(11, 11, 21, 21)
	case 21:
		return NewContrastCurve(21, 21, 21, 21)
	default:
		return NewContrastCurve(defaultContrast, defaultContrast, 7, 21)
	}
}

func validateExtendedColor(
	originalColor *Color,
	specVersion Version,
	extendedColor *Color,
) {
	if originalColor.Name != extendedColor.Name {
		panic(fmt.Sprintf(
			"Attempting to extend color %s with color %s of different name for spec version %v.",
			originalColor.Name,
			extendedColor.Name,
			specVersion,
		))
	}
	if originalColor.IsBackground != extendedColor.IsBackground {
		panic(fmt.Sprintf(
			"Attempting to extend color %s as a %s with color %s as a %s for spec version %v.",
			originalColor.Name,
			boolToBackgroundForeground(originalColor.IsBackground),
			extendedColor.Name,
			boolToBackgroundForeground(extendedColor.IsBackground),
			specVersion,
		))
	}
}

func boolToBackgroundForeground(isBackground bool) string {
	if isBackground {
		return "background"
	}
	return "foreground"
}

func extendSpecVersion(
	originalColor *Color,
	specVersion Version,
	extendedColor *Color,
) *Color {
	validateExtendedColor(originalColor, specVersion, extendedColor)

	return &Color{
		Name: originalColor.Name,
		Palette: func(s *Scheme) palettes.TonalPalette {
			if s.Version == specVersion {
				return extendedColor.Palette(s)
			}
			return originalColor.Palette(s)
		},
		Tone: func(s *Scheme) float64 {
			if s.Version == specVersion {
				return extendedColor.Tone(s)
			}
			return originalColor.Tone(s)
		},
		IsBackground: originalColor.IsBackground,
		ChromaMultiplier: func(s *Scheme) float64 {
			var chromaMultiplier ChromaMultiplier
			if s.Version == specVersion {
				chromaMultiplier = extendedColor.ChromaMultiplier
			} else {
				chromaMultiplier = originalColor.ChromaMultiplier
			}
			if chromaMultiplier != nil {
				return chromaMultiplier(s)
			}
			return 1
		},
		Background: func(s *Scheme) *Color {
			var background ColorFunc
			if s.Version == specVersion {
				background = extendedColor.Background
			} else {
				background = originalColor.Background
			}
			if background != nil {
				return background(s)
			}
			return nil
		},
		SecondBackground: func(s *Scheme) *Color {
			var secondBackground ColorFunc
			if s.Version == specVersion {
				secondBackground = extendedColor.SecondBackground
			} else {
				secondBackground = originalColor.SecondBackground
			}
			if secondBackground != nil {
				return secondBackground(s)
			}
			return nil
		},
		ContrastCurve: func(s *Scheme) *ContrastCurve {
			var contrastCurve ContrastCurveFunc
			if s.Version == specVersion {
				contrastCurve = extendedColor.ContrastCurve
			} else {
				contrastCurve = originalColor.ContrastCurve
			}
			if contrastCurve != nil {
				return contrastCurve(s)
			}
			return nil
		},
		ToneDeltaPair: func(s *Scheme) *ToneDeltaPair {
			var toneDeltaPair ToneDeltaPairFunc
			if s.Version == specVersion {
				toneDeltaPair = extendedColor.ToneDeltaPair
			} else {
				toneDeltaPair = originalColor.ToneDeltaPair
			}
			if toneDeltaPair != nil {
				return toneDeltaPair(s)
			}
			return nil
		},
	}
}

//revive:disable:exported

// MaterialSpec2025 follows ColorSpecDelegateImpl2025 in material-color-utilities
// 5b3618b. Roles it does not override come from the embedded MaterialSpec2021.
type MaterialSpec2025 struct {
	MaterialSpec2021
}

var _ MaterialColorSpec = (*MaterialSpec2025)(nil)

func isPhone(s *Scheme) bool {
	return s.Platform == PlatformPhone
}

func phoneCurve(s *Scheme, phone, watch float64) *ContrastCurve {
	if isPhone(s) {
		return GetCurve(phone)
	}
	return GetCurve(watch)
}

func containerCurve(s *Scheme) *ContrastCurve {
	if isPhone(s) && s.Contrast > 0 {
		return GetCurve(1.5)
	}
	return nil
}

func lightStandardContrast(s *Scheme) *Scheme {
	light := *s
	light.Dark = false
	light.Contrast = 0
	return &light
}

func neutralTone(s *Scheme, yellow, vibrant, other float64) float64 {
	switch {
	case s.NeutralPalette.IsYellow():
		return yellow
	case s.Variant == VariantVibrant:
		return vibrant
	default:
		return other
	}
}

func surfaceChroma(s *Scheme, neutral, tonalSpot, expressive, expressiveYellow, vibrant float64) float64 {
	switch s.Variant {
	case VariantNeutral:
		return neutral
	case VariantTonalSpot:
		return tonalSpot
	case VariantExpressive:
		if s.NeutralPalette.IsYellow() {
			return expressiveYellow
		}
		return expressive
	case VariantVibrant:
		return vibrant
	default:
		return 1
	}
}

func onSurfaceChroma(s *Scheme) float64 {
	if !isPhone(s) {
		return 1
	}
	expressiveYellow := 2.3
	if s.Dark {
		expressiveYellow = 3.0
	}
	return surfaceChroma(s, 2.2, 1.7, 1.6, expressiveYellow, 1)
}

// The embedded 2021 method would resolve the 2021 surfaces, embedding does not dispatch.
func (m MaterialSpec2025) HighestSurface(s *Scheme) *Color {
	if s.Dark {
		return m.SurfaceBright()
	}
	return m.SurfaceDim()
}

func (m MaterialSpec2025) accentBackground(s *Scheme) *Color {
	if isPhone(s) {
		return m.HighestSurface(s)
	}
	return m.SurfaceContainerHigh()
}

func (m MaterialSpec2025) containerBackground(s *Scheme) *Color {
	if isPhone(s) {
		return m.HighestSurface(s)
	}
	return nil
}

func neutralOf(s *Scheme) palettes.TonalPalette   { return s.NeutralPalette }
func primaryOf(s *Scheme) palettes.TonalPalette   { return s.PrimaryPalette }
func secondaryOf(s *Scheme) palettes.TonalPalette { return s.SecondaryPalette }
func tertiaryOf(s *Scheme) palettes.TonalPalette  { return s.TertiaryPalette }
func errorOf(s *Scheme) palettes.TonalPalette     { return s.ErrorPalette }

////////////////////////////////////////////////////////////////
// Surfaces                                                   //
////////////////////////////////////////////////////////////////

func (m MaterialSpec2025) Surface() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "surface",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			switch {
			case !isPhone(s):
				return 0
			case s.Dark:
				return 4
			default:
				return neutralTone(s, 99, 97, 98)
			}
		},
		IsBackground: true,
	})
	return extendSpecVersion(m.MaterialSpec2021.Surface(), Version2025, color)
}

func (m MaterialSpec2025) SurfaceDim() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "surface_dim",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 4
			}
			return neutralTone(s, 90, 85, 87)
		},
		IsBackground: true,
		ChromaMultiplier: func(s *Scheme) float64 {
			if s.Dark {
				return 1
			}
			return surfaceChroma(s, 2.5, 1.7, 1.75, 2.7, 1.36)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.SurfaceDim(), Version2025, color)
}

func (m MaterialSpec2025) SurfaceBright() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "surface_bright",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 18
			}
			return neutralTone(s, 99, 97, 98)
		},
		IsBackground: true,
		ChromaMultiplier: func(s *Scheme) float64 {
			if !s.Dark {
				return 1
			}
			return surfaceChroma(s, 2.5, 1.7, 1.75, 2.7, 1.36)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.SurfaceBright(), Version2025, color)
}

func (m MaterialSpec2025) SurfaceContainerLowest() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "surface_container_lowest",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 0
			}
			return 100
		},
		IsBackground: true,
	})
	return extendSpecVersion(m.MaterialSpec2021.SurfaceContainerLowest(), Version2025, color)
}

func (m MaterialSpec2025) SurfaceContainerLow() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "surface_container_low",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			switch {
			case !isPhone(s):
				return 15
			case s.Dark:
				return 6
			default:
				return neutralTone(s, 98, 95, 96)
			}
		},
		IsBackground: true,
		ChromaMultiplier: func(s *Scheme) float64 {
			if !isPhone(s) {
				return 1
			}
			return surfaceChroma(s, 1.3, 1.25, 1.15, 1.3, 1.08)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.SurfaceContainerLow(), Version2025, color)
}

func (m MaterialSpec2025) SurfaceContainer() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "surface_container",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			switch {
			case !isPhone(s):
				return 20
			case s.Dark:
				return 9
			default:
				return neutralTone(s, 96, 92, 94)
			}
		},
		IsBackground: true,
		ChromaMultiplier: func(s *Scheme) float64 {
			if !isPhone(s) {
				return 1
			}
			return surfaceChroma(s, 1.6, 1.4, 1.3, 1.6, 1.15)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.SurfaceContainer(), Version2025, color)
}

func (m MaterialSpec2025) SurfaceContainerHigh() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "surface_container_high",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			switch {
			case !isPhone(s):
				return 25
			case s.Dark:
				return 12
			default:
				return neutralTone(s, 94, 90, 92)
			}
		},
		IsBackground: true,
		ChromaMultiplier: func(s *Scheme) float64 {
			if !isPhone(s) {
				return 1
			}
			return surfaceChroma(s, 1.9, 1.5, 1.45, 1.95, 1.22)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.SurfaceContainerHigh(), Version2025, color)
}

func (m MaterialSpec2025) SurfaceContainerHighest() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "surface_container_highest",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 15
			}
			return neutralTone(s, 92, 88, 90)
		},
		IsBackground: true,
		ChromaMultiplier: func(s *Scheme) float64 {
			return surfaceChroma(s, 2.2, 1.7, 1.6, 2.3, 1.29)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.SurfaceContainerHighest(), Version2025, color)
}

func (m MaterialSpec2025) OnSurface() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "on_surface",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			if s.Variant == VariantVibrant {
				return tMaxC(s.NeutralPalette, 0, 100, 1.1)
			}
			return GetInitialToneFromBackground(m.accentBackground)(s)
		},
		ChromaMultiplier: onSurfaceChroma,
		Background:       m.accentBackground,
		ContrastCurve: func(s *Scheme) *ContrastCurve {
			if s.Dark && isPhone(s) {
				return GetCurve(11)
			}
			return GetCurve(9)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.OnSurface(), Version2025, color)
}

func (m MaterialSpec2025) OnSurfaceVariant() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:             "on_surface_variant",
		Palette:          neutralOf,
		ChromaMultiplier: onSurfaceChroma,
		Background:       m.accentBackground,
		ContrastCurve: func(s *Scheme) *ContrastCurve {
			switch {
			case !isPhone(s):
				return GetCurve(7)
			case s.Dark:
				return GetCurve(6)
			default:
				return GetCurve(4.5)
			}
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.OnSurfaceVariant(), Version2025, color)
}

func (m MaterialSpec2025) Outline() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:             "outline",
		Palette:          neutralOf,
		ChromaMultiplier: onSurfaceChroma,
		Background:       m.accentBackground,
		ContrastCurve:    func(s *Scheme) *ContrastCurve { return phoneCurve(s, 3, 4.5) },
	})
	return extendSpecVersion(m.MaterialSpec2021.Outline(), Version2025, color)
}

func (m MaterialSpec2025) OutlineVariant() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:             "outline_variant",
		Palette:          neutralOf,
		ChromaMultiplier: onSurfaceChroma,
		Background:       m.accentBackground,
		ContrastCurve:    func(s *Scheme) *ContrastCurve { return phoneCurve(s, 1.5, 3) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OutlineVariant(), Version2025, color)
}

func (m MaterialSpec2025) InverseSurface() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "inverse_surface",
		Palette: neutralOf,
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 98
			}
			return 4
		},
		IsBackground: true,
	})
	return extendSpecVersion(m.MaterialSpec2021.InverseSurface(), Version2025, color)
}

func (m MaterialSpec2025) InverseOnSurface() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "inverse_on_surface",
		Palette:       neutralOf,
		Background:    func(*Scheme) *Color { return m.InverseSurface() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.InverseOnSurface(), Version2025, color)
}

////////////////////////////////////////////////////////////////
// Primaries                                                  //
////////////////////////////////////////////////////////////////

func (m MaterialSpec2025) Primary() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "primary",
		Palette: primaryOf,
		Tone: func(s *Scheme) float64 {
			p := s.PrimaryPalette
			switch s.Variant {
			case VariantNeutral:
				switch {
				case !isPhone(s):
					return 90
				case s.Dark:
					return 80
				default:
					return 40
				}
			case VariantTonalSpot:
				switch {
				case !isPhone(s):
					return tMaxC(p, 0, 90)
				case s.Dark:
					return 80
				default:
					return tMaxC(p)
				}
			case VariantExpressive:
				if !isPhone(s) {
					return tMaxC(p)
				}
				upper := 98.0
				switch {
				case s.Dark && p.IsCyan():
					upper = 88
				case !s.Dark && p.IsYellow():
					upper = 25
				}
				return tMaxC(p, 0, upper)
			default:
				if !isPhone(s) {
					return tMaxC(p)
				}
				if p.IsCyan() {
					return tMaxC(p, 0, 88)
				}
				return tMaxC(p, 0, 98)
			}
		},
		IsBackground:  true,
		Background:    m.accentBackground,
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 4.5, 7) },
		ToneDeltaPair: func(s *Scheme) *ToneDeltaPair {
			if !isPhone(s) {
				return nil
			}
			return NewToneDeltaPair(m.PrimaryContainer(), m.Primary(), 5, TonePolarityRelativeLighter, true, ConstraintFarther)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.Primary(), Version2025, color)
}

func (m MaterialSpec2025) PrimaryDim() *Color {
	return DynamicColorFromPalette(&Color{
		Name:    "primary_dim",
		Palette: primaryOf,
		Tone: func(s *Scheme) float64 {
			switch s.Variant {
			case VariantNeutral:
				return 85
			case VariantTonalSpot:
				return tMaxC(s.PrimaryPalette, 0, 90)
			default:
				return tMaxC(s.PrimaryPalette)
			}
		},
		IsBackground:  true,
		Background:    func(*Scheme) *Color { return m.SurfaceContainerHigh() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(4.5) },
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(m.PrimaryDim(), m.Primary(), 5, TonePolarityDarker, true, ConstraintFarther)
		},
	})
}

func (m MaterialSpec2025) OnPrimary() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "on_primary",
		Palette: primaryOf,
		Background: func(s *Scheme) *Color {
			if isPhone(s) {
				return m.Primary()
			}
			return m.PrimaryDim()
		},
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 6, 7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnPrimary(), Version2025, color)
}

func (m MaterialSpec2025) PrimaryContainer() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "primary_container",
		Palette: primaryOf,
		Tone: func(s *Scheme) float64 {
			p := s.PrimaryPalette
			if !isPhone(s) {
				return 30
			}
			switch s.Variant {
			case VariantNeutral:
				if s.Dark {
					return 30
				}
				return 90
			case VariantTonalSpot:
				if s.Dark {
					return tMinC(p, 35, 93)
				}
				return tMaxC(p, 0, 90)
			case VariantExpressive:
				if s.Dark {
					return tMinC(p, 30, 93)
				}
				if p.IsCyan() {
					return tMaxC(p, 78, 88)
				}
				return tMaxC(p, 78, 90)
			default:
				if s.Dark {
					return tMinC(p, 66, 93)
				}
				if p.IsCyan() {
					return tMaxC(p, 66, 88)
				}
				return tMaxC(p, 66, 93)
			}
		},
		IsBackground: true,
		Background:   m.containerBackground,
		ToneDeltaPair: func(s *Scheme) *ToneDeltaPair {
			if isPhone(s) {
				return nil
			}
			return NewToneDeltaPair(m.PrimaryContainer(), m.PrimaryDim(), 10, TonePolarityDarker, true, ConstraintFarther)
		},
		ContrastCurve: containerCurve,
	})
	return extendSpecVersion(m.MaterialSpec2021.PrimaryContainer(), Version2025, color)
}

func (m MaterialSpec2025) OnPrimaryContainer() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_primary_container",
		Palette:       primaryOf,
		Background:    func(*Scheme) *Color { return m.PrimaryContainer() },
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 6, 7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnPrimaryContainer(), Version2025, color)
}

func (m MaterialSpec2025) PrimaryFixed() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "primary_fixed",
		Palette: primaryOf,
		Tone: func(s *Scheme) float64 {
			return m.PrimaryContainer().GetTone(lightStandardContrast(s))
		},
		IsBackground:  true,
		Background:    m.containerBackground,
		ContrastCurve: containerCurve,
	})
	return extendSpecVersion(m.MaterialSpec2021.PrimaryFixed(), Version2025, color)
}

func (m MaterialSpec2025) PrimaryFixedDim() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:         "primary_fixed_dim",
		Palette:      primaryOf,
		Tone:         func(s *Scheme) float64 { return m.PrimaryFixed().GetTone(s) },
		IsBackground: true,
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(m.PrimaryFixedDim(), m.PrimaryFixed(), 5, TonePolarityDarker, true, ConstraintExact)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.PrimaryFixedDim(), Version2025, color)
}

func (m MaterialSpec2025) OnPrimaryFixed() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_primary_fixed",
		Palette:       primaryOf,
		Background:    func(*Scheme) *Color { return m.PrimaryFixedDim() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnPrimaryFixed(), Version2025, color)
}

func (m MaterialSpec2025) OnPrimaryFixedVariant() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_primary_fixed_variant",
		Palette:       primaryOf,
		Background:    func(*Scheme) *Color { return m.PrimaryFixedDim() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(4.5) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnPrimaryFixedVariant(), Version2025, color)
}

func (m MaterialSpec2025) InversePrimary() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "inverse_primary",
		Palette:       primaryOf,
		Tone:          func(s *Scheme) float64 { return tMaxC(s.PrimaryPalette) },
		Background:    func(*Scheme) *Color { return m.InverseSurface() },
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 6, 7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.InversePrimary(), Version2025, color)
}

////////////////////////////////////////////////////////////////
// Secondaries                                                //
////////////////////////////////////////////////////////////////

func (m MaterialSpec2025) Secondary() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "secondary",
		Palette: secondaryOf,
		Tone: func(s *Scheme) float64 {
			p := s.SecondaryPalette
			switch {
			case !isPhone(s):
				if s.Variant == VariantNeutral {
					return 90
				}
				return tMaxC(p, 0, 90)
			case s.Variant == VariantNeutral:
				if s.Dark {
					return tMinC(p, 0, 98)
				}
				return tMaxC(p)
			case s.Variant == VariantVibrant:
				if s.Dark {
					return tMaxC(p, 0, 90)
				}
				return tMaxC(p, 0, 98)
			default:
				if s.Dark {
					return 80
				}
				return tMaxC(p)
			}
		},
		IsBackground:  true,
		Background:    m.accentBackground,
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 4.5, 7) },
		ToneDeltaPair: func(s *Scheme) *ToneDeltaPair {
			if !isPhone(s) {
				return nil
			}
			return NewToneDeltaPair(m.SecondaryContainer(), m.Secondary(), 5, TonePolarityRelativeLighter, true, ConstraintFarther)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.Secondary(), Version2025, color)
}

func (m MaterialSpec2025) SecondaryDim() *Color {
	return DynamicColorFromPalette(&Color{
		Name:    "secondary_dim",
		Palette: secondaryOf,
		Tone: func(s *Scheme) float64 {
			if s.Variant == VariantNeutral {
				return 85
			}
			return tMaxC(s.SecondaryPalette, 0, 90)
		},
		IsBackground:  true,
		Background:    func(*Scheme) *Color { return m.SurfaceContainerHigh() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(4.5) },
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(m.SecondaryDim(), m.Secondary(), 5, TonePolarityDarker, true, ConstraintFarther)
		},
	})
}

func (m MaterialSpec2025) OnSecondary() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "on_secondary",
		Palette: secondaryOf,
		Background: func(s *Scheme) *Color {
			if isPhone(s) {
				return m.Secondary()
			}
			return m.SecondaryDim()
		},
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 6, 7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnSecondary(), Version2025, color)
}

func (m MaterialSpec2025) SecondaryContainer() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "secondary_container",
		Palette: secondaryOf,
		Tone: func(s *Scheme) float64 {
			p := s.SecondaryPalette
			if !isPhone(s) {
				return 30
			}
			switch s.Variant {
			case VariantVibrant:
				if s.Dark {
					return tMinC(p, 30, 40)
				}
				return tMaxC(p, 84, 90)
			case VariantExpressive:
				if s.Dark {
					return 15
				}
				return tMaxC(p, 90, 95)
			default:
				if s.Dark {
					return 25
				}
				return 90
			}
		},
		IsBackground: true,
		Background:   m.containerBackground,
		ToneDeltaPair: func(s *Scheme) *ToneDeltaPair {
			if isPhone(s) {
				return nil
			}
			return NewToneDeltaPair(m.SecondaryContainer(), m.SecondaryDim(), 10, TonePolarityDarker, true, ConstraintFarther)
		},
		ContrastCurve: containerCurve,
	})
	return extendSpecVersion(m.MaterialSpec2021.SecondaryContainer(), Version2025, color)
}

func (m MaterialSpec2025) OnSecondaryContainer() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_secondary_container",
		Palette:       secondaryOf,
		Background:    func(*Scheme) *Color { return m.SecondaryContainer() },
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 6, 7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnSecondaryContainer(), Version2025, color)
}

func (m MaterialSpec2025) SecondaryFixed() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "secondary_fixed",
		Palette: secondaryOf,
		Tone: func(s *Scheme) float64 {
			return m.SecondaryContainer().GetTone(lightStandardContrast(s))
		},
		IsBackground:  true,
		Background:    m.containerBackground,
		ContrastCurve: containerCurve,
	})
	return extendSpecVersion(m.MaterialSpec2021.SecondaryFixed(), Version2025, color)
}

func (m MaterialSpec2025) SecondaryFixedDim() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:         "secondary_fixed_dim",
		Palette:      secondaryOf,
		Tone:         func(s *Scheme) float64 { return m.SecondaryFixed().GetTone(s) },
		IsBackground: true,
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(m.SecondaryFixedDim(), m.SecondaryFixed(), 5, TonePolarityDarker, true, ConstraintExact)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.SecondaryFixedDim(), Version2025, color)
}

func (m MaterialSpec2025) OnSecondaryFixed() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_secondary_fixed",
		Palette:       secondaryOf,
		Background:    func(*Scheme) *Color { return m.SecondaryFixedDim() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnSecondaryFixed(), Version2025, color)
}

func (m MaterialSpec2025) OnSecondaryFixedVariant() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_secondary_fixed_variant",
		Palette:       secondaryOf,
		Background:    func(*Scheme) *Color { return m.SecondaryFixedDim() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(4.5) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnSecondaryFixedVariant(), Version2025, color)
}

////////////////////////////////////////////////////////////////
// Tertiaries                                                 //
////////////////////////////////////////////////////////////////

func (m MaterialSpec2025) Tertiary() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "tertiary",
		Palette: tertiaryOf,
		Tone: func(s *Scheme) float64 {
			p := s.TertiaryPalette
			switch {
			case !isPhone(s):
				if s.Variant == VariantTonalSpot {
					return tMaxC(p, 0, 90)
				}
				return tMaxC(p)
			case s.Variant == VariantExpressive || s.Variant == VariantVibrant:
				switch {
				case p.IsCyan():
					return tMaxC(p, 0, 88)
				case s.Dark:
					return tMaxC(p, 0, 98)
				default:
					return tMaxC(p, 0, 100)
				}
			default:
				if s.Dark {
					return tMaxC(p, 0, 98)
				}
				return tMaxC(p)
			}
		},
		IsBackground:  true,
		Background:    m.accentBackground,
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 4.5, 7) },
		ToneDeltaPair: func(s *Scheme) *ToneDeltaPair {
			if !isPhone(s) {
				return nil
			}
			return NewToneDeltaPair(m.TertiaryContainer(), m.Tertiary(), 5, TonePolarityRelativeLighter, true, ConstraintFarther)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.Tertiary(), Version2025, color)
}

func (m MaterialSpec2025) TertiaryDim() *Color {
	return DynamicColorFromPalette(&Color{
		Name:    "tertiary_dim",
		Palette: tertiaryOf,
		Tone: func(s *Scheme) float64 {
			if s.Variant == VariantTonalSpot {
				return tMaxC(s.TertiaryPalette, 0, 90)
			}
			return tMaxC(s.TertiaryPalette)
		},
		IsBackground:  true,
		Background:    func(*Scheme) *Color { return m.SurfaceContainerHigh() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(4.5) },
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(m.TertiaryDim(), m.Tertiary(), 5, TonePolarityDarker, true, ConstraintFarther)
		},
	})
}

func (m MaterialSpec2025) OnTertiary() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "on_tertiary",
		Palette: tertiaryOf,
		Background: func(s *Scheme) *Color {
			if isPhone(s) {
				return m.Tertiary()
			}
			return m.TertiaryDim()
		},
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 6, 7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnTertiary(), Version2025, color)
}

func (m MaterialSpec2025) TertiaryContainer() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "tertiary_container",
		Palette: tertiaryOf,
		Tone: func(s *Scheme) float64 {
			p := s.TertiaryPalette
			if !isPhone(s) {
				if s.Variant == VariantTonalSpot {
					return tMaxC(p, 0, 90)
				}
				return tMaxC(p)
			}
			switch s.Variant {
			case VariantNeutral:
				if s.Dark {
					return tMaxC(p, 0, 93)
				}
				return tMaxC(p, 0, 96)
			case VariantTonalSpot:
				if s.Dark {
					return tMaxC(p, 0, 93)
				}
				return tMaxC(p, 0, 100)
			case VariantExpressive:
				switch {
				case p.IsCyan():
					return tMaxC(p, 75, 88)
				case s.Dark:
					return tMaxC(p, 75, 93)
				default:
					return tMaxC(p, 75, 100)
				}
			default:
				if s.Dark {
					return tMaxC(p, 0, 93)
				}
				return tMaxC(p, 72, 100)
			}
		},
		IsBackground: true,
		Background:   m.containerBackground,
		ToneDeltaPair: func(s *Scheme) *ToneDeltaPair {
			if isPhone(s) {
				return nil
			}
			return NewToneDeltaPair(m.TertiaryContainer(), m.TertiaryDim(), 10, TonePolarityDarker, true, ConstraintFarther)
		},
		ContrastCurve: containerCurve,
	})
	return extendSpecVersion(m.MaterialSpec2021.TertiaryContainer(), Version2025, color)
}

func (m MaterialSpec2025) OnTertiaryContainer() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_tertiary_container",
		Palette:       tertiaryOf,
		Background:    func(*Scheme) *Color { return m.TertiaryContainer() },
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 6, 7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnTertiaryContainer(), Version2025, color)
}

func (m MaterialSpec2025) TertiaryFixed() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "tertiary_fixed",
		Palette: tertiaryOf,
		Tone: func(s *Scheme) float64 {
			return m.TertiaryContainer().GetTone(lightStandardContrast(s))
		},
		IsBackground:  true,
		Background:    m.containerBackground,
		ContrastCurve: containerCurve,
	})
	return extendSpecVersion(m.MaterialSpec2021.TertiaryFixed(), Version2025, color)
}

func (m MaterialSpec2025) TertiaryFixedDim() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:         "tertiary_fixed_dim",
		Palette:      tertiaryOf,
		Tone:         func(s *Scheme) float64 { return m.TertiaryFixed().GetTone(s) },
		IsBackground: true,
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(m.TertiaryFixedDim(), m.TertiaryFixed(), 5, TonePolarityDarker, true, ConstraintExact)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.TertiaryFixedDim(), Version2025, color)
}

func (m MaterialSpec2025) OnTertiaryFixed() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_tertiary_fixed",
		Palette:       tertiaryOf,
		Background:    func(*Scheme) *Color { return m.TertiaryFixedDim() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnTertiaryFixed(), Version2025, color)
}

func (m MaterialSpec2025) OnTertiaryFixedVariant() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_tertiary_fixed_variant",
		Palette:       tertiaryOf,
		Background:    func(*Scheme) *Color { return m.TertiaryFixedDim() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(4.5) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnTertiaryFixedVariant(), Version2025, color)
}

////////////////////////////////////////////////////////////////
// Errors                                                     //
////////////////////////////////////////////////////////////////

func (m MaterialSpec2025) Error() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "error",
		Palette: errorOf,
		Tone: func(s *Scheme) float64 {
			switch {
			case !isPhone(s):
				return tMinC(s.ErrorPalette)
			case s.Dark:
				return tMinC(s.ErrorPalette, 0, 98)
			default:
				return tMaxC(s.ErrorPalette)
			}
		},
		IsBackground:  true,
		Background:    m.accentBackground,
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 4.5, 7) },
		ToneDeltaPair: func(s *Scheme) *ToneDeltaPair {
			if !isPhone(s) {
				return nil
			}
			return NewToneDeltaPair(m.ErrorContainer(), m.Error(), 5, TonePolarityRelativeLighter, true, ConstraintFarther)
		},
	})
	return extendSpecVersion(m.MaterialSpec2021.Error(), Version2025, color)
}

func (m MaterialSpec2025) ErrorDim() *Color {
	return DynamicColorFromPalette(&Color{
		Name:          "error_dim",
		Palette:       errorOf,
		Tone:          func(s *Scheme) float64 { return tMinC(s.ErrorPalette) },
		IsBackground:  true,
		Background:    func(*Scheme) *Color { return m.SurfaceContainerHigh() },
		ContrastCurve: func(*Scheme) *ContrastCurve { return GetCurve(4.5) },
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(m.ErrorDim(), m.Error(), 5, TonePolarityDarker, true, ConstraintFarther)
		},
	})
}

func (m MaterialSpec2025) OnError() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "on_error",
		Palette: errorOf,
		Background: func(s *Scheme) *Color {
			if isPhone(s) {
				return m.Error()
			}
			return m.ErrorDim()
		},
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 6, 7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnError(), Version2025, color)
}

func (m MaterialSpec2025) ErrorContainer() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:    "error_container",
		Palette: errorOf,
		Tone: func(s *Scheme) float64 {
			switch {
			case !isPhone(s):
				return 30
			case s.Dark:
				return tMinC(s.ErrorPalette, 30, 93)
			default:
				return tMaxC(s.ErrorPalette, 0, 90)
			}
		},
		IsBackground: true,
		Background:   m.containerBackground,
		ToneDeltaPair: func(s *Scheme) *ToneDeltaPair {
			if isPhone(s) {
				return nil
			}
			return NewToneDeltaPair(m.ErrorContainer(), m.ErrorDim(), 10, TonePolarityDarker, true, ConstraintFarther)
		},
		ContrastCurve: containerCurve,
	})
	return extendSpecVersion(m.MaterialSpec2021.ErrorContainer(), Version2025, color)
}

func (m MaterialSpec2025) OnErrorContainer() *Color {
	color := DynamicColorFromPalette(&Color{
		Name:          "on_error_container",
		Palette:       errorOf,
		Background:    func(*Scheme) *Color { return m.ErrorContainer() },
		ContrastCurve: func(s *Scheme) *ContrastCurve { return phoneCurve(s, 4.5, 7) },
	})
	return extendSpecVersion(m.MaterialSpec2021.OnErrorContainer(), Version2025, color)
}

////////////////////////////////////////////////////////////////
// Remapped colors                                            //
////////////////////////////////////////////////////////////////

func (m MaterialSpec2025) SurfaceVariant() *Color {
	color := *m.SurfaceContainerHighest()
	color.Name = "surface_variant"
	return extendSpecVersion(m.MaterialSpec2021.SurfaceVariant(), Version2025, &color)
}

func (m MaterialSpec2025) SurfaceTint() *Color {
	color := *m.Primary()
	color.Name = "surface_tint"
	return extendSpecVersion(m.MaterialSpec2021.SurfaceTint(), Version2025, &color)
}

func (m MaterialSpec2025) Background() *Color {
	color := *m.Surface()
	color.Name = "background"
	return extendSpecVersion(m.MaterialSpec2021.Background(), Version2025, &color)
}

func (m MaterialSpec2025) OnBackground() *Color {
	color := *m.OnSurface()
	color.Name = "on_background"
	color.Tone = func(s *Scheme) float64 {
		if !isPhone(s) {
			return 100
		}
		return m.OnSurface().GetTone(s)
	}
	return extendSpecVersion(m.MaterialSpec2021.OnBackground(), Version2025, &color)
}
