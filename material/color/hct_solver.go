// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package color

import (
	"math"

	"github.com/AvengeMedia/dankgo/material/num"
)

// ScaledDiscountFromLinRGB is a 3×3 transform matrix that converts
// from linear RGB to the "scaled discount" space used internally
// by the CAM16/HCT color appearance model.
//
// This transformation accounts for human visual response under
// standard viewing conditions, mapping linear RGB signals into
// a perceptual domain where hue, chroma, and tone can be computed.
var ScaledDiscountFromLinRGB = num.NewMatrix3(
	0.001200833568784504, 0.002389694492170889, 0.0002795742885861124,
	0.0005891086651375999, 0.0029785502573438758, 0.0003270666104008398,
	0.00010146692491640572, 0.0005364214359186694, 0.0032979401770712076,
)

// LinrgbFromScaledDiscount is the inverse of ScaledDiscountFromLinRGB.
//
// It converts values from the scaled discount space back into
// linear RGB, reconstructing displayable color values from the
// perceptual model (used when converting HCT → RGB).
var LinrgbFromScaledDiscount = num.NewMatrix3(
	1373.2198709594231, -1100.4251190754821, -7.278681089101213,
	-271.815969077903, 559.6580465940733, -32.46047482791194,
	1.9622899599665666, -57.173814538844006, 308.7233197812385,
)

// YFromLinRGB vector is used to get Y star of CieLab from linear RGB
var YFromLinRGB = num.NewVector3(0.2126, 0.7152, 0.0722)

// trueDelinearized delinearizes an RGB component, returning a floating-point
// number.
//
// rgbComponent: 0.0 <= rgb_component <= 100.0, represents linear R/G/B channel
// returns: 0.0 <= output <= 255.0, color channel converted to regular RGB space
func trueDelinearized(rgbComponent float64) float64 {
	normalized := rgbComponent / 100.0
	var delinearized float64
	if normalized <= 0.0031308 {
		delinearized = normalized * 12.92
	} else {
		delinearized = 1.055*math.Pow(normalized, 1.0/2.4) - 0.055
	}
	return delinearized * 255.0
}

func chromaticAdaptation(component float64) float64 {
	af := math.Pow(math.Abs(component), 0.42)
	return num.Sign(component) * 400.0 * af / (af + 27.13)
}

// hueOf returns the hue of a linear RGB color in CAM16.
//
// linrgb: The linear RGB coordinates of a color.
// returns: The hue of the color in CAM16, in radians.
func hueOf(linrgb num.Vector3) float64 {
	x, y, z := ScaledDiscountFromLinRGB.Mul(linrgb).Values()
	rA := chromaticAdaptation(x)
	gA := chromaticAdaptation(y)
	bA := chromaticAdaptation(z)
	// redness-greenness
	a := (11.0*rA + -12.0*gA + bA) / 11.0
	// yellowness-blueness
	b := (rA + gA - 2.0*bA) / 9.0
	return math.Atan2(b, a)
}

func areInCyclicOrder(a, b, c float64) bool {
	deltaAB := num.NormalizeRadian(b - a)
	deltaAC := num.NormalizeRadian(c - a)
	return deltaAB < deltaAC
}

// intercept solves the lerp equation.
//
// source: The starting number.
// mid: The number in the middle.
// target: The ending number.
// returns: A number t such that lerp(source, target, t) = mid.
func intercept(source, mid, target float64) float64 {
	return (mid - source) / (target - source)
}

func lerpPoint(source num.Vector3, t float64, target num.Vector3) num.Vector3 {
	return num.NewVector3(
		source[0]+(target[0]-source[0])*t,
		source[1]+(target[1]-source[1])*t,
		source[2]+(target[2]-source[2])*t,
	)
}

// setCoordinate intersects a segment with a plane.
//
// source: The coordinates of point A.
// coordinate: The R-, G-, or B-coordinate of the plane.
// target: The coordinates of point B.
// axis: The axis the plane is perpendicular with. (0: R, 1: G, 2: B)
// returns: The intersection point of the segment AB with the plane
// R=coordinate, G=coordinate, or B=coordinate
func setCoordinate(
	source num.Vector3,
	coordinate float64,
	target num.Vector3,
	axis int,
) num.Vector3 {
	t := intercept(source[axis], coordinate, target[axis])
	return lerpPoint(source, t, target)
}

func isBounded(x float64) bool {
	return 0.0 <= x && x <= 100.0
}

