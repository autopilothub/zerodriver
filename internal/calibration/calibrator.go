package calibration

import (
	"fmt"
	"math"
	"time"

	"github.com/autopilothub/zerodriver/internal/domain"
)

// DriveFunc applies steering (-1..1) and throttle (-1..1).
type DriveFunc func(steering, throttle float64) error

// ReadIMU returns the latest attitude sample.
type ReadIMU func() (domain.Attitude, error)

// Result holds suggested hardware tuning values.
type Result struct {
	ServoTrimUs            int
	SteeringInvert         bool
	ThrottleForwardStartUs int
	Notes                  []string
}

// Options tunes calibration behavior.
type Options struct {
	SteerTest        float64
	ForwardThrottle  float64
	ReverseThrottle  float64
	SampleInterval   time.Duration
	SteerDuration    time.Duration
	StraightDuration time.Duration
	ThrottleDuration time.Duration
	MinForwardAccel  float64 // m/s² on body X
}

func (o Options) withDefaults() Options {
	out := o
	if out.SteerTest == 0 {
		out.SteerTest = 0.5
	}
	if out.ForwardThrottle == 0 {
		out.ForwardThrottle = 0.35
	}
	if out.ReverseThrottle == 0 {
		out.ReverseThrottle = -0.35
	}
	if out.SampleInterval == 0 {
		out.SampleInterval = 50 * time.Millisecond
	}
	if out.SteerDuration == 0 {
		out.SteerDuration = 800 * time.Millisecond
	}
	if out.StraightDuration == 0 {
		out.StraightDuration = 1500 * time.Millisecond
	}
	if out.ThrottleDuration == 0 {
		out.ThrottleDuration = 1200 * time.Millisecond
	}
	if out.MinForwardAccel == 0 {
		out.MinForwardAccel = 0.15
	}
	return out
}

// Run performs IMU-assisted steering and throttle calibration.
// The vehicle should have clearance to move briefly on the ground.
func Run(drive DriveFunc, read ReadIMU, currentTrimUs, currentForwardStartUs int, opts Options) (Result, error) {
	opts = opts.withDefaults()
	res := Result{
		ServoTrimUs:            currentTrimUs,
		ThrottleForwardStartUs: currentForwardStartUs,
	}

	if err := drive(0, 0); err != nil {
		return res, err
	}
	time.Sleep(500 * time.Millisecond)

	heading0, err := avgHeading(read, opts.SampleInterval, 10)
	if err != nil {
		return res, fmt.Errorf("imu heading: %w", err)
	}

	leftGyro, err := avgGyroDuring(drive, read, -opts.SteerTest, 0, opts)
	if err != nil {
		return res, err
	}
	rightGyro, err := avgGyroDuring(drive, read, opts.SteerTest, 0, opts)
	if err != nil {
		return res, err
	}
	_ = drive(0, 0)

	if leftGyro > 0 && rightGyro < 0 {
		res.SteeringInvert = true
		res.Notes = append(res.Notes, "steering response inverted → steering_invert: true")
		leftGyro, rightGyro = -leftGyro, -rightGyro
	}
	if leftGyro >= 0 || rightGyro <= 0 {
		res.Notes = append(res.Notes, fmt.Sprintf("unexpected steer response (left gyro=%.1f right=%.1f°/s)", leftGyro, rightGyro))
	} else {
		res.Notes = append(res.Notes, fmt.Sprintf("steer OK (left %.1f°/s, right %.1f°/s)", leftGyro, rightGyro))
	}

	if err := drive(0, opts.ForwardThrottle); err != nil {
		return res, err
	}
	headingSum, headingN, err := sampleWhile(drive, read, 0, opts.ForwardThrottle, opts.StraightDuration, opts.SampleInterval, headingSample)
	_ = drive(0, 0)
	if err != nil {
		return res, err
	}
	heading1 := headingSum / float64(headingN)
	drift := headingDelta(heading0, heading1)
	trimAdjust := int(math.Round(-drift * 3.0))
	if abs(drift) > 1.5 {
		res.ServoTrimUs = currentTrimUs + trimAdjust
		res.Notes = append(res.Notes, fmt.Sprintf("straight drift %.1f° → servo_trim_us: %d (was %d)", drift, res.ServoTrimUs, currentTrimUs))
	} else {
		res.Notes = append(res.Notes, fmt.Sprintf("straight drift %.1f° (trim OK)", drift))
	}

	fwdAccel, hasFwd := avgAccelDuring(drive, read, 0, opts.ForwardThrottle, opts)
	revAccel, hasRev := avgAccelDuring(drive, read, 0, opts.ReverseThrottle, opts)
	_ = drive(0, 0)

	if hasFwd && fwdAccel < opts.MinForwardAccel {
		start := currentForwardStartUs
		if start == 0 {
			start = 1500
		}
		res.ThrottleForwardStartUs = start + 30
		res.Notes = append(res.Notes, fmt.Sprintf("weak forward accel %.2f m/s² → throttle_forward_start_us: %d", fwdAccel, res.ThrottleForwardStartUs))
	} else if hasFwd {
		res.Notes = append(res.Notes, fmt.Sprintf("forward accel %.2f m/s² OK", fwdAccel))
	} else {
		res.Notes = append(res.Notes, "no accelerometer — set throttle_forward_start_us manually")
	}

	if hasRev && revAccel > -opts.MinForwardAccel {
		res.Notes = append(res.Notes, fmt.Sprintf("weak reverse accel %.2f m/s² — check throttle_reverse_us", revAccel))
	} else if hasRev {
		res.Notes = append(res.Notes, fmt.Sprintf("reverse accel %.2f m/s² OK", revAccel))
	}

	return res, nil
}

