package control

import (
	"context"
	"time"

	"github.com/autopilothub/zerodriver/internal/config"
	"github.com/autopilothub/zerodriver/internal/domain"
	"github.com/autopilothub/zerodriver/internal/hal"
)

// Controller orchestrates Pure Pursuit, IMU fusion, and motor output.
type Controller struct {
	purePursuit  *PurePursuit
	headingPID   *PIDController
	lineLostPID  *PIDController
	stateMachine *StateMachine
	motor        hal.Motor
	cfg          *config.Config
}

func NewController(cfg *config.Config, motor hal.Motor) *Controller {
	return &Controller{
		purePursuit:  NewPurePursuit(cfg.Control.PurePursuit.Lookahead, cfg.Control.PurePursuit.Gain),
		headingPID:   NewPID(cfg.PID.Steering),
		lineLostPID:  NewPID(config.PIDGains{
			Kp: cfg.Control.IMUFusion.LineLostHeadingKp,
			Ki: cfg.PID.Steering.Ki,
			Kd: cfg.PID.Steering.Kd,
		}),
		stateMachine: NewStateMachine(cfg.Obstacle.StopDistanceCM),
		motor:        motor,
		cfg:          cfg,
	}
}

// Arm holds throttle at zero so the ESC can initialize (typical 1-2s).
func (c *Controller) Arm(ctx context.Context) error {
	delay := time.Duration(c.cfg.Control.ESCArmDelaySec) * time.Second
	if delay <= 0 {
		return nil
	}
	c.motor.Drive(0, 0)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
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
		c.headingPID.Reset()
		c.lineLostPID.Reset()

	case domain.StateTracing:
		steering := c.computeSteering(input, dt)
		throttle := c.cfg.Control.BaseSpeed
		if !input.LineDetected {
			throttle = c.cfg.Control.LineLostSpeed
		}
		if c.cfg.Control.CornerThreshold > 0 && abs(steering) > c.cfg.Control.CornerThreshold {
			if throttle > c.cfg.Control.CornerSpeed {
				throttle = c.cfg.Control.CornerSpeed
			}
		}
		cmd = domain.ControlCommand{Steering: steering, Throttle: throttle}

	default:
		cmd = domain.ControlCommand{Steering: 0, Throttle: 0}
	}

	c.applyMotor(cmd)
	return cmd
}

func (c *Controller) computeSteering(input domain.FusedInput, dt float64) float64 {
	imu := c.cfg.Control.IMUFusion
	yawDamp := -imu.YawDamping * input.YawRate

	if input.LineDetected {
		pp := c.purePursuit.Steering(input.LookaheadOffset)
		headingErrNorm := input.HeadingError / 45.0
		imuCorr := imu.HeadingKp*headingErrNorm + yawDamp
		return clamp(pp+imuCorr, -1, 1)
	}

	// Line lost: hold reference heading from 9-axis IMU.
	headingErrNorm := input.HeadingError / 45.0
	hold := c.lineLostPID.Compute(headingErrNorm, dt) + yawDamp
	return clamp(hold, -1, 1)
}

func (c *Controller) State() domain.RaceState {
	return c.stateMachine.State()
}

func (c *Controller) applyMotor(cmd domain.ControlCommand) {
	c.motor.Drive(cmd.Steering, cmd.Throttle)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
