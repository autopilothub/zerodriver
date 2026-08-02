package hal

import (
	"fmt"

	"github.com/autopilothub/zerodriver/internal/config"
)

// NewIMU creates an IMU device based on config mode.
func NewIMU(cfg *config.Config) (IMU, error) {
	switch cfg.Mode {
	case "mock":
		return NewMockIMU(), nil
	case "hardware":
		return newHardwareIMU(cfg)
	default:
		return nil, fmt.Errorf("unknown mode: %q", cfg.Mode)
	}
}
