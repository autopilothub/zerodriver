//go:build linux && arm

package hal

import "github.com/autopilothub/zerodriver/internal/config"

func newHardwareMotor(cfg *config.Config) (Motor, error) {
	return NewPiMotor(
		cfg.Hardware.MotorLPWM,
		cfg.Hardware.MotorLDir,
		cfg.Hardware.MotorRPWM,
		cfg.Hardware.MotorRDir,
	)
}
