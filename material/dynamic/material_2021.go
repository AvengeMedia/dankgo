// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package dynamic

import (
	"math"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/AvengeMedia/dankgo/material/dislike"
	"github.com/AvengeMedia/dankgo/material/palettes"
)

// IsFidelity returns whether the scheme is a fidelity scheme
func IsFidelity(scheme *Scheme) bool {
	return scheme.Variant == VariantFidelity || scheme.Variant == VariantContent
}

// IsMonochrome returns whether the scheme is monochrome
func IsMonochrome(scheme *Scheme) bool {
	return scheme.Variant == VariantMonochrome
}

// FindDesiredChromaByTone finds a tone where the chroma is as close as possible
// to the requested value
func FindDesiredChromaByTone(
	hue, chroma, tone float64,
	byDecreasingTone bool,
) float64 {
	answer := tone

	closestToChroma := color.NewHct(hue, chroma, tone)
	if closestToChroma.Chroma < chroma {
		chromaPeak := closestToChroma.Chroma
		for closestToChroma.Chroma < chroma {
			if byDecreasingTone {
				answer -= 1.0
			} else {
				answer += 1.0
			}
			potentialSolution := color.NewHct(hue, chroma, answer)
			if chromaPeak > potentialSolution.Chroma {
				break
			}
			if math.Abs(potentialSolution.Chroma-chroma) < 0.4 {
				break
			}

			potentialDelta := math.Abs(potentialSolution.Chroma - chroma)
			currentDelta := math.Abs(closestToChroma.Chroma - chroma)
			if potentialDelta < currentDelta {
				closestToChroma = potentialSolution
			}
			chromaPeak = math.Max(chromaPeak, potentialSolution.Chroma)
		}
	}

	return answer
}

//revive:disable:exported

type MaterialColorSpec interface {
	Background() *Color
	Error() *Color
	ErrorContainer() *Color
	ErrorDim() *Color
	ErrorPaletteKeyColor() *Color
	HighestSurface(s *Scheme) *Color
	InverseOnSurface() *Color
	InversePrimary() *Color
	InverseSurface() *Color
	NeutralPaletteKeyColor() *Color
	NeutralVariantPaletteKeyColor() *Color
	OnBackground() *Color
	OnError() *Color
	OnErrorContainer() *Color
	OnPrimary() *Color
	OnPrimaryContainer() *Color
	OnPrimaryFixed() *Color
	OnPrimaryFixedVariant() *Color
	OnSecondary() *Color
	OnSecondaryContainer() *Color
	OnSecondaryFixed() *Color
	OnSecondaryFixedVariant() *Color
	OnSurface() *Color
	OnSurfaceVariant() *Color
	OnTertiary() *Color
	OnTertiaryContainer() *Color
	OnTertiaryFixed() *Color
	OnTertiaryFixedVariant() *Color
	Outline() *Color
	OutlineVariant() *Color
	Primary() *Color
	PrimaryContainer() *Color
	PrimaryDim() *Color
	PrimaryFixed() *Color
	PrimaryFixedDim() *Color
	PrimaryPaletteKeyColor() *Color
	Scrim() *Color
	Secondary() *Color
	SecondaryContainer() *Color
	SecondaryDim() *Color
	SecondaryFixed() *Color
	SecondaryFixedDim() *Color
	SecondaryPaletteKeyColor() *Color
	Shadow() *Color
	Surface() *Color
	SurfaceBright() *Color
	SurfaceContainer() *Color
	SurfaceContainerHigh() *Color
	SurfaceContainerHighest() *Color
	SurfaceContainerLow() *Color
	SurfaceContainerLowest() *Color
	SurfaceDim() *Color
	SurfaceTint() *Color
	SurfaceVariant() *Color
	Tertiary() *Color
	TertiaryContainer() *Color
	TertiaryDim() *Color
	TertiaryFixed() *Color
	TertiaryFixedDim() *Color
	TertiaryPaletteKeyColor() *Color
}

func DynamicColorFromPalette(args *Color) *Color {
	dc := &Color{
		Name:             args.Name,
		Palette:          args.Palette,
		Tone:             args.Tone,
		IsBackground:     args.IsBackground,
		ChromaMultiplier: args.ChromaMultiplier,
		Background:       args.Background,
		SecondBackground: args.SecondBackground,
		ContrastCurve:    args.ContrastCurve,
		ToneDeltaPair:    args.ToneDeltaPair,
	}
	if dc.Tone == nil {
		if args.Background == nil {
			dc.Tone = func(*Scheme) float64 { return 50 }
		} else {
			dc.Tone = func(s *Scheme) float64 {
				bg := args.Background(s)
				if bg != nil {
					return bg.GetTone(s)
				}
				return 50
			}
		}
	}
	return dc
}

