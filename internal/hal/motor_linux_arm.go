//go:build linux && arm

package hal

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

// PiMotor drives a dual-motor differential drive via GPIO soft-PWM.
type PiMotor struct {
	mu        sync.Mutex
	leftPWM   gpio.PinOut
	leftDir   gpio.PinOut
	rightPWM  gpio.PinOut
	rightDir  gpio.PinOut
	leftDuty  float64
	rightDuty float64
	stopCh    chan struct{}
}

// NewPiMotor creates a motor driver using the given BCM pin numbers.
func NewPiMotor(lPWM, lDir, rPWM, rDir int) (*PiMotor, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("periph init: %w", err)
	}

	leftPWM, err := openPin(lPWM)
	if err != nil {
		return nil, fmt.Errorf("left PWM pin: %w", err)
	}
	leftDir, err := openPin(lDir)
	if err != nil {
		return nil, fmt.Errorf("left DIR pin: %w", err)
	}
	rightPWM, err := openPin(rPWM)
	if err != nil {
		return nil, fmt.Errorf("right PWM pin: %w", err)
	}
	rightDir, err := openPin(rDir)
	if err != nil {
		return nil, fmt.Errorf("right DIR pin: %w", err)
	}

	m := &PiMotor{
		leftPWM:  leftPWM,
		leftDir:  leftDir,
		rightPWM: rightPWM,
		rightDir: rightDir,
		stopCh:   make(chan struct{}),
	}

	go m.pwmLoop()
	return m, nil
}

func openPin(bcm int) (gpio.PinOut, error) {
	pin := gpioreg.ByName("GPIO" + strconv.Itoa(bcm))
	if pin == nil {
		return nil, fmt.Errorf("GPIO%d not found", bcm)
	}
	return pin.Out(gpio.Low)
}

func (m *PiMotor) Set(left, right float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.leftDuty = clamp(left, -1, 1)
	m.rightDuty = clamp(right, -1, 1)
	return nil
}

func (m *PiMotor) Stop() error {
	return m.Set(0, 0)
}

func (m *PiMotor) Close() error {
	close(m.stopCh)
	m.Stop()
	return nil
}

func (m *PiMotor) pwmLoop() {
	const freq = 1000
	period := time.Second / freq
	halfPeriod := period / 2

	for {
		select {
		case <-m.stopCh:
			m.leftPWM.Out(gpio.Low)
			m.rightPWM.Out(gpio.Low)
			return
		default:
		}

		m.mu.Lock()
		left := m.leftDuty
		right := m.rightDuty
		m.mu.Unlock()

		m.driveSide(m.leftPWM, m.leftDir, left, halfPeriod)
		m.driveSide(m.rightPWM, m.rightDir, right, halfPeriod)
	}
}

func (m *PiMotor) driveSide(pwm, dir gpio.PinOut, duty float64, halfPeriod time.Duration) {
	absDuty := duty
	forward := true
	if absDuty < 0 {
		absDuty = -absDuty
		forward = false
	}

	if forward {
		dir.Out(gpio.High)
	} else {
		dir.Out(gpio.Low)
	}

	onTime := time.Duration(absDuty * float64(halfPeriod))
	if onTime > 0 {
		pwm.Out(gpio.High)
		time.Sleep(onTime)
	}
	pwm.Out(gpio.Low)
	time.Sleep(halfPeriod - onTime)
}
