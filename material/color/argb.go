// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package color

import (
	"encoding"
	"errors"
	"fmt"
	"image/color"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/AvengeMedia/dankgo/material/num"
)

// ARGB represents a 32-bit color in ARGB format (Alpha, Red, Green, Blue)
// packed into a uint32. It implements the color.Color interface.
type ARGB uint32

// Offset constants indicate bit offsets of color components in ARGB format
const (
	blueOffset ARGB = iota * 8
	greenOffset
	redOffset
	alphaOffset
)

// NewARGB creates an ARGB color from individual 8-bit alpha, red, green, and
// blue components. The components are packed into a uint32 in ARGB order.
func NewARGB(a, r, g, b uint8) ARGB {
	return ARGB(a)<<alphaOffset |
		ARGB(r)<<redOffset |
		ARGB(g)<<greenOffset |
		ARGB(b)<<blueOffset
}

// ARGBFromInterface converts any color.Color implementation to ARGB. It handles
// the 16-bit to 8-bit conversion automatically.
func ARGBFromInterface(c color.Color) ARGB {
	if argb, ok := c.(ARGB); ok {
		return argb
	}

	r16, g16, b16, a16 := c.RGBA()

	// Convert from [0, 65535] to [0, 255]
	r8 := uint8(r16 >> 8)
	g8 := uint8(g16 >> 8)
	b8 := uint8(b16 >> 8)
	a8 := uint8(a16 >> 8)

	return NewARGB(a8, r8, g8, b8)
}

// ARGBFromLstar converts a CIE L* (lightness) value to an ARGB grayscale color.
// lstar is lightness value in the L*a*b* color space (0-100) Returns an ARGB
// grayscale color matching the specified lightness.
func ARGBFromLstar(lstar float64) ARGB {
	y := YFromLstar(lstar)
	component := Delinearized(y)
	return ARGBFromRGB(component, component, component)
}

// ARGBFromXYZ creates an ARGB color from XYZ color space coordinates. x, y, z
// are the coordinates in the CIE 1931 XYZ color space Returns the corresponding
// ARGB color.
func ARGBFromXYZ(x, y, z float64) ARGB {
	return NewXYZ(x, y, z).ToARGB()
}

// ARGBFromRGB creates an opaque ARGB color (alpha=255) from 8-bit RGB
// components. r, g, b are the red, green, and blue components (0-255) Returns
// the corresponding opaque ARGB color.
func ARGBFromRGB(r, g, b uint8) ARGB {
	return NewARGB(0xFF, r, g, b)
}

// ARGBFromLinearRGB creates an opaque ARGB color from linear RGB components. r,
// g, b are linear RGB components (0-100). Returns the corresponding opaque ARGB
// color after delinearization.
func ARGBFromLinearRGB(r, g, b float64) ARGB {
	dr, dg, db := Delinearized(r), Delinearized(g), Delinearized(b)
	return ARGBFromRGB(dr, dg, db)
}

// ToCam16 converts the ARGB to CAM16 color appearance model.
func (c ARGB) ToCam16() Cam16 {
	return c.ToXYZ().ToCam16()
}

// ToHct converts the ARGB to HCT (Hue-Chroma-Tone) color model.
func (c ARGB) ToHct() Hct {
	cam := c.ToCam16()
	return Hct{cam.Hue, cam.Chroma, c.LStar()}
}

// ToLinearRGB converts the ARGB to Linear RGB color model.
func (c ARGB) ToLinearRGB() LinearRGB {
	return LinearRGBFromARGB(c)
}

// ToXYZ converts the ARGB to CIE XYZ color model.
func (c ARGB) ToXYZ() XYZ {
	r, g, b := c.Red(), c.Green(), c.Blue()

	// Convert RGB channel to linear color (0-1.0)
	vec := num.NewVector3(Linearized(r), Linearized(g), Linearized(b))

	xyz := RGB_TO_XYZ.Mul(vec)
	return NewXYZ(xyz.Values())
}

// ToLab converts the ARGB to CIE L*a*b* color model.
func (c ARGB) ToLab() Lab {
	return c.ToXYZ().ToLab()
}

//revive:disable:function-result-limit

