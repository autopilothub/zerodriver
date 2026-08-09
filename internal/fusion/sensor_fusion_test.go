package fusion

import (
	"testing"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestFusion_TracingWithLine(t *testing.T) {
	f := New()
	line := domain.LinePosition{
		Offset: 0.3, LookaheadOffset: 0.25, Detected: true, Timestamp: time.Now(),
	}
	att := domain.Attitude{GyroZ: 5.0, HasMag: true, Heading: 90}
	obs := domain.ObstacleScan{FrontDistance: 999}

	input := f.Fuse(line, att, obs, domain.StateTracing)
	if !input.LineDetected {
		t.Fatal("line should be detected")
	}
	if input.LookaheadOffset != 0.25 {
		t.Fatalf("expected lookahead 0.25, got %f", input.LookaheadOffset)
	}
	if input.YawRate != 5.0 {
		t.Fatalf("expected yaw rate 5, got %f", input.YawRate)
	}
}

func TestFusion_LineLostUsesHeading(t *testing.T) {
	f := New()
	line := domain.LinePosition{Offset: 0.05, LookaheadOffset: 0.05, Detected: true}
	att := domain.Attitude{HasMag: true, Heading: 90, GyroZ: 2.0}
	f.Fuse(line, att, domain.ObstacleScan{FrontDistance: 999}, domain.StateTracing)

	lost := f.Fuse(
		domain.LinePosition{Detected: false},
		domain.Attitude{HasMag: true, Heading: 100, GyroZ: 2.0},
		domain.ObstacleScan{FrontDistance: 999},
		domain.StateTracing,
	)
	if lost.LineDetected {
		t.Fatal("line should not be detected")
	}
	if lost.HeadingError == 0 {
		t.Fatal("should have heading error toward reference")
	}
	if lost.HeadingError >= 0 {
		t.Fatalf("heading 100 vs ref 90 should be negative error, got %f", lost.HeadingError)
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
}

func TestHeadingErrorDeg(t *testing.T) {
	if d := headingErrorDeg(90, 100); d != -10 {
		t.Fatalf("expected -10, got %f", d)
	}
	if d := headingErrorDeg(350, 10); d != -20 {
		t.Fatalf("wrap expected -20, got %f", d)
	}
}
