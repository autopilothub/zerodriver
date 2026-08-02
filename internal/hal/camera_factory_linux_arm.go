//go:build linux && arm

package hal

import "github.com/autopilothub/zerodriver/internal/config"

func newHardwareCamera(cfg *config.Config) (Camera, error) {
	return NewV4L2Camera(cfg.Hardware.CameraDevice, cfg.Camera.Width, cfg.Camera.Height)
}
