package hal

// ThrottleMapMode selects how normalized throttle maps to ESC pulse width.
type ThrottleMapMode string

const (
	ThrottleMapNeutralForward ThrottleMapMode = "neutral_forward" // 0=neutral, 1=max (1500–2000)
	ThrottleMapLinear         ThrottleMapMode = "linear"          // 0=min, 1=max (1000–2000)
	ThrottleMapBidirectional  ThrottleMapMode = "bidirectional"   // -1=reverse, 0=neutral, 1=forward
)

type ThrottlePulseConfig struct {
	MinUs          int
	NeutralUs      int
	MaxUs          int
	ReverseUs      int
	TrimUs         int
	ForwardStartUs int
	ReverseStartUs int
	Map            ThrottleMapMode
}

func (c ThrottlePulseConfig) withDefaults() ThrottlePulseConfig {
	out := c
	if out.NeutralUs == 0 {
		out.NeutralUs = pca9685DefaultCenterUs
	}
	if out.MinUs == 0 {
		out.MinUs = pca9685DefaultMinUs
	}
	if out.MaxUs == 0 {
		out.MaxUs = pca9685DefaultMaxUs
	}
	if out.ReverseUs == 0 {
		out.ReverseUs = pca9685DefaultMinUs
	}
	if out.Map == "" {
		out.Map = ThrottleMapLinear
	}
	return out
}

// ThrottleToPulseUs maps normalized throttle to ESC pulse width.
func ThrottleToPulseUs(throttle float64, cfg ThrottlePulseConfig) int {
	cfg = cfg.withDefaults()
	if throttle < -1 {
		throttle = -1
	}
	if throttle > 1 {
		throttle = 1
	}

	var pulse int
	switch cfg.Map {
	case ThrottleMapLinear:
		if throttle < 0 {
			throttle = 0
		}
		pulse = cfg.MinUs + int(float64(cfg.MaxUs-cfg.MinUs)*throttle)
	case ThrottleMapBidirectional:
		if throttle >= 0 {
			pulse = cfg.NeutralUs + int(float64(cfg.MaxUs-cfg.NeutralUs)*throttle)
		} else {
			pulse = cfg.NeutralUs + int(float64(cfg.NeutralUs-cfg.ReverseUs)*throttle)
		}
		if throttle < 0 && cfg.ReverseStartUs > 0 && pulse > cfg.ReverseStartUs {
			pulse = cfg.ReverseStartUs
		}
	default: // neutral_forward
		if throttle < 0 {
			throttle = 0
		}
		pulse = cfg.NeutralUs + int(float64(cfg.MaxUs-cfg.NeutralUs)*throttle)
	}

	if throttle > 0 && cfg.ForwardStartUs > 0 && pulse < cfg.ForwardStartUs {
		pulse = cfg.ForwardStartUs
	}

	pulse += cfg.TrimUs
	if pulse < cfg.MinUs {
		return cfg.MinUs
	}
	if pulse > cfg.MaxUs {
		return cfg.MaxUs
	}
	return pulse
}
