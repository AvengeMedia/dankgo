package dislike

import (
	"testing"

	"github.com/AvengeMedia/dankgo/material/color"
)

func TestDislikeAnalyzer(t *testing.T) {
	t.Run("likes Monk Skin Tone Scale colors", func(t *testing.T) {
		// From https://skintone.google#/get-started
		monkSkinToneScaleColors := []color.ARGB{
			0xfff6ede4,
			0xfff3e7db,
			0xfff7ead0,
			0xffeadaba,
			0xffd7bd96,
			0xffa07e56,
			0xff825c43,
			0xff604134,
			0xff3a312a,
			0xff292420,
		}

		for _, color := range monkSkinToneScaleColors {
			hct := color.ToHct()
			if IsDisliked(hct) {
				t.Errorf(
					"Expected Monk Skin Tone color 0x%x to not be disliked",
					color,
				)
			}
		}
	})

	t.Run("dislikes bile colors", func(t *testing.T) {
		unlikable := []color.ARGB{
			0xff95884B,
			0xff716B40,
			0xffB08E00,
			0xff4C4308,
			0xff464521,
		}

		for _, color := range unlikable {
			hct := color.ToHct()
			if !IsDisliked(hct) {
				t.Errorf("Expected bile color 0x%x to be disliked", color)
			}
		}
	})

	t.Run("makes bile colors likable", func(t *testing.T) {
		unlikable := []color.ARGB{
			0xff95884B,
			0xff716B40,
			0xffB08E00,
			0xff4C4308,
			0xff464521,
		}

		for _, color := range unlikable {
			hct := color.ToHct()
			if !IsDisliked(hct) {
				t.Errorf("Expected bile color 0x%x to be disliked", color)
			}

			likable := FixIfDisliked(hct)
			if IsDisliked(likable) {
				t.Errorf("Expected fixed color to not be disliked")
			}
		}
	})

	t.Run("likes tone 67 colors", func(t *testing.T) {
		color := color.NewHct(100.0, 50.0, 67.0)

		if IsDisliked(color) {
			t.Error("Expected tone 67 color to not be disliked")
		}

		fixed := FixIfDisliked(color)
		if fixed.ToARGB() != color.ToARGB() {
			t.Error(
				"Expected fixIfDisliked to not modify a color that isn't disliked",
			)
		}
	})
}
