package fusion

import (
	"github.com/autopilothub/zerodriver/internal/domain"
)

// Fusion combines sensor data into a single control input.
type Fusion struct {
	imuYawBaseline float64
}

func New() *Fusion {
	return &Fusion{}
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
	}

	switch state {
	case domain.StateAvoiding:
		// LiDAR only during obstacle avoidance
		input.LineDetected = false
		input.LineOffset = 0

	case domain.StateTracing:
		if line.Detected {
			input.LineDetected = true
			input.LineOffset = line.Offset
			// IMU yaw correction: compensate for vehicle drift
			input.YawCorrection = -attitude.GyroZ * 0.01
		} else {
			// Line lost: rely on IMU dead reckoning
			input.LineDetected = false
			input.LineOffset = attitude.Yaw * 0.02
			input.YawCorrection = -attitude.GyroZ * 0.02
		}

	default:
		input.LineDetected = false
		input.LineOffset = 0
	}

	return input
}
