package dynamic

import (
	"fmt"
	"testing"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/AvengeMedia/dankgo/material/contrast"
	"github.com/stretchr/testify/require"
)

var seedColors = []color.Hct{
	color.ARGB(0xFFFF0000).ToHct(),
	color.ARGB(0xFFFFFF00).ToHct(),
	color.ARGB(0xFF00FF00).ToHct(),
	color.ARGB(0xFF0000FF).ToHct(),
}

type pair struct {
	fgName, bgName string
}

func materialDynamicColors(m MaterialColorSpec) []*Color {
	return []*Color{
		m.Background(),
		m.OnBackground(),
		m.Surface(),
		m.SurfaceDim(),
		m.SurfaceBright(),
		m.SurfaceContainerLowest(),
		m.SurfaceContainerLow(),
		m.SurfaceContainer(),
		m.SurfaceContainerHigh(),
		m.SurfaceContainerHighest(),
		m.OnSurface(),
		m.SurfaceVariant(),
		m.OnSurfaceVariant(),
		m.InverseSurface(),
		m.InverseOnSurface(),
		m.Outline(),
		m.OutlineVariant(),
		m.Shadow(),
		m.Scrim(),
		m.SurfaceTint(),
		m.Primary(),
		m.OnPrimary(),
		m.PrimaryContainer(),
		m.OnPrimaryContainer(),
		m.InversePrimary(),
		m.Secondary(),
		m.OnSecondary(),
		m.SecondaryContainer(),
		m.OnSecondaryContainer(),
		m.Tertiary(),
		m.OnTertiary(),
		m.TertiaryContainer(),
		m.OnTertiaryContainer(),
		m.Error(),
		m.OnError(),
		m.ErrorContainer(),
		m.OnErrorContainer(),
		m.PrimaryFixed(),
		m.PrimaryFixedDim(),
		m.OnPrimaryFixed(),
		m.OnPrimaryFixedVariant(),
		m.SecondaryFixed(),
		m.SecondaryFixedDim(),
		m.OnSecondaryFixed(),
		m.OnSecondaryFixedVariant(),
		m.TertiaryFixed(),
		m.TertiaryFixedDim(),
		m.OnTertiaryFixed(),
		m.OnTertiaryFixedVariant(),
	}
}

func colorByName(m MaterialColorSpec) map[string]*Color {
	byName := map[string]*Color{}
	for _, c := range materialDynamicColors(m) {
		byName[c.Name] = c
	}
	return byName
}

var textSurfacePairs = []pair{
	{"on_primary", "primary"},
	{"on_primary_container", "primary_container"},
	{"on_secondary", "secondary"},
	{"on_secondary_container", "secondary_container"},
	{"on_tertiary", "tertiary"},
	{"on_tertiary_container", "tertiary_container"},
	{"on_error", "error"},
	{"on_error_container", "error_container"},
	{"on_background", "background"},
	{"on_surface_variant", "surface_bright"},
	{"on_surface_variant", "surface_dim"},
}

func getMinRequirement(curve *ContrastCurve, level float64) float64 {
	if level >= 1 {
		return curve.high
	}
	if level >= 0.5 {
		return curve.medium
	}
	if level >= 0 {
		return curve.normal
	}
	return curve.low
}

func getPairs(resp bool, fores, backs []string) [][2]string {
	var ans [][2]string
	if resp {
		for i := range fores {
			ans = append(ans, [2]string{fores[i], backs[i]})
		}
		return ans
	}
	for _, f := range fores {
		for _, b := range backs {
			ans = append(ans, [2]string{f, b})
		}
	}
	return ans
}

func dynamicColorSchemes() []*Scheme {
	var schemes []*Scheme
	for _, seed := range seedColors {
		for _, contrastLevel := range []float64{-1.0, -0.75, -0.5, -0.25, 0.0, 0.25, 0.5, 0.75, 1.0} {
			for _, isDark := range []bool{false, true} {
				for _, variant := range []Variant{
					VariantContent,
					VariantExpressive,
					VariantFidelity,
					VariantMonochrome,
					VariantNeutral,
					VariantTonalSpot,
					VariantVibrant,
				} {
					schemes = append(schemes, NewDynamicScheme(seed, variant, contrastLevel, isDark, PlatformPhone, Version2021))
				}
			}
		}
	}
	return schemes
}