// nthVertex returns the nth possible vertex of the polygonal intersection.
//
// y: The Y value of the plane.
// n: The zero-based index of the point. 0 <= n <= 11.
// returns: The nth possible vertex of the polygonal intersection of the y plane
// and the RGB cube, in linear RGB coordinates, if it exists. If this possible
// vertex lies outside of the cube,
// [-1.0, -1.0, -1.0] is returned.
func nthVertex(y float64, n int) num.Vector3 {
	kR := YFromLinRGB[0]
	kG := YFromLinRGB[1]
	kB := YFromLinRGB[2]
	coordA := 0.0
	if n%4 > 1 {
		coordA = 100.0
	}
	coordB := 0.0
	if n%2 == 1 {
		coordB = 100.0
	}
	if n < 4 {
		g := coordA
		b := coordB
		r := (y - g*kG - b*kB) / kR
		if isBounded(r) {
			return num.NewVector3(r, g, b)
		}
		return num.NewVector3(-1.0, -1.0, -1.0)
	} else if n < 8 {
		b := coordA
		r := coordB
		g := (y - r*kR - b*kB) / kG
		if isBounded(g) {
			return num.NewVector3(r, g, b)
		}
		return num.NewVector3(-1.0, -1.0, -1.0)
	}
	r := coordA
	g := coordB
	b := (y - r*kR - g*kG) / kB
	if isBounded(b) {
		return num.NewVector3(r, g, b)
	}
	return num.NewVector3(-1.0, -1.0, -1.0)
}

// bisectToSegment finds the segment containing the desired color.
//
// y: The Y value of the color.
// targetHue: The hue of the color.
// returns: A list of two sets of linear RGB coordinates, each corresponding to
// an endpoint
// of the segment containing the desired color.
func bisectToSegment(y float64, targetHue float64) [2]num.Vector3 {
	left := num.NewVector3(-1.0, -1.0, -1.0)
	right := num.NewVector3(-1.0, -1.0, -1.0)
	leftHue := 0.0
	rightHue := 0.0
	initialized := false
	uncut := true
	for n := range 12 {
		mid := nthVertex(y, n)
		if mid[0] < 0 {
			continue
		}
		midHue := hueOf(mid)
		if !initialized {
			left = mid
			right = mid
			leftHue = midHue
			rightHue = midHue
			initialized = true
			continue
		}
		if uncut || areInCyclicOrder(leftHue, midHue, rightHue) {
			uncut = false
			if areInCyclicOrder(leftHue, targetHue, midHue) {
				right = mid
				rightHue = midHue
			} else {
				left = mid
				leftHue = midHue
			}
		}
	}
	return [2]num.Vector3{left, right}
}

func midpoint(a num.Vector3, b num.Vector3) num.Vector3 {
	return num.NewVector3((a[0]+b[0])/2, (a[1]+b[1])/2, (a[2]+b[2])/2)
}

func criticalPlaneBelow(x float64) int {
	return int(math.Floor(x - 0.5))
}

func criticalPlaneAbove(x float64) int {
	return int(math.Ceil(x - 0.5))
}

// bisectToLimit finds a color with the given Y and hue on the boundary of the
// cube.
//
// y: The Y value of the color.
// targetHue: The hue of the color.
// returns: The desired color, in linear RGB coordinates.
func bisectToLimit(y float64, targetHue float64) num.Vector3 {
	segment := bisectToSegment(y, targetHue)
	left := segment[0]
	leftHue := hueOf(left)
	right := segment[1]
	for axis := range 3 {
		if left[axis] != right[axis] {
			var lPlane, rPlane int
			if left[axis] < right[axis] {
				lPlane = criticalPlaneBelow(trueDelinearized(left[axis]))
				rPlane = criticalPlaneAbove(trueDelinearized(right[axis]))
			} else {
				lPlane = criticalPlaneAbove(trueDelinearized(left[axis]))
				rPlane = criticalPlaneBelow(trueDelinearized(right[axis]))
			}

			for range 8 {
				if math.Abs(float64(rPlane-lPlane)) <= 1 {
					break
				}

				mPlane := int(math.Floor(float64(lPlane+rPlane) / 2.0))
				midPlaneCoordinate := CriticalPlanes[mPlane]
				mid := setCoordinate(left, midPlaneCoordinate, right, axis)
				midHue := hueOf(mid)
				if areInCyclicOrder(leftHue, targetHue, midHue) {
					right = mid
					rPlane = mPlane
				} else {
					left = mid
					leftHue = midHue
					lPlane = mPlane
				}
			}
		}
	}
	return midpoint(left, right)
}

