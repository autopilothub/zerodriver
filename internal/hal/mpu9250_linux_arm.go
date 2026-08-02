//go:build linux && arm

package hal

import (
	"fmt"
	"math"
	"sync"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"

	"github.com/autopilothub/zerodriver/internal/domain"
)

const (
	mpu9250RegWHOAmI     = 0x75
	mpu9250RegPWRMgmt1   = 0x6B
	mpu9250RegAccelXOutH = 0x3B
	mpu9250WhoAmIValue   = 0x71
	mpu9250AccelScale    = 16384.0 // ±2g
	mpu9250GyroScale     = 131.0   // ±250°/s
)

// MPU9250 reads orientation from the MPU-9250 IMU over I2C.
type MPU9250 struct {
	mu       sync.Mutex
	bus      i2c.BusCloser
	dev      i2c.Dev
	yaw      float64
	lastRead time.Time
}

// NewMPU9250 opens and initializes the MPU-9250 on the given I2C bus.
func NewMPU9250(busName string, addr int) (*MPU9250, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("periph init: %w", err)
	}

	bus, err := i2creg.Open(busName)
	if err != nil {
		return nil, fmt.Errorf("open i2c bus %q: %w", busName, err)
	}

	dev := i2c.Dev{Addr: uint16(addr), Bus: bus}

	who, err := i2cReadReg(&dev, mpu9250RegWHOAmI)
	if err != nil {
		bus.Close()
		return nil, fmt.Errorf("read WHO_AM_I: %w", err)
	}
	if who != mpu9250WhoAmIValue {
		bus.Close()
		return nil, fmt.Errorf("unexpected WHO_AM_I: 0x%02X (want 0x%02X)", who, mpu9250WhoAmIValue)
	}

	if err := i2cWriteReg(&dev, mpu9250RegPWRMgmt1, 0x00); err != nil {
		bus.Close()
		return nil, fmt.Errorf("wake device: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	return &MPU9250{
		bus:      bus,
		dev:      dev,
		lastRead: time.Now(),
	}, nil
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
	if err := m.dev.Tx([]byte{mpu9250RegAccelXOutH}, buf); err != nil {
		return domain.Attitude{}, fmt.Errorf("read sensors: %w", err)
	}

	ax := int16(buf[0])<<8 | int16(buf[1])
	ay := int16(buf[2])<<8 | int16(buf[3])
	az := int16(buf[4])<<8 | int16(buf[5])
	_ = int16(buf[8])<<8 | int16(buf[9])  // gx unused
	_ = int16(buf[10])<<8 | int16(buf[11]) // gy unused
	gz := int16(buf[12])<<8 | int16(buf[13])

	now := time.Now()
	dt := now.Sub(m.lastRead).Seconds()
	m.lastRead = now

	gyroZ := float64(gz) / mpu9250GyroScale
	m.yaw += gyroZ * dt

	roll := math.Atan2(float64(ay), float64(az)) * 180 / math.Pi
	pitch := math.Atan2(-float64(ax), math.Sqrt(float64(ay*ay+az*az))) * 180 / math.Pi

	return domain.Attitude{
		Yaw:   m.yaw,
		Pitch: pitch,
		Roll:  roll,
		GyroZ: gyroZ,
	}, nil
}

func (m *MPU9250) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.bus != nil {
		return m.bus.Close()
	}
	return nil
}
