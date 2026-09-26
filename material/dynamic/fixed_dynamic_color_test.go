package dynamic

import (
	"testing"

	"github.com/AvengeMedia/dankgo/material/color"
	"github.com/stretchr/testify/require"
)

func TestFixedColors(t *testing.T) {
	red := color.ARGB(0xFFFF0000).ToHct()

	t.Run("fixed colors in non-monochrome schemes", func(t *testing.T) {
		scheme := NewDynamicScheme(red, VariantTonalSpot, 0.0, true, PlatformPhone, Version2021)
		m := scheme.MaterialColor

		require.InDelta(t, 90.0, m.PrimaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 80.0, m.PrimaryFixedDim().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 10.0, m.OnPrimaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 30.0, m.OnPrimaryFixedVariant().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 90.0, m.SecondaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 80.0, m.SecondaryFixedDim().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 10.0, m.OnSecondaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 30.0, m.OnSecondaryFixedVariant().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 90.0, m.TertiaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 80.0, m.TertiaryFixedDim().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 10.0, m.OnTertiaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 30.0, m.OnTertiaryFixedVariant().GetHct(scheme).Tone, 0.5)
	})

	t.Run("fixed ARGB colors in non-monochrome schemes", func(t *testing.T) {
		scheme := NewDynamicScheme(red, VariantTonalSpot, 0.0, true, PlatformPhone, Version2021)
		m := scheme.MaterialColor

		require.Equal(t, color.ARGB(0xFFFFDAD4), m.PrimaryFixed().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFFFFB4A8), m.PrimaryFixedDim().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFF3A0905), m.OnPrimaryFixed().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFF73342A), m.OnPrimaryFixedVariant().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFFFFDAD4), m.SecondaryFixed().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFFE7BDB6), m.SecondaryFixedDim().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFF2C1512), m.OnSecondaryFixed().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFF5D3F3B), m.OnSecondaryFixedVariant().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFFFBDFA6), m.TertiaryFixed().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFFDEC48C), m.TertiaryFixedDim().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFF251A00), m.OnTertiaryFixed().GetArgb(scheme))
		require.Equal(t, color.ARGB(0xFF564419), m.OnTertiaryFixedVariant().GetArgb(scheme))
	})

	t.Run("fixed colors in light monochrome schemes", func(t *testing.T) {
		scheme := NewDynamicScheme(red, VariantMonochrome, 0.0, false, PlatformPhone, Version2021)
		m := scheme.MaterialColor

		require.InDelta(t, 40.0, m.PrimaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 30.0, m.PrimaryFixedDim().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 100.0, m.OnPrimaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 90.0, m.OnPrimaryFixedVariant().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 80.0, m.SecondaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 70.0, m.SecondaryFixedDim().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 10.0, m.OnSecondaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 25.0, m.OnSecondaryFixedVariant().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 40.0, m.TertiaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 30.0, m.TertiaryFixedDim().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 100.0, m.OnTertiaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 90.0, m.OnTertiaryFixedVariant().GetHct(scheme).Tone, 0.5)
	})

	t.Run("fixed colors in dark monochrome schemes", func(t *testing.T) {
		scheme := NewDynamicScheme(red, VariantMonochrome, 0.0, true, PlatformPhone, Version2021)
		m := scheme.MaterialColor

		require.InDelta(t, 40.0, m.PrimaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 30.0, m.PrimaryFixedDim().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 100.0, m.OnPrimaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 90.0, m.OnPrimaryFixedVariant().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 80.0, m.SecondaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 70.0, m.SecondaryFixedDim().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 10.0, m.OnSecondaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 25.0, m.OnSecondaryFixedVariant().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 40.0, m.TertiaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 30.0, m.TertiaryFixedDim().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 100.0, m.OnTertiaryFixed().GetHct(scheme).Tone, 0.5)
		require.InDelta(t, 90.0, m.OnTertiaryFixedVariant().GetHct(scheme).Tone, 0.5)
	})
}
