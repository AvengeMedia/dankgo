package palettes

import (
	"math"
	"testing"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/stretchr/testify/require"
)

func TestTonalPalette(t *testing.T) {
	t.Run("ofBlue", func(t *testing.T) {
		blue := NewFromARGB(0xff0000ff)

		require.Equal(t, color.ARGB(0xffffffff), blue.Tone(100))
		require.Equal(t, color.ARGB(0xfff1efff), blue.Tone(95))
		require.Equal(t, color.ARGB(0xffe0e0ff), blue.Tone(90))
		require.Equal(t, color.ARGB(0xffbec2ff), blue.Tone(80))
		require.Equal(t, color.ARGB(0xff9da3ff), blue.Tone(70))
		require.Equal(t, color.ARGB(0xff7c84ff), blue.Tone(60))
		require.Equal(t, color.ARGB(0xff5a64ff), blue.Tone(50))
		require.Equal(t, color.ARGB(0xff343dff), blue.Tone(40))
		require.Equal(t, color.ARGB(0xff0000ef), blue.Tone(30))
		require.Equal(t, color.ARGB(0xff0001ac), blue.Tone(20))
		require.Equal(t, color.ARGB(0xff00006e), blue.Tone(10))
		require.Equal(t, color.ARGB(0xff000000), blue.Tone(0))
	})
}

func TestKeyColor(t *testing.T) {
	t.Run("Key color with exact chroma", func(t *testing.T) {
		palette := FromHueAndChroma(50.0, 60.0)
		result := palette.KeyColor

		require.Less(t, math.Abs(result.Hue-50.0), 10.0)
		require.Less(t, math.Abs(result.Chroma-60.0), 0.5)
		require.Greater(t, result.Tone, 0.0)
		require.Less(t, result.Tone, 100.0)
	})

	t.Run("key color with unusually high chroma", func(t *testing.T) {
		palette := FromHueAndChroma(149.0, 200.0)
		result := palette.KeyColor

		require.Less(t, math.Abs(result.Hue-149.0), 10.0)
		require.Greater(t, result.Chroma, 89.0)
		require.Greater(t, result.Tone, 0.0)
		require.Less(t, result.Tone, 100.0)
	})

	t.Run("key color with unusually low chroma", func(t *testing.T) {
		palette := FromHueAndChroma(50.0, 3.0)
		result := palette.KeyColor

		require.Less(t, math.Abs(result.Hue-50.0), 10.0)
		require.Less(t, math.Abs(result.Chroma-3.0), 0.5)
		require.Less(t, math.Abs(result.Tone-50.0), 0.5)
	})
}
