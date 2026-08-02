package hal

import (
	"fmt"

	"github.com/autopilothub/zerodriver/internal/config"
)

// NewMotor creates a motor driver based on config mode.
func NewMotor(cfg *config.Config) (Motor, error) {
	switch cfg.Mode {
	case "mock":
		return NewMockMotor(), nil
	case "hardware":
		return newHardwareMotor(cfg)
	default:
		return nil, fmt.Errorf("unknown mode: %q", cfg.Mode)
	}
}
