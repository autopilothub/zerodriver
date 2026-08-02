//go:build linux && arm

package hal

import "github.com/autopilothub/zerodriver/internal/config"

func newHardwareIMU(cfg *config.Config) (IMU, error) {
	return NewMPU9250(cfg.Hardware.I2CBus, cfg.Hardware.I2CAddr)
}
