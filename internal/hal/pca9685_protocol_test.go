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

func TestDriveInputs(t *testing.T) {
	s, throttle := DriveInputs(0.4, 0.6)
	if math.Abs(s-0.1) > 0.001 || math.Abs(throttle-0.5) > 0.001 {
		t.Fatalf("got steering=%.2f throttle=%.2f", s, throttle)
	}
}
