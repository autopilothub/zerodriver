package hal

import (
	"fmt"

	"github.com/autopilothub/zerodriver/internal/config"
)

// Devices holds all HAL device instances.
type Devices struct {
	IMU    IMU
	Lidar  Lidar
	Camera Camera
	Motor  Motor
}

// NewDevices creates HAL devices based on the configured mode.
func NewDevices(cfg *config.Config) (*Devices, error) {
	switch cfg.Mode {
	case "mock":
		return newMockDevices(cfg)
	case "hardware":
		return newHardwareDevices(cfg)
	default:
		return nil, fmt.Errorf("unknown mode: %q (use mock or hardware)", cfg.Mode)
	}
}

func newMockDevices(cfg *config.Config) (*Devices, error) {
	return &Devices{
		IMU:    NewMockIMU(),
		Lidar:  NewMockLidar(),
		Camera: NewMockCamera(cfg.Camera.Width, cfg.Camera.Height),
		Motor:  NewMockMotor(),
	}, nil
}