type MaterialSpec2021 struct{}

var _ MaterialColorSpec = (*MaterialSpec2021)(nil)

// HighestSurface returns the highest surface color based on dark mode
func (m MaterialSpec2021) HighestSurface(s *Scheme) *Color {
	if s.Dark {
		return m.SurfaceBright()
	}
	return m.SurfaceDim()
}

func (m MaterialSpec2021) PrimaryPaletteKeyColor() *Color {
	return FromPalette(
		"primary_palette_key_color",
		func(s *Scheme) palettes.TonalPalette { return s.PrimaryPalette },
		func(s *Scheme) float64 { return s.PrimaryPalette.KeyColor.Tone },
	)
}

func (m MaterialSpec2021) SecondaryPaletteKeyColor() *Color {
	return FromPalette(
		"secondary_palette_key_color",
		func(s *Scheme) palettes.TonalPalette { return s.SecondaryPalette },
		func(s *Scheme) float64 { return s.SecondaryPalette.KeyColor.Tone },
	)
}

func (m MaterialSpec2021) TertiaryPaletteKeyColor() *Color {
	return FromPalette(
		"tertiary_palette_key_color",
		func(s *Scheme) palettes.TonalPalette { return s.TertiaryPalette },
		func(s *Scheme) float64 { return s.TertiaryPalette.KeyColor.Tone },
	)
}

func (m MaterialSpec2021) NeutralPaletteKeyColor() *Color {
	return FromPalette(
		"neutral_palette_key_color",
		func(s *Scheme) palettes.TonalPalette { return s.NeutralPalette },
		func(s *Scheme) float64 { return s.NeutralPalette.KeyColor.Tone },
	)
}

func (m MaterialSpec2021) NeutralVariantPaletteKeyColor() *Color {
	return FromPalette(
		"neutral_variant_palette_key_color",
		func(s *Scheme) palettes.TonalPalette { return s.NeutralVariantPalette },
		func(s *Scheme) float64 { return s.NeutralVariantPalette.KeyColor.Tone },
	)
}

func (m MaterialSpec2021) ErrorPaletteKeyColor() *Color {
	return FromPalette(
		"error_palette_key_color",
		func(s *Scheme) palettes.TonalPalette { return s.ErrorPalette },
		func(s *Scheme) float64 { return s.ErrorPalette.KeyColor.Tone },
	)
}

func (m MaterialSpec2021) Background() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "background",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 6.0
			}
			return 98.0
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) OnBackground() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_background",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 90.0
			}
			return 10.0
		},
		IsBackground: false,
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 3.0, 4.5, 7.0)
		},
		Background: func(*Scheme) *Color { return m.Background() },
	})
}

func (m MaterialSpec2021) Surface() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 6.0
			}
			return 98.0
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) SurfaceDim() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface_dim",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			cc := NewContrastCurve(87.0, 87.0, 80.0, 75.0).Get(s.Contrast)
			if s.Dark {
				return 6.0
			}
			return cc
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) SurfaceBright() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface_bright",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return NewContrastCurve(
					24.0,
					24.0,
					29.0,
					34.0,
				).Get(s.Contrast)
			}
			return 98.0
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) SurfaceContainerLowest() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface_container_lowest",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return NewContrastCurve(4.0, 4.0, 2.0, 0).Get(s.Contrast)
			}
			return 100.0
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) SurfaceContainerLow() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface_container_low",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return NewContrastCurve(
					10.0,
					10.0,
					11.0,
					12.0,
				).Get(s.Contrast)
			}
			return NewContrastCurve(96.0, 96.0, 96.0, 95.0).Get(s.Contrast)
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) SurfaceContainer() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface_container",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return NewContrastCurve(
					12.0,
					12.0,
					16.0,
					20.0,
				).Get(s.Contrast)
			}
			return NewContrastCurve(94.0, 94.0, 92.0, 90.0).Get(s.Contrast)
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) SurfaceContainerHigh() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface_container_high",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return NewContrastCurve(
					17.0,
					17.0,
					21.0,
					25.0,
				).Get(s.Contrast)
			}
			return NewContrastCurve(92.0, 92.0, 88.0, 85.0).Get(s.Contrast)
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) SurfaceContainerHighest() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface_container_highest",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return NewContrastCurve(
					22.0,
					22.0,
					26.0,
					30.0,
				).Get(s.Contrast)
			}
			return NewContrastCurve(90.0, 90.0, 84.0, 80.0).Get(s.Contrast)
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) OnSurface() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_surface",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 90.0
			}
			return 10.0
		},
		IsBackground: false,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(4.5, 7.0, 11.0, 21.0)
		},
	})
}

