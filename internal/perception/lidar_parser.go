package perception

import (
	"github.com/autopilothub/zerodriver/internal/domain"
	"github.com/autopilothub/zerodriver/internal/hal"
)

// LidarParser wraps the HAL Lidar interface.
type LidarParser struct {
	lidar hal.Lidar
}

func NewLidarParser(lidar hal.Lidar) *LidarParser {
	return &LidarParser{lidar: lidar}
}

func (p *LidarParser) Scan() (domain.ObstacleScan, error) {
	return p.lidar.Scan()
}
