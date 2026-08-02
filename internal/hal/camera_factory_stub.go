//go:build !linux || !arm

package hal

import (
	"fmt"
	"runtime"

	"github.com/autopilothub/zerodriver/internal/config"
)

func newHardwareCamera(cfg *config.Config) (Camera, error) {
	return nil, fmt.Errorf(
		"hardware camera requires linux/arm build (current: %s/%s)",
		runtime.GOOS, runtime.GOARCH,
	)
}
