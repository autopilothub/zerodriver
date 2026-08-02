//go:build !linux || !arm

package hal

import (
	"fmt"
	"runtime"

	"github.com/autopilothub/zerodriver/internal/config"
)

func newHardwareLidar(cfg *config.Config) (Lidar, error) {
	return nil, fmt.Errorf(
		"hardware lidar requires linux/arm build (current: %s/%s)",
		runtime.GOOS, runtime.GOARCH,
	)
}
