package web

import (
	"sync"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

// Status is the live vehicle state exposed over HTTP.
type Status struct {
	Mode            string    `json:"mode"`
	DriveMode       string    `json:"drive_mode"`
	State           string    `json:"state"`
	LineOffset      float64   `json:"line_offset"`
	LookaheadOffset float64   `json:"lookahead_offset"`
	LineDetected    bool      `json:"line_detected"`
	HeadingError    float64   `json:"heading_error_deg"`
	YawRate         float64   `json:"yaw_rate"`
	Steering        float64   `json:"steering"`
	Throttle        float64   `json:"throttle"`
	FrontDistance   float64   `json:"front_distance_cm"`
	Roll            float64   `json:"roll"`
	Pitch           float64   `json:"pitch"`
	Yaw             float64   `json:"yaw"`
	Heading         float64   `json:"heading"`
	HasMag          bool      `json:"has_mag"`
	CamErrors       uint64    `json:"cam_errors"`
	IMUErrors       uint64    `json:"imu_errors"`
	LidarErrors     uint64    `json:"lidar_errors"`
	UptimeSec       float64   `json:"uptime_sec"`
	Timestamp       time.Time `json:"timestamp"`
}

// UpdateInput bundles sensor readings for a status refresh.
type UpdateInput struct {
	Mode        string
	State       domain.RaceState
	Fused       domain.FusedInput
	Command     domain.ControlCommand
	Attitude    domain.Attitude
	CamErrors   uint64
	IMUErrors   uint64
	LidarErrors uint64
	StartedAt   time.Time
}

// Store holds the latest status and optional camera frame for the dashboard.
type Store struct {
	mu             sync.RWMutex
	status         Status
	frame          []byte
	frameW         int
	frameH         int
	driveMode      DriveMode
	manualSteer    float64
	manualThrottle float64
	manualLastCmd  time.Time
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Update(in UpdateInput) {
	s.mu.Lock()
	defer s.mu.Unlock()

	heading := in.Attitude.Yaw
	if in.Attitude.HasMag {
		heading = in.Attitude.Heading
	}

	uptime := 0.0
	if !in.StartedAt.IsZero() {
		uptime = time.Since(in.StartedAt).Seconds()
	}

	dm := s.driveMode
	if dm == "" {
		dm = DriveAuto
	}

	s.status = Status{
		Mode:            in.Mode,
		DriveMode:       string(dm),
		State:           in.State.String(),
		LineOffset:      in.Fused.LineOffset,
		LookaheadOffset: in.Fused.LookaheadOffset,
		LineDetected:    in.Fused.LineDetected,
		HeadingError:    in.Fused.HeadingError,
		YawRate:         in.Fused.YawRate,
		Steering:        in.Command.Steering,
		Throttle:        in.Command.Throttle,
		FrontDistance:   in.Fused.FrontDistance,
		Roll:            in.Attitude.Roll,
		Pitch:           in.Attitude.Pitch,
		Yaw:             in.Attitude.Yaw,
		Heading:         heading,
		HasMag:          in.Attitude.HasMag,
		CamErrors:       in.CamErrors,
		IMUErrors:       in.IMUErrors,
		LidarErrors:     in.LidarErrors,
		UptimeSec:       uptime,
		Timestamp:       time.Now(),
	}
}

func (s *Store) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Store) SetFrame(rgb []byte, width, height int) {
	if len(rgb) == 0 || width <= 0 || height <= 0 {
		return
	}
	need := width * height * 3
	if len(rgb) < need {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if cap(s.frame) < need {
		s.frame = make([]byte, need)
	} else {
		s.frame = s.frame[:need]
	}
	copy(s.frame, rgb[:need])
	s.frameW = width
	s.frameH = height
}

func (s *Store) Frame() (rgb []byte, width, height int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.frame) == 0 {
		return nil, 0, 0
	}
	out := make([]byte, len(s.frame))
	copy(out, s.frame)
	return out, s.frameW, s.frameH
}
