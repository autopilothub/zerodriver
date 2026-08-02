//go:build !linux || !arm

package hal

import (
	"fmt"
	"runtime"

	"github.com/autopilothub/zerodriver/internal/config"
)

func newHardwareMotor(cfg *config.Config) (Motor, error) {
	return nil, fmt.Errorf(
		"hardware motor requires linux/arm build (current: %s/%s)",
		runtime.GOOS, runtime.GOARCH,
	)
}
