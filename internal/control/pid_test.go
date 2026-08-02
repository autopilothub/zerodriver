package control

import (
	"math"
	"testing"

	"github.com/autopilothub/zerodriver/internal/config"
)

func TestPID_ProportionalOnly(t *testing.T) {
	pid := NewPID(config.PIDGains{Kp: 1.0})
	out := pid.Compute(0.5, 0.02)
	if math.Abs(out-0.5) > 0.01 {
		t.Fatalf("expected ~0.5, got %f", out)
	}
}

func TestPID_ClampOutput(t *testing.T) {
	pid := NewPID(config.PIDGains{Kp: 10.0})
	out := pid.Compute(1.0, 0.02)
	if out > 1.0 {
		t.Fatalf("output should be clamped to 1.0, got %f", out)
	}
}

func TestPID_Reset(t *testing.T) {
	pid := NewPID(config.PIDGains{Ki: 1.0})
	pid.Compute(1.0, 1.0)
	pid.Reset()
	out := pid.Compute(0, 0.02)
	if out != 0 {
		t.Fatalf("expected 0 after reset, got %f", out)
	}
}
