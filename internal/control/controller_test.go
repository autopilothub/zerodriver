package control

import (
	"testing"

	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/domain"
	"github.com/autopilothub/zerodriver/internal/hal"
)

type mockMotor struct {
	steering float64
	throttle float64
}

func (m *mockMotor) Drive(steering, throttle float64) error {
	m.steering = steering
	m.throttle = throttle
	return nil
}
func (m *mockMotor) SetThrottleUS(us int) error { return nil }
func (m *mockMotor) Stop() error                { return m.Drive(0, 0) }
func (m *mockMotor) Close() error { return nil }

var _ hal.Motor = (*mockMotor)(nil)

func TestController_TracingDrivesMotor(t *testing.T) {
	motor := &mockMotor{}
	cfg := &config.Config{
		Control: config.ControlConfig{
			BaseSpeed:       0.6,
			CornerSpeed:     0.3,
			CornerThreshold: 1.0,
			LineLostSpeed:   0.2,
			MaxThrottleSlew: 100,
			PurePursuit:     config.PurePursuitConfig{Lookahead: 0.35, Gain: 1.0},
		},
		PID: config.PIDConfig{Steering: config.PIDGains{Kp: 0.8}},
		Obstacle: config.ObstacleConfig{StopDistanceCM: 20},
	}
	ctrl := NewController(cfg, motor)
	ctrl.Start()

	input := domain.FusedInput{
		LineOffset:      0.2,
		LookaheadOffset: 0.2,
		LineDetected:    true,
		FrontDistance:   999,
		State:           domain.StateTracing,
	}
	cmd := ctrl.Tick(input, 0.02)

	if cmd.Throttle != 0.6 {
		t.Fatalf("expected throttle 0.6, got %.2f", cmd.Throttle)
	}
	if motor.throttle != 0.6 {
		t.Fatalf("motor throttle %.2f", motor.throttle)
	}
	if motor.steering == 0 {
		t.Fatal("expected non-zero steering")
	}
}

func TestController_LineLostReducesThrottle(t *testing.T) {
	motor := &mockMotor{}
	cfg := &config.Config{
		Control: config.ControlConfig{
			BaseSpeed:       0.6,
			LineLostSpeed:   0.2,
			CornerThreshold: 0.5,
			MaxThrottleSlew: 100,
		},
		PID:      config.PIDConfig{Steering: config.PIDGains{Kp: 0.8}},
		Obstacle: config.ObstacleConfig{StopDistanceCM: 20},
	}
	ctrl := NewController(cfg, motor)
	ctrl.Start()

	input := domain.FusedInput{
		LineDetected:  false,
		HeadingError:  10,
		YawRate:       0,
		FrontDistance: 999,
	}
	cmd := ctrl.Tick(input, 0.02)

	if cmd.Throttle != 0.2 {
		t.Fatalf("expected line lost throttle 0.2, got %.2f", cmd.Throttle)
	}
}

func TestController_ObstacleStops(t *testing.T) {
	motor := &mockMotor{}
	cfg := &config.Config{
		Control:  config.ControlConfig{BaseSpeed: 0.6, LineLostSpeed: 0.2, MaxThrottleSlew: 100},
		PID:      config.PIDConfig{Steering: config.PIDGains{Kp: 0.8}},
		Obstacle: config.ObstacleConfig{StopDistanceCM: 20},
	}
	ctrl := NewController(cfg, motor)
	ctrl.Start()

	ctrl.Tick(domain.FusedInput{LineDetected: true, FrontDistance: 10}, 0.02)
	if motor.throttle != 0 {
		t.Fatalf("expected stop on obstacle, throttle=%.2f", motor.throttle)
	}
}
