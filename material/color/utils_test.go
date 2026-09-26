package color

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArgbFromRgb(t *testing.T) {
	t.Run("returns correct value for black", func(t *testing.T) {
		require.Equal(t, ARGB(0xffffffff), ARGBFromRGB(255, 255, 255))
		require.Equal(t, ARGB(4294967295), ARGBFromRGB(255, 255, 255))
	})

	t.Run("returns correct value for white", func(t *testing.T) {
		require.Equal(t, ARGB(0xff000000), ARGBFromRGB(0, 0, 0))
		require.Equal(t, ARGB(4278190080), ARGBFromRGB(0, 0, 0))
	})

	t.Run("returns correct value for random color", func(t *testing.T) {
		require.Equal(t, ARGB(0xff3296fa), ARGBFromRGB(50, 150, 250))
		require.Equal(t, ARGB(4281505530), ARGBFromRGB(50, 150, 250))
	})
}

func TestYFromLstar(t *testing.T) {
	t.Run("satisfies given values", func(t *testing.T) {
		cases := []struct{ lstar, y float64 }{
			{0.0, 0.0},
			{0.1, 0.0110705},
			{0.2, 0.0221411},
			{0.3, 0.0332116},
			{0.4, 0.0442822},
			{0.5, 0.0553528},
			{1.0, 0.1107056},
			{2.0, 0.2214112},
			{3.0, 0.3321169},
			{4.0, 0.4428225},
			{5.0, 0.5535282},
			{8.0, 0.8856451},
			{10.0, 1.1260199},
			{15.0, 1.9085832},
			{20.0, 2.9890524},
			{25.0, 4.4154767},
			{30.0, 6.2359055},
			{40.0, 11.2509737},
			{50.0, 18.4186518},
			{60.0, 28.1233342},
			{70.0, 40.7494157},
			{80.0, 56.6812907},
			{90.0, 76.3033539},
			{95.0, 87.6183294},
			{99.0, 97.4360239},
			{100.0, 100.0},
		}
		for _, c := range cases {
			require.InDelta(t, c.y, YFromLstar(c.lstar), 5e-6, "lstar %v", c.lstar)
		}
	})

	t.Run("is inverse of lstarFromY", func(t *testing.T) {
		for y := 0.0; y <= 100.0; y += 0.1 {
			lstar := LstarFromY(y)
			reconstructedY := YFromLstar(lstar)
			require.InDelta(t, y, reconstructedY, 5e-9)
		}
	})
}

func TestLstarFromY(t *testing.T) {
	t.Run("satisfies given values", func(t *testing.T) {
		cases := []struct{ y, lstar float64 }{
			{0.0, 0.0},
			{0.1, 0.9032962},
			{0.2, 1.8065925},
			{0.3, 2.7098888},
			{0.4, 3.6131851},
			{0.5, 4.5164814},
			{0.8856451, 8.0},
			{1.0, 8.9914424},
			{2.0, 15.4872443},
			{3.0, 20.0438970},
			{4.0, 23.6714419},
			{5.0, 26.7347653},
			{10.0, 37.8424304},
			{15.0, 45.6341970},
			{20.0, 51.8372115},
			{25.0, 57.0754208},
			{30.0, 61.6542222},
			{40.0, 69.4695307},
			{50.0, 76.0692610},
			{60.0, 81.8381891},
			{70.0, 86.9968642},
			{80.0, 91.6848609},
			{90.0, 95.9967686},
			{95.0, 98.0335184},
			{99.0, 99.6120372},
			{100.0, 100.0},
		}
		for _, c := range cases {
			require.InDelta(t, c.lstar, LstarFromY(c.y), 5e-6, "y %v", c.y)
		}
	})

	t.Run("is inverse of yFromLstar", func(t *testing.T) {
		for lstar := 0.0; lstar <= 100.0; lstar += 0.1 {
			y := YFromLstar(lstar)
			reconstructedLstar := LstarFromY(y)
			require.InDelta(t, lstar, reconstructedLstar, 5e-9)
		}
	})
}
