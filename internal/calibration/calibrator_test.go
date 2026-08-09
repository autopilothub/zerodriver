package calibration

import (
	"testing"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestHeadingDelta(t *testing.T) {
	if d := headingDelta(350, 10); d != 20 {
		t.Fatalf("wrap: %f", d)
	}
	if d := headingDelta(10, 350); d != -20 {
		t.Fatalf("wrap neg: %f", d)
	}
}

func TestRun_SteeringInvertDetection(t *testing.T) {
	gyro := -10.0
	read := func() (domain.Attitude, error) {
		return domain.Attitude{GyroZ: gyro, HasMag: true, Heading: 90, HasAccel: true, AccelX: 0.5}, nil
	}
	drive := func(steering, throttle float64) error {
		if steering < 0 {
			gyro = 10
		} else if steering > 0 {
			gyro = -10
		}
		return nil
	}

	res, err := Run(drive, read, 0, 1580, Options{
		SteerDuration:    100 * time.Millisecond,
		StraightDuration: 100 * time.Millisecond,
		ThrottleDuration: 100 * time.Millisecond,
		SampleInterval:   20 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.SteeringInvert {
		t.Fatal("expected steering invert")
	}
}
