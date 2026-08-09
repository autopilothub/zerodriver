package web

import (
	"errors"

	"github.com/autopilothub/zerodriver/internal/calibration"
)

// CalibrationState is the IMU calibration lifecycle.
type CalibrationState string

const (
	CalibrationIdle    CalibrationState = "idle"
	CalibrationRunning CalibrationState = "running"
	CalibrationDone    CalibrationState = "done"
	CalibrationError   CalibrationState = "error"
)

// CalibrationResultJSON is the suggested hardware tuning from IMU calibration.
type CalibrationResultJSON struct {
	ServoTrimUs            int      `json:"servo_trim_us"`
	SteeringInvert         bool     `json:"steering_invert"`
	ThrottleForwardStartUs int      `json:"throttle_forward_start_us"`
	Notes                  []string `json:"notes"`
}

// CalibrationHooks wires motor and IMU into the web calibration runner.
type CalibrationHooks struct {
	Drive                  calibration.DriveFunc
	ReadIMU                calibration.ReadIMU
	Stop                   func() error
	ServoTrimUs            int
	ThrottleForwardStartUs int
}

// SetCalibrationHooks provides motor/IMU access for web-triggered calibration.
func (s *Store) SetCalibrationHooks(h CalibrationHooks) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calHooks = h
}

func (s *Store) IsCalibrating() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.calState == CalibrationRunning
}

func (s *Store) calibrationSnapshot() (CalibrationState, string, string, *CalibrationResultJSON) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.calState, s.calPhase, s.calError, s.calResult
}

func (s *Store) setCalibrationPhase(phase string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calPhase = phase
}

func (s *Store) finishCalibration(res calibration.Result, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.calState = CalibrationError
		s.calError = err.Error()
		s.calPhase = ""
		return
	}
	s.calState = CalibrationDone
	s.calPhase = "done"
	s.calError = ""
	s.calResult = &CalibrationResultJSON{
		ServoTrimUs:            res.ServoTrimUs,
		SteeringInvert:         res.SteeringInvert,
		ThrottleForwardStartUs: res.ThrottleForwardStartUs,
		Notes:                  append([]string(nil), res.Notes...),
	}
}

// StartCalibration runs IMU-assisted calibration in the background.
func (s *Store) StartCalibration() error {
	s.mu.Lock()
	if s.calHooks.Drive == nil || s.calHooks.ReadIMU == nil {
		s.mu.Unlock()
		return errCalibrationUnavailable
	}
	if s.calState == CalibrationRunning {
		s.mu.Unlock()
		return errCalibrationRunning
	}
	hooks := s.calHooks
	s.calState = CalibrationRunning
	s.calPhase = "starting"
	s.calError = ""
	s.calResult = nil
	s.driveMode = DriveAuto
	s.manualSteer = 0
	s.manualThrottle = 0
	s.mu.Unlock()

	go func() {
		res, err := calibration.Run(
			hooks.Drive,
			hooks.ReadIMU,
			hooks.ServoTrimUs,
			hooks.ThrottleForwardStartUs,
			calibration.Options{
				OnPhase: s.setCalibrationPhase,
			},
		)
		if hooks.Stop != nil {
			_ = hooks.Stop()
		}
		s.finishCalibration(res, err)
	}()
	return nil
}

func (s *Store) ResetCalibration() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.calState == CalibrationRunning {
		return
	}
	s.calState = CalibrationIdle
	s.calPhase = ""
	s.calError = ""
	s.calResult = nil
}

var (
	errCalibrationUnavailable = errors.New("calibration not available")
	errCalibrationRunning     = errors.New("calibration already running")
	errCalibrating            = errors.New("calibration in progress")
)

type calibrationFields struct {
	State  CalibrationState
	Phase  string
	Error  string
	Result *CalibrationResultJSON
}

func (s *Store) calibrationStatusLocked() calibrationFields {
	out := calibrationFields{State: s.calState}
	if s.calState == "" {
		out.State = CalibrationIdle
	}
	out.Phase = s.calPhase
	out.Error = s.calError
	out.Result = s.calResult
	return out
}
