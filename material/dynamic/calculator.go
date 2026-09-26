// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package dynamic

import (
	"math"
	"strings"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/AvengeMedia/dankgo/material/contrast"
	"github.com/AvengeMedia/dankgo/material/num"
	"github.com/AvengeMedia/dankgo/material/palettes"
)

// ColorCalculationDelegate provides the HCT and tone of a color within a
// scheme.
//
// This interface allows different implementations of color calculation logic
// for different Material Design spec versions.
type ColorCalculationDelegate interface {
	// GetHct returns the HCT color for the given scheme and color.
	GetHct(scheme *Scheme, dc *Color) color.Hct
	// GetTone returns the tone value for the given scheme and color.
	GetTone(scheme *Scheme, dc *Color) float64
}

type (
	colorCalculationDelegateImpl2021 struct{}
	colorCalculationDelegateImpl2025 struct{}
)

var (
	// ColorCalculation2025 is a delegate implementing color calculation logic
	// following the Material Design 2025 specification.
	ColorCalculation2025 = colorCalculationDelegateImpl2025{}
	// ColorCalculation2021 is a delegate implementing color calculation logic
	// following the Material Design 2021 specification.
	ColorCalculation2021 = colorCalculationDelegateImpl2021{}
)

// GetHct returns the HCT color for the given scheme and color.
func (d colorCalculationDelegateImpl2021) GetHct(
	scheme *Scheme,
	dc *Color,
) color.Hct {
	tone := dc.GetTone(scheme)
	palette := dc.Palette(scheme)
	return palette.GetHct(tone)
}

// GetTone returns the tone value for the given scheme and color.
func (d colorCalculationDelegateImpl2021) GetTone(
	scheme *Scheme,
	dc *Color,
) float64 {
	decreasingContrast := scheme.Contrast < 0

	var toneDeltaPair *ToneDeltaPair
	if dc.ToneDeltaPair != nil {
		toneDeltaPair = dc.ToneDeltaPair(scheme)
	}
	if toneDeltaPair != nil {
		toneDeltaPair := dc.ToneDeltaPair(scheme)
		roleA := toneDeltaPair.RoleA
		roleB := toneDeltaPair.RoleB
		delta := toneDeltaPair.Delta
		polarity := toneDeltaPair.Polarity
		stayTogether := toneDeltaPair.StayTogether

		aIsNearer := polarity == TonePolarityNearer ||
			(polarity == TonePolarityLighter && !scheme.Dark) ||
			(polarity == TonePolarityDarker && scheme.Dark)
		nearer := roleA
		farther := roleB
		if !aIsNearer {
			nearer = roleB
			farther = roleA
		}
		amNearer := dc.Name == nearer.Name
		expansionDir := -1.0
		if scheme.Dark {
			expansionDir = 1.0
		}
		nTone := nearer.Tone(scheme)
		fTone := farther.Tone(scheme)

		if dc.Background != nil && nearer.ContrastCurve != nil &&
			farther.ContrastCurve != nil {
			bg := dc.Background(scheme)
			nCurve := nearer.ContrastCurve(scheme)
			fCurve := farther.ContrastCurve(scheme)
			if bg != nil && nCurve != nil && fCurve != nil {
				bgTone := bg.GetTone(scheme)
				nContrast := nCurve.Get(scheme.Contrast)
				fContrast := fCurve.Get(scheme.Contrast)

				if contrast.RatioOfTones(bgTone, nTone) < nContrast {
					nTone = ForegroundTone(bgTone, nContrast)
				}
				if contrast.RatioOfTones(bgTone, fTone) < fContrast {
					fTone = ForegroundTone(bgTone, fContrast)
				}
				if decreasingContrast {
					nTone = ForegroundTone(bgTone, nContrast)
					fTone = ForegroundTone(bgTone, fContrast)
				}
			}
		}

		if (fTone-nTone)*expansionDir < delta {
			fTone = num.Clamp(0, 100, nTone+delta*expansionDir)
			if (fTone-nTone)*expansionDir < delta {
				nTone = num.Clamp(0, 100, fTone-delta*expansionDir)
			}
		}

		if nTone >= 50 && nTone < 60 {
			if expansionDir > 0 {
				nTone = 60
				fTone = math.Max(fTone, nTone+delta*expansionDir)
			} else {
				nTone = 49
				fTone = math.Min(fTone, nTone+delta*expansionDir)
			}
		} else if fTone >= 50 && fTone < 60 {
			if stayTogether {
				if expansionDir > 0 {
					nTone = 60
					fTone = math.Max(fTone, nTone+delta*expansionDir)
				} else {
					nTone = 49
					fTone = math.Min(fTone, nTone+delta*expansionDir)
				}
			} else {
				if expansionDir > 0 {
					fTone = 60
				} else {
					fTone = 49
				}
			}
		}

		if amNearer {
			return nTone
		}
		return fTone
	}

	answer := dc.Tone(scheme)
	bg := dc.Background
	contrastCurve := dc.ContrastCurve

	if bg == nil || bg(scheme) == nil || contrastCurve == nil ||
		contrastCurve(scheme) == nil {
		return answer
	}

	bgTone := bg(scheme).GetTone(scheme)
	desiredRatio := contrastCurve(scheme).Get(scheme.Contrast)

	if contrast.RatioOfTones(bgTone, answer) < desiredRatio ||
		decreasingContrast {
		answer = ForegroundTone(bgTone, desiredRatio)
	}

	if dc.IsBackground && answer >= 50 && answer < 60 {
		if contrast.RatioOfTones(49, bgTone) >= desiredRatio {
			answer = 49
		} else {
			answer = 60
		}
	}

	if dc.SecondBackground == nil || dc.SecondBackground(scheme) == nil {
		return answer
	}

	bg1 := dc.Background(scheme)
	bg2 := dc.SecondBackground(scheme)
	bgTone1 := bg1.GetTone(scheme)
	bgTone2 := bg2.GetTone(scheme)
	upper := math.Max(bgTone1, bgTone2)
	lower := math.Min(bgTone1, bgTone2)

	if contrast.RatioOfTones(upper, answer) >= desiredRatio &&
		contrast.RatioOfTones(lower, answer) >= desiredRatio {
		return answer
	}

	lightOption := contrast.Lighter(upper, desiredRatio)
	darkOption := contrast.Darker(lower, desiredRatio)

	availables := []float64{}
	if lightOption != -1 {
		availables = append(availables, lightOption)
	}
	if darkOption != -1 {
		availables = append(availables, darkOption)
	}

	prefersLight := TonePrefersLightForeground(bgTone1) ||
		TonePrefersLightForeground(bgTone2)

	if prefersLight {
		if lightOption < 0 {
			return 100
		}
		return lightOption
	}
	if len(availables) == 1 {
		return availables[0]
	}
	if darkOption < 0 {
		return 0
	}
	return darkOption
}

