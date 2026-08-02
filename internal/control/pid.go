package control

import "github.com/autopilothub/zerodriver/internal/config"

// PIDController implements a discrete PID controller.
type PIDController struct {
	kp, ki, kd float64
	integral   float64
	prevError  float64
}

func NewPID(gains config.PIDGains) *PIDController {
	return &PIDController{kp: gains.Kp, ki: gains.Ki, kd: gains.Kd}
}

// Compute calculates the control output for the given error and dt (seconds).
func (p *PIDController) Compute(error, dt float64) float64 {
	p.integral += error * dt
	derivative := (error - p.prevError) / dt
	if dt == 0 {
		derivative = 0
	}
	output := p.kp*error + p.ki*p.integral + p.kd*derivative
	p.prevError = error
	return clamp(output, -1, 1)
}

// Reset clears integral and previous error state.
func (p *PIDController) Reset() {
	p.integral = 0
	p.prevError = 0
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
