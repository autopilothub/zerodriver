package hal

import (
	"math"
	"testing"
)

func TestComputeTiltFromAccelFlat(t *testing.T) {
	pitch, roll := computeTiltFromAccel(0, 0, 16384)
	if math.Abs(pitch) > 0.1 || math.Abs(roll) > 0.1 {
		t.Fatalf("flat board: pitch=%.2f roll=%.2f", pitch, roll)
	}
}

func TestComputeTiltFromAccelNoInt16Overflow(t *testing.T) {
	// int16 ay*ay+az*az overflows to 0 and produced NaN pitch before fix.
	pitch, roll := computeTiltFromAccel(0, 16384, 16384)
	if math.IsNaN(pitch) || math.IsNaN(roll) {
		t.Fatalf("got NaN pitch=%v roll=%v", pitch, roll)
	}
}

func TestComputeTiltFromAccelUprightX(t *testing.T) {
	pitch, _ := computeTiltFromAccel(-16384, 0, 0)
	if math.Abs(pitch-90) > 1 {
		t.Fatalf("X-up mount: pitch=%.2f want ~90", pitch)
	}
}

func TestComputeMagHeadingNorth(t *testing.T) {
	h := computeMagHeading(20, 0, 40, 0, 0)
	if h < 0 || h > 90 {
		t.Fatalf("heading=%.1f", h)
	}
}
