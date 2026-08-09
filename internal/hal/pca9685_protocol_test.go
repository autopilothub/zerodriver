package hal

import (
	"math"
	"testing"
)

func TestPrescale(t *testing.T) {
	if Prescale(50) != 121 {
		t.Fatalf("expected prescale 121 for 50Hz, got %d", Prescale(50))
	}
}

func TestSteeringToPulseUs(t *testing.T) {
	cases := []struct {
		steer float64
		want  int
	}{
		{-1, 1000},
		{0, 1500},
		{1, 2000},
	}
	for _, c := range cases {
		got := SteeringToPulseUs(c.steer, 1000, 1500, 2000)
		if got != c.want {
			t.Fatalf("steer %.1f: want %d, got %d", c.steer, c.want, got)
		}
	}
}

func TestApplySteeringTrimInvert(t *testing.T) {
	if got := ApplySteering(0, 1000, 1500, 2000, 80, false); got != 1580 {
		t.Fatalf("trim: got %d", got)
	}
	if got := ApplySteering(-1, 1000, 1500, 2000, 0, true); got != 2000 {
		t.Fatalf("invert: got %d", got)
	}
}

func TestThrottleToPulseUs(t *testing.T) {
	if ThrottleToPulseUs(0, 1500, 2000) != 1500 {
		t.Fatal("throttle 0 should be neutral")
	}
	if ThrottleToPulseUs(1, 1500, 2000) != 2000 {
		t.Fatal("throttle 1 should be max")
	}
	got := ThrottleToPulseUs(0.5, 1500, 2000)
	if math.Abs(float64(got-1750)) > 1 {
		t.Fatalf("throttle 0.5: want 1750, got %d", got)
	}
}