func (m MaterialSpec2021) SurfaceVariant() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface_variant",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralVariantPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 30.0
			}
			return 90.0
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) OnSurfaceVariant() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_surface_variant",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralVariantPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 80.0
			}
			return 30.0
		},
		IsBackground: false,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 11.0)
		},
	})
}

func (m MaterialSpec2021) InverseSurface() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "inverse_surface",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 90.0
			}
			return 20.0
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) InverseOnSurface() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "inverse_on_surface",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 20.0
			}
			return 95.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.InverseSurface()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(4.5, 7.0, 11.0, 21.0)
		},
	})
}

func (m MaterialSpec2021) Outline() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "outline",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralVariantPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 60.0
			}
			return 50.0
		},
		IsBackground: false,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.5, 3.0, 4.5, 7.0)
		},
	})
}

func (m MaterialSpec2021) OutlineVariant() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "outline_variant",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralVariantPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 30.0
			}
			return 80.0
		},
		IsBackground: false,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
	})
}

func (m MaterialSpec2021) Shadow() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "shadow",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone:         func(*Scheme) float64 { return 0 },
		IsBackground: false,
	})
}

func (m MaterialSpec2021) Scrim() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "scrim",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.NeutralPalette
		},
		Tone:         func(*Scheme) float64 { return 0 },
		IsBackground: false,
	})
}

func (m MaterialSpec2021) SurfaceTint() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "surface_tint",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 80.0
			}
			return 40.0
		},
		IsBackground: true,
	})
}

func (m MaterialSpec2021) Primary() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "primary",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				if s.Dark {
					return 100.0
				}
				return 0.0
			}
			if s.Dark {
				return 80.0
			}
			return 40.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 7.0)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.PrimaryContainer(),
				m.Primary(),
				10,
				TonePolarityNearer,
				false,
			)
		},
	})
}

func (m MaterialSpec2021) OnPrimary() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_primary",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				if s.Dark {
					return 10.0
				}
				return 90.0
			}
			if s.Dark {
				return 20.0
			}
			return 100.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.Primary()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(4.5, 7.0, 11.0, 21.0)
		},
	})
}

func (m MaterialSpec2021) PrimaryContainer() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "primary_container",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsFidelity(s) {
				return s.SourceColorHct.Tone
			}
			if IsMonochrome(s) {
				if s.Dark {
					return 85.0
				}
				return 25.0
			}
			if s.Dark {
				return 30.0
			}
			return 90.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.PrimaryContainer(),
				m.Primary(),
				10,
				TonePolarityNearer,
				false,
			)
		},
	})
}

func (m MaterialSpec2021) PrimaryDim() *Color {
	return nil
}

func (m MaterialSpec2021) OnPrimaryContainer() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_primary_container",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsFidelity(s) {
				return ForegroundTone(m.PrimaryContainer().Tone(s), 4.5)
			}
			if IsMonochrome(s) {
				if s.Dark {
					return 0.0
				}
				return 100.0
			}
			if s.Dark {
				return 90.0
			}
			return 30.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.PrimaryContainer()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 11.0)
		},
	})
}

func (m MaterialSpec2021) InversePrimary() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "inverse_primary",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 40.0
			}
			return 80.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.InverseSurface()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 7.0)
		},
	})
}

func (m MaterialSpec2021) Secondary() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "secondary",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.SecondaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 80.0
			}
			return 40.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 7.0)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.SecondaryContainer(),
				m.Secondary(),
				10,
				TonePolarityNearer,
				false,
			)
		},
	})
}

func (m MaterialSpec2021) OnSecondary() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_secondary",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.SecondaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				if s.Dark {
					return 10.0
				}
				return 100.0
			}
			if s.Dark {
				return 20.0
			}
			return 100.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.Secondary()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(4.5, 7.0, 11.0, 21.0)
		},
	})
}

