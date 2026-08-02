package telemetry

import (
	"testing"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestNewSnapshot(t *testing.T) {
	fused := domain.FusedInput{
		LineOffset:    0.25,
		LineDetected:  true,
		FrontDistance: 50,
		State:         domain.StateTracing,
	}
	cmd := domain.ControlCommand{Steering: 0.2, Throttle: 0.6}
	att := domain.Attitude{Yaw: 10.5}

	snap := NewSnapshot(domain.StateTracing, fused, cmd, att)
	if snap.State != "TRACING" {
		t.Fatalf("expected TRACING, got %s", snap.State)
	}
	if snap.LineOffset != 0.25 {
		t.Fatalf("expected line offset 0.25, got %f", snap.LineOffset)
	}
}

func TestSnapshotJSON(t *testing.T) {
	snap := Snapshot{
		State:     "TRACING",
		Timestamp: time.Now(),
	}
	data, err := snap.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestLogPublisher(t *testing.T) {
	pub := NewLogPublisher()
	err := pub.Publish(Snapshot{State: "TRACING"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoopPublisher(t *testing.T) {
	pub := NewNoopPublisher()
	if err := pub.Publish(Snapshot{}); err != nil {
		t.Fatal(err)
	}
}
