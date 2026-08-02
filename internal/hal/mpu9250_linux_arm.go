//go:build linux && arm

package hal

import (
	"fmt"
	"sync"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"

	"github.com/autopilothub/zerodriver/internal/domain"
)

const (
	mpuRegWHOAmI     = 0x75
	mpuRegPWRMgmt1   = 0x6B
	mpuRegAccelXOutH = 0x3B
	mpuGyroScale     = 131.0 // ±250°/s
)

// MPU9250 reads orientation from MPU-6500/9250/9255 over I2C.
type MPU9250 struct {
	mu       sync.Mutex
	bus      i2c.BusCloser
	dev      i2c.Dev
	mag      i2c.Dev
	chip     string
	whoAmI   byte
	hasMag   bool
	yaw      float64
	lastRead time.Time
}

// NewMPU9250 opens and initializes an MPU-6500/9250/9255 on the given I2C bus.
func NewMPU9250(busName string, addr int) (*MPU9250, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("periph init: %w", err)
	}

	bus, err := i2creg.Open(busName)
	if err != nil {
		return nil, fmt.Errorf("open i2c bus %q: %w", busName, err)
	}

	dev := i2c.Dev{Addr: uint16(addr), Bus: bus}

	who, err := i2cReadReg(&dev, mpuRegWHOAmI)
	if err != nil {
		bus.Close()
		return nil, fmt.Errorf("read WHO_AM_I: %w", err)
	}
	if !IsSupportedMPU(who) {
		bus.Close()
		return nil, fmt.Errorf("unsupported IMU WHO_AM_I: 0x%02X (expected MPU-6500/9250/9255)", who)
	}

	if err := i2cWriteReg(&dev, mpuRegPWRMgmt1, 0x00); err != nil {
		bus.Close()
		return nil, fmt.Errorf("wake device: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	m := &MPU9250{
		bus:      bus,
		dev:      dev,
		chip:     MPUChipName(who),
		whoAmI:   who,
		lastRead: time.Now(),
	}

	mag, ok, err := initAK8963(&dev, bus)
	if err != nil {
		bus.Close()
		return nil, err
	}
	if ok {
		m.mag = mag
		m.hasMag = true
		if who == mpuWHOAmIMP6500 {
			m.chip = "MPU-9250"
		}
	}

	return m, nil
}

func i2cReadReg(dev *i2c.Dev, reg byte) (byte, error) {
	buf := make([]byte, 1)
	if err := dev.Tx([]byte{reg}, buf); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func i2cWriteReg(dev *i2c.Dev, reg, val byte) error {
	return dev.Tx([]byte{reg, val}, nil)
}

func (m *MPU9250) Read() (domain.Attitude, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	buf := make([]byte, 14)
	if err := m.dev.Tx([]byte{mpuRegAccelXOutH}, buf); err != nil {
		return domain.Attitude{}, fmt.Errorf("read sensors: %w", err)
	}

	ax := int16(buf[0])<<8 | int16(buf[1])
	ay := int16(buf[2])<<8 | int16(buf[3])
	az := int16(buf[4])<<8 | int16(buf[5])
	_ = int16(buf[8])<<8 | int16(buf[9])   // gx unused
	_ = int16(buf[10])<<8 | int16(buf[11]) // gy unused
	gz := int16(buf[12])<<8 | int16(buf[13])

	now := time.Now()
	dt := now.Sub(m.lastRead).Seconds()
	m.lastRead = now

	gyroZ := float64(gz) / mpuGyroScale
	m.yaw += gyroZ * dt

	pitch, roll := computeTiltFromAccel(ax, ay, az)

	att := domain.Attitude{
		Yaw:   m.yaw,
		Pitch: pitch,
		Roll:  roll,
		GyroZ: gyroZ,
	}

	if m.hasMag {
		mx, my, mz, ok, err := readAK8963(&m.mag)
		if err != nil {
			return domain.Attitude{}, fmt.Errorf("read magnetometer: %w", err)
		}
		if ok {
			att.HasMag = true
			att.MagX, att.MagY, att.MagZ = mx, my, mz
			att.Heading = computeMagHeading(mx, my, mz, pitch, roll)
		}
	}

	return att, nil
}

func (m *MPU9250) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.bus != nil {
		return m.bus.Close()
	}
	return nil
}

// ChipName returns the detected IMU model (e.g. "MPU-9250").
func (m *MPU9250) ChipName() string {
	return m.chip
}

// WhoAmI returns the raw WHO_AM_I register value (0x75).
func (m *MPU9250) WhoAmI() byte {
	return m.whoAmI
}

// HasMagnetometer reports whether AK8963 compass reads are active.
func (m *MPU9250) HasMagnetometer() bool {
	return m.hasMag
}
