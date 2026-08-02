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

// Motor controls steering servo and throttle (RC car).
type Motor interface {
	Drive(steering, throttle float64) error
	Stop() error
	Close() error
}
