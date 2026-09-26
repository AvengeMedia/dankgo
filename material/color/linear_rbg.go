// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package color

import "github.com/AvengeMedia/dankgo/material/num"

// LinearRGB represents a color in the linear RGB color space.
//
// Each component (R, G, and B) is expressed as a float64 value in the range [0,
// 100], where 0 corresponds to no intensity and 100 corresponds to full
// intensity. Unlike gamma-corrected RGB, LinearRGB stores values in a linear
// light space, making it suitable for accurate color calculations and
// conversions.
type LinearRGB struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
}

var _ Model = (*LinearRGB)(nil)

// NewLinearRGB creates a linear RGB color model
func NewLinearRGB(r, g, b float64) LinearRGB {
	return LinearRGB{r, g, b}
}

// LinearRGBFromARGB to converts ARGB (sRGB) to linear RGB.
func LinearRGBFromARGB(c ARGB) LinearRGB {
	r, g, b := c.Red(), c.Green(), c.Blue()
	lr, lg, lb := Linearized(r), Linearized(g), Linearized(b)
	return NewLinearRGB(lr, lg, lb)
}

// LinearRGBFromXYZ to converts ARGB (sRGB) to linear RGB.
func LinearRGBFromXYZ(c XYZ) LinearRGB {
	vec := num.NewVector3(c.Values())
	rgb := XYZ_TO_RGB.Mul(vec)
	return NewLinearRGB(rgb.Values())
}

// ToARGB converts LinearRGB to ARGB.
func (c LinearRGB) ToARGB() ARGB {
	return ARGBFromLinearRGB(c.Values())
}

// ToXYZ converts LinearRGB to XYZ color model.
func (c LinearRGB) ToXYZ() XYZ {
	vec := num.NewVector(c)
	xyz := RGB_TO_XYZ.Mul(vec)
	return NewXYZ(xyz.Values())
}

// Hash returns the hash of the LinearRGB color
func (c LinearRGB) Hash() Hash {
	return getHash(c.R, c.G, c.B)
}

// String returns a formatted string representation of linear sRGB color.
func (c LinearRGB) String() string {
	return modelString("sRGB", c)
}

// Values returns R, G, B components of LinearRGB.
func (c LinearRGB) Values() (float64, float64, float64) {
	return c.R, c.G, c.B
}
