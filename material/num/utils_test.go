package num

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func rotationDirection(from, to float64) float64 {
	a := to - from
	b := to - from + 360.0
	c := to - from - 360.0
	aAbs := math.Abs(a)
	bAbs := math.Abs(b)
	cAbs := math.Abs(c)
	switch {
	case aAbs <= bAbs && aAbs <= cAbs:
		return signOf(a)
	case bAbs <= aAbs && bAbs <= cAbs:
		return signOf(b)
	default:
		return signOf(c)
	}
}

func signOf(x float64) float64 {
	if x >= 0.0 {
		return 1.0
	}
	return -1.0
}

func TestRotationDirection(t *testing.T) {
	t.Run("is identical to the original implementation", func(t *testing.T) {
		for from := 0.0; from < 360.0; from += 15.0 {
			for to := 7.5; to < 360.0; to += 15.0 {
				expectedAnswer := rotationDirection(from, to)
				actualAnswer := RotationDirection(from, to)
				require.Equal(t, expectedAnswer, actualAnswer)
				require.Equal(t, 1.0, math.Abs(actualAnswer))
			}
		}
	})
}
