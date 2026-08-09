package fusion

import (
	"github.com/autopilothub/zerodriver/internal/domain"
)

// Fusion combines camera Pure Pursuit targets with 9-axis IMU heading.
type Fusion struct {
	referenceHeading float64
	hasReference     bool
	headingLockTol   float64
}

func New() *Fusion {
	return &Fusion{headingLockTol: 0.15}
}

// NewWithHeadingLock creates fusion with a custom |line offset| threshold to lock heading.
func NewWithHeadingLock(tol float64) *Fusion {
	if tol <= 0 {
		tol = 0.15
	}
	return &Fusion{headingLockTol: tol}
}

func headingDeg(att domain.Attitude) float64 {
	if att.HasMag {
		return att.Heading
	}
	h := att.Yaw
	for h < 0 {
		h += 360
	}
	for h >= 360 {
		h -= 360
	}
	return h
}

func headingErrorDeg(reference, current float64) float64 {
	d := reference - current
	for d > 180 {
		d -= 360
	}
	for d < -180 {
		d += 360
	}
	return d
}

// Fuse merges line, IMU, and LiDAR data based on current state.
func (f *Fusion) Fuse(
	line domain.LinePosition,
	attitude domain.Attitude,
	obstacle domain.ObstacleScan,
	state domain.RaceState,
) domain.FusedInput {
	input := domain.FusedInput{
		FrontDistance: obstacle.FrontDistance,
		State:         state,
		YawRate:       attitude.GyroZ,
	}

	heading := headingDeg(attitude)

	switch state {
	case domain.StateAvoiding:
		input.LineDetected = false

	case domain.StateTracing:
		if line.Detected {
			input.LineDetected = true
			input.LineOffset = line.Offset
			input.LookaheadOffset = line.LookaheadOffset
			if input.LookaheadOffset == 0 {
				input.LookaheadOffset = line.Offset
			}

			if mathAbs(line.Offset) < f.headingLockTol {
				f.referenceHeading = heading
				f.hasReference = true
			}
			if f.hasReference {
				input.HeadingError = headingErrorDeg(f.referenceHeading, heading)
			}
		} else {
			input.LineDetected = false
			if f.hasReference {
				input.HeadingError = headingErrorDeg(f.referenceHeading, heading)
			}
		}

	default:
		input.LineDetected = false
	}

	return input
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
