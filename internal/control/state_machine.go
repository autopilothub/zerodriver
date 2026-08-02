package control

import "github.com/autopilothub/zerodriver/internal/domain"

// StateMachine manages RaceState transitions.
type StateMachine struct {
	state            domain.RaceState
	stopDistanceCM   float64
}

func NewStateMachine(stopDistanceCM float64) *StateMachine {
	return &StateMachine{
		state:          domain.StateIdle,
		stopDistanceCM: stopDistanceCM,
	}
}

func (sm *StateMachine) State() domain.RaceState {
	return sm.state
}

// Start transitions from IDLE to TRACING.
func (sm *StateMachine) Start() {
	if sm.state == domain.StateIdle || sm.state == domain.StateStopped {
		sm.state = domain.StateTracing
	}
}

// Stop transitions to STOPPED.
func (sm *StateMachine) Stop() {
	sm.state = domain.StateStopped
}

// Update evaluates sensor data and transitions state if needed.
func (sm *StateMachine) Update(frontDistance float64) domain.RaceState {
	switch sm.state {
	case domain.StateTracing:
		if frontDistance < sm.stopDistanceCM {
			sm.state = domain.StateAvoiding
		}
	case domain.StateAvoiding:
		if frontDistance >= sm.stopDistanceCM*2 {
			sm.state = domain.StateTracing
		}
	}
	return sm.state
}
