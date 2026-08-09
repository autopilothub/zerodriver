//go:build linux && arm

package hal

import (
	"sync"

	"github.com/autopilothub/zerodriver/internal/config"
)

// PCA9685Motor drives a steering servo and ESC throttle via PCA9685 (2 channels).
type PCA9685Motor struct {
	mu  sync.Mutex
	pwm *PCA9685
	cfg PCA9685Config
}

// NewPCA9685Motor creates an RC car motor driver using PCA9685.
func NewPCA9685Motor(i2cBus string, hw config.HardwareConfig) (*PCA9685Motor, error) {
	cfg := PCA9685Config{
		Addr:            hw.PCA9685Addr,
		FreqHz:          hw.PCA9685FreqHz,
		SteeringChannel: hw.SteeringChannel,
		ThrottleChannel: hw.ThrottleChannel,
		ServoMinUs:      hw.ServoMinUs,
		ServoCenterUs:   hw.ServoCenterUs,
		ServoMaxUs:      hw.ServoMaxUs,
		ServoTrimUs:     hw.ServoTrimUs,
		SteeringInvert:  hw.SteeringInvert,
		ThrottleMinUs:   hw.ThrottleMinUs,
		ThrottleMaxUs:   hw.ThrottleMaxUs,
	}.withDefaults()

	pwm, err := NewPCA9685(i2cBus, cfg.Addr, cfg.FreqHz)
	if err != nil {
		return nil, err
	}

	m := &PCA9685Motor{pwm: pwm, cfg: cfg}
	m.Stop()
	return m, nil
}

func (m *PCA9685Motor) Drive(steering, throttle float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	steerPulse := ApplySteering(steering, m.cfg.ServoMinUs, m.cfg.ServoCenterUs, m.cfg.ServoMaxUs, m.cfg.ServoTrimUs, m.cfg.SteeringInvert)
	if err := m.pwm.SetPulseUs(m.cfg.SteeringChannel, steerPulse); err != nil {
		return err
	}

	throttlePulse := ThrottleToPulseUs(throttle, m.cfg.ThrottleMinUs, m.cfg.ThrottleMaxUs)
	return m.pwm.SetPulseUs(m.cfg.ThrottleChannel, throttlePulse)
}

func (m *PCA9685Motor) Stop() error {
	return m.Drive(0, 0)
}

func (m *PCA9685Motor) Close() error {
	m.Stop()
	return m.pwm.Close()
}
