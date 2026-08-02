package telemetry

import (
	"encoding/json"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

// Snapshot is a single telemetry data point.
type Snapshot struct {
	State         string    `json:"state"`
	LineOffset    float64   `json:"line_offset"`
	LineDetected  bool      `json:"line_detected"`
	Steering      float64   `json:"steering"`
	Throttle      float64   `json:"throttle"`
	FrontDistance float64   `json:"front_distance_cm"`
	Yaw           float64   `json:"yaw"`
	Timestamp     time.Time `json:"timestamp"`
}

// Publisher sends telemetry snapshots to a backend.
type Publisher interface {
	Publish(Snapshot) error
	Close() error
}

// NewSnapshot builds a telemetry snapshot from current vehicle state.
func NewSnapshot(
	state domain.RaceState,
	fused domain.FusedInput,
	cmd domain.ControlCommand,
	att domain.Attitude,
) Snapshot {
	return Snapshot{
		State:         state.String(),
		LineOffset:    fused.LineOffset,
		LineDetected:  fused.LineDetected,
		Steering:      cmd.Steering,
		Throttle:      cmd.Throttle,
		FrontDistance: fused.FrontDistance,
		Yaw:           att.Yaw,
		Timestamp:     time.Now(),
	}
}

// JSON returns the snapshot as JSON bytes.
func (s Snapshot) JSON() ([]byte, error) {
	return json.Marshal(s)
}