func (m MaterialSpec2021) SecondaryContainer() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "secondary_container",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.SecondaryPalette
		},
		Tone: func(s *Scheme) float64 {
			initialTone := 90.0
			if s.Dark {
				initialTone = 30.0
			}
			if IsMonochrome(s) {
				if s.Dark {
					return 30.0
				}
				return 85.0
			}
			if !IsFidelity(s) {
				return initialTone
			}
			return FindDesiredChromaByTone(
				s.SecondaryPalette.Hue,
				s.SecondaryPalette.Chroma,
				initialTone,
				!s.Dark,
			)
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.SecondaryContainer(),
				m.Secondary(),
				10,
				TonePolarityNearer,
				false,
			)
		},
	})
}

func (m MaterialSpec2021) SecondaryDim() *Color {
	return nil
}

func (m MaterialSpec2021) OnSecondaryContainer() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_secondary_container",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.SecondaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				if s.Dark {
					return 90.0
				}
				return 10.0
			}
			if !IsFidelity(s) {
				if s.Dark {
					return 90.0
				}
				return 30.0
			}
			return ForegroundTone(m.SecondaryContainer().Tone(s), 4.5)
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.SecondaryContainer()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 11.0)
		},
	})
}

func (m MaterialSpec2021) Tertiary() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "tertiary",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.TertiaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				if s.Dark {
					return 90.0
				}
				return 25.0
			}
			if s.Dark {
				return 80.0
			}
			return 40.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 7.0)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.TertiaryContainer(),
				m.Tertiary(),
				10.0,
				TonePolarityNearer,
				false,
			)
		},
	})
}

func (m MaterialSpec2021) OnTertiary() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_tertiary",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.TertiaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				if s.Dark {
					return 10.0
				}
				return 90.0
			}
			if s.Dark {
				return 20.0
			}
			return 100.0
		},
		IsBackground: false,
		Background:   func(*Scheme) *Color { return m.Tertiary() },
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(4.5, 7.0, 11.0, 21.0)
		},
	})
}

func (m MaterialSpec2021) TertiaryContainer() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "tertiary_container",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.TertiaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				if s.Dark {
					return 60.0
				}
				return 49.0
			}
			if !IsFidelity(s) {
				if s.Dark {
					return 30.0
				}
				return 90.0
			}
			proposed := s.TertiaryPalette.Tone(s.SourceColorHct.Tone).ToHct()
			return dislike.FixIfDisliked(proposed).Tone
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.TertiaryContainer(),
				m.Tertiary(),
				10.0,
				TonePolarityNearer,
				false,
			)
		},
	})
}

func (m MaterialSpec2021) TertiaryDim() *Color {
	return nil
}

func (m MaterialSpec2021) OnTertiaryContainer() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_tertiary_container",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.TertiaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				if s.Dark {
					return 0.0
				}
				return 100.0
			}
			if !IsFidelity(s) {
				if s.Dark {
					return 90.0
				}
				return 30.0
			}
			return ForegroundTone(m.TertiaryContainer().Tone(s), 4.5)
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.TertiaryContainer()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 11.0)
		},
	})
}

func (m MaterialSpec2021) Error() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "error",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.ErrorPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 80.0
			}
			return 40.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 7.0)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.ErrorContainer(),
				m.Error(),
				10.0,
				TonePolarityNearer,
				false,
			)
		},
	})
}

func (m MaterialSpec2021) OnError() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_error",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.ErrorPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 20.0
			}
			return 100.0
		},
		IsBackground: false,
		Background:   func(*Scheme) *Color { return m.Error() },
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(4.5, 7.0, 11.0, 21.0)
		},
	})
}

func (m MaterialSpec2021) ErrorContainer() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "error_container",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.ErrorPalette
		},
		Tone: func(s *Scheme) float64 {
			if s.Dark {
				return 30.0
			}
			return 90.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.ErrorContainer(),
				m.Error(),
				10.0,
				TonePolarityNearer,
				false,
			)
		},
	})
}

func (m MaterialSpec2021) ErrorDim() *Color {
	return nil
}

func (m MaterialSpec2021) OnErrorContainer() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_error_container",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.ErrorPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				if s.Dark {
					return 90.0
				}
				return 10.0
			}
			if s.Dark {
				return 90.0
			}
			return 30.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.ErrorContainer()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 11.0)
		},
	})
}

