package calibration

import (
	"testing"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestRunDeadzone_ForwardAndReverse(t *testing.T) {
	lastPulse := 1500
	read := func() (domain.Attitude, error) {
		var ax float64
		if lastPulse >= 1620 {
			ax = 0.4
		} else if lastPulse <= 1380 {
			ax = -0.4
		}
		return domain.Attitude{HasAccel: true, AccelX: ax}, nil
	}
	setUS := func(us int) error {
		lastPulse = us
		return nil
	}

	res, err := RunDeadzone(setUS, read, DeadzoneConfig{
		NeutralUs: 1500, MaxUs: 2000, ReverseUs: 1000,
	}, DeadzoneOptions{
		StepUs:         20,
		HoldDuration:   10 * time.Millisecond,
		SampleInterval: 5 * time.Millisecond,
		SamplesPerStep: 2,
		MarginUs:       5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ThrottleForwardStartUs != 1625 {
		t.Fatalf("forward start: %d", res.ThrottleForwardStartUs)
	}
	if res.ThrottleReverseStartUs != 1375 {
		t.Fatalf("reverse start: %d", res.ThrottleReverseStartUs)
	}
}

func TestPulseForThrottle_BidirectionalDeadband(t *testing.T) {
	if got := PulseForThrottle(0.1, 1500, 2000, 1000, 1600, 1400); got != 1600 {
		t.Fatalf("forward deadband: %d", got)
	}
	if got := PulseForThrottle(-0.1, 1500, 2000, 1000, 1600, 1400); got != 1400 {
		t.Fatalf("reverse deadband: %d", got)
	}
}
