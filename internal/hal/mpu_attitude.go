package hal

import "math"

// computeTiltFromAccel derives pitch/roll (degrees) from accelerometer raw counts.
// Uses float64 before squaring to avoid int16 overflow (ay*ay wraps and can yield NaN pitch).
func computeTiltFromAccel(ax, ay, az int16) (pitch, roll float64) {
	axf := float64(ax)
	ayf := float64(ay)
	azf := float64(az)

	roll = math.Atan2(ayf, azf) * 180 / math.Pi
	horiz := math.Sqrt(ayf*ayf + azf*azf)
	pitch = math.Atan2(-axf, horiz) * 180 / math.Pi
	return pitch, roll
}

// computeMagHeading returns tilt-compensated compass heading in degrees [0, 360).
func computeMagHeading(mx, my, mz, pitchDeg, rollDeg float64) float64 {
	pitch := pitchDeg * math.Pi / 180
	roll := rollDeg * math.Pi / 180

	xh := mx*math.Cos(pitch) + mz*math.Sin(pitch)
	yh := mx*math.Sin(roll)*math.Sin(pitch) + my*math.Cos(roll) - mz*math.Sin(roll)*math.Cos(pitch)

	heading := math.Atan2(yh, xh) * 180 / math.Pi
	if heading < 0 {
		heading += 360
	}
	return heading
}
