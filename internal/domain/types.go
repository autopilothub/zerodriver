package domain

import "time"

// RaceState represents the current operating state of the line tracer.
type RaceState int

const (
	StateIdle RaceState = iota
	StateTracing
	StateAvoiding
	StateStopped
)

func (s RaceState) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateTracing:
		return "TRACING"
	case StateAvoiding:
		return "AVOIDING"
	case StateStopped:
		return "STOPPED"
	default:
		return "UNKNOWN"
	}
}

// LinePosition is the detected line offset normalized to [-1.0, +1.0].
// Negative = line is to the left, positive = to the right.
type LinePosition struct {
	Offset           float64 // centroid at bottom ROI
	LookaheadOffset  float64 // centroid at lookahead row (for Pure Pursuit)
	Detected         bool
	Timestamp        time.Time
}

// Attitude holds IMU orientation data in degrees.
type Attitude struct {
	Yaw      float64
	Pitch    float64
	Roll     float64
	GyroZ    float64 // yaw rate deg/s
	AccelX   float64 // m/s² body frame (BNO085)
	AccelY   float64
	AccelZ   float64
	HasAccel bool
	MagX     float64 // µT (MPU-9250 AK8963 only)
	MagY     float64
	MagZ     float64
	Heading  float64 // compass heading 0-360° (magnetometer)
	HasMag   bool
}

// ObstacleScan holds LiDAR distance readings.
type ObstacleScan struct {
	FrontDistance float64   // nearest obstacle in front (cm), 999 = clear
	Distances     []float64 // full scan distances (cm)
	Angles        []float64 // corresponding angles (degrees)
	Timestamp     time.Time
}

// ControlCommand is the motor output command.
type ControlCommand struct {
	Steering float64 // -1.0 (left) to +1.0 (right)
	Throttle float64 // 0.0 to 1.0
}

// PerceptionSnapshot aggregates all sensor readings at a point in time.
type PerceptionSnapshot struct {
	Line     LinePosition
	Attitude Attitude
	Obstacle ObstacleScan
	State    RaceState
}

// FusedInput is the output of sensor fusion, fed into the controller.
type FusedInput struct {
	LineOffset      float64
	LookaheadOffset float64 // Pure Pursuit lateral error at lookahead
	LineDetected    bool
	HeadingError    float64 // degrees: reference heading − current (9-axis)
	YawRate         float64 // deg/s (gyro Z)
	FrontDistance   float64
	State           RaceState
}
