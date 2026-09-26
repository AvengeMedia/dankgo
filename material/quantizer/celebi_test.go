package quantizer

import (
	"testing"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/stretchr/testify/require"
)

const (
	red   color.ARGB = 0xffff0000
	green color.ARGB = 0xff00ff00
	blue  color.ARGB = 0xff0000ff
)

func TestQuantizerCelebi(t *testing.T) {
	t.Run("1R", func(t *testing.T) {
		answer := QuantizeCelebi([]color.ARGB{red}, 128)
		require.Len(t, answer, 1)
		require.Equal(t, 1, answer[red])
	})

	t.Run("1G", func(t *testing.T) {
		answer := QuantizeCelebi([]color.ARGB{green}, 128)
		require.Len(t, answer, 1)
		require.Equal(t, 1, answer[green])
	})

	t.Run("1B", func(t *testing.T) {
		answer := QuantizeCelebi([]color.ARGB{blue}, 128)
		require.Len(t, answer, 1)
		require.Equal(t, 1, answer[blue])
	})

	t.Run("5B", func(t *testing.T) {
		answer := QuantizeCelebi([]color.ARGB{blue, blue, blue, blue, blue}, 128)
		require.Len(t, answer, 1)
		require.Equal(t, 5, answer[blue])
	})

	t.Run("2R 3G", func(t *testing.T) {
		answer := QuantizeCelebi([]color.ARGB{red, red, green, green, green}, 128)
		require.Len(t, answer, 2)
		require.Equal(t, 2, answer[red])
		require.Equal(t, 3, answer[green])
	})

	t.Run("1R 1G 1B", func(t *testing.T) {
		answer := QuantizeCelebi([]color.ARGB{red, green, blue}, 128)
		require.Len(t, answer, 3)
		require.Equal(t, 1, answer[red])
		require.Equal(t, 1, answer[green])
		require.Equal(t, 1, answer[blue])
	})
}
