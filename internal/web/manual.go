package web

import (
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

// DriveMode is auto (line trace) or manual (web control).
type DriveMode string

const (
	DriveAuto   DriveMode = "auto"
	DriveManual DriveMode = "manual"
)

func (s *Store) SetDriveMode(mode DriveMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if mode != DriveAuto && mode != DriveManual {
		mode = DriveAuto
	}
	s.driveMode = mode
	if mode == DriveAuto {
		s.manualSteer = 0
		s.manualThrottle = 0
	} else if mode == DriveManual {
		s.manualSteer = 0
		s.manualThrottle = 0
		s.manualLastCmd = time.Now()
	}
}

func (s *Store) DriveMode() DriveMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.driveMode == "" {
		return DriveAuto
	}
	return s.driveMode
}

func (s *Store) SetManualDrive(steering, throttle float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.driveMode != DriveManual {
		return errNotManual
	}
	s.manualSteer = clamp(steering, -1, 1)
	s.manualThrottle = clamp(throttle, -1, 1)
	s.manualLastCmd = time.Now()
	return nil
}

func (s *Store) StopManual() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.manualSteer = 0
	s.manualThrottle = 0
	s.manualLastCmd = time.Now()
}

// ManualCommand returns the active manual drive command, or false in auto mode.
// Throttle and steering zero out if no command arrived within timeout (safety).
func (s *Store) ManualCommand(timeout time.Duration) (bool, domain.ControlCommand) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.driveMode != DriveManual {
		return false, domain.ControlCommand{}
	}
	cmd := domain.ControlCommand{
		Steering: s.manualSteer,
		Throttle: s.manualThrottle,
	}
	if timeout > 0 && !s.manualLastCmd.IsZero() && time.Since(s.manualLastCmd) > timeout {
		cmd.Steering = 0
		cmd.Throttle = 0
	}
	return true, cmd
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
