package hal

import "github.com/autopilothub/zerodriver/internal/domain"

// IMU provides orientation data from MPU-9250.
type IMU interface {
	Read() (domain.Attitude, error)
	Close() error
}

// Lidar provides distance scan data.
type Lidar interface {
	Scan() (domain.ObstacleScan, error)
	Close() error
}

// Camera captures image frames.
type Camera interface {
	Capture() ([]byte, error)
	Width() int
	Height() int
	Close() error
}

// Motor controls left/right wheel speeds.
type Motor interface {
	Set(left, right float64) error
	Stop() error
	Close() error
}
