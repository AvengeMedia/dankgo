// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package color

import (
	"math"

	"github.com/AvengeMedia/dankgo/material/num"
)

// Environment encapsulates all constants needed for CAM16 color conversions.
// These are intermediate values derived from the viewing environment and are
// used
// throughout the CAM16 model to compute perceptual color attributes.
type Environment struct {
	// N is the relative luminance of the background relative to the reference
	// white.
	N float64
	// Aw is the achromatic response to the white point.
	Aw float64
	// Nbb is the brightness induction factor (background).
	Nbb float64
	// Ncb is the chromatic induction factor (background).
	Ncb float64
	// C is the surround exponential non-linearity factor.
	C float64
	// Nc is the chromatic induction factor (surround).
	Nc float64
	// RgbD is the degree of adaptation for each RGB channel after discounting
	// the illuminant.
	RgbD num.Vector3
	// Fl is the luminance-level adaptation factor (nonlinear response).
	Fl float64
	// FlRoot is the fourth root of Fl, used in CAM16 computations.
	FlRoot float64
	// Z is a base exponential factor used in the CAM16 J calculation.
	Z float64
}

var adaptingLuminance = (200 / math.Pi) * YFromLstar(50) / 100

// DefaultEnvironment returns the default sRGB-like viewing conditions.
var DefaultEnvironment = NewEnvironment(adaptingLuminance, 50, 2, false)

// NewEnvironment creates a ViewingConditions instance with the specified
// parameters.
func NewEnvironment(
	adaptingLuminance float64,
	backgroundLstar float64,
	surround float64,
	discountingIlluminant bool,
) Environment {
	if backgroundLstar < 30.0 {
		backgroundLstar = 30.0
	}

	rW, gW, bW := CatMatrix.Mul(WhitePointD65).Values()

	f := 0.8 + surround/10
	var c float64
	if f >= 0.9 {
		c = num.Lerp(0.59, 0.69, (f-0.9)*10)
	} else {
		c = num.Lerp(0.525, 0.59, (f-0.8)*10)
	}

	var d float64
	if discountingIlluminant {
		d = 1
	} else {
		d = f * (1 - (1/3.6)*math.Exp((-adaptingLuminance-42)/92))
	}

	if d > 1 {
		d = 1
	} else if d < 0 {
		d = 0
	}

	nc := f
	rgbD := num.NewVector3(
		d*(100/rW)+1-d,
		d*(100/gW)+1-d,
		d*(100/bW)+1-d,
	)

	k := 1 / (5*adaptingLuminance + 1)
	k4 := k * k * k * k
	k4F := 1 - k4
	fl := k4*adaptingLuminance + 0.1*k4F*k4F*math.Cbrt(5*adaptingLuminance)

	n := YFromLstar(backgroundLstar) / WhitePointD65[1]
	z := 1.48 + math.Sqrt(n)
	nbb := 0.725 / math.Pow(n, 0.2)
	ncb := nbb

	rgbAFactors := num.NewVector3(
		math.Pow((fl*rgbD[0]*rW)/100, 0.42),
		math.Pow((fl*rgbD[1]*gW)/100, 0.42),
		math.Pow((fl*rgbD[2]*bW)/100, 0.42),
	)

	rgbA := num.NewVector3(
		(400*rgbAFactors[0])/(rgbAFactors[0]+27.13),
		(400*rgbAFactors[1])/(rgbAFactors[1]+27.13),
		(400*rgbAFactors[2])/(rgbAFactors[2]+27.13),
	)

	aw := (2.0*rgbA[0] + rgbA[1] + 0.05*rgbA[2]) * nbb

	return Environment{
		N: n, Aw: aw, Nbb: nbb,
		Ncb: ncb, C: c, Nc: nc,
		RgbD: rgbD, Fl: fl, Z: z,
		FlRoot: math.Pow(fl, 0.25),
	}
}