// Components returns the individual 8-bit components of the ARGB color. Returns
// alpha, red, green, blue components in order (0-255).
func (c ARGB) Components() (alpha uint8, red uint8, green uint8, blue uint8) {
	return c.Alpha(), c.Red(), c.Green(), c.Blue()
}

// RGBA implements the color.Color interface. Returns the red, green, blue, and
// alpha values in the 0-65535 range.
func (c ARGB) RGBA() (red uint32, green uint32, blue uint32, alpha uint32) {
	a, r, g, b := c.Components()
	// Convert from 8-bit to 16-bit by scaling: v * 0x101 == v * 257
	const m = 0x101
	return uint32(r) * m, uint32(g) * m, uint32(b) * m, uint32(a) * m
}

//revive:enable:function-result-limit

// LStar calculates the CIE L* (lightness) value of the color. Returns the L*
// value (0-100) representing the perceived lightness.
func (c ARGB) LStar() float64 {
	r, g, b := c.Red(), c.Green(), c.Blue()
	// Convert RGB channel to linear color (0-1.0)
	lr, lg, lb := Linearized(r), Linearized(g), Linearized(b)
	linearRGB := num.NewVector3(lr, lg, lb)

	// Only calculate Y value of XYZ for LStar
	secondRow := RGB_TO_XYZ.Row(1)
	y := secondRow.Dot(linearRGB)
	return LstarFromY(y)
}

// AnsiFg wraps the given text with ANSI escape codes for this foreground color.
// text is the string to be colored Returns the string wrapped with ANSI
// foreground color codes.
func (c ARGB) AnsiFg(text string) string {
	_, r, g, b := c.Components()
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, text)
}

// AnsiBg wraps the given text with ANSI escape codes for this background color.
// text is the string to be colored Returns the string wrapped with ANSI
// background color codes.
func (c ARGB) AnsiBg(text string) string {
	_, r, g, b := c.Components()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm%s\x1b[0m", r, g, b, text)
}

// String returns a string representation of the color hex code. Eg. #FF00FF
func (c ARGB) String() string {
	return c.HexRGB()
}

var (
	_ fmt.Formatter            = (*ARGB)(nil)
	_ encoding.TextMarshaler   = (*ARGB)(nil)
	_ encoding.TextUnmarshaler = (*ARGB)(nil)
)

// Format implements fmt.Formatter.
func (c ARGB) Format(f fmt.State, verb rune) {
	switch verb {
	case 'a': // this for debuging only
		_, r, g, b := c.Components()
		lab := c.ToLab()
		if lab.L < 50 {
			lab.L += 50
		} else {
			lab.L -= 50
		}
		_, fr, fg, fb := lab.ToARGB().Components()
		fmt.Fprintf(f, "\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm%s\x1b[0m", fr, fg, fb, r, g, b, c.String()) //nolint:errcheck
	default:
		io.WriteString(f, c.String()) //nolint:errcheck
	}
}

// MarshalText implements the encoding.TextMarshaler interface. Returns the
// hexadecimal representation of the color (#RRGGBBAA format).
func (c ARGB) MarshalText() ([]byte, error) {
	return []byte(c.HexRGB()), nil
}

// UnmarshalText implements the encoding.UnmarshalText interface. Accepts
// #RRGGBB, #RRGGBBAA, RRGGBB, or RRGGBBAA formats for performance reasons.
// Returns an error if the string cannot be parsed as a valid color.
func (c *ARGB) UnmarshalText(data []byte) error {
	s := string(data)

	// Remove optional leading '#'
	s = strings.TrimPrefix(s, "#")

	if len(s) != 6 && len(s) != 8 {
		return fmt.Errorf(
			"invalid color format: %q (expected RRGGBB or RRGGBBAA)",
			s,
		)
	}

	// Parse RRGGBB or RRGGBBAA directly
	val, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return fmt.Errorf("invalid hex digits in color %q: %w", s, err)
	}

	switch len(s) {
	case 6: // RRGGBB → assume alpha=255
		*c = NewARGB(0xFF,
			uint8(val>>16), // RR
			uint8(val>>8),  // GG
			uint8(val),     // BB
		)
	case 8: // RRGGBBAA
		*c = NewARGB(
			uint8(val),     // AA
			uint8(val>>24), // RR
			uint8(val>>16), // GG
			uint8(val>>8),  // BB
		)
	}

	return nil
}

