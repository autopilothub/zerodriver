package calibration

import (
	"testing"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestRunDeadzone_Standard(t *testing.T) {
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
		NeutralUs: 1500, ForwardEndUs: 2000, ReverseEndUs: 1000,
	}, DeadzoneOptions{
		StepUs: 20, HoldDuration: 10 * time.Millisecond,
		SampleInterval: 5 * time.Millisecond, SamplesPerStep: 2, MarginUs: 5,
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

func TestRunDeadzone_Inverted(t *testing.T) {
	lastPulse := 1560
	read := func() (domain.Attitude, error) {
		var ax float64
		if lastPulse <= 1450 {
			ax = 0.4
		} else if lastPulse >= 1800 {
			ax = -0.4
		}
		return domain.Attitude{HasAccel: true, AccelX: ax}, nil
	}
	setUS := func(us int) error {
		lastPulse = us
		return nil
	}

	res, err := RunDeadzone(setUS, read, DeadzoneConfig{
		NeutralUs: 1560, ForwardEndUs: 1000, ReverseEndUs: 2000, Inverted: true,
	}, DeadzoneOptions{
		StepUs: 10, HoldDuration: 10 * time.Millisecond,
		SampleInterval: 5 * time.Millisecond, SamplesPerStep: 2, MarginUs: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ThrottleForwardStartUs != 1445 {
		t.Fatalf("forward start: %d", res.ThrottleForwardStartUs)
	}
	if res.ThrottleReverseStartUs != 1805 {
		t.Fatalf("reverse start: %d", res.ThrottleReverseStartUs)
	}
}