func inverseChromaticAdaptation(adapted float64) float64 {
	adaptedAbs := math.Abs(adapted)
	base := math.Max(0, 27.13*adaptedAbs/(400.0-adaptedAbs))
	return num.Sign(adapted) * math.Pow(base, 1.0/0.42)
}

// findResultByJ finds a color with the given hue, chroma, and Y.
//
// hueRadians: The desired hue in radians.
// chroma: The desired chroma.
// y: The desired Y.
// returns: The desired color as a hexadecimal integer, if found; 0 otherwise.
func findResultByJ(hueRadians float64, chroma float64, y float64) ARGB {
	// Initial estimate of j.
	j := math.Sqrt(y) * 11.0

	env := DefaultEnvironment
	tInnerCoeff := 1 / math.Pow(1.64-math.Pow(0.29, env.N), 0.73)
	eHue := 0.25 * (math.Cos(hueRadians+2.0) + 3.8)
	p1 := eHue * (50000.0 / 13.0) * env.Nc * env.Ncb
	hSin := math.Sin(hueRadians)
	hCos := math.Cos(hueRadians)

	for iterationRound := range 5 {
		jNormalized := j / 100.0
		alpha := chroma / math.Sqrt(jNormalized)
		if chroma == 0.0 || j == 0.0 {
			alpha = 0.0
		}

		t := math.Pow(alpha*tInnerCoeff, 1.0/0.9)
		ac := env.Aw * math.Pow(jNormalized, 1.0/env.C/env.Z)
		p2 := ac / env.Nbb
		gamma := (23.0 * (p2 + 0.305) * t) / (23.0*p1 + 11*t*hCos + 108.0*t*hSin)
		a := gamma * hCos
		b := gamma * hSin
		rA := (460.0*p2 + 451.0*a + 288.0*b) / 1403.0
		gA := (460.0*p2 - 891.0*a - 261.0*b) / 1403.0
		bA := (460.0*p2 - 220.0*a - 6300.0*b) / 1403.0
		rCScaled := inverseChromaticAdaptation(rA)
		gCScaled := inverseChromaticAdaptation(gA)
		bCScaled := inverseChromaticAdaptation(bA)

		vec := num.NewVector3(rCScaled, gCScaled, bCScaled)
		linrgb := LinrgbFromScaledDiscount.Mul(vec)

		if linrgb[0] < 0 || linrgb[1] < 0 || linrgb[2] < 0 {
			return 0
		}

		kR, kG, kB := YFromLinRGB.Values()
		fnj := kR*linrgb[0] + kG*linrgb[1] + kB*linrgb[2]
		if fnj <= 0 {
			return 0
		}
		if iterationRound == 4 || math.Abs(fnj-y) < 0.002 {
			if linrgb[0] > 100.01 || linrgb[1] > 100.01 || linrgb[2] > 100.01 {
				return 0
			}
			return ARGBFromLinearRGB(linrgb.Values())
		}
		// Iterates with Newton method,
		// Using 2 * fn(j) / j as the approximation of fn'(j)
		j = j - (fnj-y)*j/(2*fnj)
	}
	return 0
}

// SolveToARGB finds an sRGB color with the given hue, chroma, and L*, if
// possible.
//
// hueDegrees: The desired hue, in degrees.
// chroma: The desired chroma.
// lstar: The desired L*.
//
// returns A hexadecimal representing the sRGB color. The color has sufficiently
// close hue, chroma, and L* to the desired values, if possible; otherwise, the
// hue and L* will be sufficiently close, and chroma will be maximized.
func solveToARGB(hueDegrees float64, chroma float64, lstar float64) ARGB {
	if chroma < 0.0001 || lstar < 0.0001 || lstar > 99.9999 {
		return ARGBFromLstar(lstar)
	}

	hueDegrees = num.NormalizeDegree(hueDegrees)
	hueRadians := num.Radian(hueDegrees)
	y := YFromLstar(lstar)
	exactAnswer := findResultByJ(hueRadians, chroma, y)
	if exactAnswer != 0 {
		return exactAnswer
	}
	linrgb := bisectToLimit(y, hueRadians)
	return ARGBFromLinearRGB(linrgb.Values())
}
