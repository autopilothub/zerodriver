package calibration

import (
	"fmt"
	"time"
)

// SetThrottleUS sets the ESC pulse width directly (bypasses throttle map).
type SetThrottleUS func(us int) error

// DeadzoneConfig holds ESC pulse limits for deadzone sweep.
type DeadzoneConfig struct {
	NeutralUs      int
	MaxUs          int
	ReverseUs      int
	ForwardStartUs int
	ReverseStartUs int
}

// DeadzoneResult holds suggested ESC deadzone values.
type DeadzoneResult struct {
	ThrottleForwardStartUs int
	ThrottleReverseStartUs int
	Notes                  []string
}

// DeadzoneOptions tunes deadzone sweep behavior.
type DeadzoneOptions struct {
	StepUs         int
	HoldDuration   time.Duration
	SampleInterval time.Duration
	SamplesPerStep int
	MinAccel       float64 // m/s² on body X
	MarginUs       int     // added to forward / subtracted from reverse threshold
	OnProgress     func(phase string, pulseUs int)
}

func (o DeadzoneOptions) withDefaults() DeadzoneOptions {
	out := o
	if out.StepUs == 0 {
		out.StepUs = 10
	}
	if out.HoldDuration == 0 {
		out.HoldDuration = 450 * time.Millisecond
	}
	if out.SampleInterval == 0 {
		out.SampleInterval = 50 * time.Millisecond
	}
	if out.SamplesPerStep == 0 {
		out.SamplesPerStep = 6
	}
	if out.MinAccel == 0 {
		out.MinAccel = 0.12
	}
	if out.MarginUs == 0 {
		out.MarginUs = 5
	}
	return out
}

// RunDeadzone sweeps ESC pulse from neutral to find forward/reverse deadzones.
// The vehicle must be on the ground with room to roll slightly.
func RunDeadzone(setUS SetThrottleUS, read ReadIMU, cfg DeadzoneConfig, opts DeadzoneOptions) (DeadzoneResult, error) {
	opts = opts.withDefaults()
	res := DeadzoneResult{
		ThrottleForwardStartUs: cfg.ForwardStartUs,
		ThrottleReverseStartUs: cfg.ReverseStartUs,
	}

	if cfg.NeutralUs == 0 {
		cfg.NeutralUs = 1500
	}
	if cfg.MaxUs == 0 {
		cfg.MaxUs = 2000
	}
	if cfg.ReverseUs == 0 {
		cfg.ReverseUs = 1000
	}

	progress(opts, "prepare", cfg.NeutralUs)
	if err := setUS(cfg.NeutralUs); err != nil {
		return res, err
	}
	time.Sleep(800 * time.Millisecond)

	baseline, hasAccel, err := avgAccelX(read, opts.SampleInterval, 8)
	if err != nil {
		return res, fmt.Errorf("imu: %w", err)
	}
	if !hasAccel {
		res.Notes = append(res.Notes, "no accelerometer — use pcatest or manual pulse sweep")
		return res, fmt.Errorf("accelerometer required for deadzone calibration")
	}

	fwdPulse, fwdOK := sweepForward(setUS, read, cfg, opts, baseline)
	progress(opts, "prepare", cfg.NeutralUs)
	_ = setUS(cfg.NeutralUs)
	time.Sleep(500 * time.Millisecond)

	revPulse, revOK := sweepReverse(setUS, read, cfg, opts, baseline)
	progress(opts, "prepare", cfg.NeutralUs)
	_ = setUS(cfg.NeutralUs)

	if fwdOK {
		res.ThrottleForwardStartUs = fwdPulse + opts.MarginUs
		if res.ThrottleForwardStartUs > cfg.MaxUs {
			res.ThrottleForwardStartUs = cfg.MaxUs
		}
		res.Notes = append(res.Notes, fmt.Sprintf("forward deadzone at %dµs → throttle_forward_start_us: %d", fwdPulse, res.ThrottleForwardStartUs))
	} else {
		res.Notes = append(res.Notes, fmt.Sprintf("no forward movement detected (%d–%dµs) — check ESC arm / wiring", cfg.NeutralUs, cfg.MaxUs))
	}

	if revOK {
		res.ThrottleReverseStartUs = revPulse - opts.MarginUs
		if res.ThrottleReverseStartUs < cfg.ReverseUs {
			res.ThrottleReverseStartUs = cfg.ReverseUs
		}
		res.Notes = append(res.Notes, fmt.Sprintf("reverse deadzone at %dµs → throttle_reverse_start_us: %d", revPulse, res.ThrottleReverseStartUs))
	} else {
		res.Notes = append(res.Notes, fmt.Sprintf("no reverse movement detected (%d–%dµs) — check throttle_map: bidirectional", cfg.ReverseUs, cfg.NeutralUs))
	}

	progress(opts, "done", cfg.NeutralUs)
	return res, nil
}

