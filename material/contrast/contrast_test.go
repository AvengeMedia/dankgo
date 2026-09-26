package contrast

import (
	"math"
	"testing"
)

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestRatioOfTones_OutOfBoundsInput(t *testing.T) {
	got := RatioOfTones(-10.0, 110.0)
	want := 21.0
	if !almostEqual(got, want, 0.001) {
		t.Errorf("RatioOfTones(-10.0, 110.0) = %v, want %v", got, want)
	}
}

func TestLighter_ImpossibleRatioErrors(t *testing.T) {
	got := Lighter(90.0, 10.0)
	want := -1.0
	if !almostEqual(got, want, 0.001) {
		t.Errorf("Lighter(90.0, 10.0) = %v, want %v", got, want)
	}
}

func TestLighter_OutOfBoundsInputAboveErrors(t *testing.T) {
	got := Lighter(110.0, 2.0)
	want := -1.0
	if !almostEqual(got, want, 0.001) {
		t.Errorf("Lighter(110.0, 2.0) = %v, want %v", got, want)
	}
}

func TestLighter_OutOfBoundsInputBelowErrors(t *testing.T) {
	got := Lighter(-10.0, 2.0)
	want := -1.0
	if !almostEqual(got, want, 0.001) {
		t.Errorf("Lighter(-10.0, 2.0) = %v, want %v", got, want)
	}
}

func TestLighterUnsafe_ReturnsMaxTone(t *testing.T) {
	got := LighterUnsafe(100.0, 2.0)
	want := 100.0
	if !almostEqual(got, want, 0.001) {
		t.Errorf("LighterUnsafe(100.0, 2.0) = %v, want %v", got, want)
	}
}

func TestDarker_ImpossibleRatioErrors(t *testing.T) {
	got := Darker(10.0, 20.0)
	want := -1.0
	if !almostEqual(got, want, 0.001) {
		t.Errorf("Darker(10.0, 20.0) = %v, want %v", got, want)
	}
}

func TestDarker_OutOfBoundsInputAboveErrors(t *testing.T) {
	got := Darker(110.0, 2.0)
	want := -1.0
	if !almostEqual(got, want, 0.001) {
		t.Errorf("Darker(110.0, 2.0) = %v, want %v", got, want)
	}
}

func TestDarker_OutOfBoundsInputBelowErrors(t *testing.T) {
	got := Darker(-10.0, 2.0)
	want := -1.0
	if !almostEqual(got, want, 0.001) {
		t.Errorf("Darker(-10.0, 2.0) = %v, want %v", got, want)
	}
}

func TestDarkerUnsafe_ReturnsMinTone(t *testing.T) {
	got := DarkerUnsafe(0.0, 2.0)
	want := 0.0
	if !almostEqual(got, want, 0.001) {
		t.Errorf("DarkerUnsafe(0.0, 2.0) = %v, want %v", got, want)
	}
}