// GetHct returns the HCT color for the given scheme and color.
func (d colorCalculationDelegateImpl2025) GetHct(
	scheme *Scheme,
	dc *Color,
) color.Hct {
	palette := dc.Palette(scheme)
	tone := dc.GetTone(scheme)
	multiplier := 1.0
	if dc.ChromaMultiplier != nil {
		multiplier = dc.ChromaMultiplier(scheme)
	}
	if multiplier == 1 {
		return palette.GetHct(tone)
	}
	chroma := palette.Chroma * multiplier
	if tone == 99 && color.IsYellow(palette.Hue) {
		return palettes.FromHueAndChroma(palette.Hue, chroma).GetHct(tone)
	}
	return color.NewHct(palette.Hue, chroma, tone)
}

// GetTone returns the tone value for the given scheme and color.
func (d colorCalculationDelegateImpl2025) GetTone(
	scheme *Scheme,
	dc *Color,
) float64 {
	var toneDeltaPair *ToneDeltaPair
	if dc.ToneDeltaPair != nil {
		toneDeltaPair = dc.ToneDeltaPair(scheme)
	}

	// Case 0: tone delta constraint.
	if toneDeltaPair != nil {
		roleA := toneDeltaPair.RoleA
		roleB := toneDeltaPair.RoleB
		polarity := toneDeltaPair.Polarity
		constraint := toneDeltaPair.Constraint
		delta := toneDeltaPair.Delta

		absoluteDelta := delta
		if polarity == TonePolarityDarker ||
			(polarity == TonePolarityRelativeLighter && scheme.Dark) ||
			(polarity == TonePolarityRelativeDarker && !scheme.Dark) {
			absoluteDelta = -delta
		}

		amRoleA := dc.Name == roleA.Name
		selfRole := roleA
		refRole := roleB
		if !amRoleA {
			selfRole = roleB
			refRole = roleA
		}

		selfTone := selfRole.Tone(scheme)
		refTone := refRole.GetTone(scheme)
		relativeDelta := absoluteDelta
		if !amRoleA {
			relativeDelta *= -1
		}

		switch constraint {
		case ConstraintExact:
			selfTone = num.Clamp(0, 100, refTone+relativeDelta)
		case ConstraintNearer:
			if relativeDelta > 0 {
				selfTone = num.Clamp(
					0,
					100,
					num.Clamp(refTone, refTone+relativeDelta, selfTone),
				)
			} else {
				selfTone = num.Clamp(
					0,
					100,
					num.Clamp(refTone+relativeDelta, refTone, selfTone),
				)
			}
		case ConstraintFarther:
			if relativeDelta > 0 {
				selfTone = num.Clamp(refTone+relativeDelta, 100, selfTone)
			} else {
				selfTone = num.Clamp(0, refTone+relativeDelta, selfTone)
			}
		}

		if dc.Background != nil && dc.ContrastCurve != nil {
			bg := dc.Background(scheme)
			if bg != nil {
				cc := dc.ContrastCurve(scheme)
				if cc != nil {
					bgTone := bg.GetTone(scheme)
					desiredContrast := cc.Get(scheme.Contrast)
					if contrast.RatioOfTones(
						bgTone,
						selfTone,
					) < desiredContrast ||
						scheme.Contrast < 0 {
						selfTone = ForegroundTone(bgTone, desiredContrast)
					}
				}
			}
		}

		if dc.IsBackground && !strings.HasSuffix(dc.Name, "_fixed_dim") {
			if selfTone >= 57 {
				selfTone = num.Clamp(65, 100, selfTone)
			} else {
				selfTone = num.Clamp(0, 49, selfTone)
			}
		}
		return selfTone
	}

	// Case 1: No tone delta pair; just solve for itself.
	answer := dc.Tone(scheme)

	if dc.Background == nil || dc.ContrastCurve == nil {
		return answer
	}
	bg := dc.Background(scheme)
	cc := dc.ContrastCurve(scheme)
	if bg == nil || cc == nil {
		return answer
	}

	bgTone := bg.GetTone(scheme)
	desiredRatio := cc.Get(scheme.Contrast)
	if contrast.RatioOfTones(bgTone, answer) < desiredRatio ||
		scheme.Contrast < 0 {
		answer = ForegroundTone(bgTone, desiredRatio)
	}

	if dc.IsBackground && !strings.HasSuffix(dc.Name, "_fixed_dim") {
		if answer >= 57 {
			answer = num.Clamp(65, 100, answer)
		} else {
			answer = num.Clamp(0, 49, answer)
		}
	}

	if dc.SecondBackground == nil {
		return answer
	}
	bg2 := dc.SecondBackground(scheme)
	if bg2 == nil {
		return answer
	}

	// Case 2: Adjust for dual backgrounds.
	bg1 := bg
	bgTone1 := bg1.GetTone(scheme)
	bgTone2 := bg2.GetTone(scheme)
	upper := math.Max(bgTone1, bgTone2)
	lower := math.Min(bgTone1, bgTone2)

	if contrast.RatioOfTones(upper, answer) >= desiredRatio &&
		contrast.RatioOfTones(lower, answer) >= desiredRatio {
		return answer
	}

	lightOption := contrast.Lighter(upper, desiredRatio)
	darkOption := contrast.Darker(lower, desiredRatio)

	availables := []float64{}
	if lightOption != -1 {
		availables = append(availables, lightOption)
	}
	if darkOption != -1 {
		availables = append(availables, darkOption)
	}

	prefersLight := TonePrefersLightForeground(bgTone1) ||
		TonePrefersLightForeground(bgTone2)
	if prefersLight {
		if lightOption < 0 {
			return 100
		}
		return lightOption
	}
	if len(availables) == 1 {
		return availables[0]
	}
	if darkOption < 0 {
		return 0
	}
	return darkOption
}
