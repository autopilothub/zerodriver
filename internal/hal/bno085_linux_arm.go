//go:build linux && arm

package hal

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	bno08x "github.com/BoltyTheDog/go-bno08x"

	"github.com/autopilothub/zerodriver/internal/domain"
)

const (
	bno085ReportIntervalUs = 50000 // 20 Hz
	bno085InitWait         = 800 * time.Millisecond
	bno085FirstSampleWait  = 5 * time.Second
)

// BNO085 reads fused 9-axis orientation via CEVA SHTP (I2C 0x4A/0x4B).
type BNO085 struct {
	mu        sync.RWMutex
	sensor    *bno08x.BNO08X
	transport *bno08x.I2CTransport
	ctx       context.Context
	last      domain.Attitude
	hasLast   bool
	done      chan struct{}
}

// NewBNO085 opens and initializes a BNO085/BNO08x at the given I2C address.
func NewBNO085(busName string, addr int) (*BNO085, error) {
	transport, err := bno08x.NewI2CTransport(busName, uint16(addr))
	if err != nil {
		return nil, fmt.Errorf("open bno085 i2c: %w", err)
	}

	sensor := bno08x.NewBNO08X(transport)
	ctx := context.Background()

	if err := sensor.SoftReset(); err != nil {
		transport.Close()
		return nil, fmt.Errorf("bno085 reset: %w", err)
	}
	time.Sleep(bno085InitWait)

	if err := sensor.CheckID(ctx); err != nil {
		transport.Close()
		return nil, fmt.Errorf("bno085 check id: %w", err)
	}

	for _, feature := range []uint8{
		bno08x.SensorReportRotationVector,
		bno08x.SensorReportGyroscope,
		bno08x.SensorReportMagnetometer,
	} {
		if err := sensor.EnableFeature(feature, bno085ReportIntervalUs); err != nil {
			transport.Close()
			return nil, fmt.Errorf("bno085 enable 0x%02X: %w", feature, err)
		}
	}
	time.Sleep(200 * time.Millisecond)

	b := &BNO085{
		sensor:    sensor,
		transport: transport,
		ctx:       ctx,
		done:      make(chan struct{}),
	}
	go b.pollLoop()

	deadline := time.Now().Add(bno085FirstSampleWait)
	for time.Now().Before(deadline) {
		b.mu.RLock()
		ok := b.hasLast
		b.mu.RUnlock()
		if ok {
			return b, nil
		}
		time.Sleep(20 * time.Millisecond)
	}

	b.Close()
	return nil, fmt.Errorf("bno085 timeout waiting for rotation vector")
}

func (b *BNO085) pollLoop() {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-b.done:
			return
		case <-ticker.C:
			b.pollOnce()
		}
	}
}

func (b *BNO085) pollOnce() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.sensor.Process(b.ctx); err != nil {
		return
	}
	if _, ok := b.sensor.GetQuaternion(); !ok {
		return
	}
	b.last = b.buildAttitude()
	b.hasLast = true
}

func (b *BNO085) buildAttitude() domain.Attitude {
	q, _ := b.sensor.GetQuaternion()
	pitch, roll, yaw := q.ToEuler()
	heading := yaw
	if heading < 0 {
		heading += 360
	}

	att := domain.Attitude{
		Yaw:     yaw,
		Pitch:   pitch,
		Roll:    roll,
		Heading: heading,
		HasMag:  true,
	}

	if gyro, ok := b.sensor.GetGyroscope(); ok {
		att.GyroZ = gyro[2] * 180 / math.Pi
	}
	if mag, ok := b.sensor.GetMagnetometer(); ok {
		att.MagX, att.MagY, att.MagZ = mag[0], mag[1], mag[2]
	}
	return att
}

func (b *BNO085) Read() (domain.Attitude, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if !b.hasLast {
		return domain.Attitude{}, fmt.Errorf("bno085: no rotation vector")
	}
	return b.last, nil
}

func (b *BNO085) Close() error {
	select {
	case <-b.done:
	default:
		close(b.done)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.transport != nil {
		return b.transport.Close()
	}
	return nil
}

// ChipName returns the IMU model name.
func (b *BNO085) ChipName() string { return "BNO085" }

// HasMagnetometer reports built-in 9-axis fusion with magnetometer.
func (b *BNO085) HasMagnetometer() bool { return true }
