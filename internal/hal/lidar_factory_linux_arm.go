//go:build linux && arm

package hal

import (
	"fmt"

	"github.com/autopilothub/zerodriver/internal/config"
)

func newHardwareLidar(cfg *config.Config) (Lidar, error) {
	switch cfg.Hardware.LidarModel {
	case "rplidar_a1", "":
		return NewRPLidar(
			cfg.Hardware.LidarPort,
			cfg.Hardware.LidarBaud,
			cfg.Obstacle.ScanFrontAngle,
		)
	default:
		return nil, fmt.Errorf("unsupported lidar model: %q", cfg.Hardware.LidarModel)
	}
}