func sweepForward(setUS SetThrottleUS, read ReadIMU, cfg DeadzoneConfig, opts DeadzoneOptions, baseline float64) (int, bool) {
	for pulse := cfg.NeutralUs + opts.StepUs; pulse <= cfg.MaxUs; pulse += opts.StepUs {
		progress(opts, "forward", pulse)
		if err := setUS(pulse); err != nil {
			return 0, false
		}
		time.Sleep(opts.HoldDuration)
		accel, ok, err := avgAccelX(read, opts.SampleInterval, opts.SamplesPerStep)
		if err != nil || !ok {
			continue
		}
		if accel-baseline >= opts.MinAccel || accel >= opts.MinAccel {
			progress(opts, "forward_found", pulse)
			return pulse, true
		}
	}
	return 0, false
}

func sweepReverse(setUS SetThrottleUS, read ReadIMU, cfg DeadzoneConfig, opts DeadzoneOptions, baseline float64) (int, bool) {
	for pulse := cfg.NeutralUs - opts.StepUs; pulse >= cfg.ReverseUs; pulse -= opts.StepUs {
		progress(opts, "reverse", pulse)
		if err := setUS(pulse); err != nil {
			return 0, false
		}
		time.Sleep(opts.HoldDuration)
		accel, ok, err := avgAccelX(read, opts.SampleInterval, opts.SamplesPerStep)
		if err != nil || !ok {
			continue
		}
		if baseline-accel >= opts.MinAccel || accel <= -opts.MinAccel {
			progress(opts, "reverse_found", pulse)
			return pulse, true
		}
	}
	return 0, false
}

func avgAccelX(read ReadIMU, interval time.Duration, n int) (float64, bool, error) {
	var sum float64
	count := 0
	hasAccel := false
	for i := 0; i < n; i++ {
		att, err := read()
		if err != nil {
			return 0, false, err
		}
		if att.HasAccel {
			hasAccel = true
			sum += att.AccelX
			count++
		}
		time.Sleep(interval)
	}
	if count == 0 {
		return 0, hasAccel, nil
	}
	return sum / float64(count), true, nil
}

func progress(opts DeadzoneOptions, phase string, pulse int) {
	if opts.OnProgress != nil {
		opts.OnProgress(phase, pulse)
	}
}

// PulseForThrottle returns the ESC µs for diagnostics during deadzone tuning.
func PulseForThrottle(throttle float64, neutral, maxUs, reverseUs, forwardStart, reverseStart int) int {
	pulse := neutral
	switch {
	case throttle > 0:
		pulse = neutral + int(float64(maxUs-neutral)*throttle)
		if forwardStart > 0 && pulse < forwardStart {
			pulse = forwardStart
		}
	case throttle < 0:
		pulse = neutral + int(float64(neutral-reverseUs)*throttle)
		if reverseStart > 0 && pulse > reverseStart {
			pulse = reverseStart
		}
	}
	return pulse
}
