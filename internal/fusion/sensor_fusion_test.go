package fusion

import (
	"testing"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestFusion_TracingWithLine(t *testing.T) {
	f := New()
	line := domain.LinePosition{Offset: 0.3, Detected: true, Timestamp: time.Now()}
	att := domain.Attitude{GyroZ: 5.0}
	obs := domain.ObstacleScan{FrontDistance: 999}

	input := f.Fuse(line, att, obs, domain.StateTracing)
	if !input.LineDetected {
		t.Fatal("line should be detected")
	}
	if input.LineOffset != 0.3 {
		t.Fatalf("expected offset 0.3, got %f", input.LineOffset)
	}
}

func TestFusion_LineLost(t *testing.T) {
	f := New()
	line := domain.LinePosition{Detected: false}
	att := domain.Attitude{Yaw: 10.0, GyroZ: 2.0}
	obs := domain.ObstacleScan{FrontDistance: 999}

	input := f.Fuse(line, att, obs, domain.StateTracing)
	if input.LineDetected {
		t.Fatal("line should not be detected")
	}
	if input.LineOffset == 0 {
		t.Fatal("should use IMU for dead reckoning")
	}
}

func TestFusion_Avoiding(t *testing.T) {
	f := New()
	line := domain.LinePosition{Offset: 0.5, Detected: true}
	att := domain.Attitude{}
	obs := domain.ObstacleScan{FrontDistance: 10}

	input := f.Fuse(line, att, obs, domain.StateAvoiding)
	if input.LineDetected {
		t.Fatal("should ignore line during avoiding")
	}
	if input.FrontDistance != 10 {
		t.Fatalf("expected front distance 10, got %f", input.FrontDistance)
	}
}
