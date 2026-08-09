//go:build linux && arm

package hal

import (
	"github.com/autopilothub/zerodriver/internal/config"
)

func newHardwareIMU(cfg *config.Config) (IMU, error) {
	switch NormalizeIMUModel(cfg.Hardware.IMUModel) {
	case IMUModelBNO085:
		return NewBNO085(cfg.Hardware.I2CBus, cfg.Hardware.I2CAddr)
	case IMUModelAuto:
		addr := cfg.Hardware.I2CAddr
		if addr == 0x4A || addr == 0x4B {
			return NewBNO085(cfg.Hardware.I2CBus, addr)
		}
		return NewMPU9250(cfg.Hardware.I2CBus, addr, IMUModelAuto)
	default:
		return NewMPU9250(cfg.Hardware.I2CBus, cfg.Hardware.I2CAddr, cfg.Hardware.IMUModel)
	}
}
