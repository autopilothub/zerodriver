package hal

import "math"

const (
	pca9685DefaultFreqHz   = 50
	pca9685OscClockHz      = 25_000_000
	pca9685Steps           = 4096
	pca9685AddrAllCall     = 0x70 // PCA9685 broadcast address (shows on i2cdetect)
	pca9685DefaultMinUs    = 1000
	pca9685DefaultCenterUs = 1500
	pca9685DefaultMaxUs    = 2000
)

// PCA9685Config holds channel mapping and pulse ranges.
type PCA9685Config struct {
	Addr             int
	FreqHz           int
	SteeringChannel  int
	ThrottleChannel  int
	ServoMinUs       int
	ServoCenterUs    int
	ServoMaxUs       int
	ServoTrimUs      int
	SteeringInvert   bool
	ThrottleMinUs    int
	ThrottleMaxUs    int
}

func (c PCA9685Config) withDefaults() PCA9685Config {
	out := c
	if out.Addr == 0 {
		out.Addr = 0x40
	}
	if out.FreqHz == 0 {
		out.FreqHz = pca9685DefaultFreqHz
	}
	if out.ServoMinUs == 0 {
		out.ServoMinUs = pca9685DefaultMinUs
	}
	if out.ServoCenterUs == 0 {
		out.ServoCenterUs = pca9685DefaultCenterUs
	}
	if out.ServoMaxUs == 0 {
		out.ServoMaxUs = pca9685DefaultMaxUs
	}
	if out.ThrottleMinUs == 0 {
		out.ThrottleMinUs = pca9685DefaultMinUs
	}
	if out.ThrottleMaxUs == 0 {
		out.ThrottleMaxUs = pca9685DefaultMaxUs
	}
	return out
}

// Prescale calculates the PCA9685 prescale value for the target frequency.
func Prescale(freqHz int) byte {
	if freqHz <= 0 {
		freqHz = pca9685DefaultFreqHz
	}
	prescale := float64(pca9685OscClockHz)/float64(pca9685Steps*freqHz) - 1
	if prescale < 0 {
		prescale = 0
	}
	if prescale > 255 {
		prescale = 255
	}
	return byte(math.Floor(prescale + 0.5))
}

// PulseUsToTicks converts a pulse width in microseconds to PCA9685 ticks.
func PulseUsToTicks(pulseUs int, freqHz int) uint16 {
	if freqHz <= 0 {
		freqHz = pca9685DefaultFreqHz
	}
	periodUs := 1_000_000 / freqHz
	ticks := float64(pulseUs) * float64(pca9685Steps) / float64(periodUs)
	if ticks < 0 {
		return 0
	}
	if ticks > pca9685Steps-1 {
		return pca9685Steps - 1
	}
	return uint16(math.Round(ticks))
}

// SteeringToPulseUs maps steering -1..+1 to servo pulse width.
func SteeringToPulseUs(steering float64, minUs, centerUs, maxUs int) int {
	if steering < -1 {
		steering = -1
	}
	if steering > 1 {
		steering = 1
	}
	if steering < 0 {
		return centerUs + int(float64(centerUs-minUs)*steering)
	}
	return centerUs + int(float64(maxUs-centerUs)*steering)
}

// ApplySteering maps steering to pulse width with optional invert and trim.
func ApplySteering(steering float64, minUs, centerUs, maxUs, trimUs int, invert bool) int {
	if invert {
		steering = -steering
	}
	pulse := SteeringToPulseUs(steering, minUs, centerUs, maxUs) + trimUs
	if pulse < minUs {
		return minUs
	}
	if pulse > maxUs {
		return maxUs
	}
	return pulse
}

// ThrottleToPulseUs maps throttle 0..1 to ESC pulse width.
func ThrottleToPulseUs(throttle float64, minUs, maxUs int) int {
	if throttle < 0 {
		throttle = 0
	}
	if throttle > 1 {
		throttle = 1
	}
	return minUs + int(float64(maxUs-minUs)*throttle)
}