func schemeLabel(s *Scheme) string {
	return fmt.Sprintf("%s seed=%s dark=%v contrast=%v", s.Variant, s.SourceColorARGB(), s.Dark, s.Contrast)
}

type constraint struct {
	kind         string
	values       *ContrastCurve
	fore, back   []string
	respectively bool
	delta        float64
	polarity     string
	objects      []string
}

func TestDynamicColor(t *testing.T) {
	schemes := dynamicColorSchemes()

	t.Run("generates colors respecting contrast", func(t *testing.T) {
		for _, scheme := range schemes {
			byName := colorByName(scheme.MaterialColor)
			for _, p := range textSurfacePairs {
				foregroundTone := byName[p.fgName].GetHct(scheme).Tone
				backgroundTone := byName[p.bgName].GetHct(scheme).Tone
				ratio := contrast.RatioOfTones(foregroundTone, backgroundTone)

				minimumRequirement := 3.0
				if scheme.Contrast >= 0.0 {
					minimumRequirement = 4.5
				}

				if ratio < minimumRequirement {
					t.Errorf("%s: %s on %s is %v, needed %v",
						schemeLabel(scheme), p.fgName, p.bgName, ratio, minimumRequirement)
				}
			}
		}
	})

	t.Run("constraint conformance test", func(t *testing.T) {
		limitingSurfaces := []string{
			"surface_dim",
			"surface_bright",
		}

		constraints := []constraint{
			{
				kind:   "Contrast",
				values: NewContrastCurve(4.5, 7, 11, 21),
				fore:   []string{"on_surface"},
				back:   limitingSurfaces,
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(3, 4.5, 7, 11),
				fore:   []string{"on_surface_variant"},
				back:   limitingSurfaces,
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(3, 4.5, 7, 7),
				fore:   []string{"primary", "secondary", "tertiary", "error"},
				back:   limitingSurfaces,
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(1.5, 3, 4.5, 7),
				fore:   []string{"outline"},
				back:   limitingSurfaces,
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(0, 0, 3, 4.5),
				fore: []string{
					"primary_container",
					"primary_fixed",
					"primary_fixed_dim",
					"secondary_container",
					"secondary_fixed",
					"secondary_fixed_dim",
					"tertiary_container",
					"tertiary_fixed",
					"tertiary_fixed_dim",
					"error_container",
					"outline_variant",
				},
				back: limitingSurfaces,
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(4.5, 7, 11, 21),
				fore:   []string{"inverse_on_surface"},
				back:   []string{"inverse_surface"},
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(3, 4.5, 7, 7),
				fore:   []string{"inverse_primary"},
				back:   []string{"inverse_surface"},
			},
			{
				kind:         "Contrast",
				respectively: true,
				values:       NewContrastCurve(4.5, 7, 11, 21),
				fore:         []string{"on_primary", "on_secondary", "on_tertiary", "on_error"},
				back:         []string{"primary", "secondary", "tertiary", "error"},
			},
			{
				kind:         "Contrast",
				respectively: true,
				values:       NewContrastCurve(3, 4.5, 7, 11),
				fore: []string{
					"on_primary_container",
					"on_secondary_container",
					"on_tertiary_container",
					"on_error_container",
				},
				back: []string{
					"primary_container",
					"secondary_container",
					"tertiary_container",
					"error_container",
				},
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(4.5, 7, 11, 21),
				fore:   []string{"on_primary_fixed"},
				back:   []string{"primary_fixed", "primary_fixed_dim"},
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(4.5, 7, 11, 21),
				fore:   []string{"on_secondary_fixed"},
				back:   []string{"secondary_fixed", "secondary_fixed_dim"},
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(4.5, 7, 11, 21),
				fore:   []string{"on_tertiary_fixed"},
				back:   []string{"tertiary_fixed", "tertiary_fixed_dim"},
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(3, 4.5, 7, 11),
				fore:   []string{"on_primary_fixed_variant"},
				back:   []string{"primary_fixed", "primary_fixed_dim"},
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(3, 4.5, 7, 11),
				fore:   []string{"on_secondary_fixed_variant"},
				back:   []string{"secondary_fixed", "secondary_fixed_dim"},
			},
			{
				kind:   "Contrast",
				values: NewContrastCurve(3, 4.5, 7, 11),
				fore:   []string{"on_tertiary_fixed_variant"},
				back:   []string{"tertiary_fixed", "tertiary_fixed_dim"},
			},
			{
				kind:         "Delta",
				delta:        10,
				respectively: true,
				fore:         []string{"primary", "secondary", "tertiary", "error"},
				back: []string{
					"primary_container",
					"secondary_container",
					"tertiary_container",
					"error_container",
				},
				polarity: "farther",
			},
			{
				kind:         "Delta",
				delta:        10,
				respectively: true,
				fore:         []string{"primary_fixed_dim", "secondary_fixed_dim", "tertiary_fixed_dim"},
				back:         []string{"primary_fixed", "secondary_fixed", "tertiary_fixed"},
				polarity:     "darker",
			},
			{
				kind: "Background",
				objects: []string{
					"background",
					"error",
					"error_container",
					"primary",
					"primary_container",
					"primary_fixed",
					"primary_fixed_dim",
					"secondary",
					"secondary_container",
					"secondary_fixed",
					"secondary_fixed_dim",
					"surface",
					"surface_bright",
					"surface_container",
					"surface_container_high",
					"surface_container_highest",
					"surface_container_low",
					"surface_container_lowest",
					"surface_dim",
					"surface_tint",
					"surface_variant",
					"tertiary",
					"tertiary_container",
					"tertiary_fixed",
					"tertiary_fixed_dim",
				},
			},
		}

		for _, scheme := range schemes {
			resolvedColors := map[string]color.ARGB{}
			for _, c := range materialDynamicColors(scheme.MaterialColor) {
				resolvedColors[c.Name] = c.GetArgb(scheme)
			}
			label := schemeLabel(scheme)

			for _, cstr := range constraints {
				switch cstr.kind {
				case "Contrast":
					const contrastTolerance = 0.05

					minRequirement := getMinRequirement(cstr.values, scheme.Contrast)
					for _, pr := range getPairs(cstr.respectively, cstr.fore, cstr.back) {
						fore, back := pr[0], pr[1]
						ftone := resolvedColors[fore].LStar()
						btone := resolvedColors[back].LStar()
						ratio := contrast.RatioOfTones(ftone, btone)

						failing := ratio < minRequirement-contrastTolerance
						if minRequirement > 4.5 {
							failing = ftone != 0 && btone != 0 && ftone != 100 && btone != 100 &&
								ratio < minRequirement-contrastTolerance
						}

						if ratio < minRequirement-contrastTolerance && minRequirement <= 4.5 {
							t.Errorf("%s: Contrast %s %.2f %s %.2f %.2f %v ",
								label, fore, ftone, back, btone, ratio, minRequirement)
						}
						if failing && minRequirement > 4.5 {
							t.Errorf("%s: Contrast(stretch-goal) %s %.2f %s %.2f %.2f %v ",
								label, fore, ftone, back, btone, ratio, minRequirement)
						}
					}
				case "Delta":
					polarity := cstr.polarity
					require.Contains(t, []string{"nearer", "farther", "lighter", "darker"}, polarity)
					for _, pr := range getPairs(cstr.respectively, cstr.fore, cstr.back) {
						fore, back := pr[0], pr[1]
						ftone := resolvedColors[fore].LStar()
						btone := resolvedColors[back].LStar()

						isLighter := polarity == "lighter" ||
							(polarity == "nearer" && !scheme.Dark) ||
							(polarity == "farther" && scheme.Dark)

						observedDelta := btone - ftone
						if isLighter {
							observedDelta = ftone - btone
						}

						if observedDelta < cstr.delta-0.5 {
							t.Errorf("%s: Delta %s %.2f %s %.2f %.2f %v",
								label, fore, ftone, back, btone, observedDelta, cstr.delta)
						}
					}
				case "Background":
					for _, bg := range cstr.objects {
						bgtone := resolvedColors[bg].LStar()
						if bgtone >= 50.5 && bgtone < 59.5 {
							t.Errorf("%s: Background %s %.2f",
								label, bg, bgtone)
						}
					}
				default:
					t.Errorf("Bad constraint kind = %s", cstr.kind)
				}
			}
		}
	})
}
