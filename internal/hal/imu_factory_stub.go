//go:build !linux || !arm

package hal

import (
	"fmt"
	"runtime"

	"github.com/autopilothub/zerodriver/internal/config"
)

func newHardwareIMU(cfg *config.Config) (IMU, error) {
	return nil, fmt.Errorf(
		"hardware IMU requires linux/arm build (current: %s/%s)",
		runtime.GOOS, runtime.GOARCH,
	)
}
