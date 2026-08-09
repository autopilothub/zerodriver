package web

import (
	"errors"

	"github.com/autopilothub/zerodriver/internal/calibration"
)

// CalibrationState is the calibration lifecycle.
type CalibrationState string

const (
	CalibrationIdle    CalibrationState = "idle"
	CalibrationRunning CalibrationState = "running"
	CalibrationDone    CalibrationState = "done"
	CalibrationError   CalibrationState = "error"
)

// CalibrationType selects which calibration routine runs.
type CalibrationType string

const (
	CalibrationIMU      CalibrationType = "imu"
	CalibrationDeadzone CalibrationType = "deadzone"
)

// CalibrationResultJSON is the suggested hardware tuning from calibration.
type CalibrationResultJSON struct {
	Type                   string   `json:"type"`
	ServoTrimUs            int      `json:"servo_trim_us,omitempty"`
	SteeringInvert         bool     `json:"steering_invert,omitempty"`
	ThrottleForwardStartUs int      `json:"throttle_forward_start_us,omitempty"`
	ThrottleReverseStartUs int      `json:"throttle_reverse_start_us,omitempty"`
	Notes                  []string `json:"notes,omitempty"`
}

// CalibrationHooks wires motor and IMU into the web calibration runner.
type CalibrationHooks struct {
	Drive                  calibration.DriveFunc
	SetThrottleUS          calibration.SetThrottleUS
	ReadIMU                calibration.ReadIMU
	Stop                   func() error
	ServoTrimUs            int
	ThrottleForwardStartUs int
	ThrottleReverseStartUs int
	NeutralUs              int
	MaxUs                  int
	ReverseUs              int
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

func (s *Store) calibrationSnapshot() (CalibrationState, string, string, int, *CalibrationResultJSON) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.calState, s.calPhase, s.calError, s.calPulseUs, s.calResult
}

func (s *Store) calibrationType() CalibrationType {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.calType
}

func (s *Store) setCalibrationPhase(phase string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calPhase = phase
}

func (s *Store) setCalibrationProgress(phase string, pulseUs int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calPhase = phase
	s.calPulseUs = pulseUs
}

func (s *Store) finishCalibration(typ CalibrationType, res calibration.Result, dz calibration.DeadzoneResult, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.calState = CalibrationError
		s.calError = err.Error()
		s.calPhase = ""
		s.calPulseUs = 0
		if len(dz.Notes) > 0 {
			s.calResult = &CalibrationResultJSON{
				Type:  string(typ),
				Notes: append([]string(nil), dz.Notes...),
			}
		}
		return
	}
	s.calState = CalibrationDone
	s.calPhase = "done"
	s.calError = ""
	s.calPulseUs = 0
	switch typ {
	case CalibrationDeadzone:
		s.calResult = &CalibrationResultJSON{
			Type:                   string(CalibrationDeadzone),
			ThrottleForwardStartUs: dz.ThrottleForwardStartUs,
			ThrottleReverseStartUs: dz.ThrottleReverseStartUs,
			Notes:                  append([]string(nil), dz.Notes...),
		}
	default:
		s.calResult = &CalibrationResultJSON{
			Type:                   string(CalibrationIMU),
			ServoTrimUs:            res.ServoTrimUs,
			SteeringInvert:         res.SteeringInvert,
			ThrottleForwardStartUs: res.ThrottleForwardStartUs,
			Notes:                  append([]string(nil), res.Notes...),
		}
	}
}

// StartCalibration runs a calibration routine in the background.
func (s *Store) StartCalibration(typ CalibrationType) error {
	s.mu.Lock()
	if s.calHooks.ReadIMU == nil {
		s.mu.Unlock()
		return errCalibrationUnavailable
	}
	if typ == CalibrationDeadzone && s.calHooks.SetThrottleUS == nil {
		s.mu.Unlock()
		return errCalibrationUnavailable
	}
	if typ != CalibrationDeadzone && s.calHooks.Drive == nil {
		s.mu.Unlock()
		return errCalibrationUnavailable
	}
	if s.calState == CalibrationRunning {
		s.mu.Unlock()
		return errCalibrationRunning
	}
	hooks := s.calHooks
	s.calType = typ
	s.calState = CalibrationRunning
	s.calPhase = "starting"
	s.calError = ""
	s.calResult = nil
	s.calPulseUs = 0
	s.driveMode = DriveAuto
	s.manualSteer = 0
	s.manualThrottle = 0
	s.mu.Unlock()

	go func() {
		var (
			imuRes calibration.Result
			dzRes  calibration.DeadzoneResult
			err    error
		)
		switch typ {
		case CalibrationDeadzone:
			dzRes, err = calibration.RunDeadzone(
				hooks.SetThrottleUS,
				hooks.ReadIMU,
				calibration.DeadzoneConfig{
					NeutralUs:      hooks.NeutralUs,
					MaxUs:          hooks.MaxUs,
					ReverseUs:      hooks.ReverseUs,
					ForwardStartUs: hooks.ThrottleForwardStartUs,
					ReverseStartUs: hooks.ThrottleReverseStartUs,
				},
				calibration.DeadzoneOptions{
					OnProgress: s.setCalibrationProgress,
				},
			)
		default:
			imuRes, err = calibration.Run(
				hooks.Drive,
				hooks.ReadIMU,
				hooks.ServoTrimUs,
				hooks.ThrottleForwardStartUs,
				calibration.Options{
					OnPhase: s.setCalibrationPhase,
				},
			)
		}
		if hooks.Stop != nil {
			_ = hooks.Stop()
		}
		s.finishCalibration(typ, imuRes, dzRes, err)
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
	s.calType = ""
	s.calPhase = ""
	s.calError = ""
	s.calResult = nil
	s.calPulseUs = 0
}

var (
	errCalibrationUnavailable = errors.New("calibration not available")
	errCalibrationRunning     = errors.New("calibration already running")
	errCalibrating            = errors.New("calibration in progress")
)

type calibrationFields struct {
	State   CalibrationState
	Type    CalibrationType
	Phase   string
	Error   string
	PulseUs int
	Result  *CalibrationResultJSON
}

func (s *Store) calibrationStatusLocked() calibrationFields {
	out := calibrationFields{State: s.calState, Type: s.calType}
	if s.calState == "" {
		out.State = CalibrationIdle
	}
	out.Phase = s.calPhase
	out.Error = s.calError
	out.PulseUs = s.calPulseUs
	out.Result = s.calResult
	return out
}
