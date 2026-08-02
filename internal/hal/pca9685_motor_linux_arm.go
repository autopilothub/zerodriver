//go:build linux && arm

package hal

import (
	"sync"

	"github.com/autopilothub/zerodriver/internal/config"
)

// PCA9685Motor drives a steering servo and dual DC motors via PCA9685.
type PCA9685Motor struct {
	mu   sync.Mutex
	pwm  *PCA9685
	cfg  PCA9685Config
}

// NewPCA9685Motor creates a motor driver using PCA9685 channels.
func NewPCA9685Motor(i2cBus string, hw config.HardwareConfig) (*PCA9685Motor, error) {
	cfg := PCA9685Config{
		Addr:              hw.PCA9685Addr,
		FreqHz:            hw.PCA9685FreqHz,
		SteeringChannel:   hw.SteeringChannel,
		MotorLPWMChannel:  hw.MotorLPWMChannel,
		MotorLDirChannel:  hw.MotorLDirChannel,
		MotorRPWMChannel:  hw.MotorRPWMChannel,
		MotorRDirChannel:  hw.MotorRDirChannel,
		ServoMinUs:        hw.ServoMinUs,
		ServoCenterUs:     hw.ServoCenterUs,
		ServoMaxUs:        hw.ServoMaxUs,
	}.withDefaults()

	pwm, err := NewPCA9685(i2cBus, cfg.Addr, cfg.FreqHz)
	if err != nil {
		return nil, err
	}

	m := &PCA9685Motor{pwm: pwm, cfg: cfg}
	m.Stop()
	return m, nil
}

func (m *PCA9685Motor) Set(left, right float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	steering, _ := DriveInputs(left, right)
	pulse := SteeringToPulseUs(steering, m.cfg.ServoMinUs, m.cfg.ServoCenterUs, m.cfg.ServoMaxUs)
	if err := m.pwm.SetPulseUs(m.cfg.SteeringChannel, pulse); err != nil {
		return err
	}

	if err := m.setMotorSide(m.cfg.MotorLPWMChannel, m.cfg.MotorLDirChannel, left); err != nil {
		return err
	}
	return m.setMotorSide(m.cfg.MotorRPWMChannel, m.cfg.MotorRDirChannel, right)
}

func (m *PCA9685Motor) setMotorSide(pwmCh, dirCh int, speed float64) error {
	if speed < 0 {
		if err := m.pwm.SetDigital(dirCh, false); err != nil {
			return err
		}
		return m.pwm.SetDuty(pwmCh, -speed)
	}
	if err := m.pwm.SetDigital(dirCh, true); err != nil {
		return err
	}
	return m.pwm.SetDuty(pwmCh, speed)
}

func (m *PCA9685Motor) Stop() error {
	return m.Set(0, 0)
}

func (m *PCA9685Motor) Close() error {
	m.Stop()
	return m.pwm.Close()
}
