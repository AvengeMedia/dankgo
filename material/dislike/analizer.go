// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package dislike

import "github.com/AvengeMedia/dankgo/material/color"

// IsDisliked reports whether the color is considered visually unpleasant.
// Disliked colors are dark yellow-greens with sufficient saturation.
func IsDisliked(hct color.Hct) bool {
	huePasses := int(hct.Hue+0.5) >= 90 && int(hct.Hue+0.5) <= 111
	chromaPasses := int(hct.Chroma+0.5) > 16
	tonePasses := int(hct.Tone+0.5) < 65
	return huePasses && chromaPasses && tonePasses
}

// FixIfDisliked returns a lighter version of the color if it is disliked. If
// the color is disliked, it returns a new color with the same hue and chroma
// but with tone 70. Otherwise, it returns the original color unchanged.
func FixIfDisliked(hct color.Hct) color.Hct {
	if IsDisliked(hct) {
		return color.NewHct(hct.Hue, hct.Chroma, 70.0)
	}
	return hct
}
