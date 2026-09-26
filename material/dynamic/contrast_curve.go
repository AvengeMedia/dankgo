// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package dynamic

import "github.com/AvengeMedia/dankgo/material/num"

// ContrastCurve represents a curve that provides contrast values for different
// contrast levels.
type ContrastCurve struct {
	low, normal, medium, high float64
}

// NewContrastCurve returns a new ContrastCurve with the given values for each
// contrast level. The low value is used for contrast level -1.0, normal for
// 0.0, medium for 0.5, and high for 1.0.
func NewContrastCurve(low, normal, medium, high float64) *ContrastCurve {
	return &ContrastCurve{low, normal, medium, high}
}

// Get returns the value at the given contrast level. Contrast level 0.0 is the
// default (normal), -1.0 is the lowest, and 1.0 is the highest. The returned
// value for contrast ratios is between 1.0 and 21.0.
func (c *ContrastCurve) Get(contrast float64) float64 {
	if contrast <= -1.0 {
		return c.low
	} else if contrast < 0.0 {
		return num.Lerp(c.low, c.normal, (contrast-(-1))/1)
	} else if contrast < 0.5 {
		return num.Lerp(c.normal, c.medium, contrast/0.5)
	} else if contrast < 1.0 {
		return num.Lerp(c.medium, c.high, (contrast-0.5)/0.5)
	}
	return c.high
}
