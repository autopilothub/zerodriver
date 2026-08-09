package control

// ThrottleSlew limits how fast throttle can change for smooth acceleration.
type ThrottleSlew struct {
	out     float64
	maxRate float64 // per second (0–1 throttle units)
}

func NewThrottleSlew(maxRate float64) *ThrottleSlew {
	if maxRate <= 0 {
		maxRate = 0.25
	}
	return &ThrottleSlew{maxRate: maxRate}
}

func (s *ThrottleSlew) Step(target, dt float64) float64 {
	if target < 0 {
		target = 0
	}
	if target > 1 {
		target = 1
	}
	if dt <= 0 {
		s.out = target
		return s.out
	}
	maxDelta := s.maxRate * dt
	delta := target - s.out
	if delta > maxDelta {
		delta = maxDelta
	} else if delta < -maxDelta {
		delta = -maxDelta
	}
	s.out += delta
	return s.out
}

func (s *ThrottleSlew) Reset() {
	s.out = 0
}

func (s *ThrottleSlew) Out() float64 {
	return s.out
}
