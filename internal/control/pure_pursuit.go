package control

import "math"

// PurePursuit computes steering from a lookahead point on the path.
// δ = atan(2·sin(α)/Ld) with α = atan2(lateral, Ld), scaled to [-1, +1].
type PurePursuit struct {
	lookahead float64
	gain      float64
}

func NewPurePursuit(lookahead, gain float64) *PurePursuit {
	if lookahead <= 0 {
		lookahead = 0.35
	}
	if gain <= 0 {
		gain = 1.0
	}
	return &PurePursuit{lookahead: lookahead, gain: gain}
}

// Steering returns normalized steering for lateral offset at the lookahead point.
// lateralOffset is normalized [-1, +1] (negative = line left of center).
func (p *PurePursuit) Steering(lateralOffset float64) float64 {
	ld := p.lookahead
	alpha := math.Atan2(lateralOffset, ld)
	steer := math.Atan(2*math.Sin(alpha)/ld) * p.gain
	return clamp(steer, -1, 1)
}
