//go:build linux && arm

package hal

import (
	"fmt"
	"sync"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	pca9685RegMode1   = 0x00
	pca9685RegMode2   = 0x01
	pca9685RegPrescale = 0xFE
	pca9685RegLED0    = 0x06

	pca9685Mode1Sleep   = 0x10
	pca9685Mode1AI      = 0x20
	pca9685Mode1Restart = 0x80
	pca9685Mode2OutDrv  = 0x04
)

// PCA9685 is a 16-channel 12-bit PWM driver over I2C.
type PCA9685 struct {
	mu     sync.Mutex
	bus    i2c.BusCloser
	dev    i2c.Dev
	freqHz int
}

// NewPCA9685 opens and initializes the PCA9685 on the given I2C bus.
func NewPCA9685(busName string, addr, freqHz int) (*PCA9685, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("periph init: %w", err)
	}

	bus, err := i2creg.Open(busName)
	if err != nil {
		return nil, fmt.Errorf("open i2c bus %q: %w", busName, err)
	}

	if freqHz <= 0 {
		freqHz = pca9685DefaultFreqHz
	}

	p := &PCA9685{
		bus:    bus,
		dev:    i2c.Dev{Addr: uint16(addr), Bus: bus},
		freqHz: freqHz,
	}

	if err := p.init(); err != nil {
		bus.Close()
		return nil, err
	}

	return p, nil
}

func (p *PCA9685) init() error {
	if err := p.writeReg(pca9685RegMode2, pca9685Mode2OutDrv); err != nil {
		return fmt.Errorf("set MODE2: %w", err)
	}

	mode1, err := p.readReg(pca9685RegMode1)
	if err != nil {
		return fmt.Errorf("read MODE1: %w", err)
	}
	if err := p.writeReg(pca9685RegMode1, mode1|pca9685Mode1Sleep); err != nil {
		return fmt.Errorf("enter sleep: %w", err)
	}

	if err := p.writeReg(pca9685RegPrescale, Prescale(p.freqHz)); err != nil {
		return fmt.Errorf("set prescale: %w", err)
	}

	if err := p.writeReg(pca9685RegMode1, (mode1|pca9685Mode1AI)&^pca9685Mode1Sleep); err != nil {
		return fmt.Errorf("wake: %w", err)
	}

	return nil
}

func (p *PCA9685) readReg(reg byte) (byte, error) {
	buf := make([]byte, 1)
	if err := p.dev.Tx([]byte{reg}, buf); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (p *PCA9685) writeReg(reg, val byte) error {
	return p.dev.Tx([]byte{reg, val}, nil)
}

// SetPWM sets a channel PWM with ON=0 and OFF=ticks.
func (p *PCA9685) SetPWM(channel int, ticks uint16) error {
	if channel < 0 || channel > 15 {
		return fmt.Errorf("channel out of range: %d", channel)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	base := byte(pca9685RegLED0 + 4*channel)
	data := []byte{
		base,
		0x00, 0x00, // ON = 0
		byte(ticks), byte(ticks >> 8),
	}
	return p.dev.Tx(data, nil)
}

// SetPulseUs sets a channel to a specific pulse width in microseconds.
func (p *PCA9685) SetPulseUs(channel int, pulseUs int) error {
	return p.SetPWM(channel, PulseUsToTicks(pulseUs, p.freqHz))
}

func (p *PCA9685) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bus != nil {
		return p.bus.Close()
	}
	return nil
}
