package hal

import (
	"fmt"

	"github.com/autopilothub/zerodriver/internal/config"
)

// NewLidar creates a LiDAR device based on config mode and model.
func NewLidar(cfg *config.Config) (Lidar, error) {
	if !cfg.LidarEnabled() {
		return NewNoopLidar(), nil
	}
	switch cfg.Mode {
	case "mock":
		return NewMockLidar(), nil
	case "hardware":
		return newHardwareLidar(cfg)
	default:
		return nil, fmt.Errorf("unknown mode: %q", cfg.Mode)
	}
}
