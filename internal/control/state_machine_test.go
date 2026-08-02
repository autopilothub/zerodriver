package control

import (
	"testing"

	"github.com/autopilothub/zerodriver/internal/domain"
)

func TestStateMachine_StartStop(t *testing.T) {
	sm := NewStateMachine(20)
	if sm.State() != domain.StateIdle {
		t.Fatal("initial state should be IDLE")
	}
	sm.Start()
	if sm.State() != domain.StateTracing {
		t.Fatal("should be TRACING after start")
	}
	sm.Stop()
	if sm.State() != domain.StateStopped {
		t.Fatal("should be STOPPED after stop")
	}
}

func TestStateMachine_ObstacleAvoidance(t *testing.T) {
	sm := NewStateMachine(20)
	sm.Start()

	sm.Update(999)
	if sm.State() != domain.StateTracing {
		t.Fatal("should remain TRACING with clear path")
	}

	sm.Update(10)
	if sm.State() != domain.StateAvoiding {
		t.Fatal("should transition to AVOIDING")
	}

	sm.Update(50)
	if sm.State() != domain.StateTracing {
		t.Fatal("should return to TRACING when obstacle cleared")
	}
}