// Alpha returns the 8-bit alpha component of the color (0-255).
func (c ARGB) Alpha() uint8 {
	return uint8((c >> alphaOffset) & 0xFF)
}

// Red returns the 8-bit red component of the color (0-255).
func (c ARGB) Red() uint8 {
	return uint8((c >> redOffset) & 0xFF)
}

// Green returns the 8-bit green component of the color (0-255).
func (c ARGB) Green() uint8 {
	return uint8((c >> greenOffset) & 0xFF)
}

// Blue returns the 8-bit blue component of the color (0-255).
func (c ARGB) Blue() uint8 {
	return uint8((c >> blueOffset) & 0xFF)
}

// HexRGB returns the hexadecimal representation of the color in #RRGGBB format.
func (c ARGB) HexRGB() string {
	return fmt.Sprintf("#%02X%02X%02X", c.Red(), c.Green(), c.Blue())
}

// HexARGB returns the hexadecimal representation of the color in #AARRGGBB
// format.
func (c ARGB) HexARGB() string {
	return fmt.Sprintf(
		"#%02X%02X%02X%02X",
		c.Alpha(),
		c.Red(),
		c.Green(),
		c.Blue(),
	)
}

// HexRGBA returns the hexadecimal representation of the color in #RRGGBBAA
// format.
func (c ARGB) HexRGBA() string {
	return fmt.Sprintf(
		"#%02X%02X%02X%02X",
		c.Red(),
		c.Green(),
		c.Blue(),
		c.Alpha(),
	)
}

// ARGBFromHexMust parses a hex color string and returns an ARGB color. Panics
// if the string cannot be parsed. Supports formats: #RGB, #RGBA, #RRGGBB,
// #RRGGBBAA.
// hex: The hexadecimal color string to parse
// Returns the parsed ARGB color.
func ARGBFromHexMust(hex string) ARGB {
	c, err := ARGBFromHex(hex)
	if err != nil {
		panic(err)
	}
	return c
}

// ARGBFromHex parses a hex color string and returns an ARGB color. hex is the
// hexadecimal color string to parse. Returns the parsed ARGB color and error if
// the RGB hex is invalid.
//
// Supports formats: #RGB, #RGBA, #RRGGBB, #RRGGBBAA.
//
// TODO: optimize me...
func ARGBFromHex(hex string) (ARGB, error) {
	hex = strings.TrimPrefix(hex, "#")

	// Regex check if input is valid or not
	hexColorRegex := regexp.MustCompile(
		`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`,
	)
	if !hexColorRegex.MatchString(hex) {
		return 0, errors.New("invalid hex color format")
	}

	// Expand shorthand formats to full 6/8-char form
	switch len(hex) {
	case 3: // #RGB → #RRGGBB
		hex = fmt.Sprintf("%c%c%c%c%c%c",
			hex[0], hex[0],
			hex[1], hex[1],
			hex[2], hex[2],
		)
	case 4: // #RGBA → #RRGGBBAA
		hex = fmt.Sprintf("%c%c%c%c%c%c%c%c",
			hex[0], hex[0],
			hex[1], hex[1],
			hex[2], hex[2],
			hex[3], hex[3],
		)
	}

	var r, g, b, a uint8 = 0, 0, 0, 0xFF

	switch len(hex) {
	case 6:
		val, err := strconv.ParseUint(hex, 16, 24)
		if err != nil {
			return 0, fmt.Errorf("invalid hex color: %w", err)
		}
		r = uint8(val >> 16)
		g = uint8((val >> 8) & 0xFF)
		b = uint8(val & 0xFF)

	case 8:
		val, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid hex color: %w", err)
		}
		r = uint8(val >> 24)
		g = uint8((val >> 16) & 0xFF)
		b = uint8((val >> 8) & 0xFF)
		a = uint8(val & 0xFF)
	}

	return NewARGB(a, r, g, b), nil
}