func avgGyroDuring(drive DriveFunc, read ReadIMU, steer, throttle float64, opts Options) (float64, error) {
	if err := drive(steer, throttle); err != nil {
		return 0, err
	}
	sum, n, err := sampleWhile(drive, read, steer, throttle, opts.SteerDuration, opts.SampleInterval, func(att domain.Attitude) float64 {
		return att.GyroZ
	})
	if n == 0 {
		return 0, err
	}
	return sum / float64(n), err
}

func avgAccelDuring(drive DriveFunc, read ReadIMU, steer, throttle float64, opts Options) (float64, bool) {
	if err := drive(steer, throttle); err != nil {
		return 0, false
	}
	sum, n, err := sampleWhile(drive, read, steer, throttle, opts.ThrottleDuration, opts.SampleInterval, func(att domain.Attitude) float64 {
		if !att.HasAccel {
			return 0
		}
		return att.AccelX
	})
	if err != nil || n == 0 {
		return 0, false
	}
	att, _ := read()
	return sum / float64(n), att.HasAccel
}

func avgHeading(read ReadIMU, interval time.Duration, n int) (float64, error) {
	var sumX, sumY float64
	for i := 0; i < n; i++ {
		att, err := read()
		if err != nil {
			return 0, err
		}
		h := headingDeg(att)
		rad := h * math.Pi / 180
		sumX += math.Cos(rad)
		sumY += math.Sin(rad)
		time.Sleep(interval)
	}
	deg := math.Atan2(sumY, sumX) * 180 / math.Pi
	if deg < 0 {
		deg += 360
	}
	return deg, nil
}

func headingSample(att domain.Attitude) float64 {
	return headingDeg(att)
}

func headingDeg(att domain.Attitude) float64 {
	h := att.Heading
	if !att.HasMag {
		h = att.Yaw
		if h < 0 {
			h += 360
		}
	}
	return h
}

func sampleWhile(
	drive DriveFunc,
	read ReadIMU,
	steer, throttle float64,
	duration, interval time.Duration,
	sample func(domain.Attitude) float64,
) (sum float64, count int, err error) {
	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		if err := drive(steer, throttle); err != nil {
			return 0, 0, err
		}
		att, err := read()
		if err != nil {
			return 0, 0, err
		}
		sum += sample(att)
		count++
		time.Sleep(interval)
	}
	return sum, count, nil
}

func headingDelta(from, to float64) float64 {
	d := to - from
	for d > 180 {
		d -= 360
	}
	for d < -180 {
		d += 360
	}
	return d
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
