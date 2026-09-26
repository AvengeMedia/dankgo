package dynamic

import (
	"math"
	"testing"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/stretchr/testify/require"
)

func TestDynamicScheme(t *testing.T) {
	delta := 0.5 * math.Pow(10, -0.4)

	t.Run("0 length input", func(t *testing.T) {
		hue := GetRotatedHue(color.NewHct(43, 16, 16), []float64{}, []float64{})
		require.InDelta(t, 43, hue, delta)
	})

	t.Run("1 length input no rotation", func(t *testing.T) {
		hue := GetRotatedHue(color.NewHct(43, 16, 16), []float64{0}, []float64{0})
		require.InDelta(t, 43, hue, delta)
	})

	t.Run("input length mismatch asserts", func(t *testing.T) {
		hue := GetRotatedHue(color.NewHct(43, 16, 16), []float64{0}, []float64{0, 1})
		require.InDelta(t, 43, hue, delta)
	})

	t.Run("on boundary rotation correct", func(t *testing.T) {
		hue := GetRotatedHue(
			color.NewHct(43, 16, 16),
			[]float64{0, 42, 360},
			[]float64{0, 15, 0},
		)
		require.InDelta(t, 43+15, hue, delta)
	})

	t.Run("rotation result larger than 360 degrees wraps", func(t *testing.T) {
		hue := GetRotatedHue(
			color.NewHct(43, 16, 16),
			[]float64{0, 42, 360},
			[]float64{0, 480, 0},
		)
		require.InDelta(t, 163, hue, delta)
	})
}
