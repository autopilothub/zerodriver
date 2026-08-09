package control

import (
	"testing"

	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestController_DriveManual(t *testing.T) {
	motor := &mockMotor{}
	cfg := &config.Config{}
	ctrl := NewController(cfg, motor)

	cmd := ctrl.DriveManual(domain.ControlCommand{Steering: 0.5, Throttle: 0.3})
	if cmd.Steering != 0.5 || cmd.Throttle != 0.3 {
		t.Fatalf("cmd: %+v", cmd)
	}
	if motor.steering != 0.5 || motor.throttle != 0.3 {
		t.Fatalf("motor: steer=%f throttle=%f", motor.steering, motor.throttle)
	}
}
