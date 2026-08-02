package hal

import (
	"math"
	"sync"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

// MockIMU simulates MPU-9250 with a sinusoidal yaw drift.
type MockIMU struct {
	mu    sync.Mutex
	start time.Time
}

func NewMockIMU() *MockIMU {
	return &MockIMU{start: time.Now()}
}

func (m *MockIMU) Read() (domain.Attitude, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := time.Since(m.start).Seconds()
	return domain.Attitude{
		Yaw:   math.Sin(t*0.5) * 5,
		Pitch: 0,
		Roll:  0,
		GyroZ: math.Cos(t*0.5) * 2.5,
	}, nil
}

func (m *MockIMU) Close() error { return nil }

// MockLidar simulates a clear path with occasional obstacles.
type MockLidar struct {
	mu    sync.Mutex
	start time.Time
}

func NewMockLidar() *MockLidar {
	return &MockLidar{start: time.Now()}
}

func (m *MockLidar) Scan() (domain.ObstacleScan, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := time.Since(m.start).Seconds()
	front := 999.0
	// Simulate obstacle every 15 seconds for 3 seconds
	cycle := math.Mod(t, 15)
	if cycle > 12 && cycle < 15 {
		front = 15.0
	}
	return domain.ObstacleScan{
		FrontDistance: front,
		Distances:     []float64{front},
		Angles:        []float64{0},
		Timestamp:     time.Now(),
	}, nil
}

func (m *MockLidar) Close() error { return nil }

// MockCamera generates synthetic frames with a moving line.
type MockCamera struct {
	mu     sync.Mutex
	width  int
	height int
	start  time.Time
}

func NewMockCamera(width, height int) *MockCamera {
	return &MockCamera{width: width, height: height, start: time.Now()}
}

func (m *MockCamera) Capture() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	frame := make([]byte, m.width*m.height*3)
	t := time.Since(m.start).Seconds()
	// Simulate line offset oscillating -0.5 to +0.5
	lineX := int(float64(m.width)/2 + math.Sin(t*0.3)*float64(m.width)*0.3)
	roiY := m.height * 2 / 3
	for y := roiY; y < m.height; y++ {
		for x := lineX - 3; x <= lineX+3; x++ {
			if x >= 0 && x < m.width {
				idx := (y*m.width + x) * 3
				frame[idx], frame[idx+1], frame[idx+2] = 0, 0, 0
			}
		}
	}
	return frame, nil
}

func (m *MockCamera) Width() int  { return m.width }
func (m *MockCamera) Height() int { return m.height }
func (m *MockCamera) Close() error { return nil }

// MockMotor logs motor commands without driving hardware.
type MockMotor struct {
	mu        sync.Mutex
	steering  float64
	throttle  float64
}

func NewMockMotor() *MockMotor { return &MockMotor{} }

func (m *MockMotor) Drive(steering, throttle float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.steering = clamp(steering, -1, 1)
	m.throttle = clamp(throttle, 0, 1)
	return nil
}

func (m *MockMotor) Stop() error {
	return m.Drive(0, 0)
}

func (m *MockMotor) Close() error { return nil }

func (m *MockMotor) LastCommand() (steering, throttle float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.steering, m.throttle
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
