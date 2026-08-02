package hal

import (
	"fmt"

	"github.com/autopilothub/zerodriver/internal/config"
)

// NewCamera creates a camera device based on config mode.
func NewCamera(cfg *config.Config) (Camera, error) {
	switch cfg.Mode {
	case "mock":
		return NewMockCamera(cfg.Camera.Width, cfg.Camera.Height), nil
	case "hardware":
		return newHardwareCamera(cfg)
	default:
		return nil, fmt.Errorf("unknown mode: %q", cfg.Mode)
	}
}
