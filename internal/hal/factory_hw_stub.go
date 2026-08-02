//go:build !linux || !arm

package hal

import (
	"fmt"
	"runtime"

	"github.com/autopilothub/zerodriver/internal/config"
)

func newHardwareDevices(cfg *config.Config) (*Devices, error) {
	return nil, fmt.Errorf(
		"hardware mode requires linux/arm build (current: %s/%s); use -mode mock or cross-compile with deploy/build-arm.sh",
		runtime.GOOS, runtime.GOARCH,
	)
}
