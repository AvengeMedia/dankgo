package temperature

import (
	"math"
	"reflect"
	"testing"

	"github.com/AvengeMedia/dankgo/material/color"
)

func TestRawTemperature(t *testing.T) {
	testCases := []struct {
		name     string
		color    color.ARGB
		expected float64
		delta    float64
	}{
		{
			name:     "blue",
			color:    0xff0000ff,
			expected: -1.393,
			delta:    0.001,
		},
		{
			name:     "red",
			color:    0xffff0000,
			expected: 2.351,
			delta:    0.001,
		},
		{
			name:     "green",
			color:    0xff00ff00,
			expected: -0.267,
			delta:    0.001,
		},
		{
			name:     "white",
			color:    0xffffffff,
			expected: -0.5,
			delta:    0.001,
		},
		{
			name:     "black",
			color:    0xff000000,
			expected: -0.5,
			delta:    0.001,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hctColor := tc.color.ToHct()
			temp := RawTemperature(hctColor)
			if diff := math.Abs(tc.expected - temp); diff > 0.001 {
				t.Errorf("RawTemperature(%v) diff: %f", tc.color.String(), diff)
			}
		})
	}
}

func TestRelativeTemperature(t *testing.T) {
	testCases := []struct {
		name     string
		color    color.ARGB
		expected float64
		delta    float64
	}{
		{
			name:     "blue",
			color:    0xff0000ff,
			expected: 0.0,
			delta:    0.001,
		},
		{
			name:     "red",
			color:    0xffff0000,
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "green",
			color:    0xff00ff00,
			expected: 0.467,
			delta:    0.001,
		},
		{
			name:     "white",
			color:    0xffffffff,
			expected: 0.5,
			delta:    0.001,
		},
		{
			name:     "black",
			color:    0xff000000,
			expected: 0.5,
			delta:    0.001,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewCache(tc.color.ToHct())
			temp := cache.InputRelativeTemperature()
			if diff := math.Abs(tc.expected - temp); diff > 0.001 {
				t.Errorf(
					"RelativeTemperature(%v) diff: %v",
					tc.color.String(),
					diff,
				)
			}
		})
	}
}

func TestComplement(t *testing.T) {
	testCases := []struct {
		name     string
		color    color.ARGB
		expected color.ARGB
	}{
		{
			name:     "blue",
			color:    0xff0000ff,
			expected: 0xff9d0002,
		},
		{
			name:     "red",
			color:    0xffff0000,
			expected: 0xff007bfc,
		},
		{
			name:     "green",
			color:    0xff00ff00,
			expected: 0xffffd2c9,
		},
		{
			name:     "white",
			color:    0xffffffff,
			expected: 0xffffffff,
		},
		{
			name:     "black",
			color:    0xff000000,
			expected: 0xff000000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hctColor := tc.color.ToHct()
			cache := NewCache(hctColor)
			complement := cache.Complement().ToARGB()
			if tc.expected != complement {
				t.Errorf(
					"Complement(%v) = %s, want %s",
					tc.color,
					complement.String(),
					tc.expected.String(),
				)
			}
		})
	}
}

func TestAnalogous(t *testing.T) {
	testCases := []struct {
		name     string
		color    color.ARGB
		expected []color.ARGB
	}{
		{
			name:  "blue",
			color: 0xff0000ff,
			expected: []color.ARGB{
				0xff00590c,
				0xff00564e,
				0xff0000ff,
				0xff6700cc,
				0xff81009f,
			},
		},
		{
			name:  "red",
			color: 0xffff0000,
			expected: []color.ARGB{
				0xfff60082,
				0xfffc004c,
				0xffff0000,
				0xffd95500,
				0xffaf7200,
			},
		},
		{
			name:  "green",
			color: 0xff00ff00,
			expected: []color.ARGB{
				0xffcee900,
				0xff92f500,
				0xff00ff00,
				0xff00fd6f,
				0xff00fab3,
			},
		},
		{
			name:  "black",
			color: 0xff000000,
			expected: []color.ARGB{
				0xff000000,
				0xff000000,
				0xff000000,
				0xff000000,
				0xff000000,
			},
		},
		{
			name:  "white",
			color: 0xffffffff,
			expected: []color.ARGB{
				0xffffffff,
				0xffffffff,
				0xffffffff,
				0xffffffff,
				0xffffffff,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cache := NewCache(tc.color.ToHct())
			analogous := cache.Analogous(0, 0)

			result := make([]color.ARGB, len(analogous))
			for i, color := range analogous {
				result[i] = color.ToARGB()
			}

			if !reflect.DeepEqual(tc.expected, result) {
				t.Errorf("Analogous(%s) = %v", tc.color.String(), result)
			}
		})
	}
}
