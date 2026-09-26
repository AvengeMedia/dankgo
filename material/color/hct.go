package color

// Hct represents a color in the HCT (Hue, Chroma, Tone) color space.
//
// HCT is a perceptually accurate color measurement system designed to render
// colors consistently across different lighting environments.
type Hct struct {
	// Hue is the color angle in degrees [0, 360], where 0° is red, 120° is
	// green, and 240° is blue. Values outside this range are normalized.
	Hue float64 `json:"hue"`
	// Chroma is the colorfulness of the color [0, ~150]. Higher values
	// represent more saturated colors. The maximum chroma varies depending on
	// hue and tone.
	Chroma float64 `json:"chroma"`
	// Tone is the lightness or luminance of the color [0, 100], where 0 is
	// black, 50 is mid-tone, and 100 is white. Invalid values are clamped to
	// this range.
	Tone float64 `json:"tone"`
}

var _ Model = (*Hct)(nil)

// NewHct creates an HCT color from hue, chroma, and tone values. Hue values
// outside [0, 360] are normalized. Chroma and tone values outside their valid
// ranges are adjusted accordingly.
func NewHct(hue, chroma, tone float64) Hct {
	return solveToARGB(hue, chroma, tone).ToHct()
}

// ToARGB converts HCT color to ARGB representation. Returns ARGB - 32-bit
// packed color value.
func (h Hct) ToARGB() ARGB {
	return solveToARGB(h.Hue, h.Chroma, h.Tone)
}

// ToXYZ converts HCT to CIE XYZ color model.
func (h Hct) ToXYZ() XYZ {
	return h.ToARGB().ToXYZ()
}

// ToLab converts HCT to CIE L*a*b* color model.
func (h Hct) ToLab() Lab {
	return h.ToARGB().ToXYZ().ToLab()
}

// ToCam16 converts HCT to CAM16 color appearance model.
func (h Hct) ToCam16() Cam16 {
	return h.ToARGB().ToCam16()
}

// String returns a formatted string representation of HCT color.
func (h Hct) String() string {
	return modelString("HCT", h)
}

// Values returns Hue, Chroma and Tone
func (h Hct) Values() (float64, float64, float64) {
	return h.Hue, h.Chroma, h.Tone
}

// Hash returns the hash of the HCT color
func (h Hct) Hash() Hash {
	return getHash(h.Hue, h.Chroma, h.Tone)
}

// IsBlue reports whether hue falls in the blue range [250, 270].
func IsBlue(hue float64) bool {
	return hue >= 250 && hue < 270
}

// IsYellow reports whether hue falls in the yellow range [105, 125].
func IsYellow(hue float64) bool {
	return hue >= 105 && hue < 125
}

// IsCyan reports whether hue falls in the cyan range [170, 207].
func IsCyan(hue float64) bool {
	return hue >= 170 && hue < 207
}

// IsBlue reports whether the color's hue falls in the blue range [250, 270].
func (h Hct) IsBlue() bool {
	return IsBlue(h.Hue)
}

// IsYellow reports whether the color's hue falls in the yellow range [105,
// 125].
func (h Hct) IsYellow() bool {
	return IsYellow(h.Hue)
}

// IsCyan reports whether the color's hue falls in the cyan range [170, 207].
func (h Hct) IsCyan() bool {
	return IsCyan(h.Hue)
}

// InViewingConditions adjusts color appearance for different environments. Uses
// CAM16 color appearance model to account for viewing conditions. env is
// environment containing viewing condition parameters. Returns Hct color
// adjusted for specified viewing conditions.
func (h Hct) InViewingConditions(env Environment) Hct {
	cam := h.ToARGB().ToCam16()
	viewedInEnv := cam.Viewed(env)
	newCam := viewedInEnv.ToCam16()
	return newCam.ToHct()
}
