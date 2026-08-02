package control

import (
	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/domain"
	"github.com/autopilothub/zerodriver/internal/hal"
)

// Controller orchestrates PID, state machine, and motor output.
type Controller struct {
	pid       *PIDController
	stateMachine *StateMachine
	motor     hal.Motor
	cfg       *config.Config
}

func NewController(cfg *config.Config, motor hal.Motor) *Controller {
	return &Controller{
		pid:          NewPID(cfg.PID.Steering),
		stateMachine: NewStateMachine(cfg.Obstacle.StopDistanceCM),
		motor:        motor,
		cfg:          cfg,
	}
}

func (c *Controller) Start() {
	c.stateMachine.Start()
}

func (c *Controller) Stop() {
	c.stateMachine.Stop()
	c.motor.Stop()
}

// Tick runs one control cycle.
func (c *Controller) Tick(input domain.FusedInput, dt float64) domain.ControlCommand {
	state := c.stateMachine.Update(input.FrontDistance)

	var cmd domain.ControlCommand

	switch state {
	case domain.StateAvoiding, domain.StateStopped:
		cmd = domain.ControlCommand{Steering: 0, Throttle: 0}
		c.pid.Reset()

	case domain.StateTracing:
		steering := c.pid.Compute(input.LineOffset+input.YawCorrection, dt)
		throttle := c.cfg.Control.BaseSpeed
		if abs(steering) > c.cfg.Control.CornerThreshold {
			throttle = c.cfg.Control.CornerSpeed
		}
		cmd = domain.ControlCommand{Steering: steering, Throttle: throttle}

	default:
		cmd = domain.ControlCommand{Steering: 0, Throttle: 0}
	}

	c.applyMotor(cmd)
	return cmd
}

func (c *Controller) State() domain.RaceState {
	return c.stateMachine.State()
}

func (c *Controller) applyMotor(cmd domain.ControlCommand) {
	left := cmd.Throttle - cmd.Steering*cmd.Throttle
	right := cmd.Throttle + cmd.Steering*cmd.Throttle
	c.motor.Set(left, right)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
