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
	throttleCfg := ThrottlePulseConfig{
		MinUs:          hw.ThrottleMinUs,
		MaxUs:          hw.ThrottleMaxUs,
		ReverseUs:      hw.ThrottleReverseUs,
		TrimUs:         hw.ThrottleTrimUs,
		ForwardStartUs: hw.ThrottleForwardStartUs,
		ReverseStartUs: hw.ThrottleReverseStartUs,
		Map:            ThrottleMapMode(hw.ThrottleMap),
	}
	switch throttleCfg.Map {
	case ThrottleMapNeutralForward, ThrottleMapBidirectional:
		if hw.ThrottleMinUs > 0 {
			throttleCfg.NeutralUs = hw.ThrottleMinUs
		}
	default:
		if throttleCfg.MinUs == 0 {
			throttleCfg.MinUs = pca9685DefaultMinUs
		}
	}

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
		Throttle:        throttleCfg,
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

	throttlePulse := ThrottleToPulseUs(throttle, m.cfg.Throttle)
	return m.pwm.SetPulseUs(m.cfg.ThrottleChannel, throttlePulse)
}

// ThrottlePulse returns the ESC pulse width for a throttle command (diagnostics).
func (m *PCA9685Motor) ThrottlePulse(throttle float64) int {
	return ThrottleToPulseUs(throttle, m.cfg.Throttle)
}

func (m *PCA9685Motor) SetThrottleUS(us int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	steerPulse := ApplySteering(0, m.cfg.ServoMinUs, m.cfg.ServoCenterUs, m.cfg.ServoMaxUs, m.cfg.ServoTrimUs, m.cfg.SteeringInvert)
	if err := m.pwm.SetPulseUs(m.cfg.SteeringChannel, steerPulse); err != nil {
		return err
	}

	th := m.cfg.Throttle.withDefaults()
	if us < th.MinUs {
		us = th.MinUs
	}
	if us > th.MaxUs {
		us = th.MaxUs
	}
	return m.pwm.SetPulseUs(m.cfg.ThrottleChannel, us)
}

func (m *PCA9685Motor) Stop() error {
	return m.Drive(0, 0)
}

func (m *PCA9685Motor) Close() error {
	m.Stop()
	return m.pwm.Close()
}
