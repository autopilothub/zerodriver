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

func TestThrottleToPulseUs(t *testing.T) {
	if ThrottleToPulseUs(0, 1000, 2000) != 1000 {
		t.Fatal("throttle 0 should be min")
	}
	if ThrottleToPulseUs(1, 1000, 2000) != 2000 {
		t.Fatal("throttle 1 should be max")
	}
	got := ThrottleToPulseUs(0.5, 1000, 2000)
	if math.Abs(float64(got-1500)) > 1 {
		t.Fatalf("throttle 0.5: want 1500, got %d", got)
	}
}