func (m MaterialSpec2021) PrimaryFixed() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "primary_fixed",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 40.0
			}
			return 90.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.PrimaryFixed(),
				m.PrimaryFixedDim(),
				10.0,
				TonePolarityLighter,
				true)
		},
	})
}

func (m MaterialSpec2021) PrimaryFixedDim() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "primary_fixed_dim",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 30.0
			}
			return 80.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.PrimaryFixed(),
				m.PrimaryFixedDim(),
				10.0,
				TonePolarityLighter, true)
		},
	})
}

func (m MaterialSpec2021) OnPrimaryFixed() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_primary_fixed",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 100.0
			}
			return 10.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.PrimaryFixedDim()
		},
		SecondBackground: func(*Scheme) *Color {
			return m.PrimaryFixed()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(4.5, 7.0, 11.0, 21.0)
		},
	})
}

func (m MaterialSpec2021) OnPrimaryFixedVariant() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_primary_fixed_variant",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.PrimaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 90.0
			}
			return 30.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.PrimaryFixedDim()
		},
		SecondBackground: func(*Scheme) *Color {
			return m.PrimaryFixed()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 11.0)
		},
	})
}

func (m MaterialSpec2021) SecondaryFixed() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "secondary_fixed",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.SecondaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 80.0
			}
			return 90.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.SecondaryFixed(),
				m.SecondaryFixedDim(),
				10.0,
				TonePolarityLighter,
				true,
			)
		},
	})
}

func (m MaterialSpec2021) SecondaryFixedDim() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "secondary_fixed_dim",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.SecondaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 70.0
			}
			return 80.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.SecondaryFixed(),
				m.SecondaryFixedDim(),
				10.0,
				TonePolarityLighter,
				true,
			)
		},
	})
}

func (m MaterialSpec2021) OnSecondaryFixed() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_secondary_fixed",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.SecondaryPalette
		},
		Tone:         func(*Scheme) float64 { return 10.0 },
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.SecondaryFixedDim()
		},
		SecondBackground: func(*Scheme) *Color {
			return m.SecondaryFixed()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(4.5, 7.0, 11.0, 21.0)
		},
	})
}

func (m MaterialSpec2021) OnSecondaryFixedVariant() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_secondary_fixed_variant",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.SecondaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 25.0
			}
			return 30.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.SecondaryFixedDim()
		},
		SecondBackground: func(*Scheme) *Color {
			return m.SecondaryFixed()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 11.0)
		},
	})
}

func (m MaterialSpec2021) TertiaryFixed() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "tertiary_fixed",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.TertiaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 40.0
			}
			return 90.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.TertiaryFixed(),
				m.TertiaryFixedDim(),
				10.0,
				TonePolarityLighter,
				true,
			)
		},
	})
}

func (m MaterialSpec2021) TertiaryFixedDim() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "tertiary_fixed_dim",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.TertiaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 30.0
			}
			return 80.0
		},
		IsBackground: true,
		Background: func(s *Scheme) *Color {
			return m.HighestSurface(s)
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(1.0, 1.0, 3.0, 4.5)
		},
		ToneDeltaPair: func(*Scheme) *ToneDeltaPair {
			return NewToneDeltaPair(
				m.TertiaryFixed(),
				m.TertiaryFixedDim(),
				10.0,
				TonePolarityLighter,
				true,
			)
		},
	})
}

func (m MaterialSpec2021) OnTertiaryFixed() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_tertiary_fixed",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.TertiaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 100.0
			}
			return 10.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.TertiaryFixedDim()
		},
		SecondBackground: func(*Scheme) *Color {
			return m.TertiaryFixed()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(4.5, 7.0, 11.0, 21.0)
		},
	})
}

func (m MaterialSpec2021) OnTertiaryFixedVariant() *Color {
	return DynamicColorFromPalette(&Color{
		Name: "on_tertiary_fixed_variant",
		Palette: func(s *Scheme) palettes.TonalPalette {
			return s.TertiaryPalette
		},
		Tone: func(s *Scheme) float64 {
			if IsMonochrome(s) {
				return 90.0
			}
			return 30.0
		},
		IsBackground: false,
		Background: func(*Scheme) *Color {
			return m.TertiaryFixedDim()
		},
		SecondBackground: func(*Scheme) *Color {
			return m.TertiaryFixed()
		},
		ContrastCurve: func(*Scheme) *ContrastCurve {
			return NewContrastCurve(3.0, 4.5, 7.0, 11.0)
		},
	})
}
