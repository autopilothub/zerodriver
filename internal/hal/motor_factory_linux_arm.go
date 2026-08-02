//go:build linux && arm

package hal

import "github.com/autopilothub/zerodriver/internal/config"

func newHardwareMotor(cfg *config.Config) (Motor, error) {
	return NewPCA9685Motor(cfg.Hardware.I2CBus, cfg.Hardware)
}
