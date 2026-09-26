package color

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	red   ARGB = 0xffff0000
	green ARGB = 0xff00ff00
	blue  ARGB = 0xff0000ff
	white ARGB = 0xffffffff
	black ARGB = 0xff000000
)

func TestCamToArgb(t *testing.T) {
	t.Run("red", func(t *testing.T) {
		cam := red.ToCam16()

		require.InDelta(t, 27.408, cam.Hue, 5e-4)
		require.InDelta(t, 113.358, cam.Chroma, 5e-4)
		require.InDelta(t, 46.445, cam.J, 5e-4)
		require.InDelta(t, 89.494, cam.M, 5e-4)
		require.InDelta(t, 91.890, cam.S, 5e-4)
		require.InDelta(t, 105.989, cam.Q, 5e-4)
	})

	t.Run("green", func(t *testing.T) {
		cam := green.ToCam16()

		require.InDelta(t, 142.140, cam.Hue, 5e-4)
		require.InDelta(t, 108.410, cam.Chroma, 5e-4)
		require.InDelta(t, 79.332, cam.J, 5e-4)
		require.InDelta(t, 85.588, cam.M, 5e-4)
		require.InDelta(t, 78.605, cam.S, 5e-4)
		require.InDelta(t, 138.520, cam.Q, 5e-4)
	})

	t.Run("blue", func(t *testing.T) {
		cam := blue.ToCam16()

		require.InDelta(t, 282.788, cam.Hue, 5e-4)
		require.InDelta(t, 87.231, cam.Chroma, 5e-4)
		require.InDelta(t, 25.466, cam.J, 5e-4)
		require.InDelta(t, 68.867, cam.M, 5e-4)
		require.InDelta(t, 93.675, cam.S, 5e-4)
		require.InDelta(t, 78.481, cam.Q, 5e-4)
	})

	t.Run("white", func(t *testing.T) {
		cam := white.ToCam16()

		require.InDelta(t, 209.492, cam.Hue, 5e-4)
		require.InDelta(t, 2.869, cam.Chroma, 5e-4)
		require.InDelta(t, 100.0, cam.J, 5e-4)
		require.InDelta(t, 2.265, cam.M, 5e-4)
		require.InDelta(t, 12.068, cam.S, 5e-4)
		require.InDelta(t, 155.521, cam.Q, 5e-4)
	})

	t.Run("black", func(t *testing.T) {
		cam := black.ToCam16()

		require.InDelta(t, 0.0, cam.Hue, 5e-4)
		require.InDelta(t, 0.0, cam.Chroma, 5e-4)
		require.InDelta(t, 0.0, cam.J, 5e-4)
		require.InDelta(t, 0.0, cam.M, 5e-4)
		require.InDelta(t, 0.0, cam.S, 5e-4)
		require.InDelta(t, 0.0, cam.Q, 5e-4)
	})
}

func TestCamToArgbToCam(t *testing.T) {
	t.Run("red", func(t *testing.T) {
		require.Equal(t, red, red.ToCam16().ToARGB())
	})

	t.Run("green", func(t *testing.T) {
		require.Equal(t, green, green.ToCam16().ToARGB())
	})

	t.Run("blue", func(t *testing.T) {
		require.Equal(t, blue, blue.ToCam16().ToARGB())
	})
}

func TestArgbToHct(t *testing.T) {
	t.Run("green", func(t *testing.T) {
		hct := green.ToHct()
		require.InDelta(t, 142.139, hct.Hue, 5e-3)
		require.InDelta(t, 108.410, hct.Chroma, 5e-3)
		require.InDelta(t, 87.737, hct.Tone, 5e-3)
	})

	t.Run("blue", func(t *testing.T) {
		hct := blue.ToHct()
		require.InDelta(t, 282.788, hct.Hue, 5e-3)
		require.InDelta(t, 87.230, hct.Chroma, 5e-3)
		require.InDelta(t, 32.302, hct.Tone, 5e-3)
	})

	t.Run("blue tone 90", func(t *testing.T) {
		hct := NewHct(282.788, 87.230, 90.0)
		require.InDelta(t, 282.239, hct.Hue, 5e-3)
		require.InDelta(t, 19.144, hct.Chroma, 5e-3)
		require.InDelta(t, 90.035, hct.Tone, 5e-3)
	})
}

func TestViewingConditions(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		vc := DefaultEnvironment
		require.InDelta(t, 0.184, vc.N, 5e-4)
		require.InDelta(t, 29.981, vc.Aw, 5e-4)
		require.InDelta(t, 1.017, vc.Nbb, 5e-4)
		require.InDelta(t, 1.017, vc.Ncb, 5e-4)
		require.InDelta(t, 0.69, vc.C, 5e-4)
		require.InDelta(t, 1.0, vc.Nc, 5e-4)
		require.InDelta(t, 1.021, vc.RgbD[0], 5e-4)
		require.InDelta(t, 0.986, vc.RgbD[1], 5e-4)
		require.InDelta(t, 0.934, vc.RgbD[2], 5e-4)
		require.InDelta(t, 0.388, vc.Fl, 5e-4)
		require.InDelta(t, 0.789, vc.FlRoot, 5e-4)
		require.InDelta(t, 1.909, vc.Z, 5e-4)
	})
}

func colorIsOnBoundary(argb ARGB) bool {
	return argb.Red() == 0 || argb.Red() == 255 ||
		argb.Green() == 0 || argb.Green() == 255 ||
		argb.Blue() == 0 || argb.Blue() == 255
}

func TestCamSolver(t *testing.T) {
	t.Run("returns a sufficiently close color", func(t *testing.T) {
		for hue := 15.0; hue < 360; hue += 30 {
			for chroma := 0.0; chroma <= 100; chroma += 10 {
				for tone := 20.0; tone <= 80; tone += 10 {
					hctColor := NewHct(hue, chroma, tone)

					if chroma > 0 {
						require.LessOrEqual(t, math.Abs(hctColor.Hue-hue), 4.0)
					}

					require.GreaterOrEqual(t, hctColor.Chroma, 0.0)
					require.LessOrEqual(t, hctColor.Chroma, chroma+2.5)

					if hctColor.Chroma < chroma-2.5 {
						require.True(t, colorIsOnBoundary(hctColor.ToARGB()))
					}

					require.LessOrEqual(t, math.Abs(hctColor.Tone-tone), 0.5)
				}
			}
		}
	})
}

func TestHctRoundtrip(t *testing.T) {
	t.Run("preserves original color", func(t *testing.T) {
		for r := 0; r < 296; r += 37 {
			for g := 0; g < 296; g += 37 {
				for b := 0; b < 296; b += 37 {
					argb := ARGBFromRGB(
						uint8(min(255, r)),
						uint8(min(255, g)),
						uint8(min(255, b)),
					)

					hct := argb.ToHct()
					reconstructed := NewHct(hct.Hue, hct.Chroma, hct.Tone).ToARGB()

					require.Equal(t, argb, reconstructed)
				}
			}
		}
	})
}
